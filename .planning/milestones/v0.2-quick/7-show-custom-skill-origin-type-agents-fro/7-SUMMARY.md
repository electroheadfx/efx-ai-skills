---
phase: quick-7
plan: 1
subsystem: ui
tags: [tui, manage-view, origin-label, custom-skills]

requires:
  - phase: quick-6
    provides: "custom group naming and registry label display"
provides:
  - "Origin field on SkillEntry for agents vs local provider distinction"
  - "Origin label display in manage view for custom skills"
affects: []

tech-stack:
  added: []
  patterns: ["Origin field detection via centralNames membership in loadSkillsForProvider"]

key-files:
  created: []
  modified:
    - internal/tui/manage.go
    - internal/tui/config_test.go

key-decisions:
  - "Origin detection uses centralNames map (already computed) to distinguish agents vs local provider"
  - "Registry skills leave Origin empty -- registryDisplayName handles their display"
  - "Origin label appended in same parenthetical style as registry display name"

patterns-established:
  - "Origin field: agents = present in ~/.agents/skills/, local provider = only in provider path"

requirements-completed: [QUICK-7]

duration: 1min
completed: 2026-03-06
---

# Quick Task 7: Show Custom Skill Origin Type Summary

**Origin labels (agents/local provider) on custom skills in manage view via SkillEntry.Origin field and centralNames detection**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-06T16:20:04Z
- **Completed:** 2026-03-06T16:21:14Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments
- Added `Origin` field to `SkillEntry` struct for custom skill provenance tracking
- Origin detection in `loadSkillsForProvider` using existing `centralNames` map
- Manage view displays "(agents)" or "(local provider)" labels on custom skills
- Registry skills continue to show their registry display name unchanged
- 5 new test cases validating all origin label display combinations

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: Add failing test for origin label** - `8b16589` (test)
2. **Task 1 GREEN: Implement Origin field and display** - `6c4622d` (feat)

_TDD task: test commit followed by implementation commit_

## Files Created/Modified
- `internal/tui/manage.go` - Added Origin field to SkillEntry, detection in loadSkillsForProvider, display in View()
- `internal/tui/config_test.go` - Added TestSkillEntryOriginLabel with 4 sub-tests

## Decisions Made
- Origin detection reuses the existing `centralNames` map (no additional I/O needed)
- Registry skills keep Origin as empty string -- registryDisplayName handles their label
- Origin label uses same parenthetical format as registry display names for visual consistency

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Origin labels are fully functional for custom skills
- No blockers

---
*Phase: quick-7*
*Completed: 2026-03-06*
