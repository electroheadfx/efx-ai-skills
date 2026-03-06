---
phase: quick-6
plan: 1
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/manage.go
autonomous: true
must_haves:
  truths:
    - "Skills without a known provider appear under a group named 'custom' (not '_other')"
    - "Skills in the 'custom' group do NOT show a '(Custom)' registry label"
    - "Skills from known registries (Vercel, Playbooks) still show their registry label"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "Renamed group fallback and conditional registry label"
      contains: "\"custom\""
  key_links:
    - from: "extractGroup"
      to: "buildDisplayList rendering"
      via: "group name propagation"
      pattern: "return \"custom\""
---

<objective>
Rename the fallback group from '_other' to 'custom' and hide the redundant "(Custom)" registry label for skills in the custom group.

Purpose: Skills without a known provider are already grouped under a dedicated folder -- the group name 'custom' communicates their origin clearly, making the per-skill "(Custom)" label redundant visual noise.
Output: Updated manage.go with both changes.
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
  <name>Task 1: Rename _other group to custom and conditionally hide registry label</name>
  <files>internal/tui/manage.go</files>
  <action>
Two changes in manage.go:

1. In `extractGroup()` (line 197): Change `return "_other"` to `return "custom"`.

2. In the skill rendering block (around line 719): Replace the unconditional registry label append:
   ```go
   displayName += " (" + registryDisplayName(skill.Registry) + ")"
   ```
   with a conditional that only shows the label when the registry is NOT empty (i.e., skip it for custom skills whose Registry field is ""):
   ```go
   if skill.Registry != "" {
       displayName += " (" + registryDisplayName(skill.Registry) + ")"
   }
   ```
   This hides "(Custom)" for skills in the custom group (which have Registry=="") while preserving "(Vercel)" and "(Playbooks)" labels for skills from known registries.

After changes, run: `go build -o ./bin/efx-skills ./cmd/efx-skills/`
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && grep -q '"custom"' internal/tui/manage.go && grep -q 'skill.Registry != ""' internal/tui/manage.go && echo "PASS"</automated>
  </verify>
  <done>
    - extractGroup returns "custom" instead of "_other"
    - Skills with empty Registry (custom skills) show no registry label in manage view
    - Skills with a known registry still show their label (e.g., "(Vercel)")
    - Binary builds cleanly
  </done>
</task>

</tasks>

<verification>
- `go build -o ./bin/efx-skills ./cmd/efx-skills/` compiles without errors
- `grep '"_other"' internal/tui/manage.go` returns no matches (old name removed)
- `grep '"custom"' internal/tui/manage.go` returns match in extractGroup
- `grep 'skill.Registry != ""' internal/tui/manage.go` confirms conditional label logic
</verification>

<success_criteria>
The manage view shows the group as "custom" and skills in that group display without the redundant "(Custom)" label. Known-registry skills retain their labels.
</success_criteria>

<output>
After completion, create `.planning/quick/6-rename-other-group-to-custom-and-hide-re/6-SUMMARY.md`
</output>
