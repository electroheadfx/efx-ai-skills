# Phase 2: Origin Tracking - Research

**Researched:** 2026-03-05
**Domain:** Go TUI (Bubble Tea) -- config persistence, search result display, install/uninstall metadata lifecycle
**Confidence:** HIGH

## Summary

Phase 2 connects three capabilities: (1) displaying origin info (`owner/repo`) alongside skill names in search results, (2) writing `SkillMeta` entries to config.json during install, and (3) removing those entries during uninstall/removal. The structural foundation (SkillMeta struct, ConfigData.Skills array, config load/save) was completed in Phase 1. This phase wires the data flow: search display reads origin from `api.Skill.Source`, install flow constructs and appends `SkillMeta`, and uninstall/removal flow deletes the matching entry.

The codebase already has almost everything needed. The `api.Skill` struct contains `Source` (owner/repo format like `"vercel-labs/skills"`) and `Registry` (like `"skills.sh"` or `"playbooks.com"`). The search view already displays `skill.Source` in the results table. The install flow in `search.go` already has access to the full `api.Skill` object and loads config. The main gaps are: (a) the search view display needs to show origin more prominently or in a different format, (b) the install flow needs to construct and persist a `SkillMeta` entry, (c) there is NO uninstall-from-central-storage capability at all -- only provider unlinking exists in `manage.go`, and (d) there is no `RemoveFromLock` function on Store.

