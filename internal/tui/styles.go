package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primary   = lipgloss.Color("#7C3AED") // Purple
	secondary = lipgloss.Color("#10B981") // Green
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
				Background(lipgloss.Color("#7C3AED")).
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
)

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
		Background(lipgloss.Color("#7C3AED")).
		Width(width)
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
