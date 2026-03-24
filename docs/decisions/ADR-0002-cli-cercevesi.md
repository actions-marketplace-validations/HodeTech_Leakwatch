# ADR-0002: CLI Framework — Cobra + Viper

- **Status:** Accepted
- **Date:** 2026-03-24
- **Decision Makers:** Project team

## Context

Leakwatch is a CLI tool with a rich command structure: nested commands such as `scan git`, `scan fs`, `scan image`, `verify aws` are required. Configuration must be read with hierarchical precedence from files (.leakwatch.yaml), environment variables, and command-line flags.

## Decision

**spf13/cobra** has been selected as the CLI framework, and **spf13/viper** for configuration management.

### Rationale

- Industry standard used by Kubernetes, GitHub CLI, and Hugo
- Nested command support (tree structure)
- POSIX-compliant flag management (`-f`, `--flag`)
- Native Cobra <-> Viper integration — flags bind to configuration values
- Project scaffolding with `cobra-cli`
- Automatic help text, man page, and markdown output

## Alternatives Considered

### urfave/cli

- **Pros:** Simpler API, faster onboarding
- **Cons:** Less flexible nested command support, manual Viper integration, less ecosystem support
- **Decision:** Rejected. Insufficient for Leakwatch's command complexity.

## Consequences

### Positive

- UX familiar to developers (same patterns as Kubernetes, gh CLI)
- Standardized project structure with `cobra-cli`
- Configuration hierarchy (flag > env > config file > default) is automatic

### Negative

- Cobra + Viper together add a relatively large dependency tree
- Some edge cases in Viper (nested config, type coercion) require attention
