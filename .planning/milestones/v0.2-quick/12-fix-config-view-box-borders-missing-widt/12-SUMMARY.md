---
phase: quick-12
plan: 01
subsystem: ui
tags: [bubbletea, tui, layout, width]

requires:
  - phase: quick-11
    provides: Config view section borders
provides:
  - Correct width/height propagation to config and manage sub-models on first open
affects: []

tech-stack:
  added: []
  patterns: [restore sub-model dimensions after fresh struct creation]

key-files:
  created: []
  modified: [internal/tui/app.go]

key-decisions:
  - "Width restoration uses same 0.9 scaling factor as WindowSizeMsg handler for consistency"

patterns-established:
  - "Sub-model creation pattern: always restore width/height from parent model after newXxxModel()"

requirements-completed: [QUICK-12]

duration: 1min
completed: 2026-03-06
---

# Quick 12: Fix Config View Box Borders Missing Width Summary

**Restore terminal width/height to config and manage sub-models after fresh struct creation on view open**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-06T17:26:14Z
- **Completed:** 2026-03-06T17:26:54Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Config view section boxes now render at correct terminal width on first open
- Manage view renders at correct terminal width and height on first open
- No more need to resize terminal to trigger correct layout

## Task Commits

Each task was committed atomically:

1. **Task 1: Restore width/height after sub-model creation in app.go** - `281f6dd` (fix)

**Plan metadata:** (pending)

## Files Created/Modified
- `internal/tui/app.go` - Added width/height restoration after newManageModel() and newConfigModel() calls

## Decisions Made
- Width restoration uses same 0.9 scaling factor as WindowSizeMsg handler for consistency

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Config and manage views now correctly sized on first open
- No follow-up work needed

---
*Phase: quick-12*
*Completed: 2026-03-06*
