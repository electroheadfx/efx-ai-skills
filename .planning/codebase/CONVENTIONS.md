# Coding Conventions

**Analysis Date:** 2026-03-05

## Naming Patterns

**Files:**
- Use lowercase single-word names for Go source files: `client.go`, `store.go`, `styles.go`
- Use lowercase compound names without separators for multi-word files: `skillssh.go`, `playbooks.go`
- One file per logical concern within a package (e.g., `search.go`, `preview.go`, `manage.go`, `config.go` in the `tui` package)

**Functions:**
- Use PascalCase for exported functions: `SearchAll()`, `DetectAll()`, `NewStore()`
- Use camelCase for unexported functions: `searchSkills()`, `truncate()`, `extractGroup()`
- Constructor pattern: `New{Type}()` for exported constructors (`NewClient()`, `NewStore()`)
- Constructor pattern: `new{Type}()` for unexported constructors (`newSearchModel()`, `newStatusModel()`)
- Use `{Verb}{Noun}` pattern: `LoadProviders()`, `DetectAll()`, `SearchAll()`
- TUI model initializers: `new{View}Model()` pattern (e.g., `newSearchModel()`, `newConfigModel()`)

**Variables:**
- Use camelCase for all variables: `selectedIdx`, `skillsDir`, `allSkills`
- Short variable names in tight scopes: `b` for `strings.Builder`, `p` for provider, `s` for skill, `w` for width
- Full descriptive names for struct fields: `SkillsPath`, `SkillCount`, `Configured`
- Boolean fields use adjective names: `Configured`, `Synced`, `Enabled`, `Selected`, `Linked`

**Types:**
- PascalCase for exported types: `Client`, `Skill`, `Provider`, `Store`
- camelCase for unexported types: `model`, `searchModel`, `statusModel`, `viewState`
- Message types use `{noun}{action}Msg` suffix: `searchResultsMsg`, `installDoneMsg`, `providersLoadedMsg`
- Response types use `{Source}Response` suffix: `PlaybooksResponse`, `SkillsShResponse`

**Constants:**
- camelCase for unexported constants: `searchPerPage`, `perPage`
- PascalCase iota constants are unexported (they should be): `viewStatus`, `viewSearch`, etc.
- URL constants use `{source}BaseURL` pattern: `skillsShBaseURL`, `playbooksBaseURL`

## Code Style

**Formatting:**
- `go fmt` (standard Go formatter)
- Run via `make fmt`

**Linting:**
- `golangci-lint` configured in Makefile (`make lint`)
- No `.golangci.yml` config file present -- uses defaults

**Line Length:**
- No explicit limit enforced; lines up to ~130 characters observed

**Indentation:**
- Tabs (standard Go convention)

## Import Organization

**Order:**
1. Standard library packages (`fmt`, `os`, `encoding/json`, `net/http`, etc.)
2. External dependencies (`github.com/charmbracelet/*`, `github.com/spf13/cobra`)
3. Internal project packages (`github.com/lmarques/efx-skills/internal/*`)

**Grouping:**
- Imports grouped in a single `import ()` block with blank lines separating groups
- Example from `cmd/efx-skills/main.go`:
```go
import (
    "fmt"
    "os"

    "github.com/lmarques/efx-skills/internal/tui"
    "github.com/spf13/cobra"
)
```
- Example from `internal/tui/search.go`:
```go
import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/paginator"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/lmarques/efx-skills/internal/api"
    "github.com/lmarques/efx-skills/internal/skill"
)
```

**Path Aliases:**
- `tea` for `github.com/charmbracelet/bubbletea` -- used consistently across all TUI files

## Error Handling

**Patterns:**
- Return `error` as the last return value following Go convention
- Use `fmt.Errorf()` for descriptive error messages with context: `fmt.Errorf("invalid source format: %s (expected owner/repo)", source)`
- Use `%w` verb for error wrapping (used sparingly): `fmt.Errorf("failed to read skills directory: %w", err)` and `fmt.Errorf("%w: %s", err, string(output))`
- Silently swallow errors in non-critical paths (e.g., `err == nil` guard clauses that continue on error):
```go
// From client.go SearchAll() - errors from individual registries are ignored
skillsShResults, err := SearchSkillsSh(query, limit)
if err == nil {
    allSkills = append(allSkills, skillsShResults...)
}
```
- Use `os.IsNotExist(err)` check for graceful fallback to defaults when config files are missing
- In TUI, errors are stored on model structs (`m.err = msg.err`) and rendered in the View

