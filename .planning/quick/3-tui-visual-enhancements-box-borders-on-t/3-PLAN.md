---
phase: quick-3
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/styles.go
  - internal/tui/status.go
  - internal/tui/search.go
  - internal/tui/manage.go
  - internal/tui/preview.go
  - internal/tui/config.go
autonomous: true
requirements: [VISUAL-01, VISUAL-02, VISUAL-03]
must_haves:
  truths:
    - "All view titles are wrapped in a bordered box"
    - "Home page shows ASCII art logo with version and author"
    - "Manage view uses green/gray colored bullets instead of [x]/[ ] checkboxes"
    - "Group headers use green text when at least one skill is active, gray otherwise"
  artifacts:
    - path: "internal/tui/styles.go"
      provides: "titleBoxStyle, bulletActive, bulletInactive styles, ASCII logo constant"
    - path: "internal/tui/status.go"
      provides: "ASCII art logo title on home page"
    - path: "internal/tui/manage.go"
      provides: "Colored bullet indicators for skills and groups"
  key_links:
    - from: "internal/tui/manage.go"
      to: "internal/tui/styles.go"
      via: "bulletActive/bulletInactive styles"
      pattern: "bullet(Active|Inactive)"
---

<objective>
Add three visual enhancements to the TUI: bordered title boxes on all views, ASCII art logo on the home page, and colored bullet indicators replacing checkboxes in the manage view.

Purpose: Polish the TUI appearance for v0.2.0 release
Output: Updated styles, status, search, manage, preview, and config views
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/styles.go
@internal/tui/status.go
@internal/tui/search.go
@internal/tui/manage.go
@internal/tui/preview.go
@internal/tui/config.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add styles, ASCII logo, and title box helper to styles.go</name>
  <files>internal/tui/styles.go</files>
  <action>
Add the following to styles.go:

1. A `titleBoxStyle` that wraps titles in a rounded border with `primary` (#FBBF24) border foreground, horizontal padding 2, no vertical padding (compact box around title text):
```go
titleBoxStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(primary).
    Padding(0, 2)
```

2. Bullet styles for the manage view:
```go
bulletActiveStyle = lipgloss.NewStyle().Foreground(secondary)  // green
bulletInactiveStyle = lipgloss.NewStyle().Foreground(muted)    // gray
groupActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(secondary)   // green bold
groupInactiveStyle = lipgloss.NewStyle().Bold(true).Foreground(muted)     // gray bold
```

3. A constant for the ASCII art logo. Use a compact block-letter style that renders "efx-skills" in roughly 5 lines tall. Example:
```go
const asciiLogo = `
       __                 _    _ _ _
  ___ / _|_  __       ___| | _(_) | |___
 / _ \ |_\ \/ /_____ / __| |/ / | | / __|
|  __/  _|>  <______|\__ \   <| | | \__ \
 \___|_| /_/\_\      |___/_|\_\_|_|_|___/`
```
This uses standard "figlet" style banner text. Keep it as a raw string constant.

4. A helper function `renderTitleBox(text string) string` that renders text inside the titleBoxStyle with bold+primary foreground:
```go
func renderTitleBox(text string) string {
    return titleBoxStyle.Render(titleStyle.Render(text))
}
```
Note: titleStyle already has Bold(true) and Foreground(primary) but also has MarginBottom(1). Create an inline style inside renderTitleBox to avoid the margin: `lipgloss.NewStyle().Bold(true).Foreground(primary).Render(text)` wrapped in titleBoxStyle. Then add MarginBottom(1) to titleBoxStyle or add a newline after. Actually, simplest approach: just wrap the plain bold+primary text in the box and keep MarginBottom on titleBoxStyle:
```go
titleBoxStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(primary).
    Padding(0, 2).
    MarginBottom(1)

func renderTitleBox(text string) string {
    styledText := lipgloss.NewStyle().Bold(true).Foreground(primary).Render(text)
    return titleBoxStyle.Render(styledText)
}
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build ./internal/tui/</automated>
  </verify>
  <done>styles.go has titleBoxStyle, bullet styles, group styles, asciiLogo constant, and renderTitleBox helper. Package compiles.</done>
</task>

<task type="auto">
  <name>Task 2: Apply title boxes to all views and ASCII logo on home page</name>
  <files>internal/tui/status.go, internal/tui/search.go, internal/tui/manage.go, internal/tui/preview.go, internal/tui/config.go</files>
  <action>
Replace all `titleStyle.Render(...)` calls with `renderTitleBox(...)` across all views, and add the ASCII logo to the home page.

**status.go (home page) -- line 206:**
Replace:
```go
b.WriteString(titleStyle.Render("efx-skills v0.2.0 - Laurent Marques"))
b.WriteString("\n")
```
With:
```go
b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primary).Render(asciiLogo))
b.WriteString("\n")
b.WriteString(renderTitleBox("v0.2.0 - Laurent Marques"))
b.WriteString("\n")
```
The ASCII logo renders above the title box. The logo uses bold + primary color. The version/author line sits in a bordered box below.

**search.go -- line 271:**
Replace:
```go
b.WriteString(titleStyle.Render("efx-skills v0.2.0 - Laurent Marques"))
b.WriteString("\n\n")
```
With:
```go
b.WriteString(renderTitleBox("Search Skills"))
b.WriteString("\n")
```
The search page gets its own contextual title "Search Skills" in a box instead of the app-wide branding (branding is on home only).

**manage.go -- line 622:**
Replace:
```go
b.WriteString(titleStyle.Render(fmt.Sprintf("Manage Provider: %s", m.provider.Name)))
b.WriteString("\n")
```
With:
```go
b.WriteString(renderTitleBox(fmt.Sprintf("Manage Provider: %s", m.provider.Name)))
b.WriteString("\n")
```

**preview.go -- headerView() line 172:**
Replace:
```go
title := titleStyle.Render("efx-skills v0.2.0 - Laurent Marques")
skillTitle := subtitleStyle.Render(fmt.Sprintf("Preview: %s", m.skillName))
return title + "\n" + skillTitle
```
With:
```go
title := renderTitleBox(fmt.Sprintf("Preview: %s", m.skillName))
return title
```
Simplify to just the preview title in a box -- no need for the app branding on every sub-page.

**config.go -- line 372-376:**
Replace:
```go
title := "Configuration"
if m.dirty {
    title += " *"
}
b.WriteString(titleStyle.Render(title))
b.WriteString("\n")
```
With:
```go
title := "Configuration"
if m.dirty {
    title += " *"
}
b.WriteString(renderTitleBox(title))
b.WriteString("\n")
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/</automated>
  </verify>
  <done>All five views use renderTitleBox for titles. Home page shows ASCII art logo above the version box. Binary compiles.</done>
