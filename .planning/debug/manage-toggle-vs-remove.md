---
status: diagnosed
trigger: "Diagnose two related UAT gaps: [t] toggle should ONLY toggle symlinks, new [r] remove should physically delete skill + remove from config"
created: 2026-03-06T00:00:00Z
updated: 2026-03-06T00:00:00Z
---

## Current Focus

hypothesis: Toggle (`[t]`) and apply (`[s]`) currently couple three concerns: symlink management, config metadata removal, and physical file deletion intent -- all need separation
test: Traced full code path from `[t]` through `[s]` apply to identify every side effect
expecting: Map exactly which lines perform symlink-only work vs config removal vs physical deletion
next_action: Return diagnosis with line-level artifact map

## Symptoms

expected: `[t]` toggle should ONLY link/unlink symlinks from provider. A NEW `[r]` remove keybinding should physically delete skill from skills path AND remove from config.json, with confirmation dialog.
actual: `[t]` toggle marks skills for add/remove, then `[s]` apply performs ALL of: symlink creation/removal, config metadata removal (removeSkillFromConfig), but does NOT physically delete the skill directory from central storage.
errors: none
reproduction: Open manage view, toggle a linked skill off, press `[s]` apply -- symlink removed AND config entry removed
started: By design -- toggle+apply was the only removal mechanism

## Eliminated

- hypothesis: Toggle directly deletes physical skill files
  evidence: `applySkillChanges` at manage.go:554 calls `os.RemoveAll(linkPath)` where linkPath is the SYMLINK in the provider directory, NOT the skill directory in central storage. No code path removes `~/.agents/skills/{name}/` from the manage view.
  timestamp: 2026-03-06

- hypothesis: Physical skill deletion already exists somewhere
  evidence: Searched entire codebase for `RemoveAll` in tui package -- only hit is manage.go:554 which removes symlinks. `store.go` has no `Remove`, `Delete`, or `Uninstall` function. `RemoveFromLock` is mentioned only in research notes as a future need.
  timestamp: 2026-03-06

## Evidence

- timestamp: 2026-03-06
  checked: manage.go `[t]` key handler (lines 424-433)
  found: |
    `[t]` toggles `skill.Selected` boolean. For groups, calls `toggleGroup()` which flips all skills in the group.
    This is ONLY an in-memory state change -- no filesystem or config mutation happens here.
    The actual side effects happen later when user presses `[s]`.
  implication: The `[t]` key itself is benign -- the problem is in `applySkillChanges`

- timestamp: 2026-03-06
  checked: manage.go `[s]` key handler (lines 444-449)
  found: |
    `[s]` calls `applySkillChanges(m.provider, m.skills)` then reloads skills.
    This is the function that performs ALL side effects.
  implication: `applySkillChanges` is the single function that needs surgical separation

- timestamp: 2026-03-06
  checked: manage.go `applySkillChanges` function (lines 538-578) -- CRITICAL FUNCTION
  found: |
    Three distinct operations happen in sequence:

    **Operation 1 -- Symlink management (lines 546-556):**
    For each skill:
    - If `Selected && !Linked`: creates symlink from provider to central storage (line 552)
    - If `!Selected && Linked`: removes symlink from provider via `os.RemoveAll(linkPath)` (line 554)
    linkPath = provider path + skill name (the SYMLINK, not the actual skill dir)

    **Operation 2 -- Config metadata removal (lines 558-577):**
    For skills where `!Selected && Linked` (just unlinked):
    - Calls `detectProviders()` to get ALL providers (line 559)
    - Checks each configured provider for remaining symlinks via `os.Lstat` (lines 564-570)
    - If skill is NOT linked to ANY provider, calls `removeSkillFromConfig(skill.Name)` (line 574)

    **Operation 3 -- Physical deletion: DOES NOT EXIST**
    No code anywhere in this function (or elsewhere in manage.go) removes the actual skill directory from `~/.agents/skills/`.
  implication: |
    GAP 1 (Test 4): Operation 2 (config removal) is coupled to Operation 1 (symlink toggle) -- they should be independent.
    GAP 2 (Test 5): Same issue -- unlinking from all providers triggers config removal automatically.
    MISSING: Physical skill deletion (`os.RemoveAll` on `~/.agents/skills/{name}/`) does not exist anywhere.

- timestamp: 2026-03-06
  checked: config.go `removeSkillFromConfig` (lines 548-563)
  found: |
    Loads config, filters out the SkillMeta entry matching the skill name, saves config.
    Called from exactly one place: manage.go:574 inside `applySkillChanges`.
  implication: This function is correct on its own -- the problem is WHEN it gets called (coupled to toggle+apply instead of explicit remove)

- timestamp: 2026-03-06
  checked: store.go for physical removal capabilities
  found: |
    `UnlinkFromProvider` (line 138): removes symlink from provider -- exists and works
    `RemoveFromLock` / `RemoveSkill` / `DeleteSkill`: DO NOT EXIST
    No function to physically delete a skill directory or remove its lock entry
  implication: New store functions needed for `[r]` remove: physical dir deletion + lock entry removal

- timestamp: 2026-03-06
  checked: manage.go help bar (line 733)
  found: |
    `[space] preview  [o] open  [v] verify  [u] update  [g] update all  [t] toggle install/remove  [enter] collapse/expand  [a] all  [n] none  [s] apply/save  [<-/->] page  [esc] back`
    Note: `[t]` is labeled "toggle install/remove" -- this label itself conflates two operations.
  implication: Help bar needs `[r] remove` added; `[t]` label should change to "toggle link/unlink" or just "toggle"

