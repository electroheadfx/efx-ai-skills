# Codex Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Codex as a skills provider at `~/.codex/skills`, centralize provider definitions in one file, and release it as `0.2.1`.

**Architecture:** `internal/provider` becomes the single source of truth for provider names, default enabled state, and skills paths. TUI status/config code and legacy config defaults consume that catalog instead of duplicating provider lists. The user-facing version and docs are updated after behavior is covered by tests.

**Tech Stack:** Go 1.22, Cobra CLI, Bubble Tea TUI, standard library filesystem/JSON packages, Go test.

---

## File structure

- Modify `internal/provider/provider.go`: define reusable provider definitions, add Codex, keep existing `DetectAll`, `Get`, and `GetConfigured` APIs working.
- Create `internal/provider/provider_test.go`: verify the shared provider catalog includes Codex and resolves expected paths.
- Modify `internal/config/config.go`: build default provider config from `internal/provider` instead of a local hard-coded map.
- Create `internal/config/config_test.go`: verify default config includes Codex at `~/.codex/skills` and preserves enabled defaults.
- Modify `internal/tui/status.go`: use the shared provider catalog in `detectProviders`; update status version label to `v0.2.1`.
- Modify `internal/tui/config_test.go`: add status/provider detection tests proving Codex appears from the centralized catalog.
- Modify `cmd/efx-skills/main.go`: bump Cobra version from `0.2.0` to `0.2.1`.
- Modify `README.md`: add Codex to supported providers, directory structure, config example, and changelog.

---

### Task 1: Centralize provider catalog and add Codex

**Files:**
- Modify: `internal/provider/provider.go`
- Create: `internal/provider/provider_test.go`

- [ ] **Step 1: Write failing provider catalog tests**

Create `internal/provider/provider_test.go` with this content:

```go
package provider

import (
	"path/filepath"
	"testing"
)

func TestDefinitionsIncludeCodex(t *testing.T) {
	defs := Definitions()

	var found bool
	for _, def := range defs {
		if def.Name == "codex" {
			found = true
			if def.DefaultEnabled {
				t.Fatalf("codex DefaultEnabled = true, want false")
			}
			got := def.Path("/home/alice")
			want := filepath.Join("/home/alice", ".codex", "skills")
			if got != want {
				t.Fatalf("codex path = %q, want %q", got, want)
			}
		}
	}

	if !found {
		t.Fatalf("Definitions() did not include codex")
	}
}

func TestDefinitionsReturnCopy(t *testing.T) {
	defs := Definitions()
	defs[0].Name = "changed"

	fresh := Definitions()
	if fresh[0].Name == "changed" {
		t.Fatalf("Definitions() returned mutable shared backing array")
	}
}
```

- [ ] **Step 2: Run provider tests to verify they fail**

Run:

```bash
go test ./internal/provider
```

Expected: FAIL because `Definitions`, `DefaultEnabled`, and `Path` do not exist yet.

- [ ] **Step 3: Replace the local anonymous provider list with exported definitions**

Edit `internal/provider/provider.go` so the file content is:

