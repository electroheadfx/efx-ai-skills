# Phase 4: Browser Integration - Research

**Researched:** 2026-03-05
**Domain:** Cross-platform browser open from Bubble Tea TUI views
**Confidence:** HIGH

## Summary

Phase 4 adds `[o]` keybinding to three existing TUI views (search, manage/provider, config) to open skill/registry/repo URLs in the default browser. The implementation is straightforward: each view already has the data needed to construct URLs, and opening a browser is a 15-line utility function using `os/exec` with `open` (macOS) and `xdg-open` (Linux).

The main complexity is URL construction -- different views have different data shapes, and playbooks.com skills need a skill-specific URL format (`https://playbooks.com/skills/{owner}/{repo}/{skill-name}`) with fallback to `https://playbooks.com`. The Bubble Tea keybinding pattern is well-established in the codebase (search has `[i]` for install, `[p]` for preview; manage has `[t]` for toggle, `[s]` for save; config has `[a]` for add, `[d]` for delete).

**Primary recommendation:** Create a single `browser.go` utility file in `internal/tui/` with an `openInBrowser(url string) error` function and a `urlForSkill(skill)` URL resolver, then add `[o]` keybinding to each view's `Update` method following existing key handling patterns.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| BRWS-01 | User can press [o] in search view to open selected skill's source URL in browser | Search view has `api.Skill` with `Source`, `Name`, `Registry` fields -- URL derivable. Add `"o"` case to search `Update` method |
| BRWS-02 | User can press [o] in manage/provider view to open skill or group URL in browser | Manage view has `SkillEntry` with `Name` and `Group`. Need to look up URL from config.json `skills` array (SkillMeta.URL) or derive from repos |
| BRWS-03 | User can press [o] in config view to open registry URL or repo GitHub URL in browser | Config view has `Registry.URL` (raw API URL -- need base domain) and `RepoSource` with `Owner`/`Repo`/`URL`. URL data is directly available |
| BRWS-04 | Browser open uses `open` (macOS) / `xdg-open` (Linux) for cross-platform support | Use `runtime.GOOS` switch + `exec.Command`. No third-party dependency needed. Windows explicitly out of scope per REQUIREMENTS.md |
| BRWS-05 | Playbooks.com skills open skill-specific URL if available, fallback to playbooks.com domain | PlaybooksSkill API has `RepoOwner`, `RepoName`, `SkillSlug` fields. URL pattern: `https://playbooks.com/skills/{repoOwner}/{repoName}/{skillSlug}`. Fallback to `https://playbooks.com` |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `os/exec` | stdlib | Execute `open`/`xdg-open` commands | Go stdlib, no dependency needed |
| `runtime` | stdlib | Detect OS via `runtime.GOOS` | Go stdlib, standard approach for platform detection |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `fmt` | stdlib | URL string construction | Building GitHub/playbooks URLs |
| `strings` | stdlib | String manipulation | Parsing owner/repo from Source field |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Raw `exec.Command` | `github.com/pkg/browser` | Adds dependency for 15 lines of code; project only needs macOS+Linux |

**Installation:**
```bash
# No new dependencies needed -- all stdlib
```

## Architecture Patterns

### Recommended Project Structure
```
internal/tui/
  browser.go       # NEW: openInBrowser() + URL resolution helpers
  browser_test.go  # NEW: URL construction tests (not browser launch)
  search.go        # MODIFY: add [o] key handler
  manage.go        # MODIFY: add [o] key handler
  config.go        # MODIFY: add [o] key handler
```

### Pattern 1: Browser Open Utility
**What:** A single utility function that opens a URL in the default browser, dispatching by OS.
**When to use:** Called from any view's Update method when user presses `[o]`.
**Example:**
```go
// Source: Go stdlib pattern, verified via https://go.dev/src/cmd/internal/browser/browser.go
package tui

import (
    "fmt"
    "os/exec"
    "runtime"
)

// openInBrowser opens the given URL in the user's default browser.
// Supports macOS (open) and Linux (xdg-open). Returns an error if
// the command fails or the OS is unsupported.
func openInBrowser(url string) error {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "darwin":
        cmd = exec.Command("open", url)
    case "linux":
        cmd = exec.Command("xdg-open", url)
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
    return cmd.Start()
}
```

