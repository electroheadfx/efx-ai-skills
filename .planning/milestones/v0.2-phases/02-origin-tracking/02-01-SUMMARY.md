---
phase: 02-origin-tracking
plan: 01
subsystem: config
tags: [go, config, tdd, metadata, install, uninstall]

# Dependency graph
requires:
  - phase: 01-config-metadata-schema
    provides: "ConfigData with Skills []SkillMeta, SkillsPath, loadConfigFromFile, defaultSkillsPath"
provides:
  - "saveConfigData: standalone config writer independent of configModel"
  - "addSkillToConfig: idempotent skill metadata append"
  - "removeSkillFromConfig: skill metadata removal by name"
  - "skillMetaFromAPISkill: api.Skill to SkillMeta mapping"
affects: [02-02-PLAN (wire into install/uninstall flows), 03-config-page, 04-browser-integration, 05-update-system, 06-doctor]

# Tech tracking
tech-stack:
  added: []
  patterns: [load-modify-save for standalone config mutation, idempotent append with owner+name dedup]

key-files:
  created: []
  modified:
    - internal/tui/config.go
    - internal/tui/config_test.go

key-decisions:
  - "saveConfigData operates on ConfigData pointer, not configModel -- decoupled from TUI state"
  - "addSkillToConfig deduplicates by Owner+Name pair (both must match)"
  - "removeSkillFromConfig filters by Name only (sufficient for unique skill identification)"
  - "skillMetaFromAPISkill maps Source directly to Owner (preserves owner/repo format)"

patterns-established:
  - "Load-modify-save: standalone config mutations load from disk, modify in memory, save back"
  - "Idempotent writes: addSkillToConfig checks for existing entry before append"
  - "No-op on missing: removeSkillFromConfig returns nil when config or skill absent"

requirements-completed: [ORIG-05, ORIG-06]

# Metrics
duration: 2min
completed: 2026-03-05
---

# Phase 2 Plan 1: Config Mutation Functions Summary

**TDD config mutation functions (saveConfigData, addSkillToConfig, removeSkillFromConfig, skillMetaFromAPISkill) enabling install/uninstall metadata lifecycle**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-05T20:08:34Z
- **Completed:** 2026-03-05T20:10:33Z
- **Tasks:** 1 (TDD feature with RED + GREEN commits)
- **Files modified:** 2

## Accomplishments
- Four exported functions in config.go for standalone config mutation without TUI coupling
- Full test coverage with 7 new tests covering happy paths, edge cases, and idempotency
- saveConfigData creates config directory, ensures null-safety for Skills array, defaults SkillsPath
- addSkillToConfig is idempotent with Owner+Name deduplication
- removeSkillFromConfig is a safe no-op when config or skill is missing
- All 15 tests pass across 6 packages, project builds cleanly

## Task Commits

Each task was committed atomically:

1. **Feature: Config mutation functions (RED)** - `ff89d9d` (test: 7 failing tests)
2. **Feature: Config mutation functions (GREEN)** - `86adf06` (feat: implementation passing all tests)

_TDD task with two commits (RED: test, GREEN: feat)_

## Files Created/Modified
- `internal/tui/config.go` - Added saveConfigData, addSkillToConfig, removeSkillFromConfig, skillMetaFromAPISkill; added api import
- `internal/tui/config_test.go` - Added 7 new tests with temp directory isolation via HOME override; added api import

## Decisions Made
- saveConfigData takes *ConfigData (not configModel) so search/manage views can write metadata without TUI state
- Deduplication in addSkillToConfig uses Owner+Name pair (both must match to be considered duplicate)
- removeSkillFromConfig filters by Name only, sufficient since skill names are unique within an installation
- skillMetaFromAPISkill maps api.Skill.Source directly to SkillMeta.Owner, preserving the "owner/repo" format

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All four mutation functions ready for Plan 02-02 to wire into install/uninstall flows
- saveConfigData can be called from any view (search, manage) without configModel dependency
- skillMetaFromAPISkill provides the bridge between API search results and config persistence
- No blockers or concerns

## Self-Check: PASSED

- All 3 files found (config.go, config_test.go, 02-01-SUMMARY.md)
- Both commits verified (ff89d9d RED, 86adf06 GREEN)
- All 4 functions present in config.go
- All 7 test functions present in config_test.go
- key_links patterns verified (saveConfigData, skillMetaFromAPISkill)

---
*Phase: 02-origin-tracking*
*Completed: 2026-03-05*
