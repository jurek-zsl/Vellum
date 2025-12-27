package tui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
			Render

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			MarginBottom(1)
	
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			MarginTop(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

const logo = `
 _____     _ _           
|  |  |___| | |_ _ _____ 
|  |  | -_| | | | |     |
 \___/|___|_|_|_|_|_|_|_|
`
