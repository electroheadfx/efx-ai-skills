---
phase: quick-6
plan: 1
subsystem: ui
tags: [tui, manage-view, bubbletea]

requires:
  - phase: quick-5
    provides: registryDisplayName function and manage view registry labels
provides:
  - "Renamed fallback group from _other to custom"
  - "Conditional registry label (hidden for custom skills)"
affects: [manage-view, tui]

tech-stack:
  added: []
  patterns:
    - "Conditional registry label display based on Registry field emptiness"

key-files:
  created: []
  modified:
    - internal/tui/manage.go

key-decisions:
  - "Hide registry label for skills with empty Registry field rather than checking group name"

patterns-established:
  - "Registry label display is conditional: only shown when skill.Registry is non-empty"

requirements-completed: []

duration: 1min
completed: 2026-03-06
---

# Quick Task 6: Rename _other Group to Custom and Hide Redundant Registry Label Summary

**Fallback group renamed from _other to custom and registry label hidden for custom skills with empty Registry field**

## Performance

- **Duration:** 40s
- **Started:** 2026-03-06T16:02:27Z
- **Completed:** 2026-03-06T16:03:07Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Renamed the fallback skill group from "_other" to "custom" in extractGroup
- Made registry label conditional -- only displayed when skill.Registry is non-empty
- Skills from known registries (Vercel, Playbooks) retain their labels as before

## Task Commits

Each task was committed atomically:

1. **Task 1: Rename _other group to custom and conditionally hide registry label** - `1222202` (feat)

## Files Created/Modified
- `internal/tui/manage.go` - Changed extractGroup fallback return and wrapped registry label in conditional

## Decisions Made
- Used `skill.Registry != ""` check instead of comparing group name to "custom" -- more robust since it directly checks the data source rather than a derived value

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Manage view now shows cleaner custom skill display
- No blockers

## Self-Check: PASSED

- FOUND: internal/tui/manage.go
- FOUND: commit 1222202
- FOUND: 6-SUMMARY.md

---
*Phase: quick-6*
*Completed: 2026-03-06*
