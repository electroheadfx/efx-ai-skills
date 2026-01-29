package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SkillEntry represents a skill in the management view
type SkillEntry struct {
	Name     string
	Group    string
	Linked   bool
	Selected bool
}

// SkillGroup represents a group of skills
type SkillGroup struct {
	Name   string
	Skills []int // indices into skills slice
}

// manageModel handles the provider management view
type manageModel struct {
	provider    Provider
	skills      []SkillEntry
	groups      []SkillGroup
	displayList []displayItem // flat list for rendering (groups + skills)
	selectedIdx int
	width       int
	paginator   paginator.Model
	loading     bool
	err         error
}

type displayItem struct {
	isGroup    bool
	groupIdx   int // index into groups slice
	skillIdx   int // index into skills slice
	groupName  string
}

type skillsLoadedMsg struct {
	skills []SkillEntry
}

const perPage = 18

func newManageModel(provider Provider) manageModel {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = perPage
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render("● ")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("○ ")

	return manageModel{
		provider:  provider,
		loading:   true,
		paginator: p,
	}
}

func (m manageModel) Init() tea.Cmd {
	return func() tea.Msg {
		skills := loadSkillsForProvider(m.provider)
		return skillsLoadedMsg{skills: skills}
	}
}

func loadSkillsForProvider(provider Provider) []SkillEntry {
	home := os.Getenv("HOME")
	skillsDir := filepath.Join(home, ".agents", "skills")

	var skills []SkillEntry

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return skills
	}

	// Get linked skills for this provider
	linkedSkills := make(map[string]bool)
	if provider.Configured {
		providerEntries, err := os.ReadDir(provider.Path)
		if err == nil {
			for _, e := range providerEntries {
				if e.Name() != ".DS_Store" {
					linkedSkills[e.Name()] = true
				}
			}
		}
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".DS_Store" {
			group := extractGroup(entry.Name())
			skills = append(skills, SkillEntry{
				Name:     entry.Name(),
				Group:    group,
				Linked:   linkedSkills[entry.Name()],
				Selected: linkedSkills[entry.Name()],
			})
		}
	}

	// Sort by group then name
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Group != skills[j].Group {
			return skills[i].Group < skills[j].Group
		}
		return skills[i].Name < skills[j].Name
	})

	return skills
}

func extractGroup(name string) string {
	// Common prefixes to look for
	parts := strings.Split(name, "-")
	if len(parts) >= 2 {
		return parts[0]
	}
	return "_other"
}

func (m *manageModel) buildDisplayList() {
	m.displayList = nil
	m.groups = nil

	if len(m.skills) == 0 {
		return
	}

	// Group skills
	groupMap := make(map[string][]int)
	for i, skill := range m.skills {
		groupMap[skill.Group] = append(groupMap[skill.Group], i)
	}

	// Sort group names
	var groupNames []string
	for name := range groupMap {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	// Build groups and display list
	for gi, groupName := range groupNames {
		skillIndices := groupMap[groupName]
		
		m.groups = append(m.groups, SkillGroup{
			Name:   groupName,
			Skills: skillIndices,
		})

		// Add group header to display list
		m.displayList = append(m.displayList, displayItem{
			isGroup:   true,
			groupIdx:  gi,
			groupName: groupName,
		})

		// Add skills in this group
		for _, skillIdx := range skillIndices {
			m.displayList = append(m.displayList, displayItem{
				isGroup:   false,
				skillIdx:  skillIdx,
				groupName: groupName,
			})
		}
	}

	m.paginator.SetTotalPages(len(m.displayList))
}

func (m *manageModel) isGroupAllSelected(groupIdx int) bool {
	group := m.groups[groupIdx]
	for _, skillIdx := range group.Skills {
		if !m.skills[skillIdx].Selected {
			return false
		}
	}
	return true
}

func (m *manageModel) isGroupPartialSelected(groupIdx int) bool {
	group := m.groups[groupIdx]
	hasSelected := false
	hasUnselected := false
	for _, skillIdx := range group.Skills {
		if m.skills[skillIdx].Selected {
			hasSelected = true
		} else {
			hasUnselected = true
		}
	}
	return hasSelected && hasUnselected
}

func (m *manageModel) toggleGroup(groupIdx int) {
	group := m.groups[groupIdx]
	allSelected := m.isGroupAllSelected(groupIdx)
	
	for _, skillIdx := range group.Skills {
		m.skills[skillIdx].Selected = !allSelected
	}
}

func (m manageModel) Update(msg tea.Msg) (manageModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case skillsLoadedMsg:
		m.loading = false
		m.skills = msg.skills
		m.buildDisplayList()
		m.selectedIdx = 0

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
				m.paginator.Page = m.selectedIdx / perPage
			}
		case "down", "j":
			if m.selectedIdx < len(m.displayList)-1 {
				m.selectedIdx++
				m.paginator.Page = m.selectedIdx / perPage
			}
		case "left", "h", "pgup":
			m.paginator.PrevPage()
			m.selectedIdx = m.paginator.Page * perPage
		case "right", "l", "pgdown":
			m.paginator.NextPage()
			m.selectedIdx = m.paginator.Page * perPage
		case "home":
			m.selectedIdx = 0
			m.paginator.Page = 0
		case "end":
			m.selectedIdx = len(m.displayList) - 1
			m.paginator.Page = m.paginator.TotalPages - 1
		case " ":
			// Toggle selection
			if len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if item.isGroup {
					m.toggleGroup(item.groupIdx)
				} else {
					m.skills[item.skillIdx].Selected = !m.skills[item.skillIdx].Selected
				}
			}
		case "a":
			// Select all
			for i := range m.skills {
				m.skills[i].Selected = true
			}
		case "n":
			// Select none
			for i := range m.skills {
				m.skills[i].Selected = false
			}
		case "enter":
			// Apply changes
			return m, func() tea.Msg {
				applySkillChanges(m.provider, m.skills)
				return skillsLoadedMsg{skills: loadSkillsForProvider(m.provider)}
			}
		}
	}

	m.paginator, cmd = m.paginator.Update(msg)
	return m, cmd
}

