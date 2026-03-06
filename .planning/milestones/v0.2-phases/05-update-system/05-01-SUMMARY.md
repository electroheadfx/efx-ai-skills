---
phase: 05-update-system
plan: 01
subsystem: api
tags: [github-api, commit-hash, update-detection, httptest, tdd]

# Dependency graph
requires:
  - phase: 01-config-metadata-schema
    provides: LockEntry struct and lock file read/write
provides:
  - FetchLatestCommitHash function for GitHub API commit SHA fetching
  - CheckForUpdate comparing stored vs upstream commit hashes
  - UpdateSkill re-downloading and updating lock entries
  - UpdateAllSkills batch update with graceful error handling
  - CommitHash field in LockEntry for version tracking
  - search.go install flow stores commit hash on new installs
affects: [05-02, tui-wiring, verify-command, update-command, global-update]

# Tech tracking
tech-stack:
  added: [net/http/httptest]
  patterns: [package-level var for testable base URL, httptest mock server]

key-files:
  created: []
  modified:
    - internal/skill/store.go
    - internal/skill/store_test.go
    - internal/tui/search.go

key-decisions:
  - "gitHubAPIBaseURL as package-level var for httptest overrides"
  - "AddToLock changed from 2-arg to 3-arg (skillName, source, commitHash)"
  - "Empty commitHash on legacy installs treated as always-updateable"
  - "UpdateAllSkills collects individual errors gracefully without stopping"

patterns-established:
  - "Package-level var for external URLs enables httptest mocking without interfaces"
  - "Graceful error collection in batch operations (continue on individual failure)"

requirements-completed: [UPDT-01, UPDT-02, UPDT-03, UPDT-04]

# Metrics
duration: 3min
completed: 2026-03-05
---

# Phase 05 Plan 01: Update System Core Summary

**Commit hash fetching, storage, comparison, and batch update functions with httptest-mocked GitHub API**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-05T21:27:50Z
- **Completed:** 2026-03-05T21:31:01Z
- **Tasks:** 1 feature (5 phases: CommitHash field, FetchLatestCommitHash, CheckForUpdate, UpdateSkill, UpdateAllSkills)
- **Files modified:** 3

## Accomplishments
- LockEntry now tracks upstream commit hash for version detection
- FetchLatestCommitHash calls GitHub API and returns HEAD commit SHA
- CheckForUpdate compares stored vs upstream hash with legacy install support
- UpdateSkill re-downloads and updates lock entry with new commit hash
- UpdateAllSkills iterates all locked skills with graceful per-skill error handling
- search.go install flow now fetches and stores commit hash on new installs
- 12 new tests covering all functions, using httptest mock server

## Task Commits

Each task was committed atomically:

1. **RED: Failing tests for update system** - `429a6ac` (test)
2. **GREEN: Implement update system core** - `d9419bd` (feat)

_TDD: RED then GREEN. No refactor needed._

## Files Created/Modified
- `internal/skill/store.go` - Added CommitHash field, FetchLatestCommitHash, CheckForUpdate, UpdateSkill, UpdateAllSkills, updated AddToLock signature
- `internal/skill/store_test.go` - 12 new tests covering all update system functions with httptest mocks
- `internal/tui/search.go` - Updated AddToLock call site to pass commit hash on install

## Decisions Made
- Used package-level `gitHubAPIBaseURL` var for testable GitHub API base URL (avoids interface complexity)
- Changed AddToLock from 2-arg to 3-arg signature (breaking change, all callers updated in same commit)
- Empty stored commitHash treated as always-updateable (backward compatible with legacy installs)
- UpdateAllSkills collects individual errors into a combined error rather than stopping on first failure

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Update system core is complete and tested
- Ready for TUI wiring of [v] verify, [u] update, and [g] global update commands
- All functions are pure store methods, easily callable from TUI handlers

## Self-Check: PASSED

- All 3 modified files exist on disk
- Both task commits (429a6ac, d9419bd) verified in git log
- All 50 tests pass across 6 packages
- Project builds cleanly

---
*Phase: 05-update-system*
*Completed: 2026-03-05*
