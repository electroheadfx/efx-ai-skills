package tui

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

// previewModel handles the preview view
type previewModel struct {
	skillName string
	content   string
	viewport  viewport.Model
	ready     bool
	loading   bool
	err       error
}

// Message types for preview
type previewContentMsg struct {
	content string
}

type previewErrMsg struct {
	err error
}

func newPreviewModel(skillName string, width, height int) previewModel {
	// Initialize viewport with provided dimensions
	headerHeight := 3
	footerHeight := 2
	viewportHeight := height - headerHeight - footerHeight
	if viewportHeight < 10 {
		viewportHeight = 10
	}
	if width < 40 {
		width = 80
	}

	vp := viewport.New(width, viewportHeight)
	vp.YPosition = headerHeight

	return previewModel{
		skillName: skillName,
		loading:   true,
		viewport:  vp,
		ready:     true, // Viewport is ready immediately
	}
}

func (m previewModel) Init() tea.Cmd {
	return func() tea.Msg {
		content, err := fetchSkillContent(m.skillName)
		if err != nil {
			return previewErrMsg{err: err}
		}
		return previewContentMsg{content: content}
	}
}

func (m previewModel) Update(msg tea.Msg) (previewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case previewContentMsg:
		m.loading = false
		// Render markdown with glamour
		width := m.viewport.Width
		if width < 40 {
			width = 80
		}
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-4),
		)
		if err == nil {
			rendered, err := renderer.Render(msg.content)
			if err == nil {
				m.content = rendered
			} else {
				m.content = msg.content
			}
		} else {
			m.content = msg.content
		}
		// Set content in viewport
		m.viewport.SetContent(m.content)
		m.viewport.GotoTop()

	case previewErrMsg:
		m.loading = false
		m.err = msg.err

	case tea.WindowSizeMsg:
		headerHeight := 3
		footerHeight := 2
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		if m.content != "" {
			m.viewport.SetContent(m.content)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m previewModel) headerView() string {
	title := titleStyle.Render("efx-skills v0.1.1 - Laurent Marques")
	skillTitle := subtitleStyle.Render(fmt.Sprintf("Preview: %s", m.skillName))
	return title + "\n" + skillTitle
}

func (m previewModel) footerView() string {
	info := helpStyle.Render(fmt.Sprintf("  %3.f%% • [j/k/↑/↓] scroll  [space/b] page  [g/G] top/bottom  [esc] back", m.viewport.ScrollPercent()*100))
	return info
}

func (m previewModel) View() string {
	if m.loading {
		return fmt.Sprintf("%s\n\n   Loading...", m.headerView())
	}

	if m.err != nil {
		return fmt.Sprintf("%s\n\n   %s", 
			m.headerView(),
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

// fetchSkillContent fetches SKILL.md content
func fetchSkillContent(skillName string) (string, error) {
	// Parse the skill name: "source/skillname" format
	parts := strings.Split(skillName, "/")

	if len(parts) >= 2 {
		owner := parts[0]
		repo := parts[1]
		skillPath := ""
		
		// If format is "owner/repo/skillname", extract skillname
		if len(parts) >= 3 {
			skillPath = strings.Join(parts[2:], "/")
		} else {
			skillPath = repo
		}

		// Try multiple path patterns for GitHub
		paths := []string{
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s/SKILL.md", owner, repo, skillPath),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/skills/%s/SKILL.md", owner, repo, skillPath),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/SKILL.md", owner, repo),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/README.md", owner, repo),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/%s/SKILL.md", owner, repo, skillPath),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/skills/%s/SKILL.md", owner, repo, skillPath),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/SKILL.md", owner, repo),
			fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/README.md", owner, repo),
		}

		client := &http.Client{Timeout: 10 * time.Second}
		for _, path := range paths {
			resp, err := client.Get(path)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}
				return string(body), nil
			}
		}
	}

	return "", fmt.Errorf("skill documentation not found for %s", skillName)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