```go
package provider

import (
	"os"
	"path/filepath"
)

// Provider represents an AI coding agent provider
type Provider struct {
	Name       string
	SkillsPath string
	Configured bool
	SkillCount int
}

// Definition describes a known provider and how to find its skills directory.
type Definition struct {
	Name           string
	DefaultEnabled bool
	Path           func(home string) string
}

var definitions = []Definition{
	{Name: "claude", DefaultEnabled: true, Path: func(h string) string { return filepath.Join(h, ".claude", "skills") }},
	{Name: "cursor", DefaultEnabled: true, Path: func(h string) string { return filepath.Join(h, ".cursor", "skills") }},
	{Name: "qoder", DefaultEnabled: true, Path: func(h string) string { return filepath.Join(h, ".qoder", "skills") }},
	{Name: "windsurf", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".windsurf", "skills") }},
	{Name: "copilot", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".copilot", "skills") }},
	{Name: "opencode", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".config", "opencode", "skills") }},
	{Name: "codex", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".codex", "skills") }},
}

// Definitions returns the known provider catalog.
func Definitions() []Definition {
	defs := make([]Definition, len(definitions))
	copy(defs, definitions)
	return defs
}

// DetectAll detects all providers on the system
func DetectAll() []Provider {
	home := os.Getenv("HOME")
	var providers []Provider

	for _, def := range definitions {
		p := Provider{
			Name:       def.Name,
			SkillsPath: def.Path(home),
		}

		if info, err := os.Stat(p.SkillsPath); err == nil && info.IsDir() {
			p.Configured = true

			if entries, err := os.ReadDir(p.SkillsPath); err == nil {
				for _, e := range entries {
					if e.Name() != ".DS_Store" {
						p.SkillCount++
					}
				}
			}
		}

		providers = append(providers, p)
	}

	return providers
}

// Get returns a specific provider by name
func Get(name string) *Provider {
	home := os.Getenv("HOME")

	for _, def := range definitions {
		if def.Name == name {
			p := &Provider{
				Name:       def.Name,
				SkillsPath: def.Path(home),
			}

			if info, err := os.Stat(p.SkillsPath); err == nil && info.IsDir() {
				p.Configured = true

				if entries, err := os.ReadDir(p.SkillsPath); err == nil {
					for _, e := range entries {
						if e.Name() != ".DS_Store" {
							p.SkillCount++
						}
					}
				}
			}

			return p
		}
	}

	return nil
}

// GetConfigured returns only configured providers
func GetConfigured() []Provider {
	all := DetectAll()
	var configured []Provider

	for _, p := range all {
		if p.Configured {
			configured = append(configured, p)
		}
	}

	return configured
}

// ListSkills returns skills installed for a provider
func (p *Provider) ListSkills() ([]string, error) {
	if !p.Configured {
		return nil, nil
	}

	entries, err := os.ReadDir(p.SkillsPath)
	if err != nil {
		return nil, err
	}

	var skills []string
	for _, e := range entries {
		if e.Name() != ".DS_Store" {
			skills = append(skills, e.Name())
		}
	}

	return skills, nil
}

// HasSkill checks if provider has a specific skill
func (p *Provider) HasSkill(skillName string) bool {
	skillPath := filepath.Join(p.SkillsPath, skillName)
	_, err := os.Stat(skillPath)
	return err == nil
}

// Configure creates the provider directory if it doesn't exist
func (p *Provider) Configure() error {
	if err := os.MkdirAll(p.SkillsPath, 0755); err != nil {
		return err
	}
	p.Configured = true
	return nil
}
```

- [ ] **Step 4: Run provider tests to verify they pass**

Run:

```bash
go test ./internal/provider
```

Expected: PASS.

- [ ] **Step 5: Commit provider catalog change**

Run:

```bash
git add internal/provider/provider.go internal/provider/provider_test.go
git commit -m "$(cat <<'EOF'
feat: centralize provider catalog

Add Codex to the shared provider definitions so provider paths are managed in one place.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Use provider catalog in config defaults

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing config tests**

Create `internal/config/config_test.go` with this content:

```go
package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfigIncludesCodexProvider(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	cfg := DefaultConfig()
	codex, ok := cfg.Providers["codex"]
	if !ok {
		t.Fatalf("DefaultConfig().Providers missing codex")
	}

	if codex.Enabled {
		t.Fatalf("codex Enabled = true, want false")
	}

	want := filepath.Join("/home/alice", ".codex", "skills")
	if codex.Path != want {
		t.Fatalf("codex Path = %q, want %q", codex.Path, want)
	}
}