**Error Message Types (TUI):**
- Dedicated message types per domain: `searchErrMsg`, `previewErrMsg`, `installErrMsg`, `errMsg`
- Each wraps a single `err error` field

## Logging

**Framework:** None -- no logging framework is used

**Patterns:**
- `fmt.Printf` / `fmt.Println` for CLI output in non-TUI commands (`RunList`, `RunInstall`)
- `fmt.Fprintln(os.Stderr, err)` for fatal errors in `main()`
- No structured logging, no log levels, no log files
- TUI views display errors inline via styled error rendering: `errorStyle.Render(fmt.Sprintf("Error: %v", m.err))`

## Comments

**When to Comment:**
- Every exported type and function has a single-line comment: `// Client is the base HTTP client for API calls`
- Comments follow Go convention: `// FunctionName does X`
- Inline comments for non-obvious logic sections within functions
- Section separator comments within `View()` methods: `// Title`, `// Help`, `// Results header`

**JSDoc/TSDoc:** Not applicable (Go project)

**Comment Style:**
- Single-line `//` comments only -- no block `/* */` comments
- No TODO tracker or convention beyond `// TODO:` prefix

## Function Design

**Size:**
- Most functions are 10-40 lines
- `View()` methods are the largest (50-100 lines) due to string building for TUI rendering
- `Update()` methods can be 80-100 lines due to message handling switch statements

**Parameters:**
- Minimal parameter counts (0-3 parameters typical)
- Use struct receivers for methods: `(c *Client)`, `(s *Store)`, `(m manageModel)`
- Value receivers for `Init()`, `View()`, `Update()` on TUI models (Bubble Tea convention)
- Pointer receivers for mutation methods: `(m *manageModel)` for `buildDisplayList()`, `toggleGroup()`

**Return Values:**
- Follow Go idiom: `(result, error)` pair
- Nil slice returns for empty results (not empty slice)
- Boolean returns for existence checks: `IsInstalled()`, `HasSkill()`

## Module Design

**Exports:**
- Export types that are used across packages: `Client`, `Skill`, `Store`, `Provider`, `Config`
- Export functions that serve as public API: `SearchAll()`, `DetectAll()`, `Run()`
- Keep TUI model types unexported: `model`, `searchModel`, `statusModel`
- Export entry-point functions from TUI: `Run()`, `RunSearch()`, `RunPreview()`

**Barrel Files:** Not used (Go does not use barrel files)

**Package Organization:**
- Each package in `internal/` has a focused responsibility
- `internal/api/` - External API clients (one file per registry + shared client)
- `internal/config/` - Configuration loading/saving
- `internal/provider/` - Provider detection and management
- `internal/skill/` - Skill storage, installation, and lock file management
- `internal/tui/` - All TUI views and styles

## TUI Architecture Pattern (Bubble Tea / Elm Architecture)

**Model-Update-View (MVU) pattern is mandatory for all TUI views:**

```go
// Model: holds all state
type searchModel struct {
    input       textinput.Model
    results     []Skill
    selectedIdx int
    loading     bool
    err         error
}

// Init: returns initial command
func (m searchModel) Init() tea.Cmd { ... }

// Update: handles messages, returns new model + command
func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) { ... }

// View: renders the model to a string
func (m searchModel) View() string { ... }
```

**Message Pattern:**
- Define message types as unexported structs with `Msg` suffix
- Use closures returning `tea.Msg` for async operations:
```go
return m, func() tea.Msg {
    results, err := searchSkills(query)
    if err != nil {
        return searchErrMsg{err: err}
    }
    return searchResultsMsg{results: results}
}
```

**View Rendering Pattern:**
- Use `strings.Builder` for building view output
- Apply lipgloss styles from `styles.go` for consistent appearance
- Include help text at the bottom of every view
- Dynamic width handling: check `m.width` with fallback to 80

## Type Aliasing

- `type Skill = api.Skill` in `internal/tui/search.go` to avoid verbose package-qualified references
- `Provider` type is redefined in `internal/tui/status.go` (does NOT use `internal/provider` package)

## Configuration Access

- Use `os.Getenv("HOME")` directly -- no abstraction for home directory
- Config file path: `~/.config/efx-skills/config.json`
- Graceful fallback to defaults when config is missing

---

*Convention analysis: 2026-03-05*
