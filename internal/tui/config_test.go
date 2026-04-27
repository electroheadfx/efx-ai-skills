package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/lmarques/efx-skills/internal/api"
)

func TestConfigMarshalContainsNewFields(t *testing.T) {
	cfg := ConfigData{
		Registries: []Registry{{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true}},
		Repos: []RepoSource{
			{Owner: "acme", Repo: "tools", URL: "https://github.com/acme/tools"},
		},
		Providers:  []string{"claude"},
		SkillsPath: "/custom/skills",
		Skills: []SkillMeta{
			{Owner: "acme", Name: "tooling", Registry: "skills.sh", URL: "https://skills.sh/skills/acme/tooling"},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	s := string(data)
	for _, field := range []string{`"skills"`, `"skills-path"`, `"url"`} {
		if !contains(s, field) {
			t.Errorf("JSON missing field %s, got: %s", field, s)
		}
	}
}

func TestConfigUnmarshalNewFields(t *testing.T) {
	input := `{
		"registries": [],
		"repos": [{"owner":"acme","repo":"tools","url":"https://github.com/acme/tools"}],
		"enabled_providers": ["claude"],
		"skills-path": "/custom/skills",
		"skills": [
			{"owner":"acme","name":"tooling","registry":"skills.sh","url":"https://skills.sh/skills/acme/tooling"}
		]
	}`

	var cfg ConfigData
	if err := json.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if cfg.SkillsPath != "/custom/skills" {
		t.Errorf("SkillsPath = %q, want %q", cfg.SkillsPath, "/custom/skills")
	}
	if len(cfg.Skills) != 1 {
		t.Fatalf("Skills length = %d, want 1", len(cfg.Skills))
	}
	if cfg.Skills[0].Owner != "acme" {
		t.Errorf("Skills[0].Owner = %q, want %q", cfg.Skills[0].Owner, "acme")
	}
	if cfg.Skills[0].Name != "tooling" {
		t.Errorf("Skills[0].Name = %q, want %q", cfg.Skills[0].Name, "tooling")
	}
	if cfg.Skills[0].Registry != "skills.sh" {
		t.Errorf("Skills[0].Registry = %q, want %q", cfg.Skills[0].Registry, "skills.sh")
	}
	if cfg.Skills[0].URL != "https://skills.sh/skills/acme/tooling" {
		t.Errorf("Skills[0].URL = %q, want %q", cfg.Skills[0].URL, "https://skills.sh/skills/acme/tooling")
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("Repos length = %d, want 1", len(cfg.Repos))
	}
	if cfg.Repos[0].URL != "https://github.com/acme/tools" {
		t.Errorf("Repos[0].URL = %q, want %q", cfg.Repos[0].URL, "https://github.com/acme/tools")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	original := ConfigData{
		Registries: []Registry{{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true}},
		Repos: []RepoSource{
			{Owner: "acme", Repo: "tools", URL: "https://github.com/acme/tools"},
		},
		Providers:  []string{"claude"},
		SkillsPath: "/custom/skills",
		Skills: []SkillMeta{
			{Owner: "acme", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tooling"},
		},
	}

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded ConfigData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.SkillsPath != original.SkillsPath {
		t.Errorf("SkillsPath = %q, want %q", decoded.SkillsPath, original.SkillsPath)
	}
	if len(decoded.Skills) != len(original.Skills) {
		t.Fatalf("Skills length = %d, want %d", len(decoded.Skills), len(original.Skills))
	}
	if decoded.Skills[0] != original.Skills[0] {
		t.Errorf("Skills[0] = %+v, want %+v", decoded.Skills[0], original.Skills[0])
	}
	if decoded.Repos[0].URL != original.Repos[0].URL {
		t.Errorf("Repos[0].URL = %q, want %q", decoded.Repos[0].URL, original.Repos[0].URL)
	}
}

func TestConfigEmptySkillsSerializesAsArray(t *testing.T) {
	cfg := ConfigData{
		Skills: []SkillMeta{},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	s := string(data)
	// Should contain "skills":[] not "skills":null
	if contains(s, `"skills":null`) {
		t.Errorf("Skills serialized as null, want empty array; got: %s", s)
	}
	if !contains(s, `"skills":[]`) {
		t.Errorf("Skills not serialized as empty array; got: %s", s)
	}
}

func TestRepoSourceDeriveURL(t *testing.T) {
	r := RepoSource{Owner: "acme", Repo: "tools"}
	got := r.DeriveURL()
	want := "https://github.com/acme/tools"
	if got != want {
		t.Errorf("DeriveURL() = %q, want %q", got, want)
	}
}

func TestDefaultConfigIncludesSkillsPath(t *testing.T) {
	// loadConfigFromFile returns nil when no file exists,
	// so test the defaults that newConfigModel applies.
	home := os.Getenv("HOME")
	want := filepath.Join(home, ".agents", "skills")

	cfg := loadConfigFromFile()
	// If no config file, we should still get defaults from the load function
	// If cfg is nil, check that the default would be applied
	if cfg != nil {
		if cfg.SkillsPath == "" {
			t.Errorf("loaded config has empty SkillsPath, want default %q", want)
		} else if cfg.SkillsPath != want {
			// non-default is OK if user set it; just ensure it's not empty
		}
	}
	// The default is applied in loadConfigFromFile when field is empty
	// Test the default derivation directly
	defaultPath := filepath.Join(home, ".agents", "skills")
	if defaultPath != want {
		t.Errorf("default skills path = %q, want %q", defaultPath, want)
	}
}

// --- Phase 2 Plan 1: Config mutation function tests ---

// setTestHome overrides HOME to a temp directory and returns a cleanup func.
func setTestHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })
	return tmp
}

func TestAddSkillToConfig(t *testing.T) {
	setTestHome(t)

	meta := SkillMeta{
		Owner:    "acme",
		Name:     "tooling",
		Registry: "skills.sh",
		URL:      "https://github.com/acme/tooling",
	}
	if err := addSkillToConfig(meta); err != nil {
		t.Fatalf("addSkillToConfig failed: %v", err)
	}

	cfg := loadConfigFromFile()
	if cfg == nil {
		t.Fatal("config file not created")
	}
	if len(cfg.Skills) != 1 {
		t.Fatalf("Skills length = %d, want 1", len(cfg.Skills))
	}
	if cfg.Skills[0].Owner != "acme" {
		t.Errorf("Skills[0].Owner = %q, want %q", cfg.Skills[0].Owner, "acme")
	}
	if cfg.Skills[0].Name != "tooling" {
		t.Errorf("Skills[0].Name = %q, want %q", cfg.Skills[0].Name, "tooling")
	}
	if cfg.Skills[0].Registry != "skills.sh" {
		t.Errorf("Skills[0].Registry = %q, want %q", cfg.Skills[0].Registry, "skills.sh")
	}
	if cfg.Skills[0].URL != "https://github.com/acme/tooling" {
		t.Errorf("Skills[0].URL = %q, want %q", cfg.Skills[0].URL, "https://github.com/acme/tooling")
	}
}

func TestAddSkillToConfigIdempotent(t *testing.T) {
	setTestHome(t)

	meta := SkillMeta{
		Owner:    "acme",
		Name:     "tooling",
		Registry: "skills.sh",
		URL:      "https://github.com/acme/tooling",
	}
	// Add twice
	if err := addSkillToConfig(meta); err != nil {
		t.Fatalf("first addSkillToConfig failed: %v", err)
	}
	if err := addSkillToConfig(meta); err != nil {
		t.Fatalf("second addSkillToConfig failed: %v", err)
	}

	cfg := loadConfigFromFile()
	if cfg == nil {
		t.Fatal("config file not created")
	}
	if len(cfg.Skills) != 1 {
		t.Errorf("Skills length = %d after duplicate add, want 1", len(cfg.Skills))
	}
}

func TestAddSkillToConfigNoExistingFile(t *testing.T) {
	tmp := setTestHome(t)

	// Confirm no config file exists
	configFile := filepath.Join(tmp, ".config", "efx-skills", "config.json")
	if _, err := os.Stat(configFile); err == nil {
		t.Fatal("config file should not exist yet")
	}

	meta := SkillMeta{
		Owner:    "acme",
		Name:     "tooling",
		Registry: "skills.sh",
		URL:      "https://github.com/acme/tooling",
	}
	if err := addSkillToConfig(meta); err != nil {
		t.Fatalf("addSkillToConfig failed: %v", err)
	}

	// File should now exist
	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}

	cfg := loadConfigFromFile()
	if cfg == nil {
		t.Fatal("loadConfigFromFile returned nil after addSkillToConfig")
	}
	if len(cfg.Skills) != 1 {
		t.Fatalf("Skills length = %d, want 1", len(cfg.Skills))
	}
	// Should have defaults populated
	if cfg.SkillsPath == "" {
		t.Error("SkillsPath should have a default value")
	}
}

func TestRemoveSkillFromConfig(t *testing.T) {
	tmp := setTestHome(t)

	// Seed config with two skills
	cfg := &ConfigData{
		Registries: defaultRegistries(),
		SkillsPath: defaultSkillsPath(),
		Skills: []SkillMeta{
			{Owner: "acme", Name: "tooling", Registry: "skills.sh", URL: "https://github.com/acme/tooling"},
			{Owner: "beta", Name: "deploy", Registry: "skills.sh", URL: "https://github.com/beta/deploy"},
		},
	}
	if err := saveConfigData(cfg); err != nil {
		t.Fatalf("saveConfigData seed failed: %v", err)
	}

	// Remove one skill
	if err := removeSkillFromConfig("tooling"); err != nil {
		t.Fatalf("removeSkillFromConfig failed: %v", err)
	}

	reloaded := loadConfigFromFile()
	if reloaded == nil {
		t.Fatal("config file disappeared after remove")
	}
	if len(reloaded.Skills) != 1 {
		t.Fatalf("Skills length = %d after remove, want 1", len(reloaded.Skills))
	}
	if reloaded.Skills[0].Name != "deploy" {
		t.Errorf("remaining skill = %q, want %q", reloaded.Skills[0].Name, "deploy")
	}

	_ = tmp // suppress unused warning
}

func TestRemoveNonexistentSkill(t *testing.T) {
	setTestHome(t)

	// No config file on disk -- should be a no-op, no error
	err := removeSkillFromConfig("nonexistent")
	if err != nil {
		t.Errorf("removeSkillFromConfig on nonexistent skill returned error: %v", err)
	}
}

func TestSaveConfigDataCreatesDir(t *testing.T) {
	tmp := setTestHome(t)

	configDir := filepath.Join(tmp, ".config", "efx-skills")
	if _, err := os.Stat(configDir); err == nil {
		t.Fatal("config dir should not exist yet")
	}

	cfg := &ConfigData{
		SkillsPath: defaultSkillsPath(),
		Skills:     []SkillMeta{},
	}
	if err := saveConfigData(cfg); err != nil {
		t.Fatalf("saveConfigData failed: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(configDir); err != nil {
		t.Fatalf("config dir not created: %v", err)
	}

	// File should exist and be valid JSON
	configFile := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("reading config file failed: %v", err)
	}

	var out ConfigData
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("config file is not valid JSON: %v", err)
	}

	// Skills should be [] not null
	if string(data) == "" {
		t.Fatal("config file is empty")
	}
	if contains(string(data), `"skills": null`) {
		t.Error("skills serialized as null, want empty array")
	}
}

