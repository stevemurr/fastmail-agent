package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("62")
	secondaryColor = lipgloss.Color("241")
	accentColor    = lipgloss.Color("205")
	errorColor     = lipgloss.Color("196")
	successColor   = lipgloss.Color("46")

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Search input styles
	searchLabelStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	searchInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// List item styles
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(accentColor).
				Bold(true)

	// Email preview styles
	subjectStyle = lipgloss.NewStyle().
			Bold(true)

	fromStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	dateStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	previewStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Thread view styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(secondaryColor).
			MarginBottom(1)

	emailHeaderStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Padding(0, 1).
				MarginTop(1)

	bodyStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2)

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginTop(1)

	statusSuccessStyle = lipgloss.NewStyle().
				Foreground(successColor)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)
)
