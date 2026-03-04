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

// openLocalPreviewWithContentMsg carries pre-read content, no fetch needed
type openLocalPreviewWithContentMsg struct {
	skillName string
	content   string
}

// SkillEntry represents a skill in the management view
type SkillEntry struct {
	Name     string
	Group    string
	Linked   bool
	Selected bool
}

// SkillGroup represents a group of skills
type SkillGroup struct {
	Name      string
	Skills    []int // indices into skills slice
	Collapsed bool  // whether the group is collapsed
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
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Render("● ")
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
	if len(m.skills) == 0 {
		m.displayList = nil
		m.groups = nil
		return
	}

	// Preserve existing collapsed state BEFORE clearing groups
	oldCollapsed := make(map[string]bool)
	oldCollapsedSet := false
	for _, g := range m.groups {
		oldCollapsed[g.Name] = g.Collapsed
		oldCollapsedSet = true
	}

	m.displayList = nil
	m.groups = nil

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

		// Determine if any skill in this group is installed (linked)
		hasInstalled := false
		for _, idx := range skillIndices {
			if m.skills[idx].Linked {
				hasInstalled = true
				break
			}
		}

		// Determine collapsed state
		collapsed := false
		if oldCollapsedSet {
			// Preserve previous collapsed state
			collapsed = oldCollapsed[groupName]
		} else {
			// First build: collapse groups with no installed skills
			collapsed = !hasInstalled
		}

		m.groups = append(m.groups, SkillGroup{
			Name:      groupName,
			Skills:    skillIndices,
			Collapsed: collapsed,
		})

		// Add group header to display list
		m.displayList = append(m.displayList, displayItem{
			isGroup:   true,
			groupIdx:  gi,
			groupName: groupName,
		})

		// Add skills based on collapsed state
		for _, skillIdx := range skillIndices {
			if collapsed && !m.skills[skillIdx].Selected {
				// Collapsed: only show selected/installed skills
				continue
			}
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
			// Preview selected skill - read directly from local disk
			if len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if !item.isGroup {
					skillName := m.skills[item.skillIdx]
					home := os.Getenv("HOME")
					skillPath := filepath.Join(home, ".agents", "skills", skillName.Name, "SKILL.md")
					return m, func() tea.Msg {
						data, err := os.ReadFile(skillPath)
						if err != nil {
							return openLocalPreviewWithContentMsg{
								skillName: skillName.Name,
								content:   fmt.Sprintf("Error reading %s: %v", skillPath, err),
							}
						}
						return openLocalPreviewWithContentMsg{
							skillName: skillName.Name,
							content:   string(data),
						}
					}
				}
			}
		case "enter":
			// Collapse/uncollapse group
			if len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if item.isGroup {
					groupName := item.groupName
					m.groups[item.groupIdx].Collapsed = !m.groups[item.groupIdx].Collapsed
					m.buildDisplayList()
					// Keep cursor on the same group header (by name, since indices may shift)
					for i, d := range m.displayList {
						if d.isGroup && d.groupName == groupName {
							m.selectedIdx = i
							m.paginator.Page = m.selectedIdx / perPage
							break
						}
					}
				}
			}
		case "t":
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
		case "s":
			// Apply/save changes
			return m, func() tea.Msg {
				applySkillChanges(m.provider, m.skills)
				return skillsLoadedMsg{skills: loadSkillsForProvider(m.provider)}
			}
		}
		// Don't pass key messages to paginator (we handle pagination manually)
		return m, nil
	}

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

			arrow := "▼"
			if m.groups[item.groupIdx].Collapsed {
				arrow = "▶"
			}

			groupLabel := fmt.Sprintf("%s %s %s (%d/%d)", arrow, checkbox, item.groupName, groupSelected, len(group.Skills))
			
			if i == m.selectedIdx {
				b.WriteString(getSelectedRowStyle(w).Render(groupLabel))
			} else {
				// Use bold purple without margin for consistent spacing
				b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FBBF24")).Render(groupLabel))
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
	b.WriteString(helpStyle.Render("  [space] preview  [t] toggle  [↵] collapse/expand  [a] all  [n] none  [s] apply/save  [←/→] page  [esc] back"))

	return b.String()
}
