package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmarques/efx-skills/internal/api"
)

// Skill is an alias for api.Skill
type Skill = api.Skill

const searchPerPage = 10

// searchModel handles the search view
type searchModel struct {
	input       textinput.Model
	results     []Skill
	selectedIdx int
	width       int
	paginator   paginator.Model
	loading     bool
	searched    bool
	err         error
	focusOnInput bool // true = focus on input, false = focus on results
}

// Message types for search
type searchResultsMsg struct {
	results []Skill
}

type searchErrMsg struct {
	err error
}

type openPreviewMsg struct {
	skill Skill
}

func newSearchModel() searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search skills..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = searchPerPage
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render("● ")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("○ ")

	return searchModel{
		input:        ti,
		paginator:    p,
		focusOnInput: true, // Start with focus on input
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case searchResultsMsg:
		m.loading = false
		m.searched = true
		m.results = msg.results
		m.selectedIdx = 0
		m.paginator.SetTotalPages(len(m.results))
		m.paginator.Page = 0
		// After search completes, switch focus to results if we have results
		if len(msg.results) > 0 {
			m.focusOnInput = false
			m.input.Blur()
		}

	case searchErrMsg:
		m.loading = false
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Toggle focus between input and results
			if len(m.results) > 0 {
				m.focusOnInput = !m.focusOnInput
				if m.focusOnInput {
					m.input.Focus()
				} else {
					m.input.Blur()
				}
			}
			return m, nil
		case "enter":
			// Enter key behavior depends on focus
			if m.focusOnInput {
				// When focused on input: search
				if !m.loading && m.input.Value() != "" {
					m.loading = true
					query := m.input.Value()
					return m, func() tea.Msg {
						results, err := searchSkills(query)
						if err != nil {
							return searchErrMsg{err: err}
						}
						return searchResultsMsg{results: results}
					}
				}
			} else if len(m.results) > 0 {
				// When focused on results: preview
				return m, func() tea.Msg {
					return openPreviewMsg{skill: m.results[m.selectedIdx]}
				}
			}
		case "up", "k":
			// Only handle navigation when focus is on results
			if !m.focusOnInput && m.selectedIdx > 0 {
				m.selectedIdx--
				m.paginator.Page = m.selectedIdx / searchPerPage
			}
		case "down", "j":
			// Only handle navigation when focus is on results
			if !m.focusOnInput && m.selectedIdx < len(m.results)-1 {
				m.selectedIdx++
				m.paginator.Page = m.selectedIdx / searchPerPage
			}
		case "left", "h", "pgup":
			// Only handle pagination when focus is on results
			if !m.focusOnInput {
				m.paginator.PrevPage()
				m.selectedIdx = m.paginator.Page * searchPerPage
			}
		case "right", "l", "pgdown":
			// Only handle pagination when focus is on results
			if !m.focusOnInput {
				m.paginator.NextPage()
				m.selectedIdx = m.paginator.Page * searchPerPage
			}
		case "i":
			// Install selected skill (only when focus is on results)
			if !m.focusOnInput && len(m.results) > 0 {
				// TODO: Trigger install
			}
		case "p":
			// Preview selected skill with 'p' key (only when focus is on results)
			if !m.focusOnInput && len(m.results) > 0 {
				return m, func() tea.Msg {
					return openPreviewMsg{skill: m.results[m.selectedIdx]}
				}
			}
		}
	}

	// Only update text input when it has focus
	if m.focusOnInput {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m searchModel) View() string {
	var b strings.Builder

	// Use dynamic width or default
	w := m.width
	if w <= 0 {
		w = 80
	}

	// Calculate column widths dynamically
	// Reserve: 8 for count, rest split between name and source
	availableWidth := w - 8
	nameWidth := availableWidth * 35 / 100  // 35% for name
	sourceWidth := availableWidth * 55 / 100 // 55% for source

	// Title
	b.WriteString(titleStyle.Render("Search Skills"))
	b.WriteString("\n\n")

	// Search input
	b.WriteString("  > ")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(spinnerStyle.Render("  Searching..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		return b.String()
	}

	if !m.searched {
		b.WriteString(statusMutedStyle.Render("  Type a query and press Enter to search"))
		b.WriteString("\n")
		b.WriteString(statusMutedStyle.Render("  Searches skills.sh and playbooks.com"))
	} else if len(m.results) == 0 {
		b.WriteString(statusMutedStyle.Render("  No skills found"))
	} else {
		// Results header
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("  Results (%d)", len(m.results))))
		b.WriteString("\n")
		b.WriteString("  " + strings.Repeat("─", w-4))
		b.WriteString("\n\n")

		// Get page bounds
		start, end := m.paginator.GetSliceBounds(len(m.results))

		// Results list
		for i := start; i < end; i++ {
			skill := m.results[i]
			// Format: name (source) - count
			popularity := ""
			if skill.Installs > 0 {
				if skill.Installs >= 1000 {
					popularity = fmt.Sprintf("%dk", skill.Installs/1000)
				} else {
					popularity = fmt.Sprintf("%d", skill.Installs)
				}
			} else if skill.Stars > 0 {
				popularity = fmt.Sprintf("%d*", skill.Stars)
			}

			// Use dynamic column widths
			nameFmt := fmt.Sprintf("%%-%ds", nameWidth)
			sourceFmt := fmt.Sprintf("%%-%ds", sourceWidth)
			
			line := fmt.Sprintf(nameFmt+" "+sourceFmt+" %6s",
				truncate(skill.Name, nameWidth),
				truncate(skill.Source, sourceWidth),
				popularity)

			if i == m.selectedIdx {
				b.WriteString(getSelectedRowStyle(w).Render(line))
			} else {
				b.WriteString(tableRowStyle.Render(line))
			}
			b.WriteString("\n")
		}

		// Pagination dots
		if m.paginator.TotalPages > 1 {
			b.WriteString("\n")
			b.WriteString("    ")
			b.WriteString(m.paginator.View())
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString("\n")
	if m.focusOnInput {
		b.WriteString(helpStyle.Render("  [↵] search  [tab] focus results  [esc] back  [q] quit"))
	} else if len(m.results) > 0 {
		b.WriteString(helpStyle.Render("  [i] install  [p/↵] preview  [↑/↓] navigate  [←/→] page  [tab] focus input  [esc] back  [q] quit"))
	} else {
		b.WriteString(helpStyle.Render("  [↵] search  [i] install  [p] preview  [←/→] page  [esc] back  [q] quit"))
	}

	return b.String()
}

// searchSkills searches both registries
func searchSkills(query string) ([]Skill, error) {
	return api.SearchAll(query, 50)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
