# ADR-0006: Container Library — go-containerregistry

- **Status:** Accepted
- **Date:** 2026-03-24
- **Decision Makers:** Project team

## Context

Container images often contain sensitive configuration data and secrets. Even if secrets are deleted from the final image, they can persist in previous layers. Leakwatch must be able to inspect image layers individually to uncover "hidden" secrets.

## Decision

**google/go-containerregistry** library has been selected.

### Rationale

- Does not require Docker daemon — lightweight, portable, works in CI/CD environments
- Full support for OCI and Docker manifest formats
- Layer-by-layer analysis — inspect each layer individually
- Used by industry tools such as crane, ko, and cosign
- Registry authentication (Docker Hub, GHCR, ECR, GCR)
- Actively maintained, backed by Google

## Alternatives Considered

### Docker SDK (docker/docker)

- **Pros:** Full access to the Docker API
- **Cons:** Requires a running Docker daemon, heavy dependency
- **Decision:** Rejected. Daemon dependency restricts portability.

### Manual tar/gzip parsing

- **Pros:** Zero dependencies
- **Cons:** Registry auth, manifest parsing, layer management must be written from scratch — significant effort
- **Decision:** Rejected.

## Consequences

### Positive

- Daemonless operation: lighter, more secure
- Layer-by-layer analysis: detect deleted secrets in previous layers
- Multi-registry support: Docker Hub, GHCR, ECR, GCR, private registries

### Negative

- Downloading layers of large images requires network bandwidth
- Authentication complications may arise with some custom registry configurations
