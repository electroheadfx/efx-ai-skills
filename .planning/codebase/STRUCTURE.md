# Codebase Structure

**Analysis Date:** 2026-03-05

## Directory Layout

```
efx-skill-management/
├── cmd/
│   └── efx-skills/
│       └── main.go              # CLI entry point, Cobra command definitions
├── internal/
│   ├── api/
│   │   ├── client.go            # Base HTTP client, unified Skill type, SearchAll(), FetchSkillContent()
│   │   ├── skillssh.go          # skills.sh registry adapter (search, trending)
│   │   └── playbooks.go         # playbooks.com registry adapter (search, trending)
│   ├── config/
│   │   └── config.go            # Config struct, load/save, defaults (partially unused by TUI)
│   ├── provider/
│   │   └── provider.go          # Provider detection and management (partially unused by TUI)
│   ├── skill/
│   │   └── store.go             # Skill install, storage, symlinks, lock file management
│   └── tui/
│       ├── app.go               # Root Bubble Tea model, view state machine, Run*() entry functions
│       ├── config.go            # Config view: registries, repos, providers (has own config types)
│       ├── manage.go            # Provider skill management view: toggle, group, apply symlinks
│       ├── preview.go           # Skill preview view: fetch SKILL.md, render markdown
│       ├── search.go            # Search view: text input, paginated results, install action
│       ├── status.go            # Status/home view: provider list, skill counts (has own Provider type)
│       └── styles.go            # Shared lipgloss styles, colors, helper render functions
├── bin/
│   ├── efx-skills               # Compiled binary (build output)
│   └── release/                 # Cross-compiled release zips (darwin/linux/windows, amd64/arm64)
├── .bin/
│   └── efx-skills               # Alternative binary location
├── RELEASE/
│   └── RELEASE_v0.1.0.md        # Release notes
├── public/
│   ├── img/                     # README screenshots
│   └── img-v0.1.4/              # v0.1.4 screenshots
├── .claude/                     # Claude Code configuration (hooks, commands, agents)
├── .planning/                   # GSD planning documents
├── go.mod                       # Go module definition (github.com/lmarques/efx-skills)
├── go.sum                       # Dependency checksums
├── Makefile                     # Build, install, test, cross-compile, dev commands
├── ARCHITECTURE.md              # Original architecture doc (describes earlier bash-based design)
├── Skills-origin-Specs.md       # Feature specs for v0.2.0 (origin tracking, browser open)
├── README.md                    # Project README with screenshots
├── LICENSE                      # License file
└── .gitignore                   # Git ignore rules
```

## Directory Purposes

**`cmd/efx-skills/`:**
- Purpose: Application entry point
- Contains: Single `main.go` with Cobra command tree
- Key files: `cmd/efx-skills/main.go`

**`internal/api/`:**
- Purpose: External API clients for skill registries
- Contains: HTTP client wrapper, registry-specific adapters, unified Skill data type
- Key files: `internal/api/client.go` (core types + SearchAll), `internal/api/skillssh.go`, `internal/api/playbooks.go`

**`internal/config/`:**
- Purpose: Application configuration management
- Contains: Config struct with registries, repos, providers; load/save to JSON
- Key files: `internal/config/config.go`
- Note: Partially superseded by `internal/tui/config.go` which has its own config types

**`internal/provider/`:**
- Purpose: AI provider detection and management
- Contains: Provider struct, known provider definitions, detection, skill listing
- Key files: `internal/provider/provider.go`
- Note: Partially superseded by `internal/tui/status.go` which has its own Provider type

**`internal/skill/`:**
- Purpose: Central skill storage, installation, and symlink management
- Contains: Store struct with install (npx + direct), link/unlink, lock file read/write
- Key files: `internal/skill/store.go`

**`internal/tui/`:**
- Purpose: All interactive terminal UI views
- Contains: Root app model + 5 sub-models (status, search, preview, manage, config) + shared styles
- Key files: `internal/tui/app.go` (state machine), `internal/tui/styles.go` (shared styles)

**`bin/`:**
- Purpose: Build output directory
- Contains: Compiled binary, release archives
- Generated: Yes (via `make build` / `make build-all`)
- Committed: Partially (release zips committed, main binary may vary)

**`public/`:**
- Purpose: Static assets for documentation
- Contains: Screenshots used in README
- Generated: No

## Key File Locations

**Entry Points:**
- `cmd/efx-skills/main.go`: CLI entry, Cobra commands, version string
- `internal/tui/app.go`: TUI entry via `Run()`, `RunSearch()`, `RunPreview()`, etc.

