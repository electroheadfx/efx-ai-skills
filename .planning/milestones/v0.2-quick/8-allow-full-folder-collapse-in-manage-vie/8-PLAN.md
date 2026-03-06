---
phase: quick-8
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/manage.go
autonomous: true
requirements: [QUICK-8]
must_haves:
  truths:
    - "Pressing [enter] on a collapsed group header hides ALL skills in that group, including active/installed ones"
    - "All groups start expanded on first load"
    - "Expanding a collapsed group reveals all skills again (active and inactive)"
    - "Toggling, removing, and other key bindings still work correctly"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "Full folder collapse behavior"
      contains: "if collapsed"
  key_links:
    - from: "buildDisplayList"
      to: "displayList"
      via: "collapsed check skips ALL skills"
      pattern: "if collapsed \\{"
---

<objective>
Allow full folder collapse in manage view -- when a group header is collapsed via [enter], hide ALL skills in that group including active/installed ones. Currently, active skills remain visible when collapsed.

Purpose: Give users full control over group visibility to reduce clutter.
Output: Updated manage.go with corrected collapse behavior.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/manage.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Fix collapse behavior to hide all skills and default groups to expanded</name>
  <files>internal/tui/manage.go</files>
  <action>
Two targeted edits in the `buildDisplayList()` method:

1. **Full collapse** (around line 276): Change the collapsed-skill filter from:
   ```go
   if collapsed && !m.skills[skillIdx].Selected {
       continue
   }
   ```
   to:
   ```go
   if collapsed {
       continue
   }
   ```
   This hides ALL skills when a group is collapsed, not just unselected ones.

2. **Default expanded** (around line 257): Change the first-build default from:
   ```go
   collapsed = !hasInstalled
   ```
   to:
   ```go
   collapsed = false
   ```
   This makes all groups start expanded on first load, regardless of whether they have installed skills.

Do NOT change any other behavior -- toggle, remove, apply/save, preview, etc. must all remain intact.

After editing, build the binary:
```
go build -o ./bin/efx-skills ./cmd/efx-skills/
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && go vet ./internal/tui/... && echo "BUILD OK"</automated>
  </verify>
  <done>
  - `buildDisplayList()` skips ALL skills (not just unselected) when group is collapsed
  - First-build default is `collapsed = false` (all groups expanded)
  - Binary builds cleanly with no errors
  </done>
</task>

</tasks>

<verification>
- `go build` succeeds
- `go vet ./internal/tui/...` passes
- Grep confirms: no `!m.skills[skillIdx].Selected` in the collapsed check
- Grep confirms: first-build default is `collapsed = false`
</verification>

<success_criteria>
Collapsing a group header with [enter] hides every skill in that group. Expanding it shows them all again. Groups start expanded on first load.
</success_criteria>

<output>
After completion, create `.planning/quick/8-allow-full-folder-collapse-in-manage-vie/8-SUMMARY.md`
</output>
