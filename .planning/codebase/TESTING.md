# Testing Patterns

**Analysis Date:** 2026-03-05

## Test Framework

**Runner:**
- Go standard `testing` package (built-in)
- No test files exist in the codebase

**Assertion Library:**
- None configured (no testify, gomega, or other assertion libraries in `go.mod`)

**Run Commands:**
```bash
make test              # Run all tests (go test -v ./...)
go test ./...          # Run all tests directly
go test -v ./...       # Run with verbose output
go test -cover ./...   # Run with coverage report
```

## Test File Organization

**Location:**
- No test files exist anywhere in the codebase. Zero `*_test.go` files found.

**Expected Go Convention (when adding tests):**
- Co-locate test files with source files in the same package directory
- Name test files as `{source}_test.go` (e.g., `client_test.go`, `store_test.go`)

**Expected Structure:**
```
internal/
├── api/
│   ├── client.go
│   ├── client_test.go          # Tests for Client, SearchAll, FetchSkillContent
│   ├── playbooks.go
│   ├── playbooks_test.go       # Tests for SearchPlaybooks, GetPlaybooksTrending
│   ├── skillssh.go
│   └── skillssh_test.go        # Tests for SearchSkillsSh, GetSkillsShTrending
├── config/
│   ├── config.go
│   └── config_test.go          # Tests for Load, Save, AddRepo, RemoveRepo
├── provider/
│   ├── provider.go
│   └── provider_test.go        # Tests for DetectAll, Get, ListSkills
├── skill/
│   ├── store.go
│   └── store_test.go           # Tests for Install, LinkToProvider, ReadLockFile
└── tui/
    └── (TUI tests are complex; see guidance below)
```

## Test Structure

**No existing tests to reference.** When adding tests, follow standard Go patterns:

**Suite Organization:**
```go
package api

import (
    "testing"
)

func TestSearchAll(t *testing.T) {
    // Arrange
    // Act
    // Assert
}

func TestSearchAll_EmptyQuery(t *testing.T) {
    // ...
}

func TestSearchAll_DeduplicatesByName(t *testing.T) {
    // ...
}
```

**Use table-driven tests for functions with multiple input combinations:**
```go
func TestTruncate(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        maxLen   int
        expected string
    }{
        {"short string", "hello", 10, "hello"},
        {"exact length", "hello", 5, "hello"},
        {"truncated", "hello world", 8, "hello..."},
        {"very short max", "hello", 2, "he"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := truncate(tt.input, tt.maxLen)
            if got != tt.expected {
                t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
            }
        })
    }
}
```

## Mocking

**Framework:** None configured

**Recommended Approach for This Codebase:**

The codebase does NOT use interfaces for its core types (`Client`, `Store`, `Provider`), which makes mocking difficult. To enable testing:

1. **HTTP calls** (`internal/api/`): Use `net/http/httptest` for HTTP server mocking:
```go
func TestSearchSkillsSh(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(SkillsShResponse{
            Skills: []SkillsShSkill{{ID: "test", Name: "test-skill"}},
        })
    }))
    defer server.Close()

    // Replace base URL (requires refactoring to accept URL parameter)
    // ...
}
```

2. **Filesystem operations** (`internal/skill/`, `internal/provider/`): Use `t.TempDir()`:
```go
func TestStoreInstallDirect(t *testing.T) {
    tmpDir := t.TempDir()
    store := &Store{
        BaseDir:  filepath.Join(tmpDir, "skills"),
        LockFile: filepath.Join(tmpDir, ".skill-lock.json"),
    }
    // ...
}
```

3. **TUI models**: Test `Update()` and `View()` by constructing models directly and sending messages:
```go
func TestSearchModelUpdate_SearchResults(t *testing.T) {
    m := newSearchModel()
    msg := searchResultsMsg{results: []Skill{{Name: "test"}}}
    updated, _ := m.Update(msg)
    if len(updated.results) != 1 {
        t.Errorf("expected 1 result, got %d", len(updated.results))
    }
}
```

**What to Mock:**
- External HTTP APIs (skills.sh, playbooks.com, GitHub raw content)
- Filesystem operations when testing provider detection and skill installation
- `exec.Command` calls (npx invocations in `Store.installViaSkills`)

**What NOT to Mock:**
- Pure logic functions: `truncate()`, `extractGroup()`, `truncateStr()`
- Config struct methods: `AddRepo()`, `RemoveRepo()`, `EnableRegistry()`
- TUI model state transitions (test these directly with message passing)

## Fixtures and Factories

**Test Data:** None exists

**Recommended fixture location:**
- `internal/api/testdata/` - Sample JSON API responses
- `internal/skill/testdata/` - Sample SKILL.md files and lock files
- `internal/config/testdata/` - Sample config.json files

**Example fixture pattern:**
```go
// Load test fixture
data, err := os.ReadFile("testdata/skillssh_response.json")
if err != nil {
    t.Fatal(err)
}
```

## Coverage

**Requirements:** None enforced -- no coverage thresholds configured

**View Coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out          # Browser view
go tool cover -func=coverage.out          # Summary
```

## Test Types

**Unit Tests:**
- Not present. Priority targets for unit tests:
  - `internal/api/client.go` - JSON parsing, URL construction, error handling
  - `internal/config/config.go` - Load/Save/AddRepo/RemoveRepo (pure logic + filesystem)
  - `internal/skill/store.go` - Lock file read/write, install logic
  - `internal/tui/search.go` - `truncate()` helper
  - `internal/tui/config.go` - `truncateStr()` helper
  - `internal/tui/manage.go` - `extractGroup()`, `buildDisplayList()` logic

**Integration Tests:**
- Not present. Would test:
  - Full search flow (API call -> parse -> deduplicate)
  - Skill installation flow (download -> write -> link)
  - Config persistence (save -> load roundtrip)

**E2E Tests:**
- Not present. Could use Bubble Tea's `teatest` package for TUI integration testing:
```go
import "github.com/charmbracelet/x/exp/teatest"

func TestApp(t *testing.T) {
    m := initialModel()
    tm := teatest.NewTestModel(t, m)
    // Send key events, assert view output
}
```

## Testability Concerns

**Current Blockers:**
1. **Hard-coded base URLs** in `internal/api/skillssh.go` and `internal/api/playbooks.go` (constants `skillsShBaseURL`, `playbooksBaseURL`) -- cannot be overridden in tests
2. **Direct `os.Getenv("HOME")` calls** scattered across packages instead of being injected -- makes testing with temp directories difficult
3. **No interfaces** for `Client`, `Store`, or provider detection -- prevents dependency injection
4. **`exec.Command("npx", ...)` in `Store.installViaSkills`** -- hard to test without actually running npx
5. **Duplicate `Provider` type** in `internal/tui/status.go` vs `internal/provider/provider.go` -- confusion about which to test

**Recommended Refactoring for Testability:**
- Accept base URLs as parameters or struct fields (not constants)
- Accept `home` directory as a parameter to constructors
- Define interfaces for external dependencies (`HTTPClient`, `FileSystem`)
- Use `internal/provider.Provider` in TUI instead of redefining it

## CI/CD Testing

- No CI/CD pipeline detected (no `.github/workflows/`, no `.gitlab-ci.yml`, no `Jenkinsfile`)
- Tests would run via `make test` if any existed
- No pre-commit hooks configured

---

*Testing analysis: 2026-03-05*
