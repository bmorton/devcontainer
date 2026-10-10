---
description: Implement one approved plan slice with TDD and verification.
---

Implement only slice $2 of plan $1 in docs/plans/.
Resolve the plan by its exact filename or unique numeric prefix.
If the plan or slice is missing or ambiguous, ask before making changes.
Read AGENTS.md, the referenced spec, and the full plan; confirm prerequisites.
Add or update a test, run it to observe the expected failure, then implement the
smallest change and rerun it. For non-testable changes, explain the alternative
verification. Use debug-loop if a test fails twice.
Finish with verify-done, showing evidence and anything untested. Do not start the
next slice or commit changes unless asked.