- timestamp: 2026-03-06
  checked: search.go install flow (lines 107-147) for comparison
  found: |
    Install does: store.Install (download) -> store.AddToLock (lock entry) -> addSkillToConfig (config meta) -> store.LinkToProvider (symlinks)
    The inverse for `[r]` remove should do: unlink all providers -> removeSkillFromConfig -> remove lock entry -> delete physical dir
  implication: Full lifecycle needs matching teardown path

## Resolution

root_cause: |
  `applySkillChanges` in manage.go (lines 538-578) conflates TWO conceptually different operations:

  1. **Symlink toggle** (link/unlink from a provider) -- should be the ONLY thing `[t]` + `[s]` does
  2. **Config metadata cleanup** (removeSkillFromConfig when unlinked from all providers) -- should ONLY happen via explicit `[r]` remove

  Additionally, physical deletion of the skill directory from `~/.agents/skills/` does not exist at all.
  The `[r]` remove operation needs to be built from scratch.

fix: |
  **What needs to change -- 4 artifacts, 7 modifications:**

  ### 1. `internal/tui/manage.go` -- 4 changes

  **Change A: Remove config cleanup from `applySkillChanges` (lines 558-577)**
  DELETE the entire block that checks multi-provider linkage and calls `removeSkillFromConfig`.
  After this change, `applySkillChanges` becomes a pure symlink manager (create/remove symlinks only).

  **Change B: Add `[r]` key handler in `Update()` (after line 433, alongside other key cases)**
  New `case "r":` that:
  - Gets the currently selected skill (skip if group header)
  - Shows a confirmation prompt/dialog: "Remove {name}? This will delete it from disk and config. [y/n]"
  - On confirm: calls a new `removeSkillFully(skillName)` function
  - On cancel: returns to normal view

  Note: Bubble Tea confirmation can be done via a `confirmingRemove bool` + `removeTarget string` field on `manageModel`,
  with a `case "y"` / `case "n"` handler when `confirmingRemove` is true. The View() renders the confirmation
  prompt when this flag is set.

  **Change C: Add `removeSkillFully` function (new function)**
  ```
  func removeSkillFully(skillName string) error {
      // 1. Unlink from ALL providers
      for _, p := range detectProviders() {
          if p.Configured {
              linkPath := filepath.Join(p.Path, skillName)
              os.Remove(linkPath)  // remove symlink, ignore error if not linked
          }
      }
      // 2. Remove from config.json
      removeSkillFromConfig(skillName)
      // 3. Remove from lock file
      store := skill.NewStore(getSkillsPath())
      store.RemoveFromLock(skillName)
      // 4. Physically delete skill directory
      skillsPath := getSkillsPath()
      os.RemoveAll(filepath.Join(skillsPath, skillName))
      return nil
  }
  ```

  **Change D: Update help bar (line 733)**
  Change `[t] toggle install/remove` to `[t] toggle`
  Add `[r] remove` to the help bar

  ### 2. `internal/skill/store.go` -- 1 change

  **Change E: Add `RemoveFromLock` function (new method on Store)**
  ```
  func (s *Store) RemoveFromLock(skillName string) error {
      lock, err := s.ReadLockFile()
      if err != nil { return err }
      delete(lock.Skills, skillName)
      return s.WriteLockFile(lock)
  }
  ```

  ### 3. `internal/tui/manage.go` model -- 1 change

  **Change F: Add confirmation state fields to `manageModel` (around line 40)**
  ```
  confirmingRemove bool
  removeTarget     string
  ```
  The `View()` method needs a confirmation overlay/prompt when `confirmingRemove` is true.
  The `Update()` method needs `case "y"` and `case "n"` handling when `confirmingRemove` is true
  (before the normal key dispatch).

  ### 4. `internal/tui/manage.go` View -- 1 change

  **Change G: Add confirmation dialog rendering in `View()`**
  When `m.confirmingRemove` is true, render a confirmation line like:
  `"Remove {skillName}? This will delete it from disk and remove from config. [y] confirm [n] cancel"`

  ### Summary of line-level artifacts:

  | File | Lines | What | Action |
  |------|-------|------|--------|
  | `internal/tui/manage.go` | 558-577 | Config cleanup in `applySkillChanges` | DELETE entire block |
  | `internal/tui/manage.go` | 424-433 | After `[t]` case | ADD new `case "r":` block |
  | `internal/tui/manage.go` | 40-52 | `manageModel` struct | ADD `confirmingRemove bool`, `removeTarget string` |
  | `internal/tui/manage.go` | 359 | Top of `tea.KeyMsg` switch | ADD early-return for confirmation y/n when `confirmingRemove` |
  | `internal/tui/manage.go` | after 578 | After `applySkillChanges` | ADD new `removeSkillFully` function |
  | `internal/tui/manage.go` | 733 | Help bar | CHANGE `[t] toggle install/remove` to `[t] toggle`, ADD `[r] remove` |
  | `internal/tui/manage.go` | ~700 (View) | Before help bar | ADD confirmation dialog rendering |
  | `internal/skill/store.go` | after 240 | After `AddToLock` | ADD `RemoveFromLock` method |

verification: N/A (diagnosis only)
files_changed: []
