---
phase: quick-9
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: [internal/tui/manage.go, internal/tui/app.go]
autonomous: false
requirements: [QUICK-9]

must_haves:
  truths:
    - "Manage view output never exceeds terminal height regardless of how many items exist"
    - "Collapsing groups does not push the title box off the top of the TUI"
    - "Expanding groups keeps content within terminal bounds"
    - "Resizing the terminal dynamically adjusts how many items are shown per page"
  artifacts:
    - path: "internal/tui/manage.go"
      provides: "Dynamic perPage based on terminal height, height field on manageModel"
      contains: "m.height"
    - path: "internal/tui/app.go"
      provides: "Height propagation to manageModel on WindowSizeMsg"
      contains: "m.manageModel.height"
  key_links:
    - from: "app.go WindowSizeMsg handler"
      to: "manageModel.height"
      via: "direct field assignment in tea.WindowSizeMsg case"
      pattern: "manageModel\\.height"
    - from: "manageModel.View()"
      to: "manageModel.effectivePerPage()"
      via: "dynamic calculation replacing const perPage in View and navigation"
      pattern: "effectivePerPage"
---

<objective>
Fix manage view layout overflow: the rendered output exceeds terminal height because perPage is a hard constant (18) that does not account for chrome lines (title box, subtitle, separators, help bar, pagination dots, status). When total rendered lines exceed terminal height, bubbletea scrolls internally, and collapsing groups leaves a stale scroll offset that pushes the header off screen.

Purpose: Make the manage view fit within the terminal by dynamically calculating how many skill items can be displayed based on actual terminal height minus chrome overhead.

Output: Modified internal/tui/manage.go (height field, dynamic perPage), modified internal/tui/app.go (height propagation)
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/manage.go
@internal/tui/app.go
</context>

<interfaces>
<!-- Key types and contracts the executor needs -->

From internal/tui/manage.go:
```go
const perPage = 18  // PROBLEM: hard constant, must become dynamic

type manageModel struct {
    // ... existing fields
    width       int        // already set by app.go
    // NOTE: no height field -- must be added
    paginator   paginator.Model
    displayList []displayItem
    selectedIdx int
}

// buildDisplayList rebuilds flat display list, calls SetTotalPages + clampPaginator
func (m *manageModel) buildDisplayList()

// clampPaginator ensures Page and selectedIdx are within bounds -- already exists
func (m *manageModel) clampPaginator()
```

From internal/tui/app.go (WindowSizeMsg handler, lines 82-89):
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    // Passes width to sub-models but NOT height:
    m.manageModel.width = int(float64(msg.Width) * 0.9)
    // m.manageModel.height is MISSING
