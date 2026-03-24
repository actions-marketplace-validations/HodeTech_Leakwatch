# ADR-0003: Git Library — go-git

- **Status:** Accepted
- **Date:** 2026-03-24
- **Decision Makers:** Project team

## Context

Leakwatch must be able to analyze the entire commit history of Git repositories. This requires opening/cloning repositories, traversing commits, reading the file tree of each commit, and accessing blob contents. Two main approaches exist: calling an external `git` binary (os/exec) or using a Go-native library.

## Decision

**go-git/go-git/v5** (pure Go implementation) has been selected.

### Rationale

- Pure Go — no CGO required, cross-compilation is seamless
- No external `git` binary dependency — git does not need to be installed on the target system
- Full programmatic control over Git objects
- Used in production by TruffleHog — proven
- Optimized scanning via `LogOptions` (`Since`, `Order`, depth limiting)
- Pluggable storage with in-memory test support

## Alternatives Considered

### git2go (libgit2 bindings)

- **Pros:** libgit2's mature and comprehensive API
- **Cons:** Requires CGO, cross-compilation becomes complex, C compiler toolchain matching required
- **Decision:** Rejected. Contradicts Go's static binary philosophy.

### os/exec with git command

- **Pros:** Access to every git feature, simple implementation
- **Cons:** External dependency, requires text parsing, performance overhead, security risk (command injection)
- **Decision:** Rejected. Not reliable for a production-quality tool.

## Consequences

### Positive

- Zero external dependencies — single binary, runs everywhere
- Direct access to Git internals (tree, blob, commit objects)
- Testability: unit testing with in-memory repo creation

### Negative

- go-git does not support all features of native git (e.g., shallow clone limitations)
- Memory consumption in very large monorepos may be higher compared to native git
- Behavioral differences in some edge cases (submodules, sparse checkout)
