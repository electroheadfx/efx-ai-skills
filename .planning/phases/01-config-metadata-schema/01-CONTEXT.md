# Phase 1: Config Metadata Schema - Context

**Gathered:** 2026-03-05
**Status:** Ready for planning

<domain>
## Phase Boundary

Extend config.json with the structural foundation to store full skill provenance. Three additions: a `skills` array with per-skill metadata, a `skills-path` field, and a `url` field on each repo entry. This phase only defines the schema and load/save logic — writing metadata during install/uninstall is Phase 2.

</domain>

<decisions>
## Implementation Decisions

### Config format freedom
- User has a backup of the existing config — no backward compatibility needed
- Config format can be changed freely for v0.2.0
- No migration logic, no fallback defaults for missing fields
- Existing config can be replaced with the new schema

### Skills array structure
- Fields per skill entry: owner, name, registry type (skills.sh / playbooks.com / github), URL
- Keep it strictly to what ORIG-02 requires — no pre-provisioning for Phase 5 (commit hash can be added when needed)
- Lock file continues to exist separately — config.json is the user-facing metadata, lock file is the install artifact

### Skills-path field
- Store the central skill storage directory path in config as `skills-path`
- Default value: `~/.agents/skills/` (current hardcoded path)
- This makes the path configurable rather than hardcoded in Store struct

### Repo URL field
- Each repo entry gets a `url` field with the GitHub URL
- Auto-derived from owner/repo: `https://github.com/{owner}/{repo}`
- No need for manual URL entry

### Config consolidation
- Claude's Discretion: Two config implementations exist (`internal/config/config.go` is dead code, `internal/tui/config.go` is actually used). Claude decides how to consolidate or extend during implementation.

### Claude's Discretion
- Whether to consolidate the two config packages or extend only the TUI version
- Exact JSON field naming conventions (kebab-case `skills-path` vs camelCase `skillsPath`)
- Internal struct design and method signatures
- How saveConfig/loadConfig handle the new fields

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/tui/config.go`: Active config implementation with `ConfigData`, `loadConfigFromFile()`, `saveConfig()` — this is what actually gets used
- `internal/config/config.go`: Dead-code config package with `Config` struct, `Load()`, `Save()`, `AddRepo()`, `RemoveRepo()` — has cleaner API but unused
- `internal/skill/store.go`: `LockFile` and `LockEntry` structs — lock file already stores source/sourceType/sourceURL per skill

### Established Patterns
- Config persisted as JSON with `json.MarshalIndent(data, "", "  ")`
- Config path: `~/.config/efx-skills/config.json`
- `os.Getenv("HOME")` used directly for path resolution
- Graceful fallback to defaults when config missing (`os.IsNotExist` check)
- `RepoSource` struct has Owner/Repo fields (no URL currently)

### Integration Points
- `internal/tui/config.go:ConfigData` — the struct that gets serialized to config.json
- `internal/tui/config.go:loadConfigFromFile()` — reads config from disk
- `internal/tui/config.go:saveConfig()` — writes config to disk
- `internal/skill/store.go:NewStore()` — hardcodes `~/.agents/skills` path, should read from config's skills-path
- `internal/tui/search.go` — install flow will need config access in Phase 2

</code_context>

<specifics>
## Specific Ideas

- User explicitly confirmed config format can change freely — they backed up existing config for v0.2.0

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-config-metadata-schema*
*Context gathered: 2026-03-05*
