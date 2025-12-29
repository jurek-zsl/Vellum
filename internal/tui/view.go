package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jurekzsl/vellum/internal/model"
)

func (m Model) View() string {
	// Logo is always shown
	logoView := logoStyle.Render(logo)

	var content string
	var currentFooter string

	if m.state == stateForm {
		// Form View
		content = listStyle.Width(m.list.Width()).Render(m.form.View())

		// Footer
		currentFooter = footerStyle.Render(m.getFormFooter())

	} else {
		// List View

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

		// List Content
		content = lipgloss.JoinVertical(lipgloss.Left,
			searchView,
			header,
			listStyle.Width(m.list.Width()).Render(m.list.View()),
		)

		// List Footer
		footerLinks := fmt.Sprintf(
			"%s • %s • %s",
			fmt.Sprintf("%s %s", keyStyle.Render("↑/↓"), descStyle.Render("navigate")),
			fmt.Sprintf("%s %s", keyStyle.Render("/"), descStyle.Render("search")),
			fmt.Sprintf("%s %s", keyStyle.Render("enter"), descStyle.Render("run")),
		)

		moreFooter := fmt.Sprintf(
			"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
			keyStyle.Render("a"), descStyle.Render("add"),
			keyStyle.Render("i"), descStyle.Render("import"),
			keyStyle.Render("l"), descStyle.Render("alias"),
			keyStyle.Render("e"), descStyle.Render("edit"),
			keyStyle.Render("d"), descStyle.Render("del"),
			keyStyle.Render("c"), descStyle.Render("copy"),
			keyStyle.Render("r"), descStyle.Render("adv run"),
			keyStyle.Render("o"), descStyle.Render("open"),
			keyStyle.Render("q"), descStyle.Render("quit"),
		)

		currentFooter = lipgloss.JoinVertical(lipgloss.Left,
			footerStyle.Render(footerLinks),
			footerStyle.Render(moreFooter),
		)
	}

	status := ""
	if m.status != "" {
		status = statusMessageStyle(m.status)
	}

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		logoView,
		content,
		status,
		currentFooter,
	))
}

func (m Model) getFormFooter() string {
	if m.form == nil {
		return ""
	}

	// Helper to render "key desc" pair
	render := func(key, desc string) string {
		return fmt.Sprintf("%s %s", keyStyle.Render(key), descStyle.Render(desc))
	}

	var keys []string

	// Arrows only relevant if we have multiple fields or select/text elements
	// Delete confirmation is single field (Left/Right to toggle), so arrows don't "navigate" focus.
	if m.formData.mode != "delete" {
		keys = append(keys, render("↑/↓", "navigate"))
	}

	// Common navigation
	keys = append(keys,
		render("tab", "next"),
		render("shift+tab", "back"),
	)

	// Contextual help based on form mode/type
	// Create/Edit Script (Has Textarea)
	isScriptForm := (m.formData.mode == "create" || m.formData.mode == "edit") &&
		(m.formData.itemType == "script" || m.formData.itemType == string(model.TypeScript))

	if isScriptForm {
		keys = append(keys,
			render("alt+enter", "new line"),
			render("ctrl+e", "editor"),
		)
	}

	keys = append(keys, render("esc", "cancel"))

	// Join with bullet
	return strings.Join(keys, " • ")
}

func pad(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}