**Primary recommendation:** Three focused changes: adjust search View to display origin (owner/repo) clearly; extend the install command handler in `search.go:installStartMsg` to append SkillMeta to config.json after successful install; add config metadata removal to the uninstall/removal flow (which itself may need to be created or extended from the manage view's `applySkillChanges`).

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| ORIG-01 | Search results display origin (owner/repo) next to each skill name | `api.Skill.Source` already contains owner/repo. The search view `View()` function at `search.go:308-311` already renders `skill.Source` in a column. May need reformatting or labeling to meet the spec's visual expectation. |
| ORIG-05 | Installing a skill writes its metadata to the `skills` array in config.json | The install handler at `search.go:106-136` already loads config and has access to the full `api.Skill` (with Source, Name, Registry fields). Needs: construct `SkillMeta` from `api.Skill`, append to `ConfigData.Skills`, and save config. |
| ORIG-06 | Uninstalling/removing a skill removes its metadata from the `skills` array in config.json | Currently NO uninstall capability exists for central storage. `manage.go:applySkillChanges` only unlinks from providers. The removal flow needs to be extended to also remove the matching `SkillMeta` entry and persist the updated config. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.22 | Language | Project's existing language |
| bubbletea | v1.2.4 | TUI framework (Elm Architecture) | Already used throughout |
| lipgloss | v1.0.0 | Terminal styling | Already used for all rendering |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | Config serialization | Already used for config load/save |
| path/filepath | stdlib | Path manipulation | Already used for skills-path, lock file paths |

### Alternatives Considered
None -- this phase uses only existing dependencies. No new libraries needed.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  tui/
    config.go         # ConfigData, SkillMeta, load/save (Phase 1 -- modify for add/remove helpers)
    search.go         # Search view + install flow (modify: display origin, write SkillMeta on install)
    manage.go         # Manage view + apply changes (modify: remove SkillMeta on unlink/remove)
    config_test.go    # Config tests (extend: add/remove SkillMeta tests)
  api/
    client.go         # Skill struct with Source, Registry fields (read-only for this phase)
  skill/
    store.go          # Store with Install, AddToLock (read-only for this phase)
```

### Pattern 1: Config Mutation via Load-Modify-Save Cycle
**What:** Read config from disk, modify the Skills array (append or filter), write back to disk. This mirrors the existing `saveConfig()` pattern.
**When to use:** Every time a skill is installed or removed.
**Example:**
```go
// Add SkillMeta to config on install
func addSkillToConfig(meta SkillMeta) error {
    cfg := loadConfigFromFile()
    if cfg == nil {
        cfg = &ConfigData{Skills: []SkillMeta{}}
    }
    // Check for duplicates before appending
    for _, s := range cfg.Skills {
        if s.Name == meta.Name && s.Owner == meta.Owner {
            return nil // already tracked
        }
    }
    cfg.Skills = append(cfg.Skills, meta)
    return saveConfigData(cfg)
}

// Remove SkillMeta from config on uninstall
func removeSkillFromConfig(skillName string) error {
    cfg := loadConfigFromFile()
    if cfg == nil {
        return nil
    }
    filtered := cfg.Skills[:0]
    for _, s := range cfg.Skills {
        if s.Name != skillName {
            filtered = append(filtered, s)
        }
    }
    cfg.Skills = filtered
    return saveConfigData(cfg)
}
```

### Pattern 2: SkillMeta Construction from api.Skill
**What:** Derive SkillMeta fields from the api.Skill struct fields available at install time.
**When to use:** In the install command handler, after successful install.
**Example:**
```go
// In the installStartMsg handler (search.go)
// s is api.Skill with: Name, Source ("owner/repo"), Registry ("skills.sh"|"playbooks.com")
meta := SkillMeta{
    Owner:    s.Source,                    // e.g., "vercel-labs/skills"
    Name:     s.Name,                     // e.g., "find-skills"
    Registry: s.Registry,                 // e.g., "skills.sh"
    URL:      deriveSkillURL(s),          // computed from source + registry
}
```

### Pattern 3: URL Derivation by Registry Type
**What:** Different registries have different URL patterns for skill pages.
**When to use:** When constructing SkillMeta.URL during install.
**Example:**
```go
func deriveSkillURL(s api.Skill) string {
    switch s.Registry {
    case "skills.sh":
        // skills.sh skills are GitHub repos: https://github.com/{source}/blob/main/skills/{name}
        return fmt.Sprintf("https://github.com/%s", s.Source)
    case "playbooks.com":
        // playbooks skills also come from GitHub repos
        return fmt.Sprintf("https://github.com/%s", s.Source)
    default:
        // Direct GitHub repos
        return fmt.Sprintf("https://github.com/%s", s.Source)
    }
}
```

### Anti-Patterns to Avoid
- **Duplicating config save logic:** The config save is already in `configModel.saveConfig()`. For install/uninstall, use a standalone function that loads, modifies, and saves -- do NOT duplicate the full `saveConfig()` method which also handles providers and repos.
- **Modifying api.Skill struct:** The api.Skill is a shared data type for search results. Do not add config-mutation concerns to it. Keep SkillMeta construction at the call site.
- **Ignoring duplicate entries:** Without dedup check, reinstalling a skill would create duplicate SkillMeta entries. Always check before appending.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON persistence | Custom file format or database | `encoding/json` with `MarshalIndent` | Already established pattern in config.go |
| Config file locking | File locks for concurrent access | Single-goroutine Bubble Tea event loop | TUI is single-threaded; config mutations happen sequentially in message handlers |

**Key insight:** The Bubble Tea architecture guarantees that Update handlers run sequentially, so concurrent config writes are not a concern. No file locking needed.

## Common Pitfalls

### Pitfall 1: Config State Desync Between configModel and Disk
**What goes wrong:** The `configModel` holds an in-memory copy of config (registries, repos, skills, skillsPath). If install/uninstall writes directly to disk via a standalone function, the `configModel`'s in-memory state becomes stale.
**Why it happens:** Install happens in `searchModel` (search view), not `configModel` (config view). The config view has its own copy loaded at init.
**How to avoid:** This is acceptable because `configModel` reloads from disk every time the user navigates to the config view (`newConfigModel()` calls `loadConfigFromFile()`). No cache invalidation needed.
**Warning signs:** If config view ever caches across navigations, stale data would appear.

### Pitfall 2: Uninstall Scope Ambiguity
**What goes wrong:** ORIG-06 says "uninstalling or removing a skill removes its metadata." But the current manage view only toggles provider symlinks -- it does NOT remove skills from central storage (`~/.agents/skills/`).
**Why it happens:** The manage view's `applySkillChanges()` function only creates/removes symlinks. There is no "delete from central storage" action.
**How to avoid:** Define "removing a skill" as unlinking it from ALL providers. When a skill is deselected from a provider and saved, check if it's still linked to any other provider. If not linked anywhere, remove its SkillMeta from config. Alternatively, add an explicit "uninstall" action. The requirement says "removes its metadata entry from the config.json skills array" -- this should happen when the skill is truly removed (not just unlinked from one provider).
**Warning signs:** SkillMeta entries accumulating in config.json for skills that are no longer installed anywhere.

### Pitfall 3: Source Field Format Inconsistency
**What goes wrong:** `api.Skill.Source` from skills.sh contains values like `"vercel-labs/skills"` (a GitHub repo path containing many skills), while the SkillMeta `Owner` field in the spec uses the same `"owner/repo"` format. For playbooks.com, `Source` is `"repoOwner/repoName"`.
**Why it happens:** Different registries structure their data differently. Some have one skill per repo, others have many skills per repo.
**How to avoid:** Use `api.Skill.Source` directly as `SkillMeta.Owner`. This is what the spec shows (e.g., `"owner": "anthropics/skills"`). The "owner" field is really "source repo", not a user/org.
**Warning signs:** URL derivation breaking because Source doesn't match expected format.

### Pitfall 4: Missing saveConfigData Standalone Function
**What goes wrong:** The existing `saveConfig()` is a method on `configModel` -- it reads from the model's fields (m.registries, m.repos, m.providers, etc.). A standalone function is needed that takes a `*ConfigData` and writes it to disk.
**Why it happens:** Phase 1 only needed the configModel's save path. Phase 2 needs to save from the install/uninstall handlers which don't have a configModel.
**How to avoid:** Extract a standalone `saveConfigData(cfg *ConfigData) error` function that writes the given ConfigData to disk. The existing `configModel.saveConfig()` can be refactored to use it, or they can coexist.
**Warning signs:** Trying to create a configModel just to save config from the install handler.

### Pitfall 5: Search Results Already Show Origin
**What goes wrong:** Looking at `search.go:308-311`, the search view already displays `skill.Source` in a column next to the skill name. ORIG-01 says "Search results display origin (owner/repo) next to each skill name" -- this may already be partially satisfied.
**Why it happens:** The existing code shows `truncate(skill.Source, sourceWidth)` which already renders the owner/repo.
**How to avoid:** Compare the current display against the spec in `Skills-origin-Specs.md`. The spec shows `find-skills   vercel-labs/skills   414k`. The current code already renders this format. The requirement may be satisfied by what exists, or may need minor formatting adjustments.
**Warning signs:** Over-engineering a change that's already implemented.

## Code Examples

Verified patterns from the existing codebase:

### Current Install Flow (search.go:106-136)
```go
// In searchModel.Update, case installStartMsg:
case installStartMsg:
    s := msg.skill
    return m, func() tea.Msg {
        cfg := loadConfigFromFile()
        skillsPath := ""
        if cfg != nil {
            skillsPath = cfg.SkillsPath
        }
        store := skill.NewStore(skillsPath)

        // Install to central storage
        if err := store.Install(s.Source, s.Name); err != nil {
            return installErrMsg{err: err}
        }

        // Update lock file
        _ = store.AddToLock(s.Name, s.Source)

        // Link to all configured providers
        providers := detectProviders()
        var linked []string
        for _, p := range providers {
            if p.Configured {
                if err := store.LinkToProvider(s.Name, p.Path); err == nil {
                    linked = append(linked, p.Name)
                }
            }
        }

        return installDoneMsg{skillName: s.Name, providers: linked}
    }
```

### Current Search View Display (search.go:288-318)
```go
// Results list
for i := start; i < end; i++ {
    skill := m.results[i]
    // Format: name (source) - count
    popularity := ""
    // ... popularity formatting ...

    nameFmt := fmt.Sprintf("%%-%ds", nameWidth)
    sourceFmt := fmt.Sprintf("%%-%ds", sourceWidth)

    line := fmt.Sprintf(nameFmt+" "+sourceFmt+" %6s",
        truncate(skill.Name, nameWidth),
        truncate(skill.Source, sourceWidth),
        popularity)
    // ... rendering ...
}
```

### Config Load/Save Standalone Pattern (to be created)
```go
// Standalone config save -- does not depend on configModel
func saveConfigData(cfg *ConfigData) error {
    home := os.Getenv("HOME")
    configDir := filepath.Join(home, ".config", "efx-skills")
    os.MkdirAll(configDir, 0755)

    configFile := filepath.Join(configDir, "config.json")

    // Ensure skills is not nil
    if cfg.Skills == nil {
        cfg.Skills = []SkillMeta{}
    }

    jsonData, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(configFile, jsonData, 0644)
}
```

### SkillMeta Construction from api.Skill
```go
// Construct SkillMeta from api.Skill during install
func skillMetaFromAPISkill(s api.Skill) SkillMeta {
    return SkillMeta{
        Owner:    s.Source,               // "vercel-labs/skills"
        Name:     s.Name,                // "find-skills"
        Registry: s.Registry,            // "skills.sh"
        URL:      fmt.Sprintf("https://github.com/%s", s.Source),
    }
}
```

### api.Skill Data Available at Install Time
```go
// From internal/api/client.go
type Skill struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Source      string `json:"source"`      // "vercel-labs/skills" (owner/repo)
    Description string `json:"description"`
    Installs    int    `json:"installs"`
    Stars       int    `json:"stars"`
    Registry    string `json:"registry"`    // "skills.sh" | "playbooks.com"
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Skills installed without metadata tracking | Phase 1 added SkillMeta struct and Skills array to config | Phase 1 (2026-03-05) | Foundation ready for Phase 2 wiring |
| Hardcoded skills path in Store | Skills-path from config via NewStore(skillsPath) | Phase 1 (2026-03-05) | Install flow already uses config-driven path |
| Lock file only tracked installs | Lock file continues alongside config skills array | Phase 1 decision | Both coexist -- lock file is install artifact, config is user-facing |

**Deprecated/outdated:**
- None for this phase -- all Phase 1 work is current.

## Open Questions

1. **What exactly does "uninstall" mean in the current codebase?**
   - What we know: The manage view toggles provider symlinks. There is NO mechanism to delete a skill from `~/.agents/skills/` (central storage). The codebase concern list confirms: "No uninstall from central storage."
   - What's unclear: Does ORIG-06 require removing from central storage, or just removing SkillMeta when a skill is fully unlinked?
   - Recommendation: Implement SkillMeta removal when the skill is unlinked from ALL providers in `applySkillChanges`. Do NOT implement full central storage deletion in this phase -- that's a bigger change beyond ORIG-06's scope. The requirement says "removes its metadata entry from the config.json skills array" which is specifically about the config metadata.

2. **Is ORIG-01 already satisfied by existing search view?**
   - What we know: `search.go:308-311` already displays `skill.Source` (which is `owner/repo`) alongside the skill name. The spec example shows `find-skills   vercel-labs/skills   414k` which matches the current format.
   - What's unclear: Whether the current display fully satisfies the requirement or needs refinement.
   - Recommendation: Verify the current output matches the spec. If it does, ORIG-01 may only need minor label or formatting adjustments. Do not over-engineer.

3. **Should SkillMeta.URL be repo-level or skill-level?**
   - What we know: The specs show mixed URLs -- some repo-level (`https://github.com/anthropics/skills`), some skill-level (`https://github.com/vercel-labs/skills/blob/main/skills/find-skills`). Phase 1's SkillMeta has a single `URL` field.
   - What's unclear: Which URL granularity to use at install time.
   - Recommendation: Use repo-level URL (`https://github.com/{source}`) for now. Skill-level URLs require knowing the repo structure (which varies by repo). Browser integration (Phase 4) can refine URLs later.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib), go test |
