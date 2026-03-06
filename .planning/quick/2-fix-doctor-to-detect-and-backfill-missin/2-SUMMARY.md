---
phase: quick-2
plan: 01
subsystem: doctor
tags: [doctor, enrichment, backfill, config, lock-file]

# Dependency graph
requires:
  - phase: 06-doctor
    provides: "RunDiagnostics, BackfillLegacySkills, DoctorReport"
provides:
  - "EnrichCandidates detection in RunDiagnostics"
  - "EnrichExistingSkills function for populating version/installed from lock"
  - "BackfillLegacySkills now sets Version/Installed from lock entries"
affects: [doctor, config]

# Tech tracking
tech-stack:
  added: []
  patterns: ["enrichment detection loop between missing-from-FS and untracked-on-FS checks"]

key-files:
  created: []
  modified:
    - "internal/tui/doctor.go"
    - "internal/tui/doctor_test.go"

key-decisions:
  - "EnrichCandidates only flags skills present on FS with lock data available"
  - "EnrichExistingSkills saves config only when at least one skill was enriched"
  - "BackfillLegacySkills sets Version/Installed only for lock-file path, not knownCorrespondences"

patterns-established:
  - "Enrichment as separate concern from backfill: backfill adds missing entries, enrich updates existing entries"

requirements-completed: [QUICK-2]

# Metrics
duration: 2min
completed: 2026-03-06
---

# Quick Task 2: Fix Doctor to Detect and Backfill Missing Version/Installed Summary

**Doctor detects pre-v0.2.0 config entries missing version/installed and enriches them from lock file CommitHash/InstalledAt**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T12:45:59Z
- **Completed:** 2026-03-06T12:47:50Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments
- RunDiagnostics now detects config entries missing version/installed when lock file has data, reporting them as EnrichCandidates with info-severity "enrich" issues
- New EnrichExistingSkills function reads config, finds entries missing version/installed, populates from lock CommitHash/InstalledAt, and persists
- BackfillLegacySkills now sets Version and Installed fields when creating new entries from lock file data
- All 7 new tests pass, all 56 total tui tests pass, go vet clean

## Task Commits

Each task was committed atomically (TDD):

1. **Task 1 RED: Failing tests** - `811b17d` (test)
2. **Task 1 GREEN: Implementation** - `3818c6d` (feat)

## Files Created/Modified
- `internal/tui/doctor.go` - Added EnrichCandidates field to DoctorReport, enrichment detection loop in RunDiagnostics, EnrichExistingSkills function, Version/Installed in BackfillLegacySkills lock-file path
- `internal/tui/doctor_test.go` - 7 new tests covering enrichment detection, EnrichExistingSkills, and backfill version/installed

## Decisions Made
- EnrichCandidates only flagged when skill exists on FS AND lock has CommitHash or InstalledAt -- prevents false positives for skills without lock data
- EnrichExistingSkills saves config only when enrichment actually occurred -- avoids unnecessary writes
- knownCorrespondences entries in BackfillLegacySkills do NOT get version/installed (no lock data) -- they show as enrich candidates later if lock data becomes available

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Doctor can now detect and fix pre-v0.2.0 skills with missing metadata
- EnrichExistingSkills ready to be wired into doctor CLI command

## Self-Check: PASSED

- FOUND: internal/tui/doctor.go
- FOUND: internal/tui/doctor_test.go
- FOUND: 811b17d (test commit)
- FOUND: 3818c6d (feat commit)

---
*Quick Task: 2*
*Completed: 2026-03-06*