### Pattern 2: URL Resolution for Skills
**What:** A function that determines the correct browser URL given skill metadata and registry type.
**When to use:** Before calling `openInBrowser` -- translates internal data to user-facing URLs.
**Example:**
```go
// urlForAPISkill returns the browser URL for a skill from search results.
// For skills.sh skills: https://github.com/{source}/tree/main/skills/{skill-name}
// For playbooks.com skills: https://playbooks.com/skills/{repoOwner}/{repoName}/{skillSlug}
// Fallback: https://github.com/{source}
func urlForAPISkill(s api.Skill) string {
    switch s.Registry {
    case "playbooks.com":
        // Playbooks URL: https://playbooks.com/skills/{owner}/{repo}/{skill-name}
        // Source is "owner/repo" format
        if s.Source != "" && s.Name != "" {
            return fmt.Sprintf("https://playbooks.com/skills/%s/%s", s.Source, s.Name)
        }
        return "https://playbooks.com"
    default:
        // GitHub-backed registries (skills.sh, github)
        // Source is "owner/repo" format
        if s.Source != "" {
            return fmt.Sprintf("https://github.com/%s", s.Source)
        }
        return ""
    }
}
```

### Pattern 3: Keybinding in Bubble Tea Update
**What:** Adding `[o]` to the existing `tea.KeyMsg` switch in each view's `Update` method.
**When to use:** Follows exact same pattern as existing keybindings (`[i]` install, `[p]` preview, etc.).
**Example:**
```go
// In search.go Update method, inside tea.KeyMsg switch:
case "o":
    // Open selected skill URL in browser (only when focus is on results)
    if !m.focusOnInput && len(m.results) > 0 {
        selected := m.results[m.selectedIdx]
        url := urlForAPISkill(selected)
        if url != "" {
            openInBrowser(url)
        }
    }
```

### Pattern 4: Manage View URL Resolution
**What:** In manage view, skills are `SkillEntry` (local filesystem data). Must look up URL from config.json `skills` array or derive from known repos.
**When to use:** When user presses `[o]` on a skill or group in manage view.
**Example:**
```go
// urlForManagedSkill returns the browser URL for a locally installed skill.
// Looks up SkillMeta from config, falls back to searching repos.
func urlForManagedSkill(skillName string) string {
    cfg := loadConfigFromFile()
    if cfg == nil {
        return ""
    }
    // Check skills array for direct URL
    for _, meta := range cfg.Skills {
        if meta.Name == skillName {
            return meta.URL
        }
    }
    return ""
}

// urlForGroup returns the repo URL for a skill group.
// Checks config repos, then tries GitHub URL from skill metadata.
func urlForGroup(groupName string, cfg *ConfigData) string {
    if cfg == nil {
        return ""
    }
    // Check if any skill in the group has a URL we can derive a repo from
    for _, meta := range cfg.Skills {
        if strings.HasPrefix(meta.Name, groupName+"-") || meta.Name == groupName {
            // Return the owner/repo base URL
            return meta.URL
        }
    }
    return ""
}
```

### Pattern 5: Config View URL Resolution
**What:** In config view, open registry base URL or repo GitHub URL.
**When to use:** When user presses `[o]` on a registry or repo row.
**Example:**
```go
// In config.go Update method:
case "o":
    switch m.section {
    case 0: // Registries
        if len(m.registries) > m.selectedIdx {
            reg := m.registries[m.selectedIdx]
            // Open the registry base URL (not API URL)
            url := registryBaseURL(reg.Name)
            openInBrowser(url)
        }
    case 1: // Repos
        if len(m.repos) > m.selectedIdx {
            repo := m.repos[m.selectedIdx]
            url := repo.DeriveURL() // Already returns https://github.com/owner/repo
            openInBrowser(url)
        }
    }
```

### Anti-Patterns to Avoid
- **Blocking on browser open:** Use `cmd.Start()` not `cmd.Run()`. The TUI must not freeze waiting for the browser to close. `Start()` launches the process and returns immediately.
- **Using `cmd.Output()` or `cmd.CombinedOutput()`:** These wait for the command to finish, which blocks indefinitely since the browser stays open.
- **Constructing API URLs as browser URLs:** The registry URL in config is the API endpoint (e.g., `https://skills.sh/api/search`). The browser URL should be the base domain (e.g., `https://skills.sh`).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Browser open | Custom URL protocol handler | `exec.Command("open"/"xdg-open", url).Start()` | OS handles browser selection, protocol dispatch |
| Playbooks URL format | Custom URL parser | `fmt.Sprintf("https://playbooks.com/skills/%s/%s", source, name)` | Playbooks URL pattern is deterministic from API fields |