```

Chrome overhead in View() (lines that are NOT skill items):
- appStyle padding: 2 lines (Padding(1,2) = 1 top + 1 bottom)
- renderTitleBox: 3 lines (border top + text + border bottom)
- "\n" after title: 1 line
- "\n" before subtitle: 1 line
- subtitleStyle line: 1 line
- "\n" after subtitle: 1 line
- helpStyle with MarginTop(1): 1 blank + 2-3 help lines = 3-4 lines
- Pagination dots (conditional): 2 lines
- Status message (conditional): 2 lines
- Group separator blank lines: ~1 per group transition on current page

Fixed chrome overhead estimate: ~13 lines minimum. Safe to use 15 as buffer.
</interfaces>

<tasks>

<task type="auto">
  <name>Task 1: Add height tracking and dynamic perPage to manageModel</name>
  <files>internal/tui/manage.go, internal/tui/app.go</files>
  <action>
1. In `internal/tui/app.go`, in the `tea.WindowSizeMsg` case (line 82-89), add height propagation to manageModel:
   ```go
   m.manageModel.height = msg.Height
   ```
   Place it right after the existing `m.manageModel.width = ...` line.

2. In `internal/tui/manage.go`:

   a. Add a `height` field to `manageModel` struct (after `width int`):
      ```go
      height      int
      ```

   b. Remove the `const perPage = 18` line entirely (line 86).

   c. Add an `effectivePerPage()` method that computes available rows for display items:
      ```go
      // effectivePerPage returns how many display items fit on one page given the
      // current terminal height. Falls back to 18 when height is unknown.
      func (m *manageModel) effectivePerPage() int {
          if m.height <= 0 {
              return 18 // sensible default before first WindowSizeMsg
          }
          // Chrome overhead:
          //   appStyle padding:     2 (top 1 + bottom 1)
          //   title box:            3 (border + text + border)
          //   newline after title:  1
          //   blank + subtitle:     2
          //   newline after subtitle: 1
          //   help bar margin+lines: 4
          //   pagination dots:       2
          //   status line:           2
          // Total fixed chrome:     ~17 lines
          const chromeLines = 17
          available := m.height - chromeLines
          if available < 5 {
              available = 5 // minimum usable
          }
          return available
      }
      ```

   d. Replace every reference to the old `perPage` constant in manage.go. There are references in:
      - `newManageModel`: `p.PerPage = perPage` -> call will be deferred (see below)
      - `buildDisplayList`: after `SetTotalPages`, update `m.paginator.PerPage` too
      - Up/down key handlers: `m.paginator.Page = m.selectedIdx / perPage`
      - Left/right key handlers: `m.selectedIdx = m.paginator.Page * perPage`

      Specifically:

      In `newManageModel` (line 92): Change `p.PerPage = perPage` to `p.PerPage = 18` (initial default; will be updated dynamically).

      In `buildDisplayList` (before the `SetTotalPages` call), add:
      ```go
      m.paginator.PerPage = m.effectivePerPage()
      ```

      In the "up"/"k" key handler (line 406):
      ```go
      m.paginator.Page = m.selectedIdx / m.effectivePerPage()
      ```

      In the "down"/"j" key handler (line 410):
      ```go
      m.paginator.Page = m.selectedIdx / m.effectivePerPage()
      ```

      In the "left"/"h"/"pgup" handler (line 415):
      ```go
      m.selectedIdx = m.paginator.Page * m.effectivePerPage()
      ```

      In the "right"/"l"/"pgdown" handler (line 418):
      ```go
      m.selectedIdx = m.paginator.Page * m.effectivePerPage()
      ```

      In the "enter" key handler, the line that sets paginator page after finding group header (inside the for-loop, line 460):
      ```go
      m.paginator.Page = m.selectedIdx / m.effectivePerPage()
      ```

   e. Add a `tea.WindowSizeMsg` case to the `Update` method so that when the terminal is resized while in the manage view, the display list is rebuilt with the new perPage:
      ```go
      case tea.WindowSizeMsg:
          m.width = int(float64(msg.Width) * 0.9)
          m.height = msg.Height
          m.buildDisplayList()
          return m, nil
      ```
      Place this case in the `switch msg := msg.(type)` block, before the `tea.KeyMsg` case. Note: app.go also sets these, but handling it here ensures manageModel stays consistent when it receives the message via delegation.

3. Build the binary:
   ```
   go build -o ./bin/efx-skills ./cmd/efx-skills/
   ```
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go build -o ./bin/efx-skills ./cmd/efx-skills/ && go vet ./internal/tui/...</automated>
  </verify>
  <done>
    - manageModel has height field, set by app.go and WindowSizeMsg handler
    - effectivePerPage() method computes available rows from terminal height minus chrome
    - All references to old perPage constant replaced with effectivePerPage() calls
    - paginator.PerPage updated dynamically in buildDisplayList
    - go build and go vet pass cleanly
    - Binary built at ./bin/efx-skills
  </done>
</task>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Task 2: Verify manage view fits terminal and collapse works correctly</name>
  <files>internal/tui/manage.go</files>
  <action>Human verifies the dynamic page sizing fix visually in the TUI.</action>
  <what-built>Dynamic page sizing in manage view -- rendered output now fits within terminal height, preventing bubbletea internal scroll offset issues when collapsing/expanding groups.</what-built>
  <how-to-verify>
    1. Run `./bin/efx-skills` and navigate to a provider's manage view
    2. Verify the title box "Manage Provider: ..." is visible at the top
    3. Press [enter] on several group headers to collapse them -- verify:
       - The title box stays visible at the top (never pushed off screen)
       - No blank space appears at the top or bottom
       - Cursor stays on the collapsed group header
    4. Press [enter] to expand groups back -- verify same: no jump, title stays visible
    5. Try collapsing ALL groups -- content should shrink to just group headers, title stays at top
    6. Resize your terminal window while in manage view -- the number of visible items should adjust
    7. Try a small terminal (e.g., 25 rows) -- should show fewer items but still be usable with no overflow
  </how-to-verify>
  <verify>Manual visual inspection per steps above</verify>
  <done>Title box stays visible after collapse/expand, content fits terminal, resize adjusts item count</done>
  <resume-signal>Type "approved" or describe any remaining layout issues</resume-signal>
</task>

</tasks>

<verification>
- `go build -o ./bin/efx-skills ./cmd/efx-skills/` succeeds
- `go vet ./internal/tui/...` reports no issues
- Manual testing confirms title box stays visible after collapse/expand
- Terminal resize dynamically adjusts visible items
</verification>

<success_criteria>
The manage view rendered output never exceeds terminal height. Collapsing and expanding groups keeps the title box visible at the top, with no stale scroll offset pushing content off screen. The number of visible items adapts to terminal size.
</success_criteria>

<output>
After completion, create `.planning/quick/9-fix-manage-view-scroll-offset-when-colla/9-SUMMARY.md`
</output>
