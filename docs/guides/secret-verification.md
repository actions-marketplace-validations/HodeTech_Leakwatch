# Leakwatch - Secret Verification Guide

> **Document Version:** 1.0
> **Date:** 2026-03-24
> **Status:** Active

---

## 1. What is Secret Verification?

Secret verification is the process of checking whether a detected secret is actually active and valid. Leakwatch does this by making controlled, read-only API calls to the service that issued the credential.

**Why it matters:**

- **Reduces false positives** -- A regex match alone cannot tell you whether a string is a real, active secret. Verification eliminates noise by confirming status with the provider.
- **Prioritizes remediation** -- Teams can focus on verified active secrets first instead of triaging hundreds of unconfirmed findings.
- **Provides context** -- Verification results include extra metadata (e.g., AWS account ID, GitHub username) that helps identify the owner of the leaked credential.

```mermaid
flowchart LR
    subgraph Scan["Detection Phase"]
        S1["Source\n(fs/git/image)"] --> E1["Detection\nEngine"]
    end

    subgraph Verify["Verification Phase"]
        E1 -->|Findings| R1{"Verifier\nRegistry"}
        R1 -->|Match| API["Provider API\n(STS, GitHub, etc.)"]
        R1 -->|No match| U["Status:\nunverified"]
        API -->|Success| VA["Status:\nverified_active"]
        API -->|Auth error| VI["Status:\nverified_inactive"]
        API -->|Network error| VE["Status:\nverify_error"]
    end

    VA --> Out["Output\n(JSON/SARIF/CSV/Table)"]
    VI --> Out
    VE --> Out
    U --> Out
```

---

## 2. Verification Statuses

Every finding in Leakwatch carries a verification status. Understanding these statuses is essential for effective triage.

| Status | Description | Action Required |
|--------|-------------|-----------------|
| `verified_active` | Secret is **valid and active** -- the provider confirmed it works | **Immediate rotation required** |
| `verified_inactive` | Secret is **invalid or revoked** -- the provider rejected it | Low priority; may still warrant cleanup |
| `unverified` | Verification was not performed (no verifier available, or `--no-verify` was used) | Manual review recommended |
| `verify_error` | An error occurred during verification (network timeout, rate limit, etc.) | Retry or verify manually |

```mermaid
stateDiagram-v2
    [*] --> Detection: Secret found
    Detection --> Verification: Verifier registered
    Detection --> unverified: No verifier available
    Verification --> verified_active: Credentials valid
    Verification --> verified_inactive: Credentials invalid/expired
    Verification --> verify_error: API error or timeout
```

---

## 3. AWS Verification

### How It Works

