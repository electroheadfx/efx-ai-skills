package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmarques/efx-skills/internal/api"
)

// Registry represents a skill registry
type Registry struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// RepoSource represents a custom GitHub repo source
type RepoSource struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	URL   string `json:"url"`
}

// DeriveURL returns the GitHub URL for this repo, derived from owner/repo.
func (r RepoSource) DeriveURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Repo)
}

// SkillMeta represents per-skill provenance metadata stored in config.
type SkillMeta struct {
	Owner     string `json:"owner"`
	Name      string `json:"name"`
	Registry  string `json:"registry"`
	URL       string `json:"url"`
	Version   string `json:"version,omitempty"`
	Installed string `json:"installed,omitempty"`
}

// ConfigData represents the persistent configuration
type ConfigData struct {
	Registries []Registry   `json:"registries"`
	Repos      []RepoSource `json:"repos"`
	Providers  []string     `json:"enabled_providers"`
	SkillsPath string       `json:"skills-path"`
	Skills     []SkillMeta  `json:"skills"`
}

// configModel handles the config view
type configModel struct {
	registries  []Registry
	repos       []RepoSource
	providers   []Provider
	skills      []SkillMeta
	skillsPath  string
	section     int // 0=registries, 1=repos, 2=providers
	selectedIdx int
	width       int
	addingRepo  bool
	textInput   textinput.Model
	dirty       bool // track unsaved changes
	err         error
}

// registryDisplayName returns a friendly label for a registry.
// Falls back to the raw Name if no mapping exists.
func registryDisplayName(name string) string {
	switch name {
	case "":
		return "Custom"
	case "skills.sh":
		return "Vercel"
	case "playbooks.com":
		return "Playbooks"
	default:
		return name
	}
}

type configSavedMsg struct{}

func defaultSkillsPath() string {
	return filepath.Join(os.Getenv("HOME"), ".agents", "skills")
}

func loadConfigFromFile() *ConfigData {
	home := os.Getenv("HOME")
	configFile := filepath.Join(home, ".config", "efx-skills", "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil
	}

	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	// Apply defaults for new fields
	if cfg.SkillsPath == "" {
		cfg.SkillsPath = defaultSkillsPath()
	}
	if cfg.Skills == nil {
		cfg.Skills = []SkillMeta{}
	}
	for i := range cfg.Repos {
		if cfg.Repos[i].URL == "" {
			cfg.Repos[i].URL = cfg.Repos[i].DeriveURL()
		}
	}

	return &cfg
}

func defaultRegistries() []Registry {
	return []Registry{
		{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true},
		{Name: "playbooks.com", URL: "https://playbooks.com/api/skills", Enabled: true},
	}
}

func defaultRepos() []RepoSource {
	return []RepoSource{
		{Owner: "yoanbernabeu", Repo: "grepai-skills"},
		{Owner: "better-auth", Repo: "skills"},
		{Owner: "awni", Repo: "mlx-skills"},
	}
}

func newConfigModel() configModel {
	ti := textinput.New()
	ti.Placeholder = "owner/repo"
	ti.CharLimit = 100
	ti.Width = 40

	// Load existing config from file
	cfg := loadConfigFromFile()

	registries := defaultRegistries()
	repos := defaultRepos()
	skills := []SkillMeta{}
	skillsPath := defaultSkillsPath()

	if cfg != nil {
		if len(cfg.Registries) > 0 {
			registries = cfg.Registries
		}
		if cfg.Repos != nil {
			repos = cfg.Repos
		}
		if cfg.Skills != nil {
			skills = cfg.Skills
		}
		if cfg.SkillsPath != "" {
			skillsPath = cfg.SkillsPath
		}
	}

	return configModel{
		registries: registries,
		repos:      repos,
		providers:  detectProviders(), // detectProviders already respects config enabled state
		skills:     skills,
		skillsPath: skillsPath,
		textInput:  ti,
	}
}

