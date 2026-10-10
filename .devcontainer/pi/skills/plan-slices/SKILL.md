---
name: plan-slices
description: Turn an agreed spec section into small, verifiable vertical slices.
---

# Plan slices

1. Read the agreed section of docs/SPEC.md and the project's commands and layout.
   If requirements are unresolved, use spec-first instead.
2. Inspect existing plans in docs/plans/; choose the next unused NNN-name.md.
   Start at 000 if no plans exist. Never overwrite a different plan.
3. Write the plan with a reference to the spec section and numbered vertical slices.
4. For each slice include:
   - One observable outcome and explicit boundaries.
   - Tests to add or update before implementation.
   - Relevant files and dependencies on earlier slices.
   - A concrete verify command using the project's existing tooling.
5. Keep each slice independently verifiable; split oversized slices.
6. Ask for plan approval before implementing anything.

Do not add code or new tooling. Avoid speculative abstractions and unrelated cleanup.
