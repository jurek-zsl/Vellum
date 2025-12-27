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
			Foreground(lipgloss.Color("#BD93F9")). // Purple
			Bold(true).
			MarginBottom(1).
			MarginTop(1)

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

const logo = "   _____     _ _           \n" +
	"  |  |  |___| | |_ _ _____ \n" +
	"  |  |  | -_| | | | |     |\n" +
	"   \\___/|___|_|_|___|_|_|_|"
