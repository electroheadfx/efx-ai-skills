# Phase 3: Config Page Redesign - Research

**Researched:** 2026-03-05
**Domain:** Go TUI rendering (Bubble Tea / Lipgloss), config view display formatting
**Confidence:** HIGH

## Summary

Phase 3 is a purely visual/rendering phase -- all three requirements (CONF-01, CONF-02, CONF-03) involve modifying the `View()` method of `configModel` in `internal/tui/config.go`. No data model changes, no new dependencies, no new API calls. The existing `Registry` struct, `RepoSource` struct, and lipgloss styling infrastructure are sufficient.

The main technical consideration is CONF-01's "friendly display names" for registries. The current `Registry.Name` field stores API-identifier strings like `"skills.sh"` and `"playbooks.com"`. The requirement example says `"Vercel"` instead of `"skills.sh"`, meaning a lookup/mapping from the registry name to a human-friendly label. Two approaches: add a `DisplayName` field to `Registry`, or use a hardcoded map function. A map function is simpler since registries are a closed set and avoids a schema migration.

**Primary recommendation:** All three changes fit in a single plan modifying only the `View()` method of `configModel` plus a small `registryDisplayName()` helper function. No structural changes needed.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CONF-01 | Config page shows registries with friendly names and bold URLs | Modify registry rendering in `View()` (lines 353-369), add `registryDisplayName()` map, use `lipgloss.NewStyle().Bold(true)` for URL |
| CONF-02 | Config page shows custom repos as `owner    repo-name` two-column format | Modify repo rendering in `View()` (lines 381-389), use `fmt.Sprintf` with separate `%-*s` format verbs for owner and repo |
| CONF-03 | Section label renamed from "Providers" to "Providers search" | Change string literals on lines 403 and 406 of `config.go` |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/bubbletea | v1.2.4 | TUI framework (Elm architecture) | Already in use, drives entire app |
| charmbracelet/lipgloss | v1.0.0 | Terminal styling (bold, colors, layout) | Already in use, provides `Bold(true)` for CONF-01 |
| charmbracelet/bubbles | v0.20.0 | UI components (textinput, paginator) | Already in use for config text input |

### Supporting
No new libraries needed. All styling is handled by lipgloss which is already a dependency.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hardcoded display name map | New `DisplayName` field on `Registry` struct | Field approach requires config schema migration + backward compat handling; map is simpler for 2 registries |

## Architecture Patterns

### Current Config View Structure (config.go)
```
configModel.View()
  +-- Title ("Configuration" + dirty indicator)
  +-- Registries section (section 0)
  |     +-- Section header ("Registries")
  |     +-- Per-registry row: [x] name    URL
  +-- Repos section (section 1)
  |     +-- Section header ("Custom GitHub Repos")
  |     +-- Per-repo row: owner/repo
  |     +-- Add repo input / hint
  +-- Providers section (section 2)
  |     +-- Section header ("Providers")      <-- CONF-03 changes this
  |     +-- Per-provider row: [x] name    path
  +-- Help bar
```

### Pattern 1: Registry Display Name Mapping
**What:** A pure function mapping `Registry.Name` to a human-friendly display string.
**When to use:** When rendering registry rows in the config view.
**Example:**
```go
// registryDisplayName returns a friendly label for a registry.
// Falls back to the raw Name if no mapping exists.
func registryDisplayName(name string) string {
    switch name {
    case "skills.sh":
        return "Vercel"
    case "playbooks.com":
        return "Playbooks"
    default:
        return name
    }
}
```

### Pattern 2: Bold URL Rendering with Lipgloss
**What:** Apply `Bold(true)` to the URL portion of registry rows.
**When to use:** For CONF-01 registry URL display.
**Example:**
```go
// Inside the registry rendering loop
boldURLStyle := lipgloss.NewStyle().Bold(true)
displayName := registryDisplayName(reg.Name)

// Non-selected row
line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, displayName, boldURLStyle.Render(reg.URL))
```

### Pattern 3: Two-Column Repo Display
**What:** Separate owner and repo into independent columns with spacing.
**When to use:** For CONF-02 repo row rendering.
**Example:**
```go
// Inside the repo rendering loop -- current format is "owner/repo"
// New format: two columns with spacing
ownerWidth := 20
line := fmt.Sprintf("  %-*s %s", ownerWidth, repo.Owner, repo.Repo)
```

