---
phase: 01-config-metadata-schema
plan: 01
subsystem: config
tags: [json, schema, config, go-structs]

# Dependency graph
requires: []
provides:
  - "ConfigData with Skills []SkillMeta, SkillsPath, and RepoSource.URL fields"
  - "SkillMeta struct for per-skill provenance metadata"
  - "RepoSource.DeriveURL() helper for GitHub URL derivation"
  - "NewStore(skillsPath) parameterized skill store constructor"
  - "Config round-trip serialization tests"
  - "Store parameterization tests"
affects: [02-install-uninstall-metadata, 03-origin-display, 04-open-command, 05-update-system, 06-doctor-command]

# Tech tracking
tech-stack:
  added: []
  patterns: [TDD red-green for struct extensions, config default derivation]

key-files:
  created:
    - internal/tui/config_test.go
    - internal/skill/store_test.go
  modified:
    - internal/tui/config.go
    - internal/skill/store.go
    - internal/tui/search.go

key-decisions:
  - "SkillMeta fields: owner, name, registry, url -- minimal set for ORIG-02"
  - "skills-path uses kebab-case JSON tag matching existing config conventions"
  - "LockFile placed at parent of skills dir (filepath.Dir(skillsPath)/.skill-lock.json)"
  - "Empty skills serializes as [] not null for clean JSON output"

patterns-established:
  - "Config defaults: defaultSkillsPath() helper used by both load and save"
  - "DeriveURL pattern: auto-derive URL from owner/repo when field is empty"
  - "Store parameterization: callers pass config values, Store handles defaults"

requirements-completed: [ORIG-02, ORIG-03, ORIG-04]

# Metrics
duration: 3min
completed: 2026-03-05
---

# Phase 1 Plan 1: Config Metadata Schema Summary

**Extended config.json schema with SkillMeta array, skills-path field, and RepoSource URL; parameterized Store to consume skills-path from config**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-05T19:14:03Z
- **Completed:** 2026-03-05T19:17:27Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- ConfigData extended with Skills ([]SkillMeta), SkillsPath (string), and RepoSource.URL (string)
- JSON round-trip serialization verified with 6 tests covering all new fields
- NewStore parameterized to accept skills-path, with empty-string fallback to default
- search.go callsite updated to load config and pass skills-path to NewStore
- Full test suite passes (8 tests across 6 packages), project compiles cleanly

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend ConfigData with skills array, skills-path, and repo URL** - `5ab57c0` (test) + `52c2b0b` (feat)
2. **Task 2: Parameterize Store with skills-path from config** - `9929b16` (test) + `f4a99c4` (feat)

_TDD tasks have two commits each (RED: test, GREEN: feat)_

## Files Created/Modified
- `internal/tui/config.go` - Extended ConfigData, RepoSource, added SkillMeta, DeriveURL, defaults
- `internal/tui/config_test.go` - 6 tests for config schema serialization, round-trip, defaults
- `internal/skill/store.go` - NewStore now accepts skillsPath parameter
- `internal/skill/store_test.go` - 2 tests for Store parameterization with custom/default path
- `internal/tui/search.go` - Updated install flow to pass config's skills-path to NewStore

## Decisions Made
- Used kebab-case `skills-path` JSON tag to match existing config field naming style
- SkillMeta has four fields (owner, name, registry, url) -- minimal for ORIG-02, no pre-provisioning for Phase 5 commit hash
- LockFile path derived from parent of skillsPath, keeping lock file alongside skills directory
- Empty skills array serializes as `[]` not `null` for clean JSON

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Config schema is ready for Phase 2 (install/uninstall metadata writing)
- SkillMeta struct available for population during install/uninstall flows
- Store parameterization allows config-driven path resolution
- No blockers or concerns

## Self-Check: PASSED

- All 6 files found (5 source + 1 SUMMARY)
- All 4 commits verified (5ab57c0, 52c2b0b, 9929b16, f4a99c4)
- config_test.go: 179 lines (min 40)
- store_test.go: 37 lines (min 20)
- config.go contains SkillsPath (9 occurrences)
- store.go contains func NewStore( (1 match)
- key_links patterns verified

---
*Phase: 01-config-metadata-schema*
*Completed: 2026-03-05*
