# Architecture Decision Records (ADR)

This directory contains the architectural decisions for the Leakwatch project in [ADR (Architecture Decision Record)](https://adr.github.io/) format.

## What is an ADR?

An ADR is a short document that records the context, rationale, and consequences of an important decision made in software architecture. Its purpose is to answer the question "why did we make this decision?" in the future.

## Format

Each ADR follows the structure below:

- **Title:** `ADR-NNNN: <Decision Title>`
- **Status:** Proposed | **Accepted** | Amended | Rejected | Deprecated
- **Context:** The situation and problem that led to the decision
- **Decision:** The decision made and its rationale
- **Alternatives Considered:** Options evaluated and reasons for rejection
- **Consequences:** Positive and negative impacts of the decision

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0001](ADR-0001-programlama-dili.md) | Programming Language: Go | Accepted | 2026-03-24 |
| [ADR-0002](ADR-0002-cli-cercevesi.md) | CLI Framework: Cobra + Viper | Accepted | 2026-03-24 |
| [ADR-0003](ADR-0003-git-kutuphanesi.md) | Git Library: go-git | Accepted | 2026-03-24 |
| [ADR-0004](ADR-0004-eklenti-mimarisi.md) | Plugin Architecture: Compile-Time | Accepted | 2026-03-24 |
| [ADR-0005](ADR-0005-desen-eslestirme.md) | Pattern Matching: Aho-Corasick Hybrid | Accepted | 2026-03-24 |
| [ADR-0006](ADR-0006-container-kutuphanesi.md) | Container Library: go-containerregistry | Accepted | 2026-03-24 |
| [ADR-0007](ADR-0007-lisans.md) | License: MIT | Accepted | 2026-03-24 |
| [ADR-0008](ADR-0008-eszamanlilik-modeli.md) | Concurrency: Worker Pool | Accepted | 2026-03-24 |
