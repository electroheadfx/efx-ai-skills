# Technology Stack

**Analysis Date:** 2026-03-05

## Languages

**Primary:**
- Go 1.22 (minimum; `go.mod` specifies `go 1.22`) - All application code

**Secondary:**
- None - Pure Go project, no secondary languages

## Runtime

**Environment:**
- Go 1.22+ (host machine runtime `go1.25.6 darwin/arm64` detected)
- Compiles to single static binary

**Package Manager:**
- Go modules (`go.mod` / `go.sum`)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- `github.com/charmbracelet/bubbletea` v1.2.4 - TUI framework (Elm Architecture)
- `github.com/charmbracelet/bubbles` v0.20.0 - TUI component library (textinput, viewport, paginator)
- `github.com/charmbracelet/lipgloss` v1.0.0 - Terminal styling/layout
- `github.com/charmbracelet/glamour` v0.8.0 - Markdown terminal rendering (Dracula theme)
- `github.com/spf13/cobra` v1.8.1 - CLI command framework

**Testing:**
- Go standard `testing` package (via `go test -v ./...`)
- No test files exist in the codebase currently

**Build/Dev:**
- `make` via `Makefile` - Build, install, test, lint, fmt, cross-compile
- `golangci-lint` - Linting (configured in `Makefile` `lint` target)

## Key Dependencies

**Critical (direct):**
- `github.com/charmbracelet/bubbletea` v1.2.4 - Entire TUI application model (Init/Update/View loop)
- `github.com/charmbracelet/bubbles` v0.20.0 - Text input, viewport scrolling, pagination widgets
- `github.com/charmbracelet/lipgloss` v1.0.0 - All terminal styling, colors, borders, layout
- `github.com/charmbracelet/glamour` v0.8.0 - Markdown preview rendering with syntax highlighting
- `github.com/spf13/cobra` v1.8.1 - CLI subcommand routing (search, status, preview, install, list, sync, config)

**Infrastructure (indirect/transitive):**
- `github.com/alecthomas/chroma/v2` v2.14.0 - Syntax highlighting (used by glamour)
- `github.com/microcosm-cc/bluemonday` v1.0.27 - HTML sanitization (used by glamour)
- `github.com/yuin/goldmark` v1.7.4 - Markdown parsing (used by glamour)
- `github.com/yuin/goldmark-emoji` v1.0.3 - Emoji support in markdown
- `github.com/atotto/clipboard` v0.1.4 - Clipboard access (bubbles dependency)
- `github.com/muesli/termenv` v0.15.3 - Terminal environment detection
- `golang.org/x/net` v0.27.0, `golang.org/x/sys` v0.27.0, `golang.org/x/term` v0.22.0 - Standard extended libraries

## Configuration

**Application Config:**
- Config file: `~/.config/efx-skills/config.json` (JSON format)
- Config is optional; defaults are hardcoded in `internal/config/config.go` via `DefaultConfig()`
- Config persists: registries (enabled/disabled), custom GitHub repos, enabled providers
- No environment variables required - uses `$HOME` only for path resolution

**Skill Storage:**
- Central storage: `~/.agents/skills/` (skill SKILL.md files stored here)
- Lock file: `~/.agents/.skill-lock.json` (tracks install metadata, version 3 format)
- Provider linking: Symlinks from provider skills directories to central storage

**Build:**
- `Makefile` at project root - primary build orchestration
- Version injected via `-ldflags "-X main.version=$(VERSION)"` at build time
- Current version: `0.1.2` in Makefile, `0.1.3` hardcoded in `cmd/efx-skills/main.go`

## Build Targets

**Key Makefile targets:**
```bash
make build          # Build binary to bin/efx-skills
make install        # Build + copy to /usr/local/bin/
make install-local  # Build + copy to ~/bin/
make test           # go test -v ./...
make fmt            # go fmt ./...
make lint           # golangci-lint run
make tidy           # go mod tidy
make dev            # go run ./cmd/efx-skills
make build-all      # Cross-compile for darwin/linux/windows (amd64+arm64)
```

**Cross-compilation targets (in `build-all`):**
- `darwin/arm64`, `darwin/amd64`
- `linux/amd64`
- `windows/amd64`

## Platform Requirements

**Development:**
- Go 1.22+
- Make
- `golangci-lint` (for `make lint`)
- Terminal with Unicode support (TUI uses Unicode box-drawing, checkboxes, icons)

**Production:**
- Single static binary - no runtime dependencies
- macOS or Linux (primary targets)
- Windows supported via cross-compilation
- Terminal with Unicode support recommended

## Release Artifacts

**Location:** `bin/release/`
- Pre-built zip archives for 6 platform/arch combos:
  - `efx-skills-darwin-arm64.zip`, `efx-skills-darwin-amd64.zip`
  - `efx-skills-linux-amd64.zip`, `efx-skills-linux-arm64.zip`
  - `efx-skills-windows-amd64.zip`, `efx-skills-windows-arm64.zip`
- Release notes: `RELEASE/RELEASE_v0.1.0.md`

---

*Stack analysis: 2026-03-05*