The AWS verifier calls [STS GetCallerIdentity](https://docs.aws.amazon.com/STS/latest/APIReference/API_GetCallerIdentity.html) using the discovered Access Key ID and Secret Access Key. This is a read-only API call that returns identity information without performing any actions on the AWS account.

- **Detector ID:** `aws-access-key-id`
- **Required data:** Both the Access Key ID (`AKIA...`) and the corresponding Secret Access Key must be found in the same context.
- **API endpoint:** `sts.amazonaws.com` (us-east-1)

### What It Reveals

When credentials are active, the verifier returns:

| Field | Description |
|-------|-------------|
| `account` | AWS account ID |
| `arn` | Full ARN of the authenticated entity |
| `user_id` | IAM user or role ID |

### IAM Permissions

The `GetCallerIdentity` API requires **no specific IAM permissions**. It works with any valid AWS credentials regardless of attached policies. This makes it ideal for verification: even a key with zero permissions will return a successful response if the credentials are active.

### Example Output

```json
{
  "verification": {
    "status": "verified_active",
    "message": "AWS credentials are active",
    "extra_data": {
      "account": "123456789012",
      "arn": "arn:aws:iam::123456789012:user/deploy-bot",
      "user_id": "AIDAEXAMPLEUSERID"
    }
  }
}
```

If the credentials are invalid or expired:

```json
{
  "verification": {
    "status": "verified_inactive",
    "message": "AWS credentials are invalid or inactive"
  }
}
```

### Conditions for Skipping

If the Secret Access Key is not found alongside the Access Key ID, verification is skipped and the finding is marked as `unverified` with the message "secret access key not found."

---

## 4. GitHub Verification

### How It Works

The GitHub verifier sends a `GET /user` request to the GitHub API using the discovered token as a Bearer token. This is a read-only call that returns information about the authenticated user.

- **Detector ID:** `github-token`
- **API endpoint:** `https://api.github.com/user`
- **Headers:** `Authorization: Bearer <token>`, `User-Agent: leakwatch-verifier`

### What It Reveals

When the token is active, the verifier returns:

| Field | Description |
|-------|-------------|
| `login` | GitHub username associated with the token |

### Response Handling

| HTTP Status | Verification Result |
|-------------|-------------------|
| `200 OK` | `verified_active` -- token is valid |
| `401 Unauthorized` | `verified_inactive` -- token is invalid or revoked |
| Other | `verify_error` -- unexpected response |

### Example Output

```json
{
  "verification": {
    "status": "verified_active",
    "message": "GitHub token is active",
    "extra_data": {
      "login": "octocat"
    }
  }
}
```

---

## 5. CLI Flags

Leakwatch provides flags to control verification behavior on all scan commands (`scan fs`, `scan git`, `scan image`).

| Flag | Default | Description |
|------|---------|-------------|
| `--no-verify` | `false` | Skip all secret verification |
| `--only-verified` | `false` | Only include findings with `verified_active` status in the output |
| `--min-severity` | `low` | Minimum severity level to report |

### Examples

```bash
# Fast scan without any verification API calls
leakwatch scan fs /path/to/project --no-verify

# Show only confirmed active secrets (highest confidence)
leakwatch scan git . --only-verified

# Combine: only verified critical findings
leakwatch scan git . --only-verified --min-severity critical
```

---

## 6. Rate Limiting and Concurrency

The verification engine manages API calls carefully to avoid overwhelming provider APIs and to stay within rate limits.

### Default Settings

| Parameter | Default | Description |
|-----------|---------|-------------|
| Concurrency | 4 workers | Number of parallel verification goroutines |
| Rate limit | 10 req/sec | Maximum verification requests per second (token bucket) |
| Timeout | 10 seconds | Per-request timeout for each verification API call |

### How It Works

The verification engine uses a worker pool pattern with a shared rate limiter:

1. **Worker pool** -- A fixed number of goroutines (default 4) process verification jobs concurrently.
2. **Token bucket rate limiter** -- Before each API call, the worker acquires a token from a `golang.org/x/time/rate` limiter. If the bucket is empty, the worker waits until a token becomes available.
3. **Per-request timeout** -- Each verification call has its own context timeout (default 10s). If the provider does not respond in time, the finding is marked `verify_error`.
4. **Context cancellation** -- If the parent context is cancelled (e.g., the user presses Ctrl+C), all pending verifications are abandoned gracefully.

### Configuration

These settings can be adjusted in the `.leakwatch.yaml` configuration file:

```yaml
verification:
  enabled: true
  timeout: 10s
  concurrency: 4
  rate_limit: 10.0
```

---

## 7. Security Considerations

Verification involves sending discovered credentials to provider APIs. Keep the following in mind:

- **Credentials are transmitted over the network** -- The raw secret value is sent to the provider's API endpoint (e.g., `sts.amazonaws.com`, `api.github.com`) over HTTPS. Ensure your network allows outbound HTTPS traffic to these endpoints.
- **Leakwatch never logs raw secrets** -- Verifiers are designed to never log, persist, or cache the raw credential values. Only redacted values appear in logs.
- **Read-only operations only** -- All verification calls are read-only. They check validity without performing any destructive or state-changing actions.
- **Network requirements** -- Verification requires outbound HTTPS access. In air-gapped or restricted environments, use `--no-verify` to skip verification entirely.
- **Provider rate limits** -- While Leakwatch applies its own rate limiting, provider-side rate limits may still apply. If you are verifying a large number of findings, consider the provider's documented rate limits.

---

## 8. Use Cases and Strategies

### CI/CD Pipeline: Two-Phase Approach

In CI/CD pipelines, speed matters. A two-phase approach balances speed with accuracy:

```bash
# Phase 1: Fast scan without verification (fail fast on regex matches)
leakwatch scan git . --since-commit HEAD~1 --no-verify --min-severity high
if [ $? -eq 1 ]; then
    echo "Potential secrets found, running verification..."

    # Phase 2: Verify only critical findings
    leakwatch scan git . --since-commit HEAD~1 --only-verified --min-severity critical
    if [ $? -eq 1 ]; then
        echo "CONFIRMED active secrets found! Blocking merge."
        exit 1
    fi
fi
```

### Triage Workflow

For security teams reviewing scan results:

1. Run a full scan with verification enabled (the default).
2. Address `verified_active` findings immediately -- these are confirmed live credentials.
3. Review `unverified` findings manually -- these may still be real secrets without an available verifier.
4. Deprioritize `verified_inactive` findings -- these are expired or revoked, but consider removing them from the codebase for hygiene.
5. Retry `verify_error` findings -- these may have failed due to transient network issues.

### Decision Tree

```mermaid
flowchart TD
    A["Start: Choose verification strategy"] --> B{"Environment?"}

    B -->|"CI/CD pipeline"| C{"Speed priority?"}
    B -->|"Security audit"| D["Full scan with verification\nleakwatch scan git . --only-verified"]
    B -->|"Air-gapped / restricted"| E["Skip verification\nleakwatch scan fs . --no-verify"]

    C -->|"Fast feedback"| F["Phase 1: No verification\nleakwatch scan git . --no-verify"]
    C -->|"Accuracy first"| D

    F --> G{"Findings?"}
    G -->|"Yes"| H["Phase 2: Verify critical\nleakwatch scan git . --only-verified --min-severity critical"]
    G -->|"No"| I["Pipeline passes"]

    H --> J{"Active secrets?"}
    J -->|"Yes"| K["Block pipeline\nNotify security team"]
    J -->|"No"| L["Pipeline passes\nLog for review"]
```

---

## 9. Next Steps

| Topic | Document |
|-------|----------|
| Getting started with Leakwatch | [Getting Started Guide](./getting-started.md) |
| Configuration file and options | [Configuration Guide](./configuration.md) |
| Running Leakwatch with Docker | [Docker Usage Guide](./docker-usage.md) |
| Architecture overview | [Architecture Document](../architecture/03-ARCHITECTURE.md) |