</task>

<task type="auto">
  <name>Task 3: Replace checkboxes with colored bullets in manage view</name>
  <files>internal/tui/manage.go</files>
  <action>
Replace `[x]`/`[ ]`/`[-]` checkbox indicators in manage.go with colored bullet indicators.

**Group headers (lines 674-679):**
Replace:
```go
checkbox := "[ ]"
if m.isGroupAllSelected(item.groupIdx) {
    checkbox = "[x]"
} else if m.isGroupPartialSelected(item.groupIdx) {
    checkbox = "[-]"
}
```
With bullet logic:
```go
bullet := bulletInactiveStyle.Render("●")
if m.isGroupAllSelected(item.groupIdx) {
    bullet = bulletActiveStyle.Render("●")
} else if m.isGroupPartialSelected(item.groupIdx) {
    bullet = bulletActiveStyle.Render("◐")
}
```
Use `◐` (half-filled circle) for partial selection in green, full `●` green for all selected, full `●` gray for none selected.

Then update the group label format (line 686):
Replace:
```go
groupLabel := fmt.Sprintf("%s %s %s (%d/%d)", arrow, checkbox, item.groupName, groupSelected, len(group.Skills))
```
With:
```go
groupLabel := fmt.Sprintf("%s %s %s (%d/%d)", arrow, bullet, item.groupName, groupSelected, len(group.Skills))
```

