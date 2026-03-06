---
phase: quick-7
plan: 1
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/manage.go
  - internal/tui/config_test.go
autonomous: true
requirements: [QUICK-7]

must_haves:
  truths:
    - "Custom skills from ~/.agents/skills/ show 'agents' origin label in manage view"
    - "Custom skills only in a provider path show 'local provider' origin label in manage view"
    - "Registry-sourced skills still show their registry display name unchanged"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "Origin field on SkillEntry, detection in loadSkillsForProvider, display in View"
      contains: "Origin"
    - path: "internal/tui/config_test.go"
      provides: "Test for origin display logic"
  key_links:
    - from: "loadSkillsForProvider"
      to: "SkillEntry.Origin"
      via: "centralNames membership check"
      pattern: "Origin.*agents|Origin.*local provider"
---

<objective>
Show custom skill origin type ("agents" or "local provider") for skills without registry metadata in the manage view.

Purpose: Custom skills currently show no origin info. Users need to distinguish between skills installed from central ~/.agents/skills/ storage vs skills that exist only in a provider's local path (e.g., ~/.claude/skills/).
Output: Updated manage view with origin labels on custom skills.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/manage.go
@internal/tui/config.go
@internal/tui/config_test.go
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add Origin field to SkillEntry and detect origin in loadSkillsForProvider</name>
  <files>internal/tui/manage.go, internal/tui/config_test.go</files>
  <behavior>
    - Test 1: A custom skill (no SkillMeta) present in central storage (~/.agents/skills/) gets Origin="agents"
    - Test 2: A custom skill only in provider path (not in central storage) gets Origin="local provider"
    - Test 3: A registry skill (has SkillMeta) gets Origin="" (origin label not used for registry skills)
  </behavior>
  <action>
    1. Add `Origin string` field to the `SkillEntry` struct (after the `Owner` field, line 29).

    2. In `loadSkillsForProvider`, update the skill creation loop (lines 158-173). After setting Registry/Owner from metaLookup, add origin detection for custom skills:
       - If the skill has no SkillMeta (the `else` branch at line 169), determine origin:
         - If `centralNames[name]` is true: set `entry.Origin = "agents"` (skill exists in ~/.agents/skills/)
         - If `centralNames[name]` is false (provider-only skill): set `entry.Origin = "local provider"`
       - If the skill HAS SkillMeta (registry skill): leave Origin as "" (unused)

    3. In `View()`, update the skill display label (lines 719-721). Replace:
       ```go
       if skill.Registry != "" {
           displayName += " (" + registryDisplayName(skill.Registry) + ")"
       }
       ```
       With:
       ```go
       if skill.Registry != "" {
           displayName += " (" + registryDisplayName(skill.Registry) + ")"
       } else if skill.Origin != "" {
           displayName += " (" + skill.Origin + ")"
       }
       ```
       This shows the registry name for registry skills, or the origin label for custom skills.

    4. Add a test `TestSkillEntryOriginLabel` in config_test.go that verifies:
       - A SkillEntry with Registry="skills.sh" and Origin="" displays "(Vercel)" via registryDisplayName
       - A SkillEntry with Registry="" and Origin="agents" would display "(agents)"
       - A SkillEntry with Registry="" and Origin="local provider" would display "(local provider)"
       - A SkillEntry with Registry="" and Origin="" displays no parenthetical (both empty)
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go test ./internal/tui/ -run TestSkillEntryOriginLabel -v && go build -o ./bin/efx-skills ./cmd/efx-skills/</automated>
  </verify>
  <done>Custom skills in manage view show "(agents)" or "(local provider)" based on their filesystem origin. Registry skills continue to show their registry display name. Build succeeds.</done>
</task>

</tasks>

<verification>
- `go test ./internal/tui/ -run TestSkillEntryOriginLabel -v` passes
- `go build -o ./bin/efx-skills ./cmd/efx-skills/` succeeds
- `go vet ./...` passes
</verification>

<success_criteria>
- Custom skills from ~/.agents/skills/ show "(agents)" label in the manage view
- Custom skills only in a provider path show "(local provider)" label
- Registry-sourced skills still show their registry display name (Vercel, Playbooks, etc.)
- No label shown for skills with both empty Registry and empty Origin
</success_criteria>

<output>
After completion, create `.planning/quick/7-show-custom-skill-origin-type-agents-fro/7-SUMMARY.md`
</output>
