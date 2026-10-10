---
name: verify-done
description: Run the project's checks and show evidence before claiming work is done.
---

# Verify done

1. Read AGENTS.md and the project's task configuration for the actual check commands.
2. If the project defines `mise run check`, run it. Otherwise run its documented
   build, test, and lint commands; never assume mise or a check task exists.
3. Show the commands, exit status, and relevant output. Summarize long output
   accurately and explicitly identify any omitted portions.
4. Inspect the diff for unintended changes and sensitive data.
5. List what remains untested, skipped, or blocked and why.
6. Only say done when acceptance criteria are met and required checks pass.

A failed or unavailable check is not a pass. Do not add testing tools just to
satisfy this skill. Use debug-loop for repeated failures.
