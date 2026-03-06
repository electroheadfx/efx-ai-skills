---
status: diagnosed
phase: 02-origin-tracking
source: [02-01-SUMMARY.md, 02-02-SUMMARY.md, 02-03-SUMMARY.md]
started: 2026-03-06T11:00:00Z
updated: 2026-03-06T11:20:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Registry Column in Search
expected: Open the TUI search view and search for any skill. Results should display 4 columns: Name, Source, Registry (friendly names like "Vercel" or "Playbooks"), and Popularity.
result: pass

### 2. Install Writes Full Metadata
expected: Install a skill from the search view. After installation, check config.json — the skill entry should include owner, name, registry, url, version (commit hash), and installed (RFC3339 timestamp).
result: pass

### 3. Registry Origin in Manage View
expected: Open the manage view for a provider. Each skill should show its registry origin next to the name (e.g., "my-skill (Vercel)").
result: pass

### 4. Toggle Install/Remove Help Bar
expected: In the manage view, the help bar should say "[t] toggle install/remove" instead of just "[t] toggle".
result: issue
reported: "[t] remove skill acts like a toggle like before. I want [r] remove (from the specs) -- remove skill physically from ~/agents/skills/ and from config. It's dangerous so add a confirm alert. [t] toggle should only link/unlink symlink from provider."
severity: major

### 5. Removal Cleans Metadata
expected: In manage view, toggle off a skill linked to only one provider and press [s] to apply. Check config.json — the skill's entry should be removed from the skills array.
result: issue
reported: "Toggle just removes the link symbol from provider and does NOT remove config metadata. ONLY [r] Remove should remove skill from ~/agents/skills/* and from config for the selected skill. Toggle should not touch config or physical files."
severity: major

## Summary

total: 5
passed: 3
issues: 2
pending: 0
skipped: 0

## Gaps

- truth: "[t] toggle only links/unlinks symlink from provider; [r] remove physically deletes skill from ~/agents/skills/ and removes from config.json with confirmation"
  status: failed
  reason: "User reported: [t] remove skill acts like a toggle like before. I want [r] remove (from the specs) -- remove skill physically from ~/agents/skills/ and from config. It's dangerous so add a confirm alert. [t] toggle should only link/unlink symlink from provider."
  severity: major
  test: 4
  root_cause: "applySkillChanges in manage.go lines 558-577 conflates symlink toggle with config cleanup. When skill unlinked from all providers, it auto-calls removeSkillFromConfig. Help bar labels [t] as 'toggle install/remove' conflating both concepts. No [r] remove keybinding exists."
  artifacts:
    - path: "internal/tui/manage.go"
      issue: "Lines 558-577: config removal block inside applySkillChanges must be removed. Line 733: help bar conflates toggle/remove. Lines 424-433: [t] handler needs [r] sibling."
    - path: "internal/tui/manage.go"
      issue: "Lines 40-52: manageModel struct needs confirmation state fields (confirmingRemove, removeTarget)"
    - path: "internal/skill/store.go"
      issue: "Missing RemoveFromLock method needed by new remove flow"
    - path: "internal/tui/config.go"
      issue: "removeSkillFromConfig correct but must only be called from [r] remove, not from applySkillChanges"
  missing:
    - "Remove lines 558-577 from applySkillChanges to make toggle symlink-only"
    - "Add [r] keybinding with confirmation dialog for full skill removal"
    - "Add RemoveFromLock to store.go"
    - "New removeSkillFully function: unlink all providers, removeSkillFromConfig, RemoveFromLock, os.RemoveAll on skill dir"
    - "Update help bar: [t] toggle link, [r] remove skill"
  debug_session: ".planning/debug/manage-toggle-vs-remove.md"

- truth: "Config metadata only removed via [r] remove action, not via [t] toggle; toggle only manages symlinks"
  status: failed
  reason: "User reported: Toggle just removes the link symbol from provider and does NOT remove config metadata. ONLY [r] Remove should remove skill from ~/agents/skills/* and from config for the selected skill. Toggle should not touch config or physical files."
  severity: major
  test: 5
  root_cause: "Same root cause as test 4: applySkillChanges lines 558-577 auto-remove config when skill unlinked from all providers. Physical deletion (os.RemoveAll on skill directory) does not exist anywhere in codebase."
  artifacts:
    - path: "internal/tui/manage.go"
      issue: "Lines 558-577: config removal triggered by toggle path instead of dedicated remove action"
    - path: "internal/skill/store.go"
      issue: "No method to physically remove skill directory or remove lock entry"
  missing:
    - "Same fixes as test 4 — both gaps share the same root cause"
  debug_session: ".planning/debug/manage-toggle-vs-remove.md"
