# Personal preferences

- Skip Go and Ruby basics; explain only non-obvious trade-offs.
- Prefer stdlib and boring tools. Ask before adding a dependency.
- Go: wrap errors with %w, context as first argument, table tests, no global state or init().
- Never run destructive git (push --force, reset --hard, clean -fd) or drop data without asking.
- When intent is unclear, ask one concrete question in prose instead of guessing.
- Keep project rules in the project's AGENTS.md, not here.
- Done means builds and tests pass, linters are clean, and command output is shown.
- Never claim a check passed unless you ran it; list anything untested.
