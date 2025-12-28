package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// MakeFormTheme creates a custom huh theme based on the config usage
func MakeFormTheme(colorStr string) *huh.Theme {
	c := lipgloss.Color(colorStr)
	t := huh.ThemeDracula() // Base on Dracula

	// ACTIVE (FOCUSED) FIELD STYLES
	t.Focused.Title = t.Focused.Title.Foreground(c)
	t.Focused.Base = t.Focused.Base.BorderForeground(c)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(c)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(c)
	t.Focused.Option = t.Focused.Option.Foreground(c)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(c)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(c)

	// Confirm Dialog (Yes/No) - Focused Field
	// FocusedButton = The currently selected choice (e.g. No) -> Should be Theme Color
	t.Focused.FocusedButton = lipgloss.NewStyle().
		Background(c).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1)

	// BlurredButton = The unselected choice (e.g. Yes) -> Should be Neutral (No Purple)
	t.Focused.BlurredButton = lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)

	// INACTIVE (BLURRED) FIELD STYLES (When user moves to another field)
	neutral := lipgloss.Color("#626262")
	t.Blurred.Title = t.Blurred.Title.Foreground(neutral)
	t.Blurred.TextInput.Prompt = t.Blurred.TextInput.Prompt.Foreground(neutral)
	t.Blurred.SelectSelector = t.Blurred.SelectSelector.Foreground(neutral)

	// Confirm Dialog - Blurred Field
	// Both buttons should be neutral
	t.Blurred.FocusedButton = lipgloss.NewStyle().
		Background(neutral).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	t.Blurred.BlurredButton = lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	return t
}

var (
	appStyle = lipgloss.NewStyle().Padding(0, 2) // Reduced vertical padding

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")). // Purple
			Bold(true).
			MarginBottom(1) // Removed MarginTop(1)

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1).
			MarginBottom(1)

	searchFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#BD93F9")). // Purple when focused
				Padding(0, 1).
				MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			MarginTop(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	// Table styles
	itemStyle = lipgloss.NewStyle().PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#BD93F9")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				PaddingLeft(4)

	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(lipgloss.Color("#BD93F9")).
				PaddingLeft(4).
				MarginBottom(1)
)

func UpdateStyles(color string) {
	c := lipgloss.Color(color)

	logoStyle = logoStyle.Foreground(c)
	listStyle = listStyle.BorderForeground(c)
	searchFocusStyle = searchFocusStyle.BorderForeground(c)
	selectedItemStyle = selectedItemStyle.Background(c)
	tableHeaderStyle = tableHeaderStyle.BorderForeground(c)
}

const logo = "   _____     _ _           \n" +
	"  |  |  |___| | |_ _ _____ \n" +
	"  |  |  | -_| | | | |     |\n" +
	"   \\___/|___|_|_|___|_|_|_|"
