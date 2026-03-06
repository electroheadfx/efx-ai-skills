package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primary   = lipgloss.Color("#FBBF24") // Yellow-Orange
	secondary = lipgloss.Color("#10B981") // Green
	accent    = lipgloss.Color("#007AE3") // Blue (selection bg)
	muted     = lipgloss.Color("#6B7280") // Gray
	danger    = lipgloss.Color("#EF4444") // Red
	warning   = lipgloss.Color("#F59E0B") // Yellow
)

// Styles
var (
	// App frame
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			MarginBottom(1)

	// Subtitle / section header
	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted).
			MarginBottom(1)

	// Table header
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4B5563")). // gray-600
				Padding(0, 1)

	// Table row
	tableRowStyle = lipgloss.NewStyle()

	// Selected item (text only) - for section headers
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			MarginBottom(1)

	// Selected row (with background)
	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(accent).
				Padding(0, 1).
				Width(60)

	// Status indicators
	statusOkStyle = lipgloss.NewStyle().
			Foreground(secondary)

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(warning)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(danger)

	statusMutedStyle = lipgloss.NewStyle().
				Foreground(muted)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(muted).
			MarginTop(1)

	// Border box
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted).
			Padding(1, 2)

	// Focused box
	focusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(1, 2)

	// Error message
	errorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Bold(true)

	// Success message
	successStyle = lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true)

	// Spinner
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primary)

	// Status line
	statusLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)

	// Title box (rounded border with primary color)
	titleBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(0, 2).
			MarginBottom(1)

	// Bullet styles for manage view
	bulletActiveStyle   = lipgloss.NewStyle().Foreground(secondary) // green
	bulletInactiveStyle = lipgloss.NewStyle().Foreground(muted)     // gray

	// Group header styles for manage view
	groupActiveStyle   = lipgloss.NewStyle().Bold(true).Foreground(secondary) // green bold
	groupInactiveStyle = lipgloss.NewStyle().Bold(true).Foreground(muted)     // gray bold

	// Config section box (inactive) - subtle muted border
	configSectionStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(muted).
				Padding(0, 1)

	// Config section box (active) - blue border for focus
	configSectionActiveStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(accent).
					Padding(0, 1)
)

const asciiLogo = `
       __                 _    _ _ _
  ___ / _|_  __       ___| | _(_) | |___
 / _ \ |_\ \/ /_____ / __| |/ / | | / __|
|  __/  _|>  <______|\__ \   <| | | \__ \
 \___|_| /_/\_\      |___/_|\_\_|_|_|___/`

// renderTitleBox renders text inside a rounded border box with bold primary foreground.
func renderTitleBox(text string) string {
	styledText := lipgloss.NewStyle().Bold(true).Foreground(primary).Render(text)
	return titleBoxStyle.Render(styledText)
}

// Helper functions
func renderStatus(synced bool, configured bool) string {
	if !configured {
		return statusMutedStyle.Render("not configured")
	}
	if synced {
		return statusOkStyle.Render("✓ synced")
	}
	return statusWarnStyle.Render("⚠ out of sync")
}

func renderProviderIcon(configured bool) string {
	if configured {
		return statusOkStyle.Render("●")
	}
	return statusMutedStyle.Render("○")
}

// getSelectedRowStyle returns a selectedRowStyle with dynamic width
func getSelectedRowStyle(width int) lipgloss.Style {
	if width <= 0 {
		width = 60
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(accent).
		Width(width)
}

// renderHelpBar renders a responsive help bar that wraps shortcut items
// across multiple lines when the terminal is too narrow for a single line.
func renderHelpBar(width int, items []string) string {
	if width <= 0 {
		width = 80
	}
	maxLineW := width - 4 // 2 indent + 2 margin
	if maxLineW < 20 {
		maxLineW = 20
	}

	var lines []string
	currentLine := ""
	for _, item := range items {
		candidate := currentLine
		if candidate != "" {
			candidate += "  " + item
		} else {
			candidate = item
		}
		if len(candidate) > maxLineW && currentLine != "" {
			lines = append(lines, "  "+currentLine)
			currentLine = item
		} else {
			currentLine = candidate
		}
	}
	if currentLine != "" {
		lines = append(lines, "  "+currentLine)
	}

	return helpStyle.Render(strings.Join(lines, "\n"))
}

// getTableHeaderStyle returns a tableHeaderStyle with dynamic width
func getTableHeaderStyle(width int) lipgloss.Style {
	if width <= 0 {
		width = 60
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4B5563")).
		Width(width)
}
