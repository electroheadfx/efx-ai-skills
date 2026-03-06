---
phase: quick-13
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/manage.go
autonomous: true
requirements: [QUICK-13]
must_haves:
  truths:
    - "Title 'Manage Provider: X' stays visible when a large group (27+ items) is expanded"
    - "Remove confirmation alert text does not overflow the terminal width"
    - "Pagination still works correctly with adjusted per-page count"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "Fixed effectivePerPage and truncated alert text"
  key_links:
    - from: "effectivePerPage()"
      to: "View() rendering"
      via: "paginator.GetSliceBounds controls how many items render per page"
      pattern: "effectivePerPage|GetSliceBounds"
---

<objective>
Fix two layout bugs in the manage view: (1) title scrolling off-screen when a group with many items is expanded, and (2) remove confirmation alert text overflowing the terminal width.

Purpose: Both bugs degrade the usability of the skill management TUI.
Output: Corrected manage.go with proper per-page calculation and responsive alert text.
</objective>

<execution_context>
@./CLAUDE.md
</execution_context>

<context>
@internal/tui/manage.go
@internal/tui/styles.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Fix effectivePerPage to account for inter-group separators and help bar wrapping</name>
  <files>internal/tui/manage.go</files>
  <action>
The `effectivePerPage()` method underestimates chrome overhead, causing the paginated content plus chrome to exceed terminal height when large groups are expanded.

Two root causes:
1. The `chromeLines = 17` constant doesn't account for inter-group separator blank lines rendered within the paginated list. Each group transition within the visible page adds a `\n` separator (line 709: `b.WriteString("\n")`). With multiple groups visible on a page, this adds 2-5 extra lines.
2. The help bar in the manage view has 13 shortcut items which can wrap to 3-4 lines at typical terminal widths, but the chrome budget only allocates ~4 lines for help.

Fix approach -- increase `chromeLines` from 17 to 20 to provide a safer margin that accounts for:
- Group separator lines within paginated content (worst case ~3-4 per page)
- Multi-line help bar wrapping (3-4 lines at 80-col widths)
- titleBoxStyle MarginBottom(1) and subtitleStyle MarginBottom(1)

Additionally, account for inter-group separators dynamically: after computing `available` from height-chromeLines, subtract an estimated 3 lines for worst-case group separators within a single page. This is simpler and more robust than trying to count separators precisely (which would require knowing page content before pagination).

Specifically in `effectivePerPage()`:
- Change `const chromeLines = 17` to `const chromeLines = 21`
- This accounts for: appStyle padding (2) + title box (3) + newline after title (1) + blank+subtitle with margin (3) + newline after subtitle (1) + help bar margin+lines (5, accounting for wrapping) + pagination dots (2) + status line (2) + group separators within page (2) = 21

Keep the minimum of 5 as-is.
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/</automated>
  </verify>
  <done>effectivePerPage returns a smaller per-page count that prevents content from overflowing terminal height, keeping the title visible even when large groups are expanded.</done>
</task>

<task type="auto">
  <name>Task 2: Make remove confirmation alert text responsive to terminal width</name>
  <files>internal/tui/manage.go</files>
  <action>
The remove confirmation message at line 515 is a long single-line string:
`"Remove %s? This will delete it from disk and config. [y] confirm  [n] cancel"`

When rendered via `statusWarnStyle.Render("  " + m.statusMsg)` (line 807), there is no width constraint, so the text overflows past the right edge of narrow terminals.

Fix approach -- use lipgloss `Width()` to constrain the confirmation message rendering:

1. In the `View()` method, where `m.confirmingRemove` is handled (lines 805-807), change the rendering to apply a max width:
```go
if m.confirmingRemove {
    b.WriteString("\n")
    alertStyle := statusWarnStyle.Width(w - 4)
    b.WriteString(alertStyle.Render("  " + m.statusMsg))
}
```
The `w - 4` accounts for the 2-char left indent ("  ") plus some margin. This makes lipgloss wrap the text within the available terminal width.

2. Also shorten the confirmation message itself to be more compact. In the "r" key handler (line 515), change to:
```go
m.statusMsg = fmt.Sprintf("Remove %s? Deletes from disk+config. [y] confirm [n] cancel", skillName)
```
This is shorter while preserving all necessary information, reducing overflow risk on very narrow terminals.
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/</automated>
  </verify>
  <done>Remove confirmation alert text wraps within terminal width instead of overflowing to the right. Text is fully visible at standard terminal widths (80+ columns).</done>
</task>

</tasks>

<verification>
1. `go build -o ./bin/efx-skills ./cmd/efx-skills/` compiles without errors
2. Manual: Open manage view, expand a group with many items -- title stays visible
3. Manual: Press [r] on a skill -- confirmation text stays within terminal bounds
</verification>

<success_criteria>
- Title "Manage Provider: X" remains visible when groups with 20+ items are expanded
- Remove confirmation alert text wraps or fits within terminal width, no text cut off
- Pagination continues to work correctly (page navigation, item counts)
- Build succeeds
</success_criteria>

<output>
After completion, create `.planning/quick/13-fix-skill-management-page-layout-title-o/13-SUMMARY.md`
</output>
