---
phase: quick-11
plan: 11
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/config.go
  - internal/tui/styles.go
autonomous: true
requirements: [config-view-ui-fixes]
must_haves:
  truths:
    - "[a] add appears only ONCE when GitHub repos section is active (inline hint only, not in bottom help bar)"
    - "[r] remove repo appears inline under repos list when a repo is selected and removes the selected repo"
    - "Active config section is wrapped in a blue-bordered box; inactive sections have muted or no border"
  artifacts:
    - path: "internal/tui/config.go"
      provides: "Config view with deduplicated help, [r] remove, section borders"
  key_links:
    - from: "config.go View()"
      to: "styles.go border styles"
      via: "lipgloss border rendering"
      pattern: "Border.*accent|focusedBoxStyle"
---

<objective>
Fix three UI issues in the config view: remove duplicate [a] add from help bar, add [r] remove repo inline shortcut, and add blue box border for the active section.

Purpose: Clean up config view UX -- deduplicate shortcuts, add missing remove action, improve visual focus indicator.
Output: Updated config.go with all three fixes, updated styles.go if needed for section border styles.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/config.go
@internal/tui/styles.go
</context>

<interfaces>
<!-- Key types and styles the executor needs -->

From internal/tui/config.go:
```go
type configModel struct {
    section     int // 0=registries, 1=repos, 2=providers
    selectedIdx int
    repos       []RepoSource
    width       int
    // ...
}
```

From internal/tui/styles.go:
```go
var accent = lipgloss.Color("#007AE3") // Blue (selection bg)
var muted  = lipgloss.Color("#6B7280") // Gray

var boxStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(muted).
    Padding(1, 2)

var focusedBoxStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(primary).
    Padding(1, 2)

func renderHelpBar(width int, items []string) string
func getSelectedRowStyle(width int) lipgloss.Style
```
</interfaces>

<tasks>

<task type="auto">
  <name>Task 1: Fix help bar duplication and add [r] remove key binding</name>
  <files>internal/tui/config.go</files>
  <action>
Two changes in config.go:

**1. Fix help bar duplication (View method, ~line 470-473):**
When `m.section == 1`, the bottom help bar currently includes `[a] add` which duplicates the inline `[a] add repo` hint. Change the section==1 help items to remove `[a] add` and `[d] delete` (since both are shown inline now):
```go
helpItems := []string{"[tab] section", "[space] toggle", "[o] open", "[s] save", "[esc] back", "[q] quit"}
if m.section == 1 {
    helpItems = []string{"[tab] section", "[o] open", "[s] save", "[esc] back", "[q] quit"}
}
```
This removes `[a] add` and `[d] delete` from the bottom bar for repos section since they appear inline.

**2. Add [r] key binding (Update method, ~line 238):**
Add a `case "r":` alongside the existing `case "d":` block. The `r` key should do the same repo removal as `d` (section==1 only):
```go
case "d", "r":
    // Delete/remove repo (only in repos section)
    if m.section == 1 && len(m.repos) > 0 {
        ...
    }
```
Merge the existing `"d"` case to also match `"r"`.

**3. Update inline hint (View method, ~line 436):**
Change the inline hint from just `[a] add repo` to show both actions when repos exist:
```go
} else if m.section == 1 {
    hints := "  [a] add repo"
    if len(m.repos) > 0 {
        hints += "  [r] remove repo"
    }
    b.WriteString(statusMutedStyle.Render(hints))
    b.WriteString("\n")
}
```
Only show `[r] remove repo` when there are repos to remove.
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && echo "BUILD OK"</automated>
  </verify>
  <done>
    - [a] add no longer appears in the bottom help bar when repos section is active
    - [r] key removes the selected repo from the list (same behavior as [d])
    - Inline hint shows "[a] add repo  [r] remove repo" when repos exist
    - [d] delete also removed from bottom help bar for repos section (shown inline)
  </done>
</task>

<task type="auto">
  <name>Task 2: Add blue box border for active section</name>
  <files>internal/tui/config.go, internal/tui/styles.go</files>
  <action>
Add section border styles and wrap each config section in a border box.

**1. In styles.go, add two new styles for config section borders:**
```go
// Config section box (inactive) - subtle muted border
configSectionStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(muted).
    Padding(0, 1)

// Config section box (active) - blue border for focus
configSectionActiveStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(accent).
    Padding(0, 1)
```
Use `accent` (blue, `#007AE3`) for active and `muted` (gray) for inactive. Padding `(0, 1)` to keep sections compact (not the larger `(1, 2)` of boxStyle).

**2. In config.go View(), wrap each section in its border box:**

For each section (registries, repos, providers), collect the section content into a string, then wrap it using the active or inactive style. The section header should be INSIDE the box.

Refactor the View() method. For each of the 3 sections:
- Build section content (header + rows + inline hints) into a `strings.Builder` or string
- Choose style: `configSectionActiveStyle` if `m.section == sectionIndex`, else `configSectionStyle`
- Apply `.Width(w - 4)` to account for border characters (2 per side)
- Write the styled section to the main builder

Example pattern for repos section:
```go
// Repos section
var reposContent strings.Builder
reposContent.WriteString(selectedStyle.Render("Custom GitHub Repos") + "\n") // or subtitleStyle
for i, repo := range m.repos {
    // ... existing row rendering ...
    reposContent.WriteString(line + "\n")
}
if m.addingRepo {
    reposContent.WriteString(fmt.Sprintf("  Add: %s\n", m.textInput.View()))
} else if m.section == 1 {
    // inline hints
}

// Wrap in border
sectionW := w - 4
if sectionW < 20 { sectionW = 20 }
if m.section == 1 {
    b.WriteString(configSectionActiveStyle.Width(sectionW).Render(reposContent.String()))
} else {
    b.WriteString(configSectionStyle.Width(sectionW).Render(reposContent.String()))
}
b.WriteString("\n")
```

Apply the same pattern to all three sections (registries at section==0, repos at section==1, providers at section==2). Remove the leading `\n` separators between sections since the box borders provide visual separation. Keep one `\n` between boxes for spacing.

Important: The section header styling (selectedStyle vs subtitleStyle) should REMAIN as-is inside the box -- the box border indicates focus, the header color also indicates focus. Both reinforce the active state.
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && echo "BUILD OK"</automated>
  </verify>
  <done>
    - Active config section has a blue (accent color) rounded border box
    - Inactive sections have a muted gray rounded border box
    - All three sections (Registries, Custom GitHub Repos, Providers search) are wrapped in border boxes
    - Section content (header, rows, inline hints) renders correctly inside borders
    - Build succeeds with no errors
  </done>
</task>

</tasks>

<verification>
```bash
cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && go vet ./internal/tui/...
```
</verification>

<success_criteria>
- go build succeeds
- Config view shows blue border on active section, muted border on inactive sections
- [a] add does NOT appear in bottom help bar when repos section is active
- [r] key removes selected repo (same as [d])
- Inline hint shows "[a] add repo  [r] remove repo" when repos section is active and repos exist
</success_criteria>

<output>
After completion, create `.planning/quick/11-config-view-fix-duplicate-add-shortcut-a/11-SUMMARY.md`

IMPORTANT: Run `go build -o ./bin/efx-skills ./cmd/efx-skills/` as final step.
</output>
