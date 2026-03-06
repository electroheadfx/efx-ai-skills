---
phase: quick-5
plan: 01
subsystem: tui
tags: [manage-view, custom-skills, registry-display]

requires:
  - phase: 03-config-page-redesign
    provides: registryDisplayName switch-based mapping
provides:
  - Custom label for skills without registry metadata in manage view
  - registryDisplayName handles empty string input
affects: []

tech-stack:
  added: []
  patterns: [registryDisplayName centralizes all registry-to-label mapping including custom]

key-files:
  created: []
  modified:
    - internal/tui/config.go
    - internal/tui/manage.go
    - internal/tui/config_test.go

key-decisions:
  - "registryDisplayName empty string returns Custom -- single source of truth for unknown-origin skills"
  - "Always show registry label in manage view (no conditional skip for empty registry)"

patterns-established:
  - "Custom skills: any skill directory in ~/.agents/skills/ without SkillMeta in config.json shows (Custom)"

requirements-completed: [custom-skills-manage]

duration: 2min
completed: 2026-03-06
---

# Quick Task 5: Manage Custom Internal Skills Summary

**Custom skills (no registry metadata) labeled "(Custom)" in manage view with registryDisplayName empty-string handling**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T15:43:36Z
- **Completed:** 2026-03-06T15:45:10Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Skills without registry metadata (manually placed in ~/.agents/skills/) now display "(Custom)" in the manage view
- registryDisplayName handles empty string input, returning "Custom" as a first-class case
- Test coverage expanded from 3 to 5 cases for registryDisplayName
- Remove [r] and toggle [t] confirmed working unchanged for custom skills

## Task Commits

Each task was committed atomically:

1. **Task 1: Label skills without registry metadata as Custom** - `c9fe94d` (feat)
2. **Task 2: Add test for registryDisplayName covering Custom case** - `a9fa0eb` (test)

## Files Created/Modified
- `internal/tui/config.go` - Added empty string case to registryDisplayName returning "Custom"
- `internal/tui/manage.go` - Explicit empty Registry for untracked skills; always show registry label in view
- `internal/tui/config_test.go` - Expanded TestRegistryDisplayName with empty string and github passthrough cases

## Decisions Made
- registryDisplayName empty string returns "Custom" -- single source of truth for unknown-origin skills
- Always show registry label in manage view (removed conditional `skill.Registry != ""` check)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Custom skills are now visually distinguishable in the manage view
- No blockers or concerns

---
*Phase: quick-5*
*Completed: 2026-03-06*
