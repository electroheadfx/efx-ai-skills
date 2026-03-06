---
phase: quick-12
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: [internal/tui/app.go]
autonomous: true
requirements: [QUICK-12]

must_haves:
  truths:
    - "Config view section boxes render at correct terminal width on first open"
    - "Manage view renders at correct terminal width on first open"
  artifacts:
    - path: "internal/tui/app.go"
      provides: "Width/height restoration after sub-model creation"
      contains: "m.configModel.width = int(float64(m.width)"
  key_links:
    - from: "app.go openConfigMsg handler"
      to: "configModel.width"
      via: "width assignment after newConfigModel()"
      pattern: "configModel.width.*float64.*m.width"
    - from: "app.go openManageMsg handler"
      to: "manageModel.width and manageModel.height"
      via: "width/height assignment after newManageModel()"
      pattern: "manageModel.width.*float64.*m.width"
---

<objective>
Fix config and manage view section box borders having wrong width on first open.

Purpose: When opening config or manage views, newConfigModel()/newManageModel() creates a fresh struct with width=0, discarding the width already set by the initial WindowSizeMsg. The View() falls back to w=80 which doesn't match the actual terminal. Resizing the terminal fires a new WindowSizeMsg which fixes it -- but first open is broken.

Output: Corrected app.go that restores width (and height for manage) after creating fresh sub-models.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/app.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Restore width/height after sub-model creation in app.go</name>
  <files>internal/tui/app.go</files>
  <action>
In internal/tui/app.go, modify two message handlers to restore width/height after creating fresh sub-models:

1. openManageMsg handler (lines 92-95): After `m.manageModel = newManageModel(msg.provider)`, add:
   ```go
   m.manageModel.width = int(float64(m.width) * 0.9)
   m.manageModel.height = m.height
   ```

2. openConfigMsg handler (lines 97-100): After `m.configModel = newConfigModel()`, add:
   ```go
   m.configModel.width = int(float64(m.width) * 0.9)
   ```

The 0.9 multiplier matches the existing WindowSizeMsg handler (lines 86-90) which applies the same scaling.

After editing, build the binary:
```bash
go build -o ./bin/efx-skills ./cmd/efx-skills/
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && echo "Build succeeded"</automated>
  </verify>
  <done>Config and manage views receive correct width/height immediately on creation, matching the terminal dimensions already captured by the parent model. Build succeeds.</done>
</task>

</tasks>

<verification>
- `go build` succeeds with no errors
- The openConfigMsg handler sets configModel.width after newConfigModel()
- The openManageMsg handler sets manageModel.width and manageModel.height after newManageModel()
- The 0.9 scaling factor matches the WindowSizeMsg handler
</verification>

<success_criteria>
- Config view section boxes render at correct width on first open (no resize needed)
- Manage view renders at correct width on first open (no resize needed)
- Binary builds cleanly
</success_criteria>

<output>
After completion, create `.planning/quick/12-fix-config-view-box-borders-missing-widt/12-SUMMARY.md`
</output>
