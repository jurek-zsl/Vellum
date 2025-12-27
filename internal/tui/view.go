package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.state == stateForm {
		return appStyle.Render(m.form.View())
	}

	// Search Bar
	var searchView string
	if m.searchInput.Focused() {
		searchView = searchFocusStyle.Render(m.searchInput.View())
	} else {
		searchView = searchStyle.Render(m.searchInput.View())
	}

	// Table Header
	width := m.list.Width()
	if width <= 0 {
		width = 80
	}

	typeWidth := 5
	lastRunWidth := 15
	gap := 2
	totalGap := gap * 3
	remaining := width - typeWidth - lastRunWidth - totalGap - 4 // Match delegate padding logic
	if remaining < 10 {
		remaining = 10
	}
	nameWidth := int(float64(remaining) * 0.3)
	cmdWidth := remaining - nameWidth

	header := tableHeaderStyle.Render(fmt.Sprintf("%s  %s  %s  %s",
		pad("Name", nameWidth),
		pad("Type", typeWidth),
		pad("Description", cmdWidth),
		pad("Last Run", lastRunWidth),
	))

	// Footer
	footer := footerStyle.Render(fmt.Sprintf(
		"%s • %s • %s",
		fmt.Sprintf("%s %s", keyStyle.Render("↑/↓"), descStyle.Render("navigate")),
		fmt.Sprintf("%s %s", keyStyle.Render("/"), descStyle.Render("search")),
		fmt.Sprintf("%s %s", keyStyle.Render("enter"), descStyle.Render("run")),
	))

	moreFooter := footerStyle.Render(fmt.Sprintf(
		"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
		keyStyle.Render("a"), descStyle.Render("add"),
		keyStyle.Render("i"), descStyle.Render("import"),
		keyStyle.Render("l"), descStyle.Render("alias"),
		keyStyle.Render("e"), descStyle.Render("edit"),
		keyStyle.Render("d"), descStyle.Render("del"),
		keyStyle.Render("c"), descStyle.Render("conf"),
		keyStyle.Render("q"), descStyle.Render("quit"),
	))

	status := ""
	if m.status != "" {
		status = statusMessageStyle(m.status)
	}

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		logoStyle.Render(logo),
		searchView,
		header,
		listStyle.Width(m.list.Width()).Render(m.list.View()),
		status,
		footer,
		moreFooter,
	))
}

func pad(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}
