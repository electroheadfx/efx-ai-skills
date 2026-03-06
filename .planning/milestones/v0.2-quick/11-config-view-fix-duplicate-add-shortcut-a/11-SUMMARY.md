---
phase: quick-11
plan: 11
subsystem: ui
tags: [tui, lipgloss, config, border-box, help-bar]

requires:
  - phase: quick-3
    provides: "TUI visual enhancements (box styles, selected/subtitle styles)"
  - phase: quick-4
    provides: "renderHelpBar responsive helper"
provides:
  - "Deduplicated config help bar (no [a]/[d] in bottom bar for repos section)"
  - "[r] remove repo key binding and inline hint"
  - "Blue border box for active config section, muted border for inactive"
affects: [config-view, tui-styles]

tech-stack:
  added: []
  patterns: ["configSectionStyle/configSectionActiveStyle for section focus borders"]

key-files:
  created: []
  modified:
    - internal/tui/config.go
    - internal/tui/styles.go

key-decisions:
  - "Merged d and r key bindings into single case for repo removal"
  - "Used Padding(0, 1) for compact config section borders vs Padding(1, 2) of boxStyle"
  - "Section width calculated as w-4 to account for border chars"

patterns-established:
  - "configSectionStyle/configSectionActiveStyle: border styles for focus indication in config view sections"

requirements-completed: [config-view-ui-fixes]

duration: 2min
completed: 2026-03-06
---

# Quick Task 11: Config View Fix Duplicate Add Shortcut Summary

**Deduplicated [a] add from help bar, added [r] remove repo shortcut, and wrapped config sections in blue/muted border boxes**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T17:16:04Z
- **Completed:** 2026-03-06T17:17:56Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Removed duplicate [a] add and [d] delete from bottom help bar when repos section is active (they appear inline instead)
- Added [r] key binding that removes selected repo (same behavior as existing [d] key)
- Inline hint now shows "[a] add repo  [r] remove repo" when repos exist
- All three config sections wrapped in rounded border boxes: blue for active, muted gray for inactive

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix help bar duplication and add [r] remove key binding** - `df1837b` (feat)
2. **Task 2: Add blue box border for active section** - `314a900` (feat)

## Files Created/Modified
- `internal/tui/config.go` - Deduplicated help items, merged d/r key binding, added inline [r] hint, refactored View() to wrap sections in border boxes
- `internal/tui/styles.go` - Added configSectionStyle (muted border) and configSectionActiveStyle (accent/blue border)

## Decisions Made
- Merged `"d"` and `"r"` into a single case statement for repo removal rather than duplicating logic
- Used `Padding(0, 1)` for config section borders to keep them compact (vs `Padding(1, 2)` used by the general `boxStyle`)
- Section content width calculated as `w - 4` to account for border characters (2 per side)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Config view UI is polished with deduplicated shortcuts and visual section focus
- No blockers

---
*Phase: quick-11*
*Completed: 2026-03-06*
