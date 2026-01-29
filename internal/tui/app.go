package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// View states
type viewState int

const (
	viewStatus viewState = iota
	viewSearch
	viewPreview
	viewManage
	viewConfig
)

// Main application model
type model struct {
	state  viewState
	width  int
	height int
	err    error

	// Sub-models
	statusModel  statusModel
	searchModel  searchModel
	previewModel previewModel
	manageModel  manageModel
	configModel  configModel
}

// Initialize the main model
func initialModel() model {
	return model{
		state:       viewStatus,
		statusModel: newStatusModel(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.statusModel.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == viewStatus {
				return m, tea.Quit
			}
			// Return to status view
			m.state = viewStatus
			return m, m.statusModel.Init()
		case "s":
			if m.state == viewStatus {
				m.state = viewSearch
				m.searchModel = newSearchModel()
				return m, m.searchModel.Init()
			}
		case "esc":
			if m.state == viewPreview {
				// Return to search view from preview
				m.state = viewSearch
				return m, nil
			} else if m.state != viewStatus {
				m.state = viewStatus
				return m, m.statusModel.Init()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Pass width to sub-models
		m.statusModel.width = int(float64(msg.Width) * 0.9)
		m.searchModel.width = int(float64(msg.Width) * 0.9)
		m.manageModel.width = int(float64(msg.Width) * 0.9)
		m.configModel.width = int(float64(msg.Width) * 0.9)

	case openManageMsg:
		m.state = viewManage
		m.manageModel = newManageModel(msg.provider)
		return m, m.manageModel.Init()

	case openConfigMsg:
		m.state = viewConfig
		m.configModel = newConfigModel()
		return m, m.configModel.Init()

	case openPreviewMsg:
		m.state = viewPreview
		m.previewModel = newPreviewModel(msg.skill.Source+"/"+msg.skill.Name, m.width, m.height)
		return m, m.previewModel.Init()
	}

	// Delegate to sub-models based on current state
	var cmd tea.Cmd
	switch m.state {
	case viewStatus:
		m.statusModel, cmd = m.statusModel.Update(msg)
	case viewSearch:
		m.searchModel, cmd = m.searchModel.Update(msg)
	case viewPreview:
		m.previewModel, cmd = m.previewModel.Update(msg)
	case viewManage:
		m.manageModel, cmd = m.manageModel.Update(msg)
	case viewConfig:
		m.configModel, cmd = m.configModel.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var content string

	switch m.state {
	case viewStatus:
		content = m.statusModel.View()
	case viewSearch:
		content = m.searchModel.View()
	case viewPreview:
		content = m.previewModel.View()
	case viewManage:
		content = m.manageModel.View()
	case viewConfig:
		content = m.configModel.View()
	}

	return appStyle.Render(content)
}

// Run starts the main TUI
func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// RunStatus starts directly in status view
func RunStatus() error {
	return Run()
}

// RunSearch starts in search view
func RunSearch(query string) error {
	m := initialModel()
	m.state = viewSearch
	m.searchModel = newSearchModel()
	m.searchModel.input.SetValue(query)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// RunPreview shows skill preview
func RunPreview(skill string) error {
	m := initialModel()
	m.state = viewPreview
	m.previewModel = newPreviewModel(skill, 80, 24) // Default size, will be updated by WindowSizeMsg

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// RunInstall installs a skill
func RunInstall(skill string, providers []string) error {
	fmt.Printf("Installing %s to providers: %v\n", skill, providers)
	// TODO: Implement installation logic
	return nil
}

// RunList lists installed skills
func RunList() error {
	providers := detectProviders()

	fmt.Println("Installed Skills")
	fmt.Println("================")

	skillsDir := filepath.Join(os.Getenv("HOME"), ".agents", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	fmt.Printf("\nCentral storage: %s\n", skillsDir)
	fmt.Printf("Total skills: %d\n\n", len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("  • %s\n", entry.Name())
		}
	}

	fmt.Println("\nProvider Status:")
	for _, p := range providers {
		fmt.Printf("  %s %s: %d skills\n",
			renderProviderIcon(p.Configured),
			p.Name,
			p.SkillCount)
	}

	return nil
}

// RunSync syncs skills across providers
func RunSync() error {
	fmt.Println("Syncing skills across all providers...")
	// TODO: Implement sync logic
	return nil
}

// RunConfig shows config view
func RunConfig() error {
	return Run()
}
