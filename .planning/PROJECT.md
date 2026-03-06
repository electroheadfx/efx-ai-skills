# efx-skills

## What This Is

A Go CLI/TUI tool for discovering, installing, and managing AI coding agent skills across providers (Claude, Cursor, OpenCode, etc.). Users search registries (skills.sh, playbooks.com) and GitHub repos, install skills to a central store, and symlink them to providers. Skills carry full provenance metadata — users can open source URLs, check for updates, and verify installation integrity via the doctor command.

## Core Value

Every installed skill knows where it came from and can be opened, updated, or verified against its upstream source.

## Requirements

### Validated

- Search skills across skills.sh and playbooks.com registries — existing
- Preview skill SKILL.md content with markdown rendering — existing
- Install skills via npx or direct GitHub download — existing
- Symlink skills to AI provider directories — existing
- Manage skill-provider assignments with toggle UI — existing
- Group skills by prefix in manage view — existing
- Config page to enable/disable registries and manage custom repos — existing
- Non-interactive list command showing installed skills — existing
- Lock file tracking installed skill metadata — existing
- ✓ Show skill origin (owner/repo) in search results — v0.2.0
- ✓ [o] open skill URL in browser from search view — v0.2.0
- ✓ [o] open skill URL in browser from manage/provider view — v0.2.0
- ✓ [o] open registry/repo URL in browser from config view — v0.2.0
- ✓ Enrich config.json with `skills` array storing full metadata per installed skill — v0.2.0
- ✓ Add `skills-path` to config.json — v0.2.0
- ✓ Add `url` field to repos in config.json — v0.2.0
- ✓ Install writes skill metadata to config.json — v0.2.0
- ✓ Uninstall removes skill metadata from config.json — v0.2.0
- ✓ Redesign config page presentation (cleaner registry/repo display) — v0.2.0
- ✓ Rename "Providers" to "Providers search" in config — v0.2.0
- ✓ [u] update individual skill from upstream (git commit hash comparison) — v0.2.0
- ✓ [g] global update all skills — v0.2.0
- ✓ [v] verify if skill has upstream changes — v0.2.0
- ✓ Doctor command: verify skills in config match skills-path, detect/fix gaps — v0.2.0
- ✓ Doctor backfills metadata for pre-v0.2.0 installed skills — v0.2.0
- ✓ Playbooks.com skill-specific URLs for [o] (fallback to playbooks.com domain) — v0.2.0

### Active

(None yet — define in next milestone)

### Out of Scope

- Authentication/API keys for registries — not needed currently
- Automatic backfill on startup — only Doctor handles metadata gaps, keeps startup fast
- Self-update of efx-skills binary — separate distribution concern
- New registry integrations beyond skills.sh, playbooks.com, GitHub — future
- Windows browser open — focus on macOS/Linux first
- Mobile app — CLI-only tool

## Context

Shipped v0.2.0 with 6,573 LOC Go across 99 files.
Tech stack: Go 1.22, Bubble Tea (Elm Architecture), Cobra CLI, Lipgloss, Glamour.
Config at `~/.config/efx-skills/config.json`, skills at `~/.agents/skills/`, lock at `~/.agents/.skill-lock.json`.
60 tests across 6 packages (config_test.go, store_test.go, browser_test.go, doctor_test.go).
Two dead-code packages (`internal/config`, `internal/provider`) remain — TUI has its own implementations.

## Constraints

- **Tech stack**: Go, Bubble Tea, Cobra — must stay consistent with existing architecture
- **Config format**: JSON at `~/.config/efx-skills/config.json` — backward compatible
- **Storage**: `~/.agents/skills/` central store with symlinks — no changes to storage model
- **Browser**: Use `open` (macOS) / `xdg-open` (Linux) for browser features
- **No database**: All persistence remains JSON files on disk

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Git commit hash for update detection | Reliable, doesn't require downloading full content | ✓ Good — works for all GitHub-hosted skills |
| Doctor-only backfill (no auto on startup) | Keeps startup fast, user controls when to fix | ✓ Good — clean separation of concerns |
| Enrich config.json (not lock file) for metadata | Config is user-facing, lock file is install artifact | ✓ Good — config.json is the source of truth for provenance |
| Playbooks URL: try skill page, fallback to domain | Best effort — depends on API supporting it | ✓ Good — graceful degradation |
| TDD approach for all phases | Ensures correctness before wiring into TUI | ✓ Good — caught edge cases early |
| Separate toggle [t] from remove [r] | Toggle is non-destructive (symlink), remove is destructive (full delete) | ✓ Good — prevents accidental data loss |
| openInBrowser uses cmd.Start() not cmd.Run() | Avoids blocking the TUI event loop | ✓ Good — responsive UI |
| Lock file as primary backfill data source | Lock file has most reliable installed skill data | ✓ Good — knownCorrespondences is fallback only |

---
*Last updated: 2026-03-06 after v0.2.0 milestone*
