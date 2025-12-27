package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.state == stateForm {
		return appStyle.Render(m.form.View())
	}

	footer := footerStyle.Render(fmt.Sprintf(
		"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
		keyStyle.Render("a"), descStyle.Render("add"),
		keyStyle.Render("i"), descStyle.Render("import"),
		keyStyle.Render("l"), descStyle.Render("alias"),
		keyStyle.Render("e"), descStyle.Render("edit"),
		keyStyle.Render("d"), descStyle.Render("delete"),
		keyStyle.Render("c"), descStyle.Render("config"),
		keyStyle.Render("enter"), descStyle.Render("run"),
		keyStyle.Render("h"), descStyle.Render("help"),
		keyStyle.Render("q"), descStyle.Render("quit"),
	))

    status := ""
    if m.status != "" {
        status = statusMessageStyle(m.status)
    }

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		logoStyle.Render(logo),
		listStyle.Render(m.list.View()),
        status,
        footer,
	))
}
