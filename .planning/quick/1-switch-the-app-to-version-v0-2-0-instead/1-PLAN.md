---
phase: quick
plan: 1
type: execute
wave: 1
depends_on: []
files_modified:
  - Makefile
  - cmd/efx-skills/main.go
  - internal/tui/search.go
  - internal/tui/status.go
  - internal/tui/preview.go
autonomous: true
requirements: []
must_haves:
  truths:
    - "efx-skills --version prints 0.2.0"
    - "All TUI views show v0.2.0 in their title bar"
    - "make build produces a binary reporting v0.2.0"
  artifacts:
    - path: "Makefile"
      contains: "VERSION=0.2.0"
    - path: "cmd/efx-skills/main.go"
      contains: '"0.2.0"'
    - path: "internal/tui/search.go"
      contains: "v0.2.0"
    - path: "internal/tui/status.go"
      contains: "v0.2.0"
    - path: "internal/tui/preview.go"
      contains: "v0.2.0"
  key_links: []
---

<objective>
Update all version strings from v0.1.x to v0.2.0 across the entire codebase.

Purpose: The v0.2.0 milestone (origin tracking) is complete. The app version should reflect this.
Output: All version references updated to 0.2.0, consistent across Makefile, main.go, and all TUI views.
</objective>

<context>
Version strings are currently scattered and inconsistent:
- Makefile: VERSION=0.1.2
- cmd/efx-skills/main.go: var version = "0.1.3"
- internal/tui/search.go line 271: hardcoded "efx-skills v0.1.3 - Laurent Marques"
- internal/tui/status.go line 206: hardcoded "efx-skills v0.1.4 - Laurent Marques"
- internal/tui/preview.go line 172: hardcoded "efx-skills v0.1.3 - Laurent Marques"
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update all version strings to 0.2.0</name>
  <files>Makefile, cmd/efx-skills/main.go, internal/tui/search.go, internal/tui/status.go, internal/tui/preview.go</files>
  <action>
Update version strings in all five files:

1. Makefile line 5: Change `VERSION=0.1.2` to `VERSION=0.2.0`
2. cmd/efx-skills/main.go line 11: Change `var version = "0.1.3"` to `var version = "0.2.0"`
3. internal/tui/search.go line 271: Change `"efx-skills v0.1.3 - Laurent Marques"` to `"efx-skills v0.2.0 - Laurent Marques"`
4. internal/tui/status.go line 206: Change `"efx-skills v0.1.4 - Laurent Marques"` to `"efx-skills v0.2.0 - Laurent Marques"`
5. internal/tui/preview.go line 172: Change `"efx-skills v0.1.3 - Laurent Marques"` to `"efx-skills v0.2.0 - Laurent Marques"`

Do NOT touch:
- README.md references to v0.1.4 image paths (those are historical screenshots)
- README.md changelog entries (those are historical)
- go.mod/go.sum dependency versions (e.g., clipboard v0.1.4 is a dependency version, not this app)
- internal/tui/config.go `Version` field (that's a JSON struct field for skill metadata, not app version)
- .planning/ files (documentation, not code)
  </action>
  <verify>
    <automated>grep -rn "0\.1\.[0-9]" Makefile cmd/efx-skills/main.go internal/tui/search.go internal/tui/status.go internal/tui/preview.go | grep -v "test" ; echo "Exit: $?"</automated>
    Expect: no matches (exit 1 from grep). All five files should only contain 0.2.0 for the app version.
    Additional: `go build ./...` compiles without errors.
  </verify>
  <done>All five files updated to 0.2.0. No 0.1.x version strings remain in production source files (excluding tests, docs, dependencies).</done>
</task>

<task type="auto">
  <name>Task 2: Build and verify version output</name>
  <files>bin/efx-skills</files>
  <action>
Run `make build` to produce a fresh binary, then verify it reports 0.2.0:
1. `make build` -- compiles with ldflags injecting VERSION=0.2.0
2. `./bin/efx-skills --version` -- should output "efx-skills version 0.2.0"
3. `go test ./...` -- all existing tests still pass
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && make build && ./bin/efx-skills --version 2>&1 | grep -q "0.2.0" && echo "VERSION OK" || echo "VERSION MISMATCH"</automated>
  </verify>
  <done>Binary builds cleanly and reports version 0.2.0. All tests pass.</done>
</task>

</tasks>

<verification>
1. `grep -rn "0\.1\.[0-9]" Makefile cmd/ internal/tui/search.go internal/tui/status.go internal/tui/preview.go` returns no matches
2. `./bin/efx-skills --version` outputs "efx-skills version 0.2.0"
3. `go test ./...` passes
</verification>

<success_criteria>
- All app version strings are 0.2.0 (Makefile, main.go, 3 TUI view files)
- Binary builds and reports 0.2.0
- No test regressions
</success_criteria>
