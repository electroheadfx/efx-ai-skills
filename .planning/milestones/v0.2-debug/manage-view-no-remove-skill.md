---
status: diagnosed
trigger: "Diagnose why there's no way to remove/uninstall a skill from the manage view"
created: 2026-03-06T00:00:00Z
updated: 2026-03-06T00:00:00Z
---

## Current Focus

hypothesis: Removal IS implemented but hidden behind a two-step toggle+apply workflow with no dedicated key
test: Traced full code path from key binding to filesystem removal
expecting: Confirm whether removal exists or is dead code
next_action: Return diagnosis

## Symptoms

expected: User should be able to remove/uninstall a skill from the manage view
actual: Help bar shows no remove/uninstall/delete key binding
errors: none
reproduction: Open manage provider view, look at help bar
started: By design - never had a dedicated remove key

## Eliminated

- hypothesis: removeSkillFromConfig is dead code
  evidence: Called at manage.go:558 inside applySkillChanges, which is triggered by 's' key at manage.go:431
  timestamp: 2026-03-06

- hypothesis: Removal logic is missing entirely
  evidence: Full removal flow exists: toggle deselects -> apply removes symlink -> multi-provider check -> removeSkillFromConfig
  timestamp: 2026-03-06

## Evidence

- timestamp: 2026-03-06
  checked: manage.go key bindings (lines 344-514)
  found: 12 key bindings exist - space(preview), o(open), v(verify), u(update), g(update all), t(toggle), enter(collapse/expand), a(select all), n(select none), s(apply/save), navigation keys
  implication: No dedicated remove/delete key exists

- timestamp: 2026-03-06
  checked: manage.go toggle behavior (lines 408-417)
  found: 't' key toggles skill.Selected boolean. When a linked skill is deselected (Selected=false, Linked=true), it is marked for removal
  implication: Toggle IS the removal mechanism - deselecting a linked skill marks it for removal

- timestamp: 2026-03-06
  checked: manage.go View rendering (lines 660-666)
  found: When skill.Linked && !skill.Selected, the UI shows "(remove)" status annotation in warn style
  implication: The UI does visually indicate pending removal, but only AFTER toggling

- timestamp: 2026-03-06
  checked: manage.go applySkillChanges (lines 522-562)
  found: On 's' key press, applySkillChanges iterates skills. For !Selected && Linked: removes symlink (os.RemoveAll), then checks all providers, calls removeSkillFromConfig if unlinked from all
  implication: Full removal pipeline works - symlink removal + config cleanup + multi-provider safety check

- timestamp: 2026-03-06
  checked: manage.go help bar (line 714)
  found: Help text is "[space] preview [o] open [v] verify [u] update [g] update all [t] toggle [enter] collapse/expand [a] all [n] none [s] apply/save [arrows] page [esc] back"
  implication: No mention of "remove" or "uninstall" anywhere in the help bar

- timestamp: 2026-03-06
  checked: manage.go initial state (lines 145-153)
  found: SkillEntry.Selected is initialized to skill.Linked - linked skills start as selected
  implication: Removing = deselecting a currently-linked skill, then applying

## Resolution

root_cause: Removal is NOT missing - it is a two-step implicit workflow (toggle to deselect + apply/save) but the UX makes this non-obvious because (1) the help bar has no "remove" hint, (2) "toggle" does not communicate that it can mark for removal, and (3) the user must know to press 's' to commit the removal
fix: N/A (diagnosis only)
verification: N/A
files_changed: []
