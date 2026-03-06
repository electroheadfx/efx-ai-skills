---
phase: quick-4
plan: 1
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/styles.go
  - internal/tui/status.go
  - internal/tui/manage.go
  - internal/tui/search.go
  - internal/tui/config.go
  - internal/tui/preview.go
autonomous: true
requirements: [QUICK-4]
must_haves:
  truths:
    - "Help bar text wraps or abbreviates gracefully at terminal widths below 80 columns"
    - "All shortcut keys remain visible (possibly across two lines) at widths as narrow as 50 columns"
    - "At full width (100+), help text displays the same content as before on a single line"
  artifacts:
    - path: "internal/tui/styles.go"
      provides: "renderHelpBar helper function for responsive help text"
      contains: "renderHelpBar"
    - path: "internal/tui/status.go"
      provides: "Responsive help bar in status view"
      contains: "renderHelpBar"
    - path: "internal/tui/manage.go"
      provides: "Responsive help bar in manage view"
      contains: "renderHelpBar"
    - path: "internal/tui/search.go"
      provides: "Responsive help bar in search view"
      contains: "renderHelpBar"
    - path: "internal/tui/config.go"
      provides: "Responsive help bar in config view"
      contains: "renderHelpBar"
    - path: "internal/tui/preview.go"
      provides: "Responsive help bar in preview view"
      contains: "renderHelpBar"
  key_links:
    - from: "internal/tui/styles.go"
      to: "all view files"
      via: "renderHelpBar function call"
      pattern: "renderHelpBar\\("
---

<objective>
Make the TUI status line shortcut keys responsive for narrow terminal widths.

Currently all views render help bars as single hardcoded strings (e.g. manage view: 13 shortcuts on one line ~110 chars). When the terminal is narrower than the help text, it either gets clipped or wraps mid-shortcut, breaking readability.

Purpose: Ensure shortcut help remains usable at any reasonable terminal width (50+).
Output: A shared `renderHelpBar` helper and updated help rendering in all 5 views.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/styles.go
@internal/tui/status.go
@internal/tui/manage.go
@internal/tui/search.go
@internal/tui/config.go
@internal/tui/preview.go
@internal/tui/app.go

<interfaces>
From internal/tui/styles.go:
```go
var helpStyle = lipgloss.NewStyle().Foreground(muted).MarginTop(1)
```

From internal/tui/app.go (width propagation):
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.statusModel.width = int(float64(msg.Width) * 0.9)
    m.searchModel.width = int(float64(msg.Width) * 0.9)
    m.manageModel.width = int(float64(msg.Width) * 0.9)
    m.configModel.width = int(float64(msg.Width) * 0.9)