func TestDefaultConfigProviderDefaults(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	cfg := DefaultConfig()
	tests := map[string]bool{
		"claude":   true,
		"cursor":   true,
		"qoder":    true,
		"windsurf": false,
		"copilot":  false,
		"opencode": false,
		"codex":    false,
	}

	for name, wantEnabled := range tests {
		provider, ok := cfg.Providers[name]
		if !ok {
			t.Fatalf("provider %q missing from DefaultConfig", name)
		}
		if provider.Enabled != wantEnabled {
			t.Fatalf("provider %q Enabled = %v, want %v", name, provider.Enabled, wantEnabled)
		}
	}
}
```

- [ ] **Step 2: Run config tests to verify they fail**

Run:

```bash
go test ./internal/config
```

Expected: FAIL because `DefaultConfig` does not include Codex yet.

- [ ] **Step 3: Build default config providers from the shared catalog**

Edit `internal/config/config.go`:

1. Add the provider package import and remove the now-unused `path/filepath` import.
2. Replace the hard-coded `Providers` map in `DefaultConfig` with a map built from `provider.Definitions()`.

The import block should become:

```go
import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lmarques/efx-skills/internal/provider"
)
```

Keep `path/filepath` because `ConfigPath` and `Save` still use it.

Replace `DefaultConfig` with:

```go
// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home := os.Getenv("HOME")
	providers := make(map[string]ProviderConfig)
	for _, def := range provider.Definitions() {
		providers[def.Name] = ProviderConfig{
			Enabled: def.DefaultEnabled,
			Path:    def.Path(home),
		}
	}

	return &Config{
		Registries: []Registry{
			{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true},
			{Name: "playbooks.com", URL: "https://playbooks.com/api/skills", Enabled: true},
		},
		Repos: []string{
			"yoanbernabeu/grepai-skills",
			"better-auth/skills",
			"awni/mlx-skills",
		},
		Providers: providers,
	}
}
```

- [ ] **Step 4: Run config tests to verify they pass**

Run:

```bash
go test ./internal/config
```

Expected: PASS.

- [ ] **Step 5: Commit config catalog usage**

Run:

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "$(cat <<'EOF'
refactor: derive config providers from catalog

Use the shared provider definitions for default config so provider metadata stays in sync.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Use provider catalog in TUI provider detection

**Files:**
- Modify: `internal/tui/status.go`
- Modify: `internal/tui/config_test.go`

- [ ] **Step 1: Write failing TUI provider detection tests**

Append these tests to `internal/tui/config_test.go` before the existing `contains` helper:

```go
func TestDetectProvidersIncludesCodexFromCatalog(t *testing.T) {
	home := setTestHome(t)
	if err := os.MkdirAll(filepath.Join(home, ".codex", "skills", "example-skill"), 0755); err != nil {
		t.Fatalf("create codex skill: %v", err)
	}

	providers := detectProviders()
	var codex *Provider
	for i := range providers {
		if providers[i].Name == "codex" {
			codex = &providers[i]
			break
		}
	}

	if codex == nil {
		t.Fatalf("detectProviders() missing codex")
	}
	if codex.Path != filepath.Join(home, ".codex", "skills") {
		t.Fatalf("codex Path = %q, want %q", codex.Path, filepath.Join(home, ".codex", "skills"))
	}
	if !codex.Configured {
		t.Fatalf("codex Configured = false, want true when directory exists and no config overrides providers")
	}
	if codex.SkillCount != 1 {
		t.Fatalf("codex SkillCount = %d, want 1", codex.SkillCount)
	}
	if !codex.Synced {
		t.Fatalf("codex Synced = false, want true")
	}
}

func TestDetectProvidersConfigCanEnableCodex(t *testing.T) {
	home := setTestHome(t)
	configDir := filepath.Join(home, ".config", "efx-skills")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	config := `{"enabled_providers":["codex"]}`
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(config), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	providers := detectProviders()
	var codex *Provider
	for i := range providers {
		if providers[i].Name == "codex" {
			codex = &providers[i]
			break
		}
	}

	if codex == nil {
		t.Fatalf("detectProviders() missing codex")
	}
	if !codex.Configured {
		t.Fatalf("codex Configured = false, want true from enabled_providers")
	}
	if codex.SkillCount != 0 {
		t.Fatalf("codex SkillCount = %d, want 0 without directory")
	}
	if codex.Synced {
		t.Fatalf("codex Synced = true, want false without directory")
	}
}
```

- [ ] **Step 2: Run the new TUI tests to verify they fail**

Run:

```bash
go test ./internal/tui -run 'TestDetectProviders.*Codex'
```

Expected: FAIL because `detectProviders` still uses a local list without Codex.

- [ ] **Step 3: Update status provider detection to use the shared catalog**

Edit `internal/tui/status.go`:

1. Add the provider import:

```go
	"github.com/lmarques/efx-skills/internal/provider"
```

2. Remove `path/filepath` from the import block if it is no longer used after the edit.

3. Replace the `providerDefs := ...` block and its loop with this code inside `detectProviders` after the config loading block:

```go
	var providers []Provider

	for _, def := range provider.Definitions() {
		path := def.Path(home)
		p := Provider{
			Name: def.Name,
			Path: path,
		}

		dirExists := false
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			dirExists = true
		}

		if enabledSet != nil {
			p.Configured = enabledSet[def.Name]
		} else {
			p.Configured = dirExists
		}

		if dirExists && p.Configured {
			if entries, err := os.ReadDir(path); err == nil {
				for _, e := range entries {
					if e.Name() != ".DS_Store" {
						p.SkillCount++
					}
				}
			}
			p.Synced = true
		}

		providers = append(providers, p)
	}

	return providers
```

The complete import block should include `encoding/json`, `fmt`, `os`, `strings`, Bubble Tea/Lipgloss/Bubbles imports, `internal/api`, and `internal/provider`.

- [ ] **Step 4: Run TUI Codex tests to verify they pass**

Run:

```bash
go test ./internal/tui -run 'TestDetectProviders.*Codex'
```

Expected: PASS.

- [ ] **Step 5: Run all TUI tests**

Run:

```bash
go test ./internal/tui
```

Expected: PASS.

- [ ] **Step 6: Commit TUI catalog usage**

Run:

```bash
git add internal/tui/status.go internal/tui/config_test.go
git commit -m "$(cat <<'EOF'
refactor: use provider catalog in TUI detection

Drive provider status and configuration lists from the shared catalog including Codex.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Bump user-facing version and docs

