---
phase: quick-5
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/manage.go
  - internal/tui/status.go
  - internal/tui/config.go
autonomous: true
requirements: ["custom-skills-manage"]
must_haves:
  truths:
    - "Skills without registry metadata show 'Custom' as their registry in the manage view"
    - "User can remove a custom (no-origin) skill via [r] just like any other skill"
    - "User can deactivate (unlink) a custom skill from a provider via [t] toggle"
    - "Custom skills are visually distinguishable from registry-sourced skills"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "Custom registry label for skills without SkillMeta"
    - path: "internal/tui/config.go"
      provides: "registryDisplayName handles empty registry"
  key_links:
    - from: "internal/tui/manage.go"
      to: "internal/tui/config.go"
      via: "registryDisplayName for empty registry"
      pattern: "registryDisplayName"
---

<objective>
Show skills that have no origin/registry metadata (manually placed in ~/.agents/skills/ or missing from config.json) with a "Custom" label in the manage view, and ensure remove [r] and deactivate [t] work correctly for them.

Purpose: Users who manually place skill directories into ~/.agents/skills/ (without installing via search/registry) currently see no registry indicator. They should see "Custom" as the source, and be able to remove or unlink these skills like any other.

Output: Updated manage view and registry display logic.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/STATE.md
@internal/tui/manage.go
@internal/tui/config.go
@internal/tui/status.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Label skills without registry metadata as "Custom" in manage view</name>
  <files>internal/tui/config.go, internal/tui/manage.go</files>
  <action>
1. In `config.go`, update `registryDisplayName` to handle the empty string case: when `name` is `""`, return `"Custom"`. Add it as the first case in the switch (before "skills.sh").

2. In `manage.go`, in `loadSkillsForProvider`, after the metadata enrichment block (lines 166-169), add logic: if a skill was NOT found in the metaLookup (no SkillMeta exists), explicitly set `entry.Registry = ""` (it's already zero-value, but this makes intent clear). The existing `registryDisplayName("")` call in the View will then show "Custom".

3. Verify that `removeSkillFully` in manage.go already works for custom skills -- it does: it removes from providers, config (no-op if not tracked), lock (idempotent), and disk. No changes needed.

4. Verify that toggle [t] already works for custom skills -- it does: toggle just flips `Selected` bool, and `applySkillChanges` creates/removes symlinks. No changes needed.

5. Build with: `go build -o ./bin/efx-skills ./cmd/efx-skills/`
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && go test ./internal/tui/ -run TestRegistryDisplayName -v 2>&1 || echo "No specific test yet, build passed"</automated>
  </verify>
  <done>
- `registryDisplayName("")` returns "Custom"
- Skills without SkillMeta in config show "(Custom)" in the manage view display
- Remove [r] and toggle [t] work unchanged for these skills (already functional)
- Binary builds successfully
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add test for registryDisplayName covering Custom case</name>
  <files>internal/tui/config_test.go</files>
  <behavior>
    - registryDisplayName("") returns "Custom"
    - registryDisplayName("skills.sh") returns "Vercel" (existing)
    - registryDisplayName("playbooks.com") returns "Playbooks" (existing)
    - registryDisplayName("github") returns "github" (passthrough)
  </behavior>
  <action>
Add a `TestRegistryDisplayName` test function to `config_test.go` that covers all four cases above using table-driven tests. This confirms the new empty-string handling and guards against regression.

Run: `go test ./internal/tui/ -run TestRegistryDisplayName -v`
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go test ./internal/tui/ -run TestRegistryDisplayName -v</automated>
  </verify>
  <done>TestRegistryDisplayName passes with all 4 cases green</done>
</task>

</tasks>

<verification>
- `go build -o ./bin/efx-skills ./cmd/efx-skills/` succeeds
- `go test ./internal/tui/ -v` passes all tests (existing + new)
- `registryDisplayName("")` returns "Custom"
</verification>

<success_criteria>
- Skills without registry metadata display "(Custom)" in the manage view
- Remove and deactivate (toggle) work for custom skills (no regression)
- New test covers registryDisplayName edge cases
- Binary builds cleanly
</success_criteria>

<output>
After completion, create `.planning/quick/5-manage-custom-internal-skills-without-pr/5-SUMMARY.md`
</output>