**Configuration:**
- `Makefile`: Build targets, version number, cross-compilation
- `go.mod`: Module path, Go version, dependencies
- Runtime config: `~/.config/efx-skills/config.json` (not in repo)

**Core Logic:**
- `internal/api/client.go`: Unified Skill type, multi-registry search, GitHub content fetch
- `internal/skill/store.go`: Install logic, central storage, provider linking, lock file
- `internal/tui/status.go`: Provider detection (the version actually used at runtime)
- `internal/tui/manage.go`: Skill grouping, toggle, apply symlink changes

**Testing:**
- No test files exist in the codebase

**Specifications:**
- `Skills-origin-Specs.md`: Feature specs for the v0.2.0 branch (origin tracking, browser open, config enrichment)
- `ARCHITECTURE.md`: Original architecture doc (describes earlier bash+gum design, not current Go implementation)

## Naming Conventions

**Files:**
- Lowercase, single-word Go files: `client.go`, `store.go`, `preview.go`
- One file per concern within a package (e.g., `skillssh.go` for skills.sh, `playbooks.go` for playbooks.com)
- No `_test.go` files exist

**Directories:**
- Lowercase, singular nouns: `api`, `config`, `provider`, `skill`, `tui`
- Standard Go layout: `cmd/` for binaries, `internal/` for private packages

**Go Types:**
- Exported types: PascalCase (`Skill`, `Store`, `Provider`, `Config`, `LockFile`)
- Unexported types: camelCase (`model`, `searchModel`, `statusModel`, `viewState`)
- Message types: camelCase with `Msg` suffix (`searchResultsMsg`, `openPreviewMsg`, `installDoneMsg`)

**Functions:**
- Exported: PascalCase (`SearchAll`, `DetectAll`, `NewStore`, `FetchSkillContent`)
- Unexported: camelCase (`searchSkills`, `detectProviders`, `loadSkillsForProvider`, `applySkillChanges`)
- Constructor pattern: `New*` for exported (`NewClient`, `NewStore`), `new*` for unexported (`newSearchModel`, `newStatusModel`)

## Where to Add New Code

**New TUI View:**
- Create: `internal/tui/{viewname}.go`
- Add viewState constant in `internal/tui/app.go` (in the `viewState` iota block)
- Add sub-model field to `model` struct in `internal/tui/app.go`
- Add message type for view transitions (e.g., `openViewNameMsg`)
- Handle in `Update()` and `View()` switch statements in `internal/tui/app.go`
- Follow pattern: `{viewname}Model` struct with `Init()`, `Update()`, `View()` methods

**New API Registry:**
- Create: `internal/api/{registryname}.go`
- Define response types and `Search{RegistryName}()` function
- Map results to `[]Skill` (unified type in `internal/api/client.go`)
- Add call in `SearchAll()` in `internal/api/client.go`

**New CLI Subcommand:**
- Add `cobra.Command` in `cmd/efx-skills/main.go`
- Add corresponding `Run{Command}()` function in `internal/tui/app.go`
- Register with `rootCmd.AddCommand()`

**New Provider:**
- Add entry in `providerDefs` slice in `internal/tui/status.go:detectProviders()`
- Add entry in `KnownProviders` slice in `internal/provider/provider.go`
- Add entry in `DefaultConfig().Providers` in `internal/config/config.go`

**Shared Styles:**
- Add to `internal/tui/styles.go`
- Follow existing pattern: package-level `var` with lipgloss style chain

**Utilities:**
- Place in the package where they are used (no shared `utils` package exists)
- Helper functions like `truncate()`, `truncateStr()` currently live in TUI files

## Special Directories

**`bin/`:**
- Purpose: Build output and release artifacts
- Generated: Yes (via Makefile)
- Committed: Release zips are committed in `bin/release/`

**`.bin/`:**
- Purpose: Alternative binary location (likely for local dev installs)
- Generated: Yes
- Committed: Possibly

**`RELEASE/`:**
- Purpose: Release documentation
- Generated: No (manual)
- Committed: Yes

**`.claude/`:**
- Purpose: Claude Code configuration (hooks, commands, agents, GSD workflow)
- Generated: No
- Committed: Yes (part of dev tooling)

**`.planning/`:**
- Purpose: GSD planning and codebase analysis documents
- Generated: By GSD workflow
- Committed: Yes

---

*Structure analysis: 2026-03-05*
