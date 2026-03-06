---
phase: quick-10
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/manage.go
autonomous: true
requirements: []
must_haves:
  truths:
    - "Removing a local provider skill (actual directory, not symlink) deletes it from the provider path"
    - "Removing a regular skill (symlink) still works correctly"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "removeSkillFully with os.RemoveAll for provider paths"
      contains: "os.RemoveAll(linkPath)"
  key_links:
    - from: "internal/tui/manage.go"
      to: "os.RemoveAll"
      via: "removeSkillFully step 1"
      pattern: "os\\.RemoveAll\\(linkPath\\)"
---

<objective>
Fix removeSkillFully to delete local provider skills that exist as actual directories (not symlinks) in provider paths.

Purpose: Currently `os.Remove(linkPath)` silently fails on non-empty directories, leaving local provider skills undeletable. `os.RemoveAll` handles both symlinks (removes just the symlink, does not follow) and directories (removes recursively).
Output: One-line fix in manage.go, rebuilt binary.
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
  <name>Task 1: Change os.Remove to os.RemoveAll in removeSkillFully step 1</name>
  <files>internal/tui/manage.go</files>
  <action>
In `removeSkillFully()` at the provider unlinking loop (step 1), change:
```go
os.Remove(linkPath) // ignore error if not linked
```
to:
```go
os.RemoveAll(linkPath) // handles both symlinks and directories
```

This is safe because `os.RemoveAll` on a symlink removes only the symlink itself (does not follow it), so behavior for regular registry-installed skills (symlinks) is unchanged. For local provider skills (actual directories), it correctly removes the directory and all contents.

After the fix, rebuild the binary:
```bash
go build -o ./bin/efx-skills ./cmd/efx-skills/
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && grep -n "os.RemoveAll(linkPath)" internal/tui/manage.go && go build -o ./bin/efx-skills ./cmd/efx-skills/ && echo "BUILD OK"</automated>
  </verify>
  <done>Line 649 uses os.RemoveAll(linkPath) instead of os.Remove(linkPath). Binary builds successfully.</done>
</task>

</tasks>

<verification>
- `grep "os.RemoveAll(linkPath)" internal/tui/manage.go` returns the changed line
- `go build -o ./bin/efx-skills ./cmd/efx-skills/` succeeds
- `grep "os.Remove(linkPath)" internal/tui/manage.go` returns nothing (old call removed)
</verification>

<success_criteria>
removeSkillFully uses os.RemoveAll for provider path cleanup, binary compiles cleanly.
</success_criteria>

<output>
After completion, create `.planning/quick/10-fix-remove-for-local-provider-skills-rem/10-SUMMARY.md`
</output>