func applySkillChanges(provider Provider, skills []SkillEntry) {
	if !provider.Configured {
		os.MkdirAll(provider.Path, 0755)
	}

	home := os.Getenv("HOME")
	skillsDir := filepath.Join(home, ".agents", "skills")

	for _, skill := range skills {
		linkPath := filepath.Join(provider.Path, skill.Name)
		targetPath := filepath.Join(skillsDir, skill.Name)
		relPath, _ := filepath.Rel(provider.Path, targetPath)

		if skill.Selected && !skill.Linked {
			os.Symlink(relPath, linkPath)
		} else if !skill.Selected && skill.Linked {
			os.Remove(linkPath)
		}
	}
}

func (m manageModel) View() string {
	var b strings.Builder

	w := m.width
	if w <= 0 {
		w = 80
	}

	// Title
	b.WriteString(titleStyle.Render(fmt.Sprintf("Manage Provider: %s", m.provider.Name)))
	b.WriteString("\n")

	if m.loading {
		b.WriteString(spinnerStyle.Render("Loading..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	// Count selected
	selected := 0
	for _, s := range m.skills {
		if s.Selected {
			selected++
		}
	}

	// Section header
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("Skills (%d selected of %d)", selected, len(m.skills))))
	b.WriteString("\n")

	// Get page bounds
	start, end := m.paginator.GetSliceBounds(len(m.displayList))

	// Track current group for separator logic
	var lastGroup string

	// Display list
	for i := start; i < end; i++ {
		item := m.displayList[i]

		// Add separator between groups (except first)
		if item.groupName != lastGroup && lastGroup != "" && i > start {
			b.WriteString("\n")
		}
		lastGroup = item.groupName

		if item.isGroup {
			// Group header
			group := m.groups[item.groupIdx]
			groupSelected := 0
			for _, skillIdx := range group.Skills {
				if m.skills[skillIdx].Selected {
					groupSelected++
				}
			}

			checkbox := "[ ]"
			if m.isGroupAllSelected(item.groupIdx) {
				checkbox = "[x]"
			} else if m.isGroupPartialSelected(item.groupIdx) {
				checkbox = "[-]"
			}

			groupLabel := fmt.Sprintf("%s %s (%d/%d)", checkbox, item.groupName, groupSelected, len(group.Skills))
			
			if i == m.selectedIdx {
				b.WriteString(getSelectedRowStyle(w).Render(groupLabel))
			} else {
				// Use bold purple without margin for consistent spacing
				b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Render(groupLabel))
			}
			b.WriteString("\n")
		} else {
			// Skill item (indented)
			skill := m.skills[item.skillIdx]
			checkbox := "[ ]"
			if skill.Selected {
				checkbox = "[x]"
			}

			// Show skill name without group prefix for cleaner display
			displayName := skill.Name
			if strings.HasPrefix(skill.Name, skill.Group+"-") {
				displayName = strings.TrimPrefix(skill.Name, skill.Group+"-")
			}

			status := ""
			if skill.Linked && !skill.Selected {
				status = " (remove)"
			} else if !skill.Linked && skill.Selected {
				status = " (add)"
			}

			// Use consistent formatting - indent is part of content
			line := fmt.Sprintf("    %s %s%s", checkbox, displayName, status)

			if i == m.selectedIdx {
				b.WriteString(getSelectedRowStyle(w).Render(line))
			} else {
				if skill.Linked && !skill.Selected {
					line = fmt.Sprintf("    %s %s%s", checkbox, displayName, statusWarnStyle.Render(" (remove)"))
				} else if !skill.Linked && skill.Selected {
					line = fmt.Sprintf("    %s %s%s", checkbox, displayName, statusOkStyle.Render(" (add)"))
				}
				b.WriteString(tableRowStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	// Pagination dots
	if m.paginator.TotalPages > 1 {
		b.WriteString("\n")
		b.WriteString("    ")
		b.WriteString(m.paginator.View())
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  [space] toggle  [a] all  [n] none  [↵] apply  [←/→] page  [esc] back"))

	return b.String()
}
