---
phase: 06-doctor
plan: 01
subsystem: diagnostics
tags: [doctor, config-consistency, backfill, lock-file, tdd]

requires:
  - phase: 01-config-metadata-schema
    provides: SkillMeta, ConfigData, loadConfigFromFile, addSkillToConfig
  - phase: 02-origin-tracking
    provides: Store, LockFile, LockEntry, ReadLockFile
provides:
  - DoctorIssue and DoctorReport types for structured diagnostics
  - RunDiagnostics function for config-filesystem consistency checking
  - BackfillLegacySkills function for adding metadata to pre-v0.2.0 skills
  - FormatReport function for human-readable diagnostic output
  - knownCorrespondences map for fallback skill metadata lookup
affects: [06-doctor]

tech-stack:
  added: []
  patterns: [lock-file-primary-backfill, known-correspondences-fallback, severity-grouped-output]

key-files:
  created:
    - internal/tui/doctor.go
    - internal/tui/doctor_test.go
  modified: []

key-decisions:
  - "Lock file is the primary data source for backfill; knownCorrespondences is fallback only"
  - "DoctorIssue uses string enums for Severity and Category rather than custom types for simplicity"
  - "BackfillLegacySkills continues on individual failures, returning only successful backfills"
  - "FormatReport uses ! for error, ? for warning, i for info severity icons"

patterns-established:
  - "Lock-file-first backfill: check Store.ReadLockFile before falling back to hardcoded correspondences"
  - "Structured diagnostic reporting: DoctorIssue with severity, category, message, fix"

requirements-completed: [DCTR-01, DCTR-02, DCTR-03, DCTR-04]

duration: 2min
completed: 2026-03-05
---

# Phase 06 Plan 01: Doctor Diagnostic Engine Summary

**Config-filesystem consistency checker with lock-file-first backfill and severity-grouped reporting**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-05T22:10:59Z
- **Completed:** 2026-03-05T22:13:35Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments
- RunDiagnostics detects skills in config missing from filesystem (DCTR-01) and skills on filesystem not tracked in config (DCTR-02)
- BackfillLegacySkills adds metadata using lock file as primary source with knownCorrespondences fallback (DCTR-03)
- FormatReport produces severity-grouped output with icons and summary line (DCTR-04)
- 10 new tests covering all diagnostic paths: missing, untracked, backfill, healthy, format

## Task Commits

Each task was committed atomically:

1. **RED: Failing tests for doctor engine** - `ef8a7d0` (test)
2. **GREEN: Implement doctor diagnostic engine** - `900494e` (feat)

_TDD task: test commit followed by implementation commit_

## Files Created/Modified
- `internal/tui/doctor.go` - DoctorIssue, DoctorReport types; RunDiagnostics, BackfillLegacySkills, FormatReport functions; knownCorrespondences map
- `internal/tui/doctor_test.go` - 10 tests covering missing-from-FS, untracked-on-FS, backfill candidates (lock file + known correspondences), healthy state, backfill execution, format output

## Decisions Made
- Lock file is the primary data source for backfill; knownCorrespondences is the fallback for pre-v0.2.0 skills without lock entries
- DoctorIssue uses string enums ("error", "warning", "info") for Severity rather than custom iota types -- keeps things simple and readable
- BackfillLegacySkills continues on individual addSkillToConfig failures, returning only the list of successfully backfilled names
- FormatReport uses ! for error, ? for warning, i for info as severity prefix icons

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Doctor diagnostic engine ready for CLI wiring in plan 06-02
- RunDiagnostics, BackfillLegacySkills, and FormatReport are pure functions ready to be called from a CLI command handler

## Self-Check: PASSED

- FOUND: internal/tui/doctor.go
- FOUND: internal/tui/doctor_test.go
- FOUND: 06-01-SUMMARY.md
- FOUND: ef8a7d0 (test commit)
- FOUND: 900494e (feat commit)

---
*Phase: 06-doctor*
*Completed: 2026-03-05*
