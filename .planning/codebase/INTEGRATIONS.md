# External Integrations

**Analysis Date:** 2026-03-05

## APIs & External Services

**Skill Registries:**

- **skills.sh** - AI agent skill registry (search and trending)
  - Base URL: `https://skills.sh`
  - Endpoints:
    - `GET /api/search?q={query}&limit={n}` - Search skills
    - `GET /api/skills?limit={n}` - Get trending/listed skills
  - SDK/Client: Custom HTTP client in `internal/api/client.go`
  - Auth: None required (public API)
  - Response format: `{"skills": [{"id", "name", "installs", "source"}]}`
  - Implementation: `internal/api/skillssh.go`

- **playbooks.com** - AI agent playbooks/skills registry (search and trending)
  - Base URL: `https://playbooks.com`
  - Endpoints:
    - `GET /api/skills?search={query}&limit={n}` - Search skills
    - `GET /api/skills?limit={n}` - Get trending/listed skills
  - SDK/Client: Custom HTTP client in `internal/api/client.go`
  - Auth: None required (public API)
  - Response format: `{"success": true, "data": [{"id", "name", "description", "repoOwner", "repoName", "path", "skillSlug", "stars", "isOfficial"}]}`
  - Implementation: `internal/api/playbooks.go`

**GitHub (raw content):**
- **GitHub Raw Content** - Fetches SKILL.md files from GitHub repositories
  - Base URL: `https://raw.githubusercontent.com`
  - Pattern: `GET /{owner}/{repo}/{branch}/{path}/SKILL.md`
  - Tries multiple path patterns: `{skillPath}/SKILL.md`, `skills/{skillPath}/SKILL.md`, `SKILL.md`, `README.md`
  - Tries both `main` and `master` branches as fallback
  - Auth: None (public repos only)
  - Implementation: `internal/api/client.go` (`FetchSkillContent`), `internal/tui/preview.go` (`fetchSkillContent`), `internal/skill/store.go` (`installDirect`)
  - HTTP timeout: 10 seconds

**npx skills CLI (optional):**
- **skills npm package** - Alternative skill installation method
  - Invoked via: `npx skills add {source} -g -y [--skill {name}]`
  - Used when `npx` is available on `$PATH`
  - Falls back to direct GitHub download if `npx` not found
  - Implementation: `internal/skill/store.go` (`installViaSkills`)

## Data Storage

**Databases:**
- None - No database used

**File Storage (local filesystem):**
- Central skill storage: `~/.agents/skills/{skill-name}/SKILL.md`
- Lock file: `~/.agents/.skill-lock.json` (JSON, tracks source, install date, update date, hashes)
- Config file: `~/.config/efx-skills/config.json` (JSON, stores registries, repos, enabled providers)
- Provider skill directories linked via symlinks from provider paths to central storage

**Caching:**
- None - All API calls are made fresh on each search/preview operation

## Authentication & Identity

**Auth Provider:**
- None - No authentication system; all external APIs used are public
- No user accounts or identity management

## AI Coding Agent Providers

The application manages skills for these AI coding agent providers by symlinking skills into their configuration directories:

| Provider  | Skills Path                          | Default Enabled |
|-----------|--------------------------------------|-----------------|
| claude    | `~/.claude/skills/`                  | Yes             |
| cursor    | `~/.cursor/skills/`                  | Yes             |
| qoder     | `~/.qoder/skills/`                   | Yes             |
| windsurf  | `~/.windsurf/skills/`                | No              |
| copilot   | `~/.copilot/skills/`                 | No              |
| opencode  | `~/.config/opencode/skills/`         | No              |

- Provider detection: `internal/provider/provider.go` (`KnownProviders`, `DetectAll`)
- Provider status view: `internal/tui/status.go` (`detectProviders`)
- Provider config (enabled/disabled): `internal/tui/config.go` (persisted in `config.json`)
- Skill linking: `internal/skill/store.go` (`LinkToProvider`, `UnlinkFromProvider`) - creates relative symlinks

## Default GitHub Repo Sources

Pre-configured custom repo sources (defined in `internal/config/config.go` and `internal/tui/config.go`):
- `yoanbernabeu/grepai-skills`
- `better-auth/skills`
- `awni/mlx-skills`

Users can add/remove custom repos via the config TUI view.

## Monitoring & Observability

**Error Tracking:**
- None - Errors displayed inline in TUI via `errorStyle`

**Logs:**
- None - No structured logging; errors handled in TUI message loop
- Some operations silently swallow errors (e.g., `_ = store.AddToLock(...)` in `internal/tui/search.go`)

## CI/CD & Deployment

**Hosting:**
- Local binary distribution (no hosted service)
- Release artifacts stored in `bin/release/` as zip files

**CI Pipeline:**
- None detected - No GitHub Actions, CircleCI, or other CI configuration files present

## Environment Configuration

**Required env vars:**
- `HOME` - Used extensively for resolving all file paths (config, skills, providers)

**Optional env vars:**
- None

**Secrets:**
- None required - All external APIs are public, no authentication needed

## HTTP Client Configuration

**Shared HTTP client** (`internal/api/client.go`):
- Timeout: 10 seconds
- No retry logic
- No rate limiting
- No caching headers
- Standard `net/http` client

**Ad-hoc HTTP clients** (in `internal/tui/preview.go`, `internal/skill/store.go`):
- Also 10-second timeout
- Created per-request (not pooled)

## Webhooks & Callbacks

**Incoming:**
- None - This is a CLI/TUI tool, not a server

**Outgoing:**
- None

---

*Integration audit: 2026-03-05*
