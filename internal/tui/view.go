package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jurekzsl/vellum/internal/model"
)

func (m Model) View() string {
	logoView := logoStyle.Render(logo)

	var content string
	var currentFooter string

	switch m.state {
	case stateForm:
		width := m.list.Width()
		if width <= 0 {
			width = 80
		}
		formTitle := " Form"
		switch m.activeFormID {
		case "create_meta":
			formTitle = " Create New Script (Step 1/2: Metadata)"
		case "create_content":
			formTitle = fmt.Sprintf(" Create New Script (Step 2/2: %s Content)", m.formData.scriptType)
		case "edit_script":
			formTitle = fmt.Sprintf(" Edit Script: %s (%s)", m.formData.name, m.formData.scriptType)
		case "alias":
			formTitle = " Create New Shell Alias"
		case "import":
			formTitle = " Import Existing Script File"
		case "exec_advanced":
			formTitle = fmt.Sprintf(" Advanced Execution: %s", m.formData.name)
		case "delete":
			formTitle = fmt.Sprintf(" Confirm Deletion: %s", m.formData.name)
		}

		header := logHeaderStyle.Width(width).Render(formTitle)
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			listStyle.Width(width).Render(m.form.WithWidth(width - 4).View()),
		)
		currentFooter = footerStyle.Render(m.getFormFooter())

	case stateLogView:
		// Log Viewer View
		width := m.list.Width()
		if width <= 0 {
			width = 80
		}

		headerText := fmt.Sprintf("Logs for: %s", m.currentLogMeta.Name)
		if len(m.logFiles) > 0 {
			headerText += fmt.Sprintf(" (%d of %d: %s)", m.currentLogIdx+1, len(m.logFiles), m.logFiles[m.currentLogIdx])
		} else {
			headerText += " (No logs recorded yet)"
		}

		statusBadge := ""
		if !m.currentLogMeta.LastRunAt.IsZero() {
			if m.currentLogMeta.LastExitCode == 0 {
				statusBadge = logSuccessBadge.Render(fmt.Sprintf(" [EXIT 0 | %dms]", m.currentLogMeta.LastDurationMs))
			} else {
				statusBadge = logFailBadge.Render(fmt.Sprintf(" [EXIT %d | %dms]", m.currentLogMeta.LastExitCode, m.currentLogMeta.LastDurationMs))
			}
		}

		header := logHeaderStyle.Width(width).Render(headerText + statusBadge)
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			listStyle.Width(width).Render(m.logViewport.View()),
		)

		logFooterLinks := fmt.Sprintf(
			"%s • %s • %s • %s • %s",
			fmt.Sprintf("%s %s", keyStyle.Render("↑/↓/pgup/pgdn"), descStyle.Render("scroll")),
			fmt.Sprintf("%s %s", keyStyle.Render("n/p"), descStyle.Render("next/prev log")),
			fmt.Sprintf("%s %s", keyStyle.Render("c"), descStyle.Render("copy log")),
			fmt.Sprintf("%s %s", keyStyle.Render("g/G"), descStyle.Render("top/bottom")),
			fmt.Sprintf("%s %s", keyStyle.Render("esc/q"), descStyle.Render("back to list")),
		)
		currentFooter = footerStyle.Render(logFooterLinks)

	default: // stateList
		var searchView string
		if m.searchInput.Focused() {
			searchView = searchFocusStyle.Render(m.searchInput.View())
		} else {
			searchView = searchStyle.Render(m.searchInput.View())
		}

		width := m.list.Width()
		if width <= 0 {
			width = 80
		}

		totalGap := ColGap * 3
		available := width - ColTypeWidth - ColLastRunWidth - totalGap - 6
		if available < 20 {
			available = 20
		}
		nameWidth := int(float64(available) * 0.35)
		if nameWidth < 8 {
			nameWidth = 8
		}
		descWidth := available - nameWidth
		if descWidth < 10 {
			descWidth = 10
		}

		nameH := "Name"
		typeH := "Type"
		descH := "Description"
		runH := "Last Run"

		switch m.config.SortOrder {
		case "name":
			nameH += " ▼"
		case "type":
			typeH += " ▼"
		case "lastrun":
			runH += " ▼"
		}

		header := tableHeaderStyle.Render(fmt.Sprintf("%s%s%s%s%s%s%s",
			padRunes(nameH, nameWidth),
			strings.Repeat(" ", ColGap),
			padRunes(typeH, ColTypeWidth),
			strings.Repeat(" ", ColGap),
			padRunes(descH, descWidth),
			strings.Repeat(" ", ColGap),
			padRunes(runH, ColLastRunWidth),
		))

		content = lipgloss.JoinVertical(lipgloss.Left,
			searchView,
			header,
			listStyle.Width(m.list.Width()).Render(m.list.View()),
		)

		footerLinks := fmt.Sprintf(
			"%s • %s • %s • %s",
			fmt.Sprintf("%s %s", keyStyle.Render("↑/↓"), descStyle.Render("navigate")),
			fmt.Sprintf("%s %s", keyStyle.Render("/"), descStyle.Render("search")),
			fmt.Sprintf("%s %s", keyStyle.Render("enter"), descStyle.Render("run")),
			fmt.Sprintf("%s %s", keyStyle.Render("v"), descStyle.Render("view logs")),
		)

		moreFooter := fmt.Sprintf(
			"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
			keyStyle.Render("a"), descStyle.Render("add"),
			keyStyle.Render("i"), descStyle.Render("import"),
			keyStyle.Render("l"), descStyle.Render("alias"),
			keyStyle.Render("e"), descStyle.Render("edit"),
			keyStyle.Render("d"), descStyle.Render("del"),
			keyStyle.Render("c"), descStyle.Render("copy"),
			keyStyle.Render("r"), descStyle.Render("adv run"),
			keyStyle.Render("x"), descStyle.Render("export"),
			keyStyle.Render("o"), descStyle.Render("open"),
			keyStyle.Render("s"), descStyle.Render("sort"),
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

	render := func(key, desc string) string {
		return fmt.Sprintf("%s %s", keyStyle.Render(key), descStyle.Render(desc))
	}

	var keys []string

	if m.formData.mode != "delete" {
		keys = append(keys, render("↑/↓", "navigate"))
	}

	keys = append(keys,
		render("tab", "next"),
		render("shift+tab", "back"),
	)

	isScriptForm := (m.formData.mode == "create" || m.formData.mode == "edit") &&
		(m.formData.itemType == "script" || m.formData.itemType == string(model.TypeScript))

	if isScriptForm {
		keys = append(keys,
			render("ctrl+]", "new line"),
			render("ctrl+e", "editor"),
		)
	}

	keys = append(keys, render("esc", "cancel"))
	return strings.Join(keys, " • ")
}