### Anti-Patterns to Avoid
- **Modifying the data model for display concerns:** Adding `DisplayName` to `Registry` struct would require config.json schema changes, backward compatibility, and migration -- overkill for two known registries.
- **Hardcoding widths without the existing `w` variable:** The view already passes dynamic width from terminal size. Keep using it.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bold text in terminal | ANSI escape codes | `lipgloss.NewStyle().Bold(true)` | Already used throughout; handles terminal compatibility |
| Column alignment | Manual space padding | `fmt.Sprintf("%-*s", width, text)` | Already the pattern used in status.go and config.go |

**Key insight:** This phase is entirely about reformatting existing rendered output. No new infrastructure needed.

## Common Pitfalls

### Pitfall 1: Bold Style Interaction with Selected Row
**What goes wrong:** When a registry row is selected, `getSelectedRowStyle(w)` applies its own styling (bold + white foreground + blue background). If you also apply `boldURLStyle` to the URL, the nested styles may conflict or produce unexpected output.
**Why it happens:** Lipgloss styles are composed; inner styles can override outer styles for the same property.
**How to avoid:** For the selected row case, render the line as plain text (no inner bold) -- the selected row style already makes everything bold. Only apply `boldURLStyle` to the non-selected row path.
**Warning signs:** URLs appearing non-bold in selected rows, or selected rows losing their background color.

### Pitfall 2: Column Width Misalignment When Names Change Length
**What goes wrong:** "Vercel" is 6 chars while "skills.sh" is 9 chars. Changing display names shifts column alignment.
**Why it happens:** The current `nameWidth` is hardcoded to 18 in the registry section. This is generous enough for both old and new names, but worth verifying.
**How to avoid:** Keep using the existing `nameWidth := 18` constant which accommodates both original and friendly names.
**Warning signs:** URLs not lining up vertically across registry rows.

### Pitfall 3: Repos Section Owner Width Assumptions
**What goes wrong:** If owner names vary significantly in length, the two-column display looks uneven.
**Why it happens:** Default repos have owners like "yoanbernabeu" (12 chars) and "awni" (4 chars).
**How to avoid:** Use a fixed owner column width (e.g., 20 chars) with left-alignment, matching the existing column patterns.
**Warning signs:** Repo names not lining up vertically.

### Pitfall 4: Truncation of Bold-Styled URLs
**What goes wrong:** The existing `truncateStr()` function truncates plain strings. If you apply lipgloss Bold to the URL first and then truncate, you'll cut through ANSI escape codes.
**Why it happens:** `truncateStr` counts raw bytes, not visible characters.
**How to avoid:** Truncate the URL string BEFORE applying `boldURLStyle.Render()`.
**Warning signs:** Garbled terminal output, broken escape sequences at line end.

## Code Examples

### Current Registry Rendering (config.go lines 353-369)
```go
for i, reg := range m.registries {
    checkbox := "[ ]"
    if reg.Enabled {
        checkbox = "[x]"
    }
    nameWidth := 18
    urlWidth := w - nameWidth - 8
    line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, reg.Name, reg.URL)

    if m.section == 0 && i == m.selectedIdx {
        b.WriteString(getSelectedRowStyle(w).Render(line))
    } else {
        line = fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, reg.Name, statusMutedStyle.Render(truncateStr(reg.URL, urlWidth)))
        b.WriteString(tableRowStyle.Render(line))
    }
    b.WriteString("\n")
}
```

### Target Registry Rendering (CONF-01)
```go
boldURLStyle := lipgloss.NewStyle().Bold(true)

for i, reg := range m.registries {
    checkbox := "[ ]"
    if reg.Enabled {
        checkbox = "[x]"
    }
    nameWidth := 18
    urlWidth := w - nameWidth - 8
    displayName := registryDisplayName(reg.Name)

    if m.section == 0 && i == m.selectedIdx {
        // Selected: getSelectedRowStyle already applies bold
        line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, displayName, truncateStr(reg.URL, urlWidth))
        b.WriteString(getSelectedRowStyle(w).Render(line))
    } else {
        // Non-selected: bold URL, friendly name
        line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, displayName, boldURLStyle.Render(truncateStr(reg.URL, urlWidth)))
        b.WriteString(tableRowStyle.Render(line))
    }
    b.WriteString("\n")
}
```

