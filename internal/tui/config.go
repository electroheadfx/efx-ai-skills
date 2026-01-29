package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
}

// ConfigData represents the persistent configuration
type ConfigData struct {
	Registries []Registry   `json:"registries"`
	Repos      []RepoSource `json:"repos"`
	Providers  []string     `json:"enabled_providers"`
}

// configModel handles the config view
type configModel struct {
	registries  []Registry
	repos       []RepoSource
	providers   []Provider
	section     int // 0=registries, 1=repos, 2=providers
	selectedIdx int
	width       int
	addingRepo  bool
	textInput   textinput.Model
	dirty       bool // track unsaved changes
	err         error
}

type configSavedMsg struct{}

func newConfigModel() configModel {
	ti := textinput.New()
	ti.Placeholder = "owner/repo"
	ti.CharLimit = 100
	ti.Width = 40

	return configModel{
		registries: []Registry{
			{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true},
			{Name: "playbooks.com", URL: "https://playbooks.com/api/skills", Enabled: true},
		},
		repos: []RepoSource{
			{Owner: "yoanbernabeu", Repo: "grepai-skills"},
			{Owner: "better-auth", Repo: "skills"},
			{Owner: "awni", Repo: "mlx-skills"},
		},
		providers: detectProviders(),
		textInput: ti,
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
		case "d":
			// Delete (for repos)
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

	data := ConfigData{
		Registries: m.registries,
		Repos:      m.repos,
		Providers:  enabledProviders,
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
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Registries section
	b.WriteString("\n")
	if m.section == 0 {
		b.WriteString(selectedStyle.Render("Registries"))
	} else {
		b.WriteString(subtitleStyle.Render("Registries"))
	}
	b.WriteString("\n")

	for i, reg := range m.registries {
		checkbox := "[ ]"
		if reg.Enabled {
			checkbox = "[x]"
		}
		// Dynamic column widths
		nameWidth := 18
		urlWidth := w - nameWidth - 8
		line := fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, reg.Name, reg.URL)

		if m.section == 0 && i == m.selectedIdx {
			b.WriteString(getSelectedRowStyle(w).Render(line))
		} else {
			line = fmt.Sprintf("%s %-*s %s", checkbox, nameWidth, reg.Name, statusMutedStyle.Render(truncateStr(reg.URL, urlWidth)))
			b.WriteString(tableRowStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Repos section
	b.WriteString("\n")
	if m.section == 1 {
		b.WriteString(selectedStyle.Render("Custom GitHub Repos"))
	} else {
		b.WriteString(subtitleStyle.Render("Custom GitHub Repos"))
	}
	b.WriteString("\n")

	for i, repo := range m.repos {
		repoName := fmt.Sprintf("%s/%s", repo.Owner, repo.Repo)
		line := fmt.Sprintf("  %s", repoName)

		if m.section == 1 && i == m.selectedIdx {
			b.WriteString(getSelectedRowStyle(w).Render(line))
		} else {
			b.WriteString(tableRowStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Show add repo input or hint
	if m.addingRepo {
		b.WriteString(fmt.Sprintf("  Add: %s\n", m.textInput.View()))
	} else if m.section == 1 {
		b.WriteString(statusMutedStyle.Render("  [a] add repo"))
		b.WriteString("\n")
	}

	// Providers section
	b.WriteString("\n")
	if m.section == 2 {
		b.WriteString(selectedStyle.Render("Providers"))
	} else {
		b.WriteString(subtitleStyle.Render("Providers"))
	}
	b.WriteString("\n")

	// Dynamic column widths for providers
	providerNameWidth := 14
	pathWidth := w - providerNameWidth - 8

	for i, p := range m.providers {
		checkbox := "[ ]"
		if p.Configured {
			checkbox = "[x]"
		}
		line := fmt.Sprintf("%s %-*s %s", checkbox, providerNameWidth, p.Name, p.Path)

		if m.section == 2 && i == m.selectedIdx {
			b.WriteString(getSelectedRowStyle(w).Render(line))
		} else {
			line = fmt.Sprintf("%s %-*s %s", checkbox, providerNameWidth, p.Name, statusMutedStyle.Render(truncateStr(p.Path, pathWidth)))
			b.WriteString(tableRowStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	helpText := "[tab] section  [space] toggle  [s] save  [esc] back  [q] quit"
	if m.section == 1 {
		helpText = "[tab] section  [a] add  [d] delete  [s] save  [esc] back  [q] quit"
	}
	b.WriteString(helpStyle.Render("  " + helpText))

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
