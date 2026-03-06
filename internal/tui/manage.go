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
	"github.com/lmarques/efx-skills/internal/skill"
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
	Registry string
	Owner    string
	Origin   string // "agents" | "local provider" | "" (empty for registry skills)
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
	height      int
	paginator   paginator.Model
	loading     bool
	err         error
	statusMsg        string // feedback message shown at bottom of view
	updating         bool   // true while an update operation is in progress
	confirmingRemove bool   // true while showing remove confirmation dialog
	removeTarget     string // skill name being confirmed for removal
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

type verifySkillMsg struct {
	skillName   string
	hasUpdate   bool
	currentHash string
	latestHash  string
	err         error
}

type updateSkillMsg struct {
	skillName string
	err       error
}

type updateAllMsg struct {
	updated []string
	err     error
}

// effectivePerPage returns how many display items fit on one page given the
// current terminal height. Falls back to 18 when height is unknown.
func (m *manageModel) effectivePerPage() int {
	if m.height <= 0 {
		return 18 // sensible default before first WindowSizeMsg
	}
	// Chrome overhead:
	//   appStyle padding:       2 (top 1 + bottom 1)
	//   title box:              3 (border + text + border)
	//   newline after title:    1
	//   blank + subtitle+margin:3
	//   newline after subtitle: 1
	//   help bar margin+lines:  5 (wraps to 3-4 lines at 80-col + marginTop)
	//   pagination dots:        2
	//   status line:            2
	//   group separators:       2 (worst-case inter-group blank lines per page)
	// Total fixed chrome:      ~21 lines
	const chromeLines = 21
	available := m.height - chromeLines
	if available < 5 {
		available = 5 // minimum usable
	}
	return available
}

func newManageModel(provider Provider) manageModel {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 18 // initial default; updated dynamically by effectivePerPage
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

	// Load config for metadata enrichment
	cfg := loadConfigFromFile()
	metaLookup := make(map[string]SkillMeta)
	if cfg != nil {
		for _, meta := range cfg.Skills {
			metaLookup[meta.Name] = meta
		}
	}

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

	// Collect all directory names first for group detection
	centralNames := make(map[string]bool)
	var allNames []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".DS_Store" {
			allNames = append(allNames, entry.Name())
			centralNames[entry.Name()] = true
		}
	}

	// Also include provider-only skills (not in central storage)
	for name := range linkedSkills {
		if !centralNames[name] {
			allNames = append(allNames, name)
		}
	}

	for _, name := range allNames {
		group := extractGroup(name, allNames)
		entry := SkillEntry{
			Name:     name,
			Group:    group,
			Linked:   linkedSkills[name],
			Selected: linkedSkills[name],
		}
		if meta, ok := metaLookup[name]; ok {
			entry.Registry = meta.Registry
			entry.Owner = meta.Owner
			// Registry skills: Origin stays "" (unused)
		} else {
			entry.Registry = "" // No SkillMeta
			if centralNames[name] {
				entry.Origin = "agents"
			} else {
				entry.Origin = "local provider"
			}
		}
		skills = append(skills, entry)
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

func extractGroup(name string, allNames []string) string {
	parts := strings.Split(name, "-")
	if len(parts) >= 2 {
		return parts[0]
	}
	// No dash: check if this name is a prefix of other skill names (i.e., a group exists)
	for _, other := range allNames {
		if other != name && strings.HasPrefix(other, name+"-") {
			return name
		}
	}
	return "custom"
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

		// Determine collapsed state
		collapsed := false
		if oldCollapsedSet {
			// Preserve previous collapsed state
			collapsed = oldCollapsed[groupName]
		}
		// else: First build -- all groups start expanded (collapsed = false)

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
			if collapsed {
				// Collapsed: hide ALL skills in this group
				continue
			}
			m.displayList = append(m.displayList, displayItem{
				isGroup:   false,
				skillIdx:  skillIdx,
				groupName: groupName,
			})
		}
	}

	m.paginator.PerPage = m.effectivePerPage()
	m.paginator.SetTotalPages(len(m.displayList))
	m.clampPaginator()
}