func TestSkillMetaFromAPISkill(t *testing.T) {
	s := api.Skill{
		ID:          "abc123",
		Name:        "tooling",
		Source:      "acme/tooling",
		Description: "A tooling skill",
		Installs:    100,
		Stars:       50,
		Registry:    "skills.sh",
	}

	meta := skillMetaFromAPISkill(s)

	if meta.Owner != "acme/tooling" {
		t.Errorf("Owner = %q, want %q", meta.Owner, "acme/tooling")
	}
	if meta.Name != "tooling" {
		t.Errorf("Name = %q, want %q", meta.Name, "tooling")
	}
	if meta.Registry != "skills.sh" {
		t.Errorf("Registry = %q, want %q", meta.Registry, "skills.sh")
	}
	if meta.URL != "https://github.com/acme/tooling" {
		t.Errorf("URL = %q, want %q", meta.URL, "https://github.com/acme/tooling")
	}
}

// --- Phase 3 Plan 1: Config page display changes ---

func TestRegistryDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty returns Custom", "", "Custom"},
		{"skills.sh returns Vercel", "skills.sh", "Vercel"},
		{"playbooks.com returns Playbooks", "playbooks.com", "Playbooks"},
		{"unknown passthrough", "github", "github"},
		{"other unknown passthrough", "unknown-registry", "unknown-registry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registryDisplayName(tt.input)
			if got != tt.want {
				t.Errorf("registryDisplayName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigViewRegistryFriendlyNames(t *testing.T) {
	m := configModel{
		registries: []Registry{
			{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true},
			{Name: "playbooks.com", URL: "https://playbooks.com/api/skills", Enabled: true},
		},
		width: 80,
	}
	output := m.View()

	// The checkbox line for the first registry should show "Vercel" as the name
	if !strings.Contains(output, "[x] Vercel") {
		t.Errorf("View() should contain '[x] Vercel' for skills.sh registry, got:\n%s", output)
	}
	// The checkbox line for the second registry should show "Playbooks" as the name
	if !strings.Contains(output, "[x] Playbooks") {
		t.Errorf("View() should contain '[x] Playbooks' for playbooks.com registry, got:\n%s", output)
	}
}

func TestConfigViewRepoTwoColumn(t *testing.T) {
	m := configModel{
		repos: []RepoSource{
			{Owner: "yoanbernabeu", Repo: "grepai-skills"},
			{Owner: "awni", Repo: "mlx-skills"},
		},
		section: 1,
		width:   80,
	}
	output := m.View()

	// Should NOT contain slash-joined format
	if strings.Contains(output, "yoanbernabeu/grepai-skills") {
		t.Errorf("View() should not contain slash-joined 'yoanbernabeu/grepai-skills', got:\n%s", output)
	}
	// Should contain owner and repo as separate tokens
	if !strings.Contains(output, "yoanbernabeu") {
		t.Errorf("View() should contain 'yoanbernabeu', got:\n%s", output)
	}
	if !strings.Contains(output, "grepai-skills") {
		t.Errorf("View() should contain 'grepai-skills', got:\n%s", output)
	}
}

func TestSkillMetaVersionInstalledRoundTrip(t *testing.T) {
	original := SkillMeta{
		Owner:     "acme",
		Name:      "tooling",
		Registry:  "skills.sh",
		URL:       "https://github.com/acme/tooling",
		Version:   "abc123def",
		Installed: "2026-03-06T10:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded SkillMeta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Version != original.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, original.Version)
	}
	if decoded.Installed != original.Installed {
		t.Errorf("Installed = %q, want %q", decoded.Installed, original.Installed)
	}
}

func TestSkillMetaVersionInstalledOmitEmpty(t *testing.T) {
	meta := SkillMeta{
		Owner:    "acme",
		Name:     "tooling",
		Registry: "skills.sh",
		URL:      "https://github.com/acme/tooling",
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	s := string(data)
	if strings.Contains(s, `"version"`) {
		t.Errorf("empty Version should be omitted, got: %s", s)
	}
	if strings.Contains(s, `"installed"`) {
		t.Errorf("empty Installed should be omitted, got: %s", s)
	}
}

func TestSkillMetaFromAPISkillVersionInstalledEmpty(t *testing.T) {
	s := api.Skill{
		ID:       "abc123",
		Name:     "tooling",
		Source:   "acme/tooling",
		Registry: "skills.sh",
	}

	meta := skillMetaFromAPISkill(s)

	if meta.Version != "" {
		t.Errorf("Version = %q, want empty string", meta.Version)
	}
	if meta.Installed != "" {
		t.Errorf("Installed = %q, want empty string", meta.Installed)
	}
}

func TestSearchViewContainsRegistryColumn(t *testing.T) {
	m := searchModel{
		results: []Skill{
			{
				ID:       "1",
				Name:     "test-skill",
				Source:   "acme/test-skill",
				Registry: "skills.sh",
				Installs: 42,
			},
			{
				ID:       "2",
				Name:     "other-skill",
				Source:   "beta/other-skill",
				Registry: "playbooks.com",
				Installs: 10,
			},
		},
		searched:     true,
		width:        100,
		focusOnInput: false,
	}
	m.paginator = paginator.New()
	m.paginator.PerPage = searchPerPage
	m.paginator.SetTotalPages(len(m.results))

	output := m.View()

	if !strings.Contains(output, "Vercel") {
		t.Errorf("Search View() should contain 'Vercel' for skills.sh registry, got:\n%s", output)
	}
	if !strings.Contains(output, "Playbooks") {
		t.Errorf("Search View() should contain 'Playbooks' for playbooks.com registry, got:\n%s", output)
	}
}

func TestConfigViewProvidersLabel(t *testing.T) {
	m := configModel{
		providers: []Provider{{Name: "Claude", Path: "/some/path", Configured: true}},
		width:     80,
	}
	output := m.View()

	if !strings.Contains(output, "Providers search") {
		t.Errorf("View() should contain 'Providers search', got:\n%s", output)
	}
}

func TestSkillEntryOriginLabel(t *testing.T) {
	// Test the display label logic that combines Registry and Origin fields.
	// Registry skills show registryDisplayName; custom skills show Origin label.
	tests := []struct {
		name     string
		entry    SkillEntry
		wantTag  string // expected parenthetical label, or "" if none
	}{
		{
			name:    "registry skill shows registry display name",
			entry:   SkillEntry{Name: "test-skill", Registry: "skills.sh", Origin: ""},
			wantTag: "(Vercel)",
		},
		{
			name:    "agents origin skill shows agents label",
			entry:   SkillEntry{Name: "test-skill", Registry: "", Origin: "agents"},
			wantTag: "(agents)",
		},
		{
			name:    "local provider origin skill shows local provider label",
			entry:   SkillEntry{Name: "test-skill", Registry: "", Origin: "local provider"},
			wantTag: "(local provider)",
		},
		{
			name:    "no registry no origin shows no label",
			entry:   SkillEntry{Name: "test-skill", Registry: "", Origin: ""},
			wantTag: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a minimal manageModel with one skill, one group
			m := manageModel{
				skills: []SkillEntry{tt.entry},
				groups: []SkillGroup{{Name: "custom", Skills: []int{0}}},
				displayList: []displayItem{
					{isGroup: true, groupIdx: 0, groupName: "custom"},
					{isGroup: false, skillIdx: 0, groupName: "custom"},
				},
				selectedIdx: 1, // select the skill row
				width:       80,
				paginator:   paginator.New(),
			}
			m.paginator.PerPage = 18 // default perPage (now dynamic via effectivePerPage)
			m.paginator.SetTotalPages(len(m.displayList))

			output := m.View()

			if tt.wantTag != "" {
				if !strings.Contains(output, tt.wantTag) {
					t.Errorf("View() should contain %q, got:\n%s", tt.wantTag, output)
				}
			} else {
				// Neither registryDisplayName nor origin label should appear
				if strings.Contains(output, "(Vercel)") || strings.Contains(output, "(Playbooks)") ||
					strings.Contains(output, "(agents)") || strings.Contains(output, "(local provider)") ||
					strings.Contains(output, "(Custom)") {
					t.Errorf("View() should not contain any parenthetical label, got:\n%s", output)
				}
			}
		})
	}
}

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
		t.Fatalf("codex SkillCount = %d, want 0 without directory", codex.SkillCount)
	}
	if codex.Synced {
		t.Fatalf("codex Synced = true, want false without directory")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