**Key insight:** This phase is pure glue code -- no complex logic, no new data structures, no new API calls. The data is already available; we just need to wire `[o]` to construct URLs and open them.

## Common Pitfalls

### Pitfall 1: Blocking the TUI on browser launch
**What goes wrong:** Using `cmd.Run()` or `cmd.CombinedOutput()` blocks the TUI until the browser process exits (which may never happen).
**Why it happens:** Confusion between `Start()` (async) and `Run()` (sync).
**How to avoid:** Always use `cmd.Start()` which returns immediately after launching the process.
**Warning signs:** TUI freezes when `[o]` is pressed.

### Pitfall 2: Opening API URLs instead of human-readable URLs
**What goes wrong:** Registry `URL` field stores `https://skills.sh/api/search` -- this is the API endpoint, not a human-readable page.
**Why it happens:** The `Registry.URL` field is named ambiguously.
**How to avoid:** Create a `registryBaseURL(name string)` function that maps registry names to browser-friendly URLs: `"skills.sh" -> "https://skills.sh"`, `"playbooks.com" -> "https://playbooks.com"`.
**Warning signs:** Browser opens an API JSON response instead of a web page.

### Pitfall 3: Missing URL for managed skills not in config
**What goes wrong:** Skills installed before v0.2.0 have no `SkillMeta` entry in config.json, so `urlForManagedSkill` returns empty string.
**Why it happens:** Doctor phase (Phase 6) handles backfilling metadata -- it has not run yet.
**How to avoid:** Silently do nothing (or show brief status message) when URL is unavailable. Do NOT attempt to guess URLs for unknown skills. The Doctor phase will handle backfilling.
**Warning signs:** Pressing `[o]` on a pre-v0.2.0 skill does nothing -- this is expected behavior.

### Pitfall 4: Playbooks SkillSlug vs Name confusion
**What goes wrong:** The `api.Skill.Name` field might not match the `skillSlug` from playbooks API, causing wrong URLs.
**Why it happens:** `api.Skill.Name` is mapped from `PlaybooksSkill.Name` while `api.Skill.ID` is mapped from `PlaybooksSkill.SkillSlug` (see `playbooks.go:60`).
**How to avoid:** For playbooks.com skills in search view, the URL pattern `https://playbooks.com/skills/{source}/{name}` uses the skill Name (which is the directory name), matching how playbooks.com structures its URLs. Verify with a live test.
**Warning signs:** Playbooks URL returns 404.

### Pitfall 5: Pressing [o] while input focused in search view
**What goes wrong:** User types "o" into the search input instead of opening browser.
**Why it happens:** Key events go to the text input when `focusOnInput` is true.
**How to avoid:** Only handle `[o]` when `!m.focusOnInput` (same guard as `[i]` and `[p]`).
**Warning signs:** "o" appears in search text.

## Code Examples

Verified patterns from the existing codebase:

### Existing Key Handler Pattern (search.go)
```go
// Source: internal/tui/search.go:212-221
case "i":
    // Install selected skill (only when focus is on results)
    if !m.focusOnInput && len(m.results) > 0 && !m.installing {
        selected := m.results[m.selectedIdx]
        m.installing = true
        m.installMsg = ""
        return m, func() tea.Msg {
            return installStartMsg{skill: selected}
        }
    }
```

### Existing Key Handler Pattern (config.go)
```go
// Source: internal/tui/config.go:228-233
case "a":
    // Add repo (only in repos section)
    if m.section == 1 {
        m.addingRepo = true
        m.textInput.Focus()
        return m, textinput.Blink
    }
```

### Help Text Update Pattern
```go
// Source: internal/tui/search.go:351-352 (results-focused help)
// Current:
b.WriteString(helpStyle.Render("  [i] install  [p/enter] preview  ..."))
// Updated (add [o]):
b.WriteString(helpStyle.Render("  [i] install  [o] open  [p/enter] preview  ..."))
```

