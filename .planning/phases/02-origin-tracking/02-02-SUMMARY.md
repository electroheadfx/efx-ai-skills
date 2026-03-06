---
phase: 02-origin-tracking
plan: 02
subsystem: tui
tags: [go, tui, install, uninstall, metadata, origin-tracking]

# Dependency graph
requires:
  - phase: 02-origin-tracking
    plan: 01
    provides: "addSkillToConfig, removeSkillFromConfig, skillMetaFromAPISkill mutation functions"
provides:
  - "Install flow writes SkillMeta to config.json on skill install"
  - "Removal flow cleans SkillMeta from config.json when skill unlinked from all providers"
  - "ORIG-01 satisfied: search view displays owner/repo (Source field) next to skill name"
affects: [03-config-page, 04-browser-integration, 05-update-system, 06-doctor]

# Tech tracking
tech-stack:
  added: []
  patterns: [post-install metadata persistence, multi-provider unlink check before metadata removal]

key-files:
  created: []
  modified:
    - internal/tui/search.go
    - internal/tui/manage.go

key-decisions:
  - "Error from addSkillToConfig discarded with _ matching existing AddToLock pattern"
  - "Removal checks all configured providers via detectProviders+Lstat before removing metadata"
  - "ORIG-01 confirmed satisfied by existing search.go Source column display -- no changes needed"

patterns-established:
  - "Post-install hook: metadata persistence follows store operations in install flow"
  - "Multi-provider check: skill metadata only removed when no provider has it linked"

requirements-completed: [ORIG-01, ORIG-05, ORIG-06]

# Metrics
duration: 1min
completed: 2026-03-05
---

# Phase 2 Plan 2: Wire Skill Metadata into Install/Removal Flows Summary

**Skill metadata lifecycle wired into TUI: install writes SkillMeta to config, removal cleans it when unlinked from all providers, search displays origin**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-05T20:13:03Z
- **Completed:** 2026-03-05T20:14:23Z
- **Tasks:** 1
- **Files modified:** 2

## Accomplishments
- Install flow in search.go now calls addSkillToConfig(skillMetaFromAPISkill(s)) after store.AddToLock
- Removal flow in manage.go checks all configured providers before calling removeSkillFromConfig
- ORIG-01 confirmed satisfied: search view already renders skill.Source (owner/repo) as a column
- All 15 tests pass, project builds cleanly

## Task Commits

Each task was committed atomically:

1. **Task 1: Wire addSkillToConfig into install flow and removeSkillFromConfig into removal flow** - `5363ade` (feat)

## Files Created/Modified
- `internal/tui/search.go` - Added addSkillToConfig call after store.AddToLock in installStartMsg handler
- `internal/tui/manage.go` - Added multi-provider check and removeSkillFromConfig call in applySkillChanges

## Decisions Made
- Error from addSkillToConfig discarded with `_` to match the existing pattern used by store.AddToLock on the line above
- Removal logic uses detectProviders() + os.Lstat to check all configured providers before removing metadata, preventing premature removal when a skill is still linked to another provider
- ORIG-01 requires no code changes -- the search view already renders `skill.Source` (owner/repo format) in a dedicated column at lines 308-311

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All three ORIG requirements for Phase 2 are fully implemented (ORIG-01, ORIG-05, ORIG-06)
- Skill metadata lifecycle is complete: install writes, removal cleans
- Ready for Phase 3 (config page) which can display the skills array from config.json
- No blockers or concerns

## Self-Check: PASSED

- All 3 files found (search.go, manage.go, 02-02-SUMMARY.md)
- Commit verified (5363ade feat)
- addSkillToConfig present in search.go (line 126)
- skillMetaFromAPISkill present in search.go (line 125)
- removeSkillFromConfig present in manage.go (line 423)
- All 15 tests pass, project builds cleanly

---
*Phase: 02-origin-tracking*
*Completed: 2026-03-05*