**IMPORTANT for selected row rendering (line 688-693):**
When the group is the selected row (highlighted with background), the bullet must NOT have color (color invisible on accent background). For the selected row case, use plain bullet text without color styles:
```go
if i == m.selectedIdx {
    // For selected row, use plain bullet (no color -- accent bg obscures it)
    plainBullet := "●"
    if m.isGroupPartialSelected(item.groupIdx) {
        plainBullet = "◐"
    }
    groupLabel := fmt.Sprintf("%s %s %s (%d/%d)", arrow, plainBullet, item.groupName, groupSelected, len(group.Skills))
    b.WriteString(getSelectedRowStyle(w).Render(groupLabel))
} else {
    // Non-selected: use colored group style based on whether any skill is active
    hasActive := groupSelected > 0
    var styledLabel string
    if hasActive {
        styledLabel = groupActiveStyle.Render(fmt.Sprintf("%s %s (%d/%d)", arrow, item.groupName, groupSelected, len(group.Skills)))
        styledLabel = bullet + " " + styledLabel
    } else {
        styledLabel = groupInactiveStyle.Render(fmt.Sprintf("%s %s (%d/%d)", arrow, item.groupName, groupSelected, len(group.Skills)))
        styledLabel = bullet + " " + styledLabel
    }
    b.WriteString(styledLabel)
}
```
Wait -- re-examine. The bullet is already styled. The group name+arrow text should be styled with groupActiveStyle/groupInactiveStyle. Build the line as: `bullet + " " + styledGroupText`. For the non-selected case, the bullet already has its color from above. The group text (arrow, name, count) gets green or gray based on `groupSelected > 0`.

Actually, simplify: keep the bullet assignment from above (which handles all/partial/none). Then for the non-selected rendering:
```go
} else {
    hasActive := groupSelected > 0
    groupText := fmt.Sprintf("%s %s (%d/%d)", arrow, item.groupName, groupSelected, len(group.Skills))
    if hasActive {
        b.WriteString(fmt.Sprintf("%s %s", bullet, groupActiveStyle.Render(groupText)))
    } else {
        b.WriteString(fmt.Sprintf("%s %s", bullet, groupInactiveStyle.Render(groupText)))
    }
}
```

**Skill items (lines 698-701):**
Replace:
```go
checkbox := "[ ]"
if skill.Selected {
    checkbox = "[x]"
}
```
With:
```go
bullet := bulletInactiveStyle.Render("●")
if skill.Selected {
    bullet = bulletActiveStyle.Render("●")
}
```

Then update the line format (line 720):
Replace:
```go
line := fmt.Sprintf("    %s %s%s", checkbox, displayName, status)
```
With:
```go
line := fmt.Sprintf("    %s %s%s", bullet, displayName, status)
```

For the selected row case (line 722-723), use a plain bullet (no color) same as groups:
```go
if i == m.selectedIdx {
    plainBullet := "●"
    line := fmt.Sprintf("    %s %s%s", plainBullet, displayName, status)
    b.WriteString(getSelectedRowStyle(w).Render(line))
}
```

For the non-selected case (lines 725-731), the bullet variable already has color. Update the line construction to use `bullet` instead of `checkbox`:
```go
} else {
    if skill.Linked && !skill.Selected {
        line = fmt.Sprintf("    %s %s%s", bullet, displayName, statusWarnStyle.Render(" (remove)"))
    } else if !skill.Linked && skill.Selected {
        line = fmt.Sprintf("    %s %s%s", bullet, displayName, statusOkStyle.Render(" (add)"))
    }
    b.WriteString(tableRowStyle.Render(line))
}
```
Note: the `line` variable already has `bullet` from the initial assignment. The if/else only overrides when there's a status suffix.

After all changes, build the binary:
```bash
go build -o ./bin/efx-skills ./cmd/efx-skills/
```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/</automated>
  </verify>
  <done>Manage view uses green bullet for active/selected skills, gray bullet for inactive, half-filled green for partial groups. Group header text is green when at least one skill is active, gray otherwise. Selected rows use plain bullets without color. Binary compiles and runs.</done>
</task>

</tasks>

<verification>
1. `go build -o ./bin/efx-skills ./cmd/efx-skills/` compiles without errors
2. Run `./bin/efx-skills` and visually confirm:
   - Home page shows ASCII art logo with version in bordered box below
   - All view titles (search, manage, config, preview) have rounded border boxes
   - Manage view shows green/gray bullets instead of [x]/[ ] checkboxes
   - Group headers show green text when skills are active, gray when not
</verification>

<success_criteria>
- All views render titles inside bordered boxes using renderTitleBox
- Home page displays ASCII art "efx-skills" logo with version/author in box
- Manage view uses colored bullet indicators (green active, gray inactive, half-filled partial)
- Group header text color reflects activation state (green if any active, gray if none)
- Binary compiles and runs without errors
</success_criteria>

<output>
After completion, create `.planning/quick/3-tui-visual-enhancements-box-borders-on-t/3-SUMMARY.md`
</output>
