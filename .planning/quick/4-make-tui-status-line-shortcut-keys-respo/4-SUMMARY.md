---
phase: quick-4
plan: 1
subsystem: ui
tags: [tui, lipgloss, responsive, help-bar]

# Dependency graph
requires:
  - phase: quick-3
    provides: helpStyle and existing TUI view structure
provides:
  - renderHelpBar helper for responsive shortcut key display
  - All 5 views updated to use responsive help bars
affects: [any future TUI views that need help bars]

# Tech tracking
tech-stack:
  added: []
  patterns: [greedy line-packing for responsive text wrapping]

key-files:
  created: []
  modified:
    - internal/tui/styles.go
    - internal/tui/status.go
    - internal/tui/manage.go
    - internal/tui/search.go
    - internal/tui/config.go
    - internal/tui/preview.go

key-decisions:
  - "Replace unicode glyphs with ASCII equivalents for consistent display width"
  - "Use helpStyle.MarginTop(1) instead of manual newline before help bars"
  - "Use m.viewport.Width for preview model since it has no width field"

patterns-established:
  - "renderHelpBar: all views use shared helper for help bar rendering"

requirements-completed: [QUICK-4]

# Metrics
duration: 2min
completed: 2026-03-06
---

# Quick Task 4: Make TUI Status Line Shortcut Keys Responsive

**Shared renderHelpBar helper with greedy line-packing for responsive help bar wrapping across all 5 TUI views**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T15:07:02Z
- **Completed:** 2026-03-06T15:08:38Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Created renderHelpBar function with greedy line-packing algorithm that wraps shortcut items across multiple lines at narrow terminal widths
- Replaced hardcoded help strings in all 5 views (status, manage, search, config, preview) with renderHelpBar calls
- Normalized unicode arrow/return glyphs to ASCII equivalents for consistent display width measurement

## Task Commits

Each task was committed atomically:

1. **Task 1: Create renderHelpBar helper in styles.go** - `9647637` (feat)
2. **Task 2: Replace hardcoded help strings with renderHelpBar calls in all views** - `79888bb` (feat)

## Files Created/Modified
- `internal/tui/styles.go` - Added renderHelpBar function and strings import
- `internal/tui/status.go` - Replaced helpStyle.Render with renderHelpBar, changed unicode arrow to [m/enter]
- `internal/tui/manage.go` - 13 shortcuts now wrap gracefully via renderHelpBar
- `internal/tui/search.go` - Replaced unicode glyphs with ASCII, uses renderHelpBar
- `internal/tui/config.go` - Conditional help items use renderHelpBar
- `internal/tui/preview.go` - Scroll percentage and shortcuts use renderHelpBar via m.viewport.Width

## Decisions Made
- Replaced unicode glyphs (arrows, return symbol) with ASCII text equivalents ([enter], [up/down], [<-/->]) for consistent character width measurement in the line-packing algorithm
- Removed manual "\n" before help bars since helpStyle already has MarginTop(1)
- Used m.viewport.Width as the width parameter for preview model since it lacks a direct width field

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All help bars across all views now use the shared renderHelpBar function
- Any future views can use renderHelpBar for consistent responsive help display

---
*Quick Task: 4*
*Completed: 2026-03-06*