### Current Repo Rendering (config.go lines 381-389)
```go
for i, repo := range m.repos {
    repoName := fmt.Sprintf("%s/%s", repo.Owner, repo.Repo)
    line := fmt.Sprintf("  %s", repoName)
    // ...
}
```

### Target Repo Rendering (CONF-02)
```go
ownerWidth := 16 // accommodates longest default owner "yoanbernabeu" (12) + padding

for i, repo := range m.repos {
    line := fmt.Sprintf("  %-*s %s", ownerWidth, repo.Owner, repo.Repo)
    // ...
}
```

### Current Providers Label (config.go lines 402-406)
```go
if m.section == 2 {
    b.WriteString(selectedStyle.Render("Providers"))
} else {
    b.WriteString(subtitleStyle.Render("Providers"))
}
```

### Target Providers Label (CONF-03)
```go
if m.section == 2 {
    b.WriteString(selectedStyle.Render("Providers search"))
} else {
    b.WriteString(subtitleStyle.Render("Providers search"))
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Raw registry name ("skills.sh") | Friendly display name ("Vercel") | This phase | User sees recognizable brand names |
| Single-string repo display ("owner/repo") | Two-column display ("owner    repo-name") | This phase | Better visual scanning |

**No deprecated APIs or breaking changes involved.** Lipgloss v1.0.0 is stable.

## Open Questions

1. **Exact friendly names for registries**
   - What we know: The requirement example says "Vercel" for skills.sh. Playbooks.com likely maps to "Playbooks".
   - What's unclear: Whether the user wants exactly these names or different ones.
   - Recommendation: Use "Vercel" for skills.sh and "Playbooks" for playbooks.com as shown in the requirement. The mapping function makes this trivially changeable.

2. **Owner column width for repos**
   - What we know: Default owners are "yoanbernabeu" (12), "better-auth" (11), "awni" (4).
   - What's unclear: What the maximum expected owner length is for user-added repos.
   - Recommendation: Use 16 chars -- comfortably fits all defaults with room for typical GitHub usernames (max 39 chars, but most are under 15).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None needed (uses `go test`) |
| Quick run command | `go test ./internal/tui/ -run TestConfig -v` |
| Full suite command | `go test ./... -v` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CONF-01 | Registry rows show friendly display name and bold URL | unit | `go test ./internal/tui/ -run TestRegistryDisplayName -v` | No - Wave 0 |
| CONF-01 | Config view renders friendly names (View output check) | unit | `go test ./internal/tui/ -run TestConfigViewRegistryFriendlyNames -v` | No - Wave 0 |
| CONF-02 | Repo rows show two-column owner/repo format | unit | `go test ./internal/tui/ -run TestConfigViewRepoTwoColumn -v` | No - Wave 0 |
| CONF-03 | Providers section label reads "Providers search" | unit | `go test ./internal/tui/ -run TestConfigViewProvidersLabel -v` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/tui/ -run TestConfig -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `TestRegistryDisplayName` in `config_test.go` -- covers CONF-01 mapping logic
- [ ] `TestConfigViewRegistryFriendlyNames` in `config_test.go` -- covers CONF-01 View output
- [ ] `TestConfigViewRepoTwoColumn` in `config_test.go` -- covers CONF-02 View output
- [ ] `TestConfigViewProvidersLabel` in `config_test.go` -- covers CONF-03 View output

Note: Testing `View()` output is straightforward since `configModel.View()` returns a plain string. Tests can construct a `configModel` with known data and assert substrings in the output. The existing test file (`config_test.go`) already has ~400 lines of tests using this pattern.

## Sources

### Primary (HIGH confidence)
- Source code analysis: `internal/tui/config.go` -- View() method lines 327-438
- Source code analysis: `internal/tui/styles.go` -- lipgloss style definitions
- Source code analysis: `internal/tui/config_test.go` -- existing test patterns
- Source code analysis: `go.mod` -- dependency versions confirmed

### Secondary (MEDIUM confidence)
- Lipgloss v1.0.0 API: `Bold(true)` method on `Style` -- verified via existing usage in `styles.go` (line 25, 46)

### Tertiary (LOW confidence)
- None -- all findings are from direct source code analysis.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new dependencies, all existing libraries
- Architecture: HIGH - pure View() rendering changes in a single file
- Pitfalls: HIGH - identified from direct code analysis of rendering pipeline

**Research date:** 2026-03-05
**Valid until:** 2026-04-05 (stable -- no external dependency changes expected)
