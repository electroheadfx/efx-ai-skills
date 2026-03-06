# Milestones

## v0.2 Origin & Metadata (Shipped: 2026-03-06)

**Phases completed:** 6 phases, 12 plans | 120 commits | 99 files changed | 6,573 LOC Go
**Timeline:** 2026-03-06 | **Git range:** main (v0.1.4) → v0.2.0-origin
**Quick tasks:** 11 post-phase fixes (UX polish, doctor fixes, visual enhancements)

**Delivered:** Every installed skill knows where it came from — origin tracking, browser integration, update system, and doctor command.

**Key accomplishments:**
- Config metadata schema: `skills` array, `skills-path`, repo URLs in config.json
- Origin tracking: install/uninstall metadata sync, search displays origin (owner/repo)
- Config page redesign: friendly registry names, two-column repos, "Providers search" label
- Browser integration: [o] opens skill/registry URLs from search, manage, and config views
- Update system: [v] verify, [u] update, [g] global update via git commit hash comparison
- Doctor command: config-filesystem consistency verification and legacy metadata backfill

---

