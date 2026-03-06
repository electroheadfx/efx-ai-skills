# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v0.2.0 — Origin & Metadata

**Shipped:** 2026-03-06
**Phases:** 6 | **Plans:** 12 | **Quick tasks:** 11

### What Was Built
- Config metadata schema: `skills` array, `skills-path`, repo URLs in config.json
- Origin tracking: install/uninstall metadata sync, search displays owner/repo
- Config page redesign: friendly registry names, two-column repos, "Providers search" label
- Browser integration: [o] opens skill/registry URLs from search, manage, config views
- Update system: [v] verify, [u] update, [g] global update via git commit hash comparison
- Doctor command: config-filesystem consistency verification and legacy metadata backfill

### What Worked
- TDD approach for all phases — tests caught edge cases before TUI wiring
- Phase dependency chain (1→2→3,4,5,6) enabled clean incremental delivery
- Config mutation functions decoupled from TUI state (saveConfigData on *ConfigData)
- Pure URL resolver functions (no disk I/O) made browser integration highly testable
- Quick task system for post-phase polish — 11 fixes shipped without heavyweight planning

### What Was Inefficient
- ROADMAP.md plan checkboxes stayed unchecked despite all plans being complete (stale progress tracking)
- Phase 6 Doctor plans listed as "TBD" in ROADMAP even after execution
- Nyquist validation was partial/missing for most phases — added overhead without full benefit
- 11 quick tasks indicate gap closure missed some UX issues during phase execution

### Patterns Established
- Confirmation dialog pattern for destructive TUI actions (confirmingRemove flag + key interception)
- Display name mapping via switch statement for known registries
- cmd.Start() (not cmd.Run()) for non-blocking browser opens from TUI
- Lock file as primary data source for backfill; hardcoded map as fallback only
- Responsive help bars using renderHelpBar helper with terminal width awareness

### Key Lessons
1. Separate toggle (non-destructive symlink) from remove (destructive full delete) early — the confusion generated multiple quick fixes
2. Width/sizing bugs in Bubble Tea appear when sub-models are created before WindowSizeMsg — initialize dimensions early
3. Unicode glyphs cause display width inconsistencies — prefer ASCII equivalents in terminal UIs
4. Post-phase UAT is essential — quick tasks 2-13 all came from real usage testing

### Cost Observations
- Model mix: 100% opus (quality profile)
- Sessions: ~5 sessions across 1 day
- Notable: All 6 phases + 11 quick tasks completed in a single day — high velocity from clear requirements and TDD foundation

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v0.2.0 | ~5 | 6 | First milestone with GSD workflow, TDD, quick task system |

### Cumulative Quality

| Milestone | Tests | Packages Tested | Zero-Dep Additions |
|-----------|-------|-----------------|-------------------|
| v0.2.0 | 60 | 6 | config_test, store_test, browser_test, doctor_test |

### Top Lessons (Verified Across Milestones)

1. TDD-first approach enables fast, confident TUI wiring
2. Post-phase UAT catches UX issues that unit tests miss