func (m configModel) Init() tea.Cmd {
	return nil
}

func (m configModel) Update(msg tea.Msg) (configModel, tea.Cmd) {
	var cmd tea.Cmd

	// Handle text input mode
	if m.addingRepo {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// Parse and add repo
				input := m.textInput.Value()
				if parts := strings.Split(input, "/"); len(parts) == 2 {
					m.repos = append(m.repos, RepoSource{Owner: parts[0], Repo: parts[1]})
					m.dirty = true
				}
				m.addingRepo = false
				m.textInput.Reset()
				return m, nil
			case "esc":
				m.addingRepo = false
				m.textInput.Reset()
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case configSavedMsg:
		m.dirty = false

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.section = (m.section + 1) % 3
			m.selectedIdx = 0
		case "shift+tab":
			m.section = (m.section + 2) % 3
			m.selectedIdx = 0
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			maxIdx := m.getMaxIndex()
			if m.selectedIdx < maxIdx {
				m.selectedIdx++
			}
		case " ":
			m.toggleItem()
			m.dirty = true
		case "a":
			// Add repo (only in repos section)
			if m.section == 1 {
				m.addingRepo = true
				m.textInput.Focus()
				return m, textinput.Blink
			}
		case "d", "r":
			// Delete/remove repo (only in repos section)
			if m.section == 1 && len(m.repos) > 0 {
				m.repos = append(m.repos[:m.selectedIdx], m.repos[m.selectedIdx+1:]...)
				if m.selectedIdx >= len(m.repos) && m.selectedIdx > 0 {
					m.selectedIdx = len(m.repos) - 1
				}
				m.dirty = true
			}
		case "s":
			// Save config
			return m, m.saveConfig
		case "o":
			// Open registry or repo URL in browser
			switch m.section {
			case 0: // Registries
				if len(m.registries) > m.selectedIdx {
					reg := m.registries[m.selectedIdx]
					if url := registryBaseURL(reg.Name); url != "" {
						openInBrowser(url)
					}
				}
			case 1: // Repos
				if len(m.repos) > m.selectedIdx {
					repo := m.repos[m.selectedIdx]
					url := repo.DeriveURL()
					if url != "" {
						openInBrowser(url)
					}
				}
			}
		}
	}

	return m, nil
}

func (m configModel) getMaxIndex() int {
	switch m.section {
	case 0:
		return len(m.registries) - 1
	case 1:
		return len(m.repos) - 1
	case 2:
		return len(m.providers) - 1
	}
	return 0
}

func (m *configModel) toggleItem() {
	switch m.section {
	case 0:
		if len(m.registries) > m.selectedIdx {
			m.registries[m.selectedIdx].Enabled = !m.registries[m.selectedIdx].Enabled
		}
	case 2:
		// Toggle provider enabled state (creates/removes skills directory)
		if len(m.providers) > m.selectedIdx {
			p := &m.providers[m.selectedIdx]
			if p.Configured {
				// Just mark as disabled (don't delete directory)
				p.Configured = false
			} else {
				// Create the skills directory
				os.MkdirAll(p.Path, 0755)
				p.Configured = true
			}
		}
	}
}

func (m configModel) saveConfig() tea.Msg {
	home := os.Getenv("HOME")
	configDir := filepath.Join(home, ".config", "efx-skills")
	os.MkdirAll(configDir, 0755)

	configFile := filepath.Join(configDir, "config.json")

	// Collect enabled providers
	var enabledProviders []string
	for _, p := range m.providers {
		if p.Configured {
			enabledProviders = append(enabledProviders, p.Name)
		}
	}

	// Derive URL for any repos missing it
	repos := make([]RepoSource, len(m.repos))
	copy(repos, m.repos)
	for i := range repos {
		if repos[i].URL == "" {
			repos[i].URL = repos[i].DeriveURL()
		}
	}

	// Ensure skills is an empty slice, not nil
	skills := m.skills
	if skills == nil {
		skills = []SkillMeta{}
	}

	skillsPath := m.skillsPath
	if skillsPath == "" {
		skillsPath = defaultSkillsPath()
	}

	data := ConfigData{
		Registries: m.registries,
		Repos:      repos,
		Providers:  enabledProviders,
		SkillsPath: skillsPath,
		Skills:     skills,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errMsg{err: err}
	}

	if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
		return errMsg{err: err}
	}

	return configSavedMsg{}
}

