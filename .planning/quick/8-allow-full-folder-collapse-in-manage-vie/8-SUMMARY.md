---
phase: quick-8
plan: 01
subsystem: ui
tags: [tui, bubbletea, manage-view, collapse]

requires:
  - phase: quick-3
    provides: "Group header collapse/expand rendering"
provides:
  - "Full folder collapse hides ALL skills in group (not just unselected)"
  - "All groups start expanded on first load"
affects: [manage-view, tui]

tech-stack:
  added: []
  patterns: ["Full collapse pattern: skip all skills when collapsed, no exceptions"]

key-files:
  created: []
  modified:
    - internal/tui/manage.go

key-decisions:
  - "Removed hasInstalled variable entirely since default is now unconditionally false"

patterns-established:
  - "Collapse = hide everything: collapsed groups show only the header, no skills"

requirements-completed: [QUICK-8]

duration: 1min
completed: 2026-03-06
---

# Quick Task 8: Allow Full Folder Collapse in Manage View Summary

**Collapsed groups now hide ALL skills (including active/installed) and all groups start expanded on first load**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-06T16:29:23Z
- **Completed:** 2026-03-06T16:30:22Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Collapsed groups hide every skill in the group, giving users full control over clutter
- All groups start expanded on first load (removed conditional `!hasInstalled` default)
- Removed dead `hasInstalled` variable that became unused after the change

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix collapse behavior to hide all skills and default groups to expanded** - `addde1b` (feat)

## Files Created/Modified
- `internal/tui/manage.go` - Changed `buildDisplayList()` to skip ALL skills when collapsed, default collapsed state to false

## Decisions Made
- Removed `hasInstalled` loop and variable entirely rather than keeping dead code, since the default collapsed state is now unconditionally `false`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed unused hasInstalled variable**
- **Found during:** Task 1 (Fix collapse behavior)
- **Issue:** After changing default from `collapsed = !hasInstalled` to `collapsed = false`, the `hasInstalled` variable became unused, causing Go compilation failure
- **Fix:** Removed the `hasInstalled` variable and its associated loop
- **Files modified:** internal/tui/manage.go
- **Verification:** `go build` and `go vet` pass cleanly
- **Committed in:** addde1b (part of task commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary for compilation. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Full collapse behavior is ready for use
- No blockers or concerns

---
*Phase: quick-8*
*Completed: 2026-03-06*