### API Skill Fields Available in Search View
```go
// Source: internal/api/client.go:55-63
type Skill struct {
    ID          string // SkillSlug for playbooks, skill ID for skills.sh
    Name        string // Skill directory name
    Source      string // "owner/repo" format
    Description string
    Installs    int
    Stars       int
    Registry    string // "skills.sh" or "playbooks.com"
}
```

### SkillMeta Fields Available in Config
```go
// Source: internal/tui/config.go:36-41
type SkillMeta struct {
    Owner    string // "owner/repo" format (from Source)
    Name     string // Skill directory name
    Registry string // "skills.sh", "playbooks.com", or "github"
    URL      string // https://github.com/{source}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `github.com/pkg/browser` dependency | Raw `exec.Command` with `runtime.GOOS` | Always (for simple macOS/Linux) | No extra dependency for 2-platform support |
| `cmd.Run()` blocking | `cmd.Start()` non-blocking | Always (TUI apps) | Prevents TUI freeze |

**Deprecated/outdated:**
- None relevant -- `os/exec` and `runtime.GOOS` are stable Go APIs.

## Open Questions

1. **Playbooks skill-specific URL verification**
   - What we know: URL pattern is `https://playbooks.com/skills/{repoOwner}/{repoName}/{skillName}` based on live verification
   - What's unclear: Whether ALL playbooks skills follow this pattern or some edge cases exist
   - Recommendation: Implement the pattern, add fallback to `https://playbooks.com` on 404 (but since we just open the URL, the browser handles 404 gracefully)

2. **Manage view: skills without metadata**
   - What we know: Pre-v0.2.0 skills lack SkillMeta in config. Doctor (Phase 6) backfills them.
   - What's unclear: How many skills will have no URL when user first uses [o]
   - Recommendation: Silently skip (no error shown) or show brief "URL not available" status message. This is a graceful degradation, not a bug.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None needed (Go convention) |
| Quick run command | `go test ./internal/tui/ -run Browser -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BRWS-01 | Search view [o] constructs correct URL for skills.sh and playbooks skills | unit | `go test ./internal/tui/ -run TestURLForAPISkill -count=1` | Wave 0 |
| BRWS-02 | Manage view [o] looks up URL from config SkillMeta | unit | `go test ./internal/tui/ -run TestURLForManagedSkill -count=1` | Wave 0 |
| BRWS-03 | Config view [o] resolves registry base URL and repo DeriveURL | unit | `go test ./internal/tui/ -run TestRegistryBaseURL -count=1` | Wave 0 |
| BRWS-04 | openInBrowser dispatches by OS (darwin/linux) | unit | `go test ./internal/tui/ -run TestOpenInBrowser -count=1` | Wave 0 |
| BRWS-05 | Playbooks URL uses skill-specific format with fallback | unit | `go test ./internal/tui/ -run TestURLForAPISkill/playbooks -count=1` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/tui/ -run Browser -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/tui/browser_test.go` -- covers BRWS-01 through BRWS-05 (URL construction tests)
- [ ] `internal/tui/browser.go` -- the utility module itself

## Sources

### Primary (HIGH confidence)
- Go stdlib `os/exec` -- exec.Command with Start() for non-blocking launch
- Go stdlib `runtime` -- runtime.GOOS for platform detection
- Go internal `cmd/internal/browser/browser.go` -- canonical pattern for browser open in Go
- Existing codebase: `internal/tui/search.go`, `manage.go`, `config.go` -- keybinding patterns, data shapes

### Secondary (MEDIUM confidence)
- Playbooks.com URL pattern: `https://playbooks.com/skills/{owner}/{repo}/{skill-name}` -- verified via live WebFetch of `https://playbooks.com/skills/vercel-labs/agent-skills/web-design-guidelines`
- PlaybooksSkill API struct fields (`RepoOwner`, `RepoName`, `SkillSlug`) -- mapped to `api.Skill` as `Source` (owner/repo) and `ID` (skillSlug)

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- Go stdlib only, no dependencies
- Architecture: HIGH -- follows exact patterns already in codebase
- Pitfalls: HIGH -- well-understood from existing TUI key handling
- URL construction: MEDIUM -- playbooks URL pattern verified but edge cases possible

**Research date:** 2026-03-05
**Valid until:** 2026-04-05 (stable -- stdlib APIs, no fast-moving dependencies)