func (m configModel) View() string {
	var b strings.Builder

	// Use dynamic width or default
	w := m.width
	if w <= 0 {
		w = 80
	}

	// Title with dirty indicator
	title := "Configuration"
	if m.dirty {
		title += " *"
	}
	b.WriteString(renderTitleBox(title))
	b.WriteString("\n")

	// Section box width (account for border chars: 2 per side + 1 padding each side)
	sectionW := w - 4
	if sectionW < 20 {
		sectionW = 20
	}

	// Registries section
	var regContent strings.Builder
	if m.section == 0 {
		regContent.WriteString(selectedStyle.Render("Registries"))
	} else {
		regContent.WriteString(subtitleStyle.Render("Registries"))
	}
	regContent.WriteString("\n")

	boldURLStyle := lipgloss.NewStyle().Bold(true)

	for i, reg := range m.registries {
		checkbox := "[ ]"
		if reg.Enabled {
			checkbox = "[x]"
		}
		nameWidth := 18
		urlWidth := sectionW - nameWidth - 8
		displayName := registryDisplayName(reg.Name)

		if m.section == 0 && i == m.selectedIdx {
			line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, displayName, truncateStr(reg.URL, urlWidth))
			regContent.WriteString(getSelectedRowStyle(sectionW).Render(line))
		} else {
			line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, displayName, boldURLStyle.Render(truncateStr(reg.URL, urlWidth)))
			regContent.WriteString(tableRowStyle.Render(line))
		}
		regContent.WriteString("\n")
	}

	if m.section == 0 {
		b.WriteString(configSectionActiveStyle.Width(sectionW).Render(regContent.String()))
	} else {
		b.WriteString(configSectionStyle.Width(sectionW).Render(regContent.String()))
	}
	b.WriteString("\n")

	// Repos section
	var reposContent strings.Builder
	if m.section == 1 {
		reposContent.WriteString(selectedStyle.Render("Custom GitHub Repos"))
	} else {
		reposContent.WriteString(subtitleStyle.Render("Custom GitHub Repos"))
	}
	reposContent.WriteString("\n")

	ownerWidth := 16
	for i, repo := range m.repos {
		line := fmt.Sprintf("  %-*s %s", ownerWidth, repo.Owner, repo.Repo)

		if m.section == 1 && i == m.selectedIdx {
			reposContent.WriteString(getSelectedRowStyle(sectionW).Render(line))
		} else {
			reposContent.WriteString(tableRowStyle.Render(line))
		}
		reposContent.WriteString("\n")
	}

	// Show add repo input or hint
	if m.addingRepo {
		reposContent.WriteString(fmt.Sprintf("  Add: %s\n", m.textInput.View()))
	} else if m.section == 1 {
		hints := "  [a] add repo"
		if len(m.repos) > 0 {
			hints += "  [r] remove repo"
		}
		reposContent.WriteString(statusMutedStyle.Render(hints))
		reposContent.WriteString("\n")
	}

	if m.section == 1 {
		b.WriteString(configSectionActiveStyle.Width(sectionW).Render(reposContent.String()))
	} else {
		b.WriteString(configSectionStyle.Width(sectionW).Render(reposContent.String()))
	}
	b.WriteString("\n")

	// Providers section
	var provContent strings.Builder
	if m.section == 2 {
		provContent.WriteString(selectedStyle.Render("Providers search"))
	} else {
		provContent.WriteString(subtitleStyle.Render("Providers search"))
	}
	provContent.WriteString("\n")

	// Dynamic column widths for providers
	providerNameWidth := 14
	pathWidth := sectionW - providerNameWidth - 8

	for i, p := range m.providers {
		checkbox := "[ ]"
		if p.Configured {
			checkbox = "[x]"
		}
		line := fmt.Sprintf("%s %-*s %s", checkbox, providerNameWidth, p.Name, p.Path)

		if m.section == 2 && i == m.selectedIdx {
			provContent.WriteString(getSelectedRowStyle(sectionW).Render(line))
		} else {
			line = fmt.Sprintf("%s %-*s %s", checkbox, providerNameWidth, p.Name, statusMutedStyle.Render(truncateStr(p.Path, pathWidth)))
			provContent.WriteString(tableRowStyle.Render(line))
		}
		provContent.WriteString("\n")
	}

	if m.section == 2 {
		b.WriteString(configSectionActiveStyle.Width(sectionW).Render(provContent.String()))
	} else {
		b.WriteString(configSectionStyle.Width(sectionW).Render(provContent.String()))
	}

	// Help
	helpItems := []string{"[tab] section", "[space] toggle", "[o] open", "[s] save", "[esc] back", "[q] quit"}
	if m.section == 1 {
		helpItems = []string{"[tab] section", "[o] open", "[s] save", "[esc] back", "[q] quit"}
	}
	b.WriteString(renderHelpBar(m.width, helpItems))

	return b.String()
}