| Config file | none -- Go convention uses `*_test.go` files |
| Quick run command | `go test ./internal/tui/ -run TestConfig -v -count=1` |
| Full suite command | `go test ./... -v -count=1 && go build ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ORIG-01 | Search results display origin (owner/repo) next to skill name | unit | `go test ./internal/tui/ -run TestSearchOriginDisplay -v -count=1` | Wave 0 |
| ORIG-05 | Installing a skill writes SkillMeta to config.json skills array | unit | `go test ./internal/tui/ -run TestInstallWritesSkillMeta -v -count=1` | Wave 0 |
| ORIG-05 | Duplicate install does not create duplicate SkillMeta entries | unit | `go test ./internal/tui/ -run TestInstallSkipsDuplicate -v -count=1` | Wave 0 |
| ORIG-06 | Removing a skill removes its SkillMeta from config.json skills array | unit | `go test ./internal/tui/ -run TestRemoveSkillFromConfig -v -count=1` | Wave 0 |
| ORIG-06 | Removing nonexistent skill from config is a no-op | unit | `go test ./internal/tui/ -run TestRemoveNonexistentSkill -v -count=1` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/tui/ -v -count=1`
- **Per wave merge:** `go test ./... -v -count=1 && go build ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/tui/config_test.go` -- extend with addSkillToConfig/removeSkillFromConfig tests (ORIG-05, ORIG-06)
- [ ] Tests for ORIG-01 may be difficult to unit test (View output is string formatting) -- consider testing the data availability rather than rendered output

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/tui/config.go`, `internal/tui/search.go`, `internal/tui/manage.go`, `internal/api/client.go`, `internal/api/skillssh.go`, `internal/api/playbooks.go`, `internal/skill/store.go`
- Phase 1 summary: `.planning/phases/01-config-metadata-schema/01-01-SUMMARY.md`
- User spec: `Skills-origin-Specs.md` (correspondence list and desired behavior)

### Secondary (MEDIUM confidence)
- None needed -- all findings are from direct codebase analysis

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all existing Go + Bubble Tea
- Architecture: HIGH -- patterns directly observed in codebase, straightforward extensions
- Pitfalls: HIGH -- identified through code analysis of actual integration points

**Research date:** 2026-03-05
**Valid until:** 2026-04-05 (stable -- no external dependency changes expected)
