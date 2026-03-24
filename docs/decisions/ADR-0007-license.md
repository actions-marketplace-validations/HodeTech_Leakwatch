# ADR-0007: License — MIT

- **Status:** Accepted
- **Date:** 2026-03-24
- **Decision Makers:** Project team

## Context

License selection directly affects project adoption, community contributions, and future commercial possibilities. The current state of the secret scanning market:

- TruffleHog uses **AGPL-3.0** — strong copyleft that deters enterprise users
- Gitleaks uses **MIT** — but its GitHub Action is commercial for private repos
- GitGuardian is fully commercial

## Decision

**MIT License** has been selected.

### Rationale

1. **Zero enterprise adoption barrier:** Many organizations that avoid AGPL (banks, defense, large tech companies) prefer MIT-licensed tools
2. **Market differentiation:** The "MIT + verification" combination is unique in the open source market
3. **Open-core model compatibility:** The core remains MIT while a SaaS/Enterprise tier can be added in the future
4. **Community contribution incentive:** Minimum restrictions, maximum flexibility
5. **Embedding/integration:** No restrictions in scenarios involving embedding or integrating into other tools

## Alternatives Considered

### AGPL-3.0

- **Pros:** Mandates sharing code changes, prevents free SaaS usage
- **Cons:** Many enterprise organizations prohibit AGPL as a policy; high adoption barrier
- **Decision:** Rejected. Leakwatch's positioning targets TruffleHog's AGPL weakness.

### Apache 2.0

- **Pros:** Includes patent protection, enterprise-friendly
- **Cons:** More complex than MIT, practical difference is minimal
- **Decision:** Rejected. MIT's simplicity and prevalence were preferred.

### BSL (Business Source License)

- **Pros:** Restricts commercial SaaS usage, transitions to open source later
- **Cons:** Not recognized as "true" open source (not OSI-approved), erodes community trust
- **Decision:** Rejected.

## Consequences

### Positive

- A natural choice for enterprise users avoiding AGPL
- Minimum community contribution barrier
- Compatible with future commercial models (SaaS tier)

### Negative

- Competitors can fork the code and create commercial products
- Cannot prevent being offered as SaaS (as with AGPL)
- This risk is mitigated through a strong brand and community