```

Each sub-model has a `width int` field already available.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create renderHelpBar helper in styles.go</name>
  <files>internal/tui/styles.go</files>
  <action>
Add a `renderHelpBar(width int, items []string) string` function to styles.go that:

1. Takes the available width and a slice of shortcut strings (e.g. `[]string{"[s] search", "[m/enter] manage", "[c] config", "[r] refresh", "[q] quit"}`).
2. Joins items with two-space separators ("  ") into lines, wrapping to the next line when adding another item would exceed `width - 4` (accounting for the 2-char left indent on each line).
3. Each line is prefixed with "  " (2 spaces indent).
4. Lines are joined with "\n".
5. The full result is styled with `helpStyle`.
6. If width <= 0, default to 80.

The logic is straightforward greedy line-packing:
```go
func renderHelpBar(width int, items []string) string {
    if width <= 0 {
        width = 80
    }
    maxLineW := width - 4 // 2 indent + 2 margin
    if maxLineW < 20 {
        maxLineW = 20
    }

    var lines []string
    currentLine := ""
    for _, item := range items {
        candidate := currentLine
        if candidate != "" {
            candidate += "  " + item
        } else {
            candidate = item
        }
        if len(candidate) > maxLineW && currentLine != "" {
            lines = append(lines, "  "+currentLine)
            currentLine = item
        } else {
            currentLine = candidate
        }
    }
    if currentLine != "" {
        lines = append(lines, "  "+currentLine)
    }

    return helpStyle.Render(strings.Join(lines, "\n"))
}
```

Add `"strings"` to the import block in styles.go (it is not currently imported there).
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build ./internal/tui/</automated>
  </verify>
  <done>renderHelpBar function exists in styles.go, compiles, and is ready to be called from all views.</done>
</task>

<task type="auto">
  <name>Task 2: Replace hardcoded help strings with renderHelpBar calls in all views</name>
  <files>internal/tui/status.go, internal/tui/manage.go, internal/tui/search.go, internal/tui/config.go, internal/tui/preview.go</files>
  <action>
Replace every `helpStyle.Render("...")` help bar call with `renderHelpBar(m.width, items)` in each view. Preserve the exact same shortcut keys and labels -- only change how they are rendered.

**status.go** (lines 294-299): Two conditional branches.
- Configured provider branch (line 296):
  ```go
  b.WriteString(renderHelpBar(m.width, []string{"[s] search", "[m/enter] manage", "[c] config", "[r] refresh", "[q] quit"}))
  ```
  Note: Change "[m/↵]" to "[m/enter]" for consistency (the ↵ character takes variable display width).
- Unconfigured provider branch (line 298):
  ```go
  b.WriteString(renderHelpBar(m.width, []string{"[s] search", "[c] configure", "[r] refresh", "[q] quit"}))
  ```
- Remove the "\n" before the help since helpStyle already has MarginTop(1).

**manage.go** (line 778): Single long line with 13 shortcuts.
  ```go
  b.WriteString(renderHelpBar(m.width, []string{
      "[space] preview", "[o] open", "[v] verify", "[u] update", "[g] update all",
      "[t] toggle", "[r] remove", "[enter] collapse/expand",
      "[a] all", "[n] none", "[s] apply/save", "[<-/->] page", "[esc] back",
  }))
  ```
- Remove the "\n" before the help since helpStyle already has MarginTop(1).

**search.go** (lines 370-376): Three conditional branches.
- focusOnInput (line 371):
  ```go
  b.WriteString(renderHelpBar(m.width, []string{"[enter] search", "[tab] focus results", "[esc] back", "[q] quit"}))
  ```
  Note: Change "[↵]" to "[enter]" for consistent display width.
- results available (line 373):
  ```go
  b.WriteString(renderHelpBar(m.width, []string{"[i] install", "[o] open", "[p/enter] preview", "[up/down] navigate", "[<-/->] page", "[tab] focus input", "[esc] back", "[q] quit"}))
  ```
  Note: Change arrow glyphs "[↑/↓]" to "[up/down]" and "[←/→]" to "[<-/->]" for consistent width. Change "[p/↵]" to "[p/enter]".
- else (line 375):
  ```go
  b.WriteString(renderHelpBar(m.width, []string{"[enter] search", "[i] install", "[p] preview", "[<-/->] page", "[esc] back", "[q] quit"}))
  ```
- Remove the "\n" before the help since helpStyle already has MarginTop(1).

**config.go** (lines 468-473): Two conditional branches.
- section == 1 (line 471):
  ```go
  helpItems := []string{"[tab] section", "[a] add", "[o] open", "[d] delete", "[s] save", "[esc] back", "[q] quit"}
  ```
- else (line 469):
  ```go
  helpItems := []string{"[tab] section", "[space] toggle", "[o] open", "[s] save", "[esc] back", "[q] quit"}
  ```
- Replace the final render:
  ```go
  b.WriteString(renderHelpBar(m.width, helpItems))
  ```
- Remove the "\n" before the help since helpStyle already has MarginTop(1).

**preview.go** (line 177): Single line in footerView().
  The preview model does not have a `width` field. It uses `m.viewport` which has a width. Use `m.viewport.Width` as the width parameter.
  ```go
  func (m previewModel) footerView() string {
      scrollPct := fmt.Sprintf("%3.0f%%", m.viewport.ScrollPercent()*100)
      return renderHelpBar(m.viewport.Width, []string{scrollPct, "[j/k/up/down] scroll", "[space/b] page", "[g/G] top/bottom", "[esc] back"})
  }
  ```
  Add `"fmt"` to preview.go imports if not already present.

After all changes, build the binary:
```bash
go build -o ./bin/efx-skills ./cmd/efx-skills/
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && echo "Build successful"</automated>
  </verify>
  <done>All 5 view files use renderHelpBar with proper item slices. Binary builds. Help bars wrap gracefully at narrow widths while displaying identically to before at wide widths.</done>
</task>

</tasks>

<verification>
1. `go build -o ./bin/efx-skills ./cmd/efx-skills/` succeeds with no errors
2. `go vet ./internal/tui/` passes
3. Manual spot-check: run `./bin/efx-skills` in an 80-column terminal -- help bar looks normal on one line
4. Resize terminal to ~50 columns -- help bar wraps to multiple lines, all shortcuts visible
</verification>

<success_criteria>
- All help bars across all 5 views use the shared renderHelpBar function
- At terminal width 100+, help text appears on a single line (same as before)
- At terminal width 50-80, help text wraps to 2+ lines with all shortcuts visible
- Binary builds and runs without errors
</success_criteria>

<output>
After completion, create `.planning/quick/4-make-tui-status-line-shortcut-keys-respo/4-SUMMARY.md`
</output>