func truncateStr(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// saveConfigData writes a ConfigData to ~/.config/efx-skills/config.json.
// It creates the config directory if it does not exist, ensures Skills is []
// not null in the output JSON, and defaults SkillsPath if empty.
func saveConfigData(cfg *ConfigData) error {
	home := os.Getenv("HOME")
	configDir := filepath.Join(home, ".config", "efx-skills")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Ensure Skills is empty slice, not nil
	if cfg.Skills == nil {
		cfg.Skills = []SkillMeta{}
	}

	// Default SkillsPath if empty
	if cfg.SkillsPath == "" {
		cfg.SkillsPath = defaultSkillsPath()
	}

	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// addSkillToConfig appends a SkillMeta to the config.json skills array.
// It is idempotent: duplicate entries (same Owner AND Name) are not added.
// If no config file exists, a new one is created with defaults.
func addSkillToConfig(meta SkillMeta) error {
	cfg := loadConfigFromFile()
	if cfg == nil {
		cfg = &ConfigData{
			Registries: defaultRegistries(),
			Repos:      defaultRepos(),
			SkillsPath: defaultSkillsPath(),
			Skills:     []SkillMeta{},
		}
	}

	// Check for duplicate by Owner+Name
	for _, existing := range cfg.Skills {
		if existing.Owner == meta.Owner && existing.Name == meta.Name {
			return nil // already tracked, idempotent
		}
	}

	cfg.Skills = append(cfg.Skills, meta)
	return saveConfigData(cfg)
}

// removeSkillFromConfig removes a SkillMeta entry from config.json by skill name.
// If no config file exists or the skill is not tracked, it is a no-op (returns nil).
func removeSkillFromConfig(skillName string) error {
	cfg := loadConfigFromFile()
	if cfg == nil {
		return nil
	}

	filtered := make([]SkillMeta, 0, len(cfg.Skills))
	for _, s := range cfg.Skills {
		if s.Name != skillName {
			filtered = append(filtered, s)
		}
	}
	cfg.Skills = filtered

	return saveConfigData(cfg)
}

// skillMetaFromAPISkill constructs a SkillMeta from an api.Skill.
// Maps: Owner=Source, Name=Name, Registry=Registry, URL=https://github.com/{Source}.
func skillMetaFromAPISkill(s api.Skill) SkillMeta {
	return SkillMeta{
		Owner:    s.Source,
		Name:     s.Name,
		Registry: s.Registry,
		URL:      fmt.Sprintf("https://github.com/%s", s.Source),
	}
}