// clampPaginator ensures the paginator Page and selectedIdx stay within valid
// bounds after any display list change (e.g., collapse/expand, skill removal).
func (m *manageModel) clampPaginator() {
	if m.paginator.TotalPages <= 0 {
		m.paginator.Page = 0
		return
	}
	if m.paginator.Page >= m.paginator.TotalPages {
		m.paginator.Page = m.paginator.TotalPages - 1
	}
	// Also clamp selectedIdx
	if m.selectedIdx >= len(m.displayList) {
		m.selectedIdx = len(m.displayList) - 1
	}
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	}
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

func getSkillsPath() string {
	cfg := loadConfigFromFile()
	if cfg != nil && cfg.SkillsPath != "" {
		return cfg.SkillsPath
	}
	return defaultSkillsPath()
}

func (m manageModel) Update(msg tea.Msg) (manageModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case skillsLoadedMsg:
		m.loading = false
		m.skills = msg.skills
		m.buildDisplayList()
		m.selectedIdx = 0

	case verifySkillMsg:
		m.updating = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error checking %s: %v", msg.skillName, msg.err)
		} else if msg.hasUpdate {
			m.statusMsg = fmt.Sprintf("Update available for %s (installed: %.7s, latest: %.7s)", msg.skillName, msg.currentHash, msg.latestHash)
		} else {
			m.statusMsg = fmt.Sprintf("%s is up to date (%.7s)", msg.skillName, msg.currentHash)
		}

	case updateSkillMsg:
		m.updating = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error updating %s: %v", msg.skillName, msg.err)
		} else {
			m.statusMsg = fmt.Sprintf("Updated %s successfully", msg.skillName)
		}

	case updateAllMsg:
		m.updating = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Update errors: %v", msg.err)
		} else if len(msg.updated) == 0 {
			m.statusMsg = "All skills are up to date"
		} else {
			m.statusMsg = fmt.Sprintf("Updated %d skills: %s", len(msg.updated), strings.Join(msg.updated, ", "))
		}

	case tea.WindowSizeMsg:
		m.width = int(float64(msg.Width) * 0.9)
		m.height = msg.Height
		m.buildDisplayList()
		return m, nil

	case tea.KeyMsg:
		// Handle confirmation dialog first (intercepts all keys when active)
		if m.confirmingRemove {
			switch msg.String() {
			case "y":
				m.confirmingRemove = false
				skillName := m.removeTarget
				m.removeTarget = ""
				return m, func() tea.Msg {
					removeSkillFully(skillName)
					return skillsLoadedMsg{skills: loadSkillsForProvider(m.provider)}
				}
			case "n", "esc":
				m.confirmingRemove = false
				m.removeTarget = ""
				m.statusMsg = ""
				return m, nil
			default:
				// Ignore all other keys during confirmation
				return m, nil
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
				m.paginator.Page = m.selectedIdx / m.effectivePerPage()
			}
		case "down", "j":
			if m.selectedIdx < len(m.displayList)-1 {
				m.selectedIdx++
				m.paginator.Page = m.selectedIdx / m.effectivePerPage()
			}
		case "left", "h", "pgup":
			m.paginator.PrevPage()
			m.selectedIdx = m.paginator.Page * m.effectivePerPage()
		case "right", "l", "pgdown":
			m.paginator.NextPage()
			m.selectedIdx = m.paginator.Page * m.effectivePerPage()
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
							m.paginator.Page = m.selectedIdx / m.effectivePerPage()
							break
						}
					}
					m.clampPaginator()
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
		case "r":
			// Remove skill (with confirmation)
			if len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if !item.isGroup {
					skillName := m.skills[item.skillIdx].Name
					m.confirmingRemove = true
					m.removeTarget = skillName
					m.statusMsg = fmt.Sprintf("Remove %s? Deletes from disk+config. [y] confirm [n] cancel", skillName)
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
		case "o":
			// Open selected skill or group URL in browser
			if len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if item.isGroup {
					// For groups: try to find a repo URL from config skills
					cfg := loadConfigFromFile()
					if cfg != nil {
						for _, meta := range cfg.Skills {
							if meta.Owner != "" && strings.HasPrefix(meta.Name, item.groupName) {
								if meta.URL != "" {
									openInBrowser(meta.URL)
									break
								}
							}
						}
					}
				} else {
					skillName := m.skills[item.skillIdx].Name
					cfg := loadConfigFromFile()
					if cfg != nil {
						if url := urlForManagedSkill(skillName, cfg.Skills); url != "" {
							openInBrowser(url)
						}
					}
				}
			}
		case "v":
			// Verify selected skill -- check for upstream update
			if !m.updating && len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if !item.isGroup {
					skillName := m.skills[item.skillIdx].Name
					m.updating = true
					m.statusMsg = "Checking for updates..."
					return m, func() tea.Msg {
						store := skill.NewStore(getSkillsPath())
						hasUpdate, currentHash, latestHash, err := store.CheckForUpdate(skillName)
						return verifySkillMsg{
							skillName:   skillName,
							hasUpdate:   hasUpdate,
							currentHash: currentHash,
							latestHash:  latestHash,
							err:         err,
						}
					}
				}
			}
		case "u":
			// Update selected skill
			if !m.updating && len(m.displayList) > 0 && m.selectedIdx < len(m.displayList) {
				item := m.displayList[m.selectedIdx]
				if !item.isGroup {
					skillName := m.skills[item.skillIdx].Name
					m.updating = true
					m.statusMsg = fmt.Sprintf("Updating %s...", skillName)
					return m, func() tea.Msg {
						store := skill.NewStore(getSkillsPath())
						err := store.UpdateSkill(skillName)
						return updateSkillMsg{
							skillName: skillName,
							err:       err,
						}
					}
				}
			}
		case "g":
			// Global update all skills
			if !m.updating {
				m.updating = true
				m.statusMsg = "Updating all skills..."
				return m, func() tea.Msg {
					store := skill.NewStore(getSkillsPath())
					updated, err := store.UpdateAllSkills()
					return updateAllMsg{
						updated: updated,
						err:     err,
					}
				}
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
			os.RemoveAll(linkPath)
		}
	}

}

