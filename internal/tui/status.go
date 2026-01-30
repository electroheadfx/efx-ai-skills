package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Provider represents an AI coding agent provider
type Provider struct {
	Name       string
	Path       string
	Configured bool
	SkillCount int
	Synced     bool
}

// statusModel handles the status view
type statusModel struct {
	providers   []Provider
	totalSkills int
	selectedIdx int
	width       int
	loading     bool
	err         error
}

// Message types
type providersLoadedMsg struct {
	providers   []Provider
	totalSkills int
}

type errMsg struct {
	err error
}

type openManageMsg struct {
	provider Provider
}

type openConfigMsg struct {
	provider Provider
}

func newStatusModel() statusModel {
	return statusModel{
		loading: true,
	}
}

func (m statusModel) Init() tea.Cmd {
	return loadProviders
}

func loadProviders() tea.Msg {
	providers := detectProviders()

	// Count total skills in central storage
	skillsDir := filepath.Join(os.Getenv("HOME"), ".agents", "skills")
	totalSkills := 0
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				totalSkills++
			}
		}
	}

	return providersLoadedMsg{
		providers:   providers,
		totalSkills: totalSkills,
	}
}

func detectProviders() []Provider {
	home := os.Getenv("HOME")

	providerDefs := []struct {
		name string
		path string
	}{
		{"claude", filepath.Join(home, ".claude", "skills")},
		{"cursor", filepath.Join(home, ".cursor", "skills")},
		{"qoder", filepath.Join(home, ".qoder", "skills")},
		{"windsurf", filepath.Join(home, ".windsurf", "skills")},
		{"copilot", filepath.Join(home, ".copilot", "skills")},
		{"opencode", filepath.Join(home, ".opencode", "skills")},
	}

	var providers []Provider

	for _, pd := range providerDefs {
		p := Provider{
			Name: pd.name,
			Path: pd.path,
		}

		// Check if directory exists
		if info, err := os.Stat(pd.path); err == nil && info.IsDir() {
			p.Configured = true

			// Count skills (symlinks)
			if entries, err := os.ReadDir(pd.path); err == nil {
				for _, e := range entries {
					if e.Name() != ".DS_Store" {
						p.SkillCount++
					}
				}
			}
			p.Synced = true // Assume synced for now
		}

		providers = append(providers, p)
	}

	return providers
}

func (m statusModel) Update(msg tea.Msg) (statusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case providersLoadedMsg:
		m.loading = false
		m.providers = msg.providers
		m.totalSkills = msg.totalSkills

	case errMsg:
		m.loading = false
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.providers)-1 {
				m.selectedIdx++
			}
		case "enter", "m":
			// Open provider management for selected provider
			if len(m.providers) > 0 {
				// Return selected provider for manage view
				return m, func() tea.Msg {
					return openManageMsg{provider: m.providers[m.selectedIdx]}
				}
			}
		case "c":
			// Open config for selected provider
			if len(m.providers) > 0 {
				return m, func() tea.Msg {
					return openConfigMsg{provider: m.providers[m.selectedIdx]}
				}
			}
		case "r":
			// Refresh
			m.loading = true
			return m, loadProviders
		}
	}

	return m, nil
}

func (m statusModel) View() string {
	var b strings.Builder

	// Use dynamic width or default
	w := m.width
	if w <= 0 {
		w = 80
	}

	// Title
	b.WriteString(titleStyle.Render("efx-skills v0.1.1 - Laurent Marques"))
	b.WriteString("\n")

	if m.loading {
		b.WriteString(spinnerStyle.Render("Loading..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	// Section header
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Provider Status"))
	b.WriteString("\n")

	// Table header - use dynamic widths based on terminal width
	providerW := 20
	skillsW := 10
	statusW := w - providerW - skillsW - 10 // Use remaining width for status
	
	header := fmt.Sprintf("  %-*s  %*s  %*s", providerW, "Provider", skillsW, "Skills", statusW, "Status")
	b.WriteString(getTableHeaderStyle(w).Render(header))
	b.WriteString("\n")

	// Provider rows
	for i, p := range m.providers {
		skillCount := "-"
		if p.Configured {
			skillCount = fmt.Sprintf("%d", p.SkillCount)
		}
		
		statusText := "not configured"
		if p.Configured {
			if p.Synced {
				statusText = "✓ synced"
			} else {
				statusText = "⚠ out of sync"
			}
		}

		if i == m.selectedIdx {
			// Selected row: plain icon (no color) so background shows through
			icon := "●"
			if !p.Configured {
				icon = "○"
			}
			// Calculate padding for right-alignment (same as non-selected)
			statusTextLen := len(statusText)
			padding := statusW - statusTextLen
			if padding < 0 {
				padding = 0
			}
			row := fmt.Sprintf("%s %-*s  %*s  %*s%s", icon, providerW, p.Name, skillsW, skillCount, padding, "", statusText)
			b.WriteString(getSelectedRowStyle(w).Render(row))
		} else {
			// Non-selected: colored icon, right-aligned status
			icon := renderProviderIcon(p.Configured)
			var statusStyled string
			if p.Configured && p.Synced {
				statusStyled = statusOkStyle.Render("✓ synced")
			} else if p.Configured {
				statusStyled = statusWarnStyle.Render("⚠ out of sync")
			} else {
				statusStyled = statusMutedStyle.Render("not configured")
			}
			// Calculate padding for right-alignment
			statusTextLen := len(statusText) // Use plain text length
			padding := statusW - statusTextLen
			if padding < 0 {
				padding = 0
			}
			row := fmt.Sprintf("%s %-*s  %*s  %*s%s", icon, providerW, p.Name, skillsW, skillCount, padding, "", statusStyled)
			b.WriteString(tableRowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Summary
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Total: %d skills in ~/.agents/skills/\n", m.totalSkills))

	// Help - show context-aware help
	b.WriteString("\n")
	if len(m.providers) > 0 && m.providers[m.selectedIdx].Configured {
		b.WriteString(helpStyle.Render("  [s] search  [m/↵] manage  [c] config  [r] refresh  [q] quit"))
	} else {
		b.WriteString(helpStyle.Render("  [s] search  [c] configure  [r] refresh  [q] quit"))
	}

	return b.String()
}
