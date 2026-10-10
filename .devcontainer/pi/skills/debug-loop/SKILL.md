---
name: debug-loop
description: Debug systematically when a test fails twice or a defect needs investigation.
---

# Debug loop

1. Reproduce with the smallest existing test or command. Capture the exact failure.
2. Read the relevant code, logs, and recent diff before proposing a fix.
3. State one falsifiable hypothesis and the evidence supporting it.
4. Make one focused change or diagnostic experiment; rerun the reproducer.
5. Record whether the hypothesis held. Undo only your own unsuccessful experiment,
   preserving the user's changes; never use destructive git.
6. Stop after three failed hypotheses. Summarize evidence and ask one concrete
   question instead of continuing to guess.
7. When fixed, add a regression test where appropriate and use verify-done.

Do not bundle unrelated fixes, weaken tests, or delegate debugging to a squad.