func removeSkillFully(skillName string) {
	// 1. Unlink from ALL configured providers
	for _, p := range detectProviders() {
		if p.Configured {
			linkPath := filepath.Join(p.Path, skillName)
			os.RemoveAll(linkPath) // handles both symlinks and directories
		}
	}
	// 2. Remove from config.json
	removeSkillFromConfig(skillName)
	// 3. Remove from lock file
	store := skill.NewStore(getSkillsPath())
	store.RemoveFromLock(skillName)
	// 4. Physically delete skill directory from central storage
	skillsPath := getSkillsPath()
	os.RemoveAll(filepath.Join(skillsPath, skillName))
}

func (m manageModel) View() string {
	var b strings.Builder

	w := m.width
	if w <= 0 {
		w = 80
	}

	// Title
	b.WriteString(renderTitleBox(fmt.Sprintf("Manage Provider: %s", m.provider.Name)))
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

			bullet := bulletInactiveStyle.Render("●")
			if m.isGroupAllSelected(item.groupIdx) {
				bullet = bulletActiveStyle.Render("●")
			} else if m.isGroupPartialSelected(item.groupIdx) {
				bullet = bulletActiveStyle.Render("◐")
			}

			arrow := "▼"
			if m.groups[item.groupIdx].Collapsed {
				arrow = "▶"
			}

			if i == m.selectedIdx {
				// For selected row, use plain bullet (no color -- accent bg obscures it)
				plainBullet := "●"
				if m.isGroupPartialSelected(item.groupIdx) {
					plainBullet = "◐"
				}
				groupLabel := fmt.Sprintf("%s %s %s (%d/%d)", arrow, plainBullet, item.groupName, groupSelected, len(group.Skills))
				b.WriteString(getSelectedRowStyle(w).Render(groupLabel))
			} else {
				hasActive := groupSelected > 0
				groupText := fmt.Sprintf("%s %s (%d/%d)", arrow, item.groupName, groupSelected, len(group.Skills))
				if hasActive {
					b.WriteString(fmt.Sprintf("%s %s", bullet, groupActiveStyle.Render(groupText)))
				} else {
					b.WriteString(fmt.Sprintf("%s %s", bullet, groupInactiveStyle.Render(groupText)))
				}
			}
			b.WriteString("\n")
		} else {
			// Skill item (indented)
			skill := m.skills[item.skillIdx]
			bullet := bulletInactiveStyle.Render("●")
			if skill.Selected {
				bullet = bulletActiveStyle.Render("●")
			}

			// Show skill name without group prefix for cleaner display
			displayName := skill.Name
			if strings.HasPrefix(skill.Name, skill.Group+"-") {
				displayName = strings.TrimPrefix(skill.Name, skill.Group+"-")
			}
			if skill.Registry != "" {
				displayName += " (" + registryDisplayName(skill.Registry) + ")"
			} else if skill.Origin != "" {
				displayName += " (" + skill.Origin + ")"
			}

			status := ""
			if skill.Linked && !skill.Selected {
				status = " (remove)"
			} else if !skill.Linked && skill.Selected {
				status = " (add)"
			}

			if i == m.selectedIdx {
				plainBullet := "●"
				line := fmt.Sprintf("    %s %s%s", plainBullet, displayName, status)
				b.WriteString(getSelectedRowStyle(w).Render(line))
			} else {
				line := fmt.Sprintf("    %s %s%s", bullet, displayName, status)
				if skill.Linked && !skill.Selected {
					line = fmt.Sprintf("    %s %s%s", bullet, displayName, statusWarnStyle.Render(" (remove)"))
				} else if !skill.Linked && skill.Selected {
					line = fmt.Sprintf("    %s %s%s", bullet, displayName, statusOkStyle.Render(" (add)"))
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

	// Status message
	if m.confirmingRemove {
		b.WriteString("\n")
		alertStyle := statusWarnStyle.Width(w - 4)
		b.WriteString(alertStyle.Render("  " + m.statusMsg))
	} else if m.updating {
		b.WriteString("\n")
		b.WriteString(spinnerStyle.Render("  " + m.statusMsg))
	} else if m.statusMsg != "" {
		b.WriteString("\n")
		switch {
		case strings.HasPrefix(m.statusMsg, "Error"):
			b.WriteString(errorStyle.Render("  " + m.statusMsg))
		case strings.HasPrefix(m.statusMsg, "Update available"):
			b.WriteString(statusWarnStyle.Render("  " + m.statusMsg))
		case strings.HasPrefix(m.statusMsg, "Updated"),
			strings.HasPrefix(m.statusMsg, "All skills"),
			strings.Contains(m.statusMsg, "up to date"):
			b.WriteString(statusOkStyle.Render("  " + m.statusMsg))
		default:
			b.WriteString(statusMutedStyle.Render("  " + m.statusMsg))
		}
	}

	// Help
	b.WriteString(renderHelpBar(m.width, []string{
		"[space] preview", "[o] open", "[v] verify", "[u] update", "[g] update all",
		"[t] toggle", "[r] remove", "[enter] collapse/expand",
		"[a] all", "[n] none", "[s] apply/save", "[<-/->] page", "[esc] back",
	}))

	return b.String()
}
