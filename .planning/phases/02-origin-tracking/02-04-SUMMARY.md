---
phase: 02-origin-tracking
plan: 04
subsystem: tui
tags: [bubbletea, symlink, confirmation-dialog, lock-file]

# Dependency graph
requires:
  - phase: 02-origin-tracking
    provides: "removeSkillFromConfig, addSkillToConfig, SkillMeta config management"
provides:
  - "RemoveFromLock method on Store for lock file cleanup"
  - "removeSkillFully function for complete skill deletion"
  - "Confirmation dialog pattern for destructive actions in TUI"
  - "Separated toggle (symlink-only) from remove (full deletion)"
affects: [06-doctor]

# Tech tracking
tech-stack:
  added: []
  patterns: ["confirmation dialog with key interception in bubbletea", "symlink-only toggle vs full removal separation"]

key-files:
  created: []
  modified:
    - internal/skill/store.go
    - internal/skill/store_test.go
    - internal/tui/manage.go

key-decisions:
  - "[Phase 02]: applySkillChanges is now symlink-only -- no config removal on unlink"
  - "[Phase 02]: removeSkillFully handles full deletion: providers, config, lock, disk"
  - "[Phase 02]: Confirmation dialog intercepts all keys via confirmingRemove flag"
  - "[Phase 02]: RemoveFromLock is idempotent -- no-op for missing skills"

patterns-established:
  - "Confirmation dialog: set confirmingRemove=true, intercept keys before main switch, [y] confirms, [n]/esc cancels"
  - "Destructive action separation: non-destructive toggle vs destructive remove with explicit confirmation"

requirements-completed: [ORIG-06]

# Metrics
duration: 15min
completed: 2026-03-06
---

# Phase 02 Plan 04: Toggle/Remove Separation Summary

**Separated [t] toggle (symlink-only) from [r] remove (full deletion with confirmation dialog) in manage view, with RemoveFromLock store method**

## Performance

- **Duration:** 15 min
- **Started:** 2026-03-06T11:52:01Z
- **Completed:** 2026-03-06T12:07:00Z
- **Tasks:** 3 (2 auto + 1 human-verify checkpoint)
- **Files modified:** 3

## Accomplishments
- `applySkillChanges` decoupled to symlink-only -- pressing [t]+[s] no longer removes config metadata or physical files
- `RemoveFromLock` method added to Store with 3 passing tests (remove, no-op for missing, empty lock)
- `removeSkillFully` function performs complete deletion: unlinks all providers, removes from config.json, removes from lock file, deletes from disk
- Help bar shows `[t] toggle` and `[r] remove` as clearly separate actions
- Confirmation dialog with warning styling intercepts all keys until user confirms or cancels

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): Add RemoveFromLock tests** - `51f77f5` (test)
2. **Task 1 (GREEN): Implement RemoveFromLock** - `4f92576` (feat)
3. **Task 1 (REFACTOR): Decouple applySkillChanges** - `7b07415` (refactor)
4. **Task 2: Add [r] remove with confirmation dialog** - `ece5b88` (feat)
5. **Task 3: Human verification** - approved by user

## Files Created/Modified
- `internal/skill/store.go` - Added RemoveFromLock method (read lock, delete entry, write back)
- `internal/skill/store_test.go` - 3 new tests for RemoveFromLock (52 lines added)
- `internal/tui/manage.go` - Confirmation dialog, [r] handler, removeSkillFully, symlink-only applySkillChanges, updated help bar

## Decisions Made
- `applySkillChanges` is now purely a symlink manager -- config removal responsibility moved to `removeSkillFully`
- `removeSkillFully` handles full deletion in order: unlink all providers, remove config, remove lock, delete disk
- Confirmation dialog uses `confirmingRemove` flag to intercept all keys before the main switch statement
- `RemoveFromLock` is idempotent -- removing a nonexistent skill is a silent no-op

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- UAT tests 4 and 5 (toggle vs remove separation) are now satisfied
- All toggle/remove semantics clearly separated for user understanding
- RemoveFromLock available for any future lock file cleanup needs

## Self-Check: PASSED

All files exist. All commits verified.

---
*Phase: 02-origin-tracking*
*Completed: 2026-03-06*