**Files:**
- Modify: `cmd/efx-skills/main.go`
- Modify: `internal/tui/status.go`
- Modify: `README.md`

- [ ] **Step 1: Write failing version test through CLI output**

Run:

```bash
go run ./cmd/efx-skills --version
```

Expected before the edit: output contains `0.2.0`, proving the CLI still needs a version bump.

- [ ] **Step 2: Bump Cobra version constant**

Edit `cmd/efx-skills/main.go` line 11:

```go
var version = "0.2.1"
```

- [ ] **Step 3: Bump TUI status version label**

Edit `internal/tui/status.go` line 209 so the rendered label is:

```go
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("v0.2.1 - Laurent Marques"))
```

- [ ] **Step 4: Update README provider docs**

Edit `README.md`:

1. In the features list near the top, change the Smart Linking bullet to:

```markdown
- 🔗 **Smart Linking** - Symlink skills to multiple providers (Claude, Cursor, Qoder, Windsurf, Copilot, Codex)
```

2. In the directory structure block, add Codex after OpenCode or near the other provider directories:

```markdown
~/.codex/skills/              # Symlinks to central storage
```

3. In Supported Providers, add this bullet:

```markdown
- **Codex** (`~/.codex/skills/`) - OpenAI Codex CLI
```

4. In the CLI/config example around `enabled_providers`, include Codex as an optional provider if the example currently lists provider names. The example should use the current `enabled_providers` shape if present; if the example still shows old `providers`, change that key to `enabled_providers` and include:

```json
"enabled_providers": [
  "claude",
  "cursor",
  "qoder",
  "codex"
]
```

5. Add a changelog entry above `v0.1.4`:

```markdown
### v0.2.1

- Added OpenAI Codex provider support at `~/.codex/skills/`
- Centralized provider definitions so future provider additions happen in one place
```

- [ ] **Step 5: Verify CLI version output**

Run:

```bash
go run ./cmd/efx-skills --version
```

Expected: output contains `0.2.1`.

- [ ] **Step 6: Verify docs mention Codex**

Run:

```bash
grep -n "Codex\|codex\|0.2.1" README.md cmd/efx-skills/main.go internal/tui/status.go
```

Expected: output shows Codex docs in `README.md`, version `0.2.1` in `cmd/efx-skills/main.go`, and `v0.2.1` in `internal/tui/status.go`.

- [ ] **Step 7: Commit version and docs**

Run:

```bash
git add cmd/efx-skills/main.go internal/tui/status.go README.md
git commit -m "$(cat <<'EOF'
chore: bump version for Codex provider

Release Codex provider support as v0.2.1 and document the new provider path.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Final verification and required build

**Files:**
- No source edits expected.
- Build output: `bin/efx-skills`

- [ ] **Step 1: Run all Go tests**

Run:

```bash
go test ./...
```

Expected: PASS for all packages.

- [ ] **Step 2: Run the required project build command**

Run:

```bash
go build -o ./bin/efx-skills ./cmd/efx-skills/
```

Expected: command exits successfully and updates `bin/efx-skills`.

- [ ] **Step 3: Verify built binary version**

Run:

```bash
./bin/efx-skills --version
```

Expected: output contains `0.2.1`.

- [ ] **Step 4: Verify no duplicate provider lists remain outside tests/docs**

Run:

```bash
grep -R "\.claude.*skills\|\.cursor.*skills\|\.qoder.*skills\|\.codex.*skills\|opencode.*skills" -n internal cmd --exclude='*_test.go'
```

Expected: provider path definitions appear in `internal/provider/provider.go` only. Version references may remain in `cmd/efx-skills/main.go` and `internal/tui/status.go`.

- [ ] **Step 5: Review git status**

Run:

```bash
git status --short
```

Expected: only `bin/efx-skills` may be modified by the required build. If it is tracked, include it in the final commit; if it is untracked or ignored, do not add it.

- [ ] **Step 6: Commit build artifact if tracked**

If `git status --short` shows `M bin/efx-skills`, run:

```bash
git add bin/efx-skills
git commit -m "$(cat <<'EOF'
build: update efx-skills binary

Rebuild the CLI after adding Codex provider support for v0.2.1.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

If `git status --short` does not show `M bin/efx-skills`, do not create a commit for this task.

---

## Self-review

- Spec coverage: Codex provider path, centralized provider definitions, version bump, README updates, tests, and required build are covered by Tasks 1-5.
- Placeholder scan: no TODO/TBD/fill-in-later placeholders remain.
- Type consistency: `provider.Definition`, `Definitions()`, `DefaultEnabled`, and `Path(home string)` are introduced in Task 1 and used consistently in Tasks 2-3.
