---
name: Always update docs before committing
description: User requires all documentation to be reviewed and updated before any commit — no outdated or incomplete docs allowed
type: feedback
---

Before committing code changes, ALWAYS check and update all related documentation first. The user does not want any commit that leaves docs outdated or with incomplete information.

**Why:** The user has strict documentation standards and wants the repo to be consistent at every commit point — not just code-complete but docs-complete.

**How to apply:** Before running `git commit`, audit README.md, CLAUDE.md, CHANGELOG.md, roadmap, guides, and any other docs that reference counts, versions, features, or capabilities that may have changed. Update them in the same commit.
