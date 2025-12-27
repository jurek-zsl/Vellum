package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/jurekzsl/vellum/internal/runner"
	"github.com/jurekzsl/vellum/internal/store"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, _ := appStyle.GetFrameSize()

		// List border/padding
		lh, _ := listStyle.GetFrameSize()

		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = msg.Width - h - lh - 4 // Adjust search width
		m.updateListHeight()

	case itemsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, item := range msg {
			items[i] = item
		}
		m.allItems = items
		cmds = append(cmds, m.list.SetItems(items))
		m.updateListHeight()

	case tea.KeyMsg:
		// If form is active, it handles keys
		if m.state == stateForm {
			break
		}

		// Search Input Handling
		if m.searchInput.Focused() {
			switch msg.String() {
			case "enter", "down":
				m.searchInput.Blur()
				return m, nil
			case "esc":
				m.searchInput.Blur()
				return m, nil
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)

				// Filter logic
				val := m.searchInput.Value()
				var filtered []list.Item
				if val == "" {
					filtered = m.allItems
				} else {
					for _, item := range m.allItems {
						if meta, ok := item.(model.Metadata); ok {
							matches := false

							// Special tags
							if strings.Contains(val, "#S") || strings.Contains(val, "#s") {
								if meta.Type == model.TypeScript {
									matches = true
								}
							} else if strings.Contains(val, "#A") || strings.Contains(val, "#a") {
								if meta.Type == model.TypeAlias {
									matches = true
								}
							} else {
								// Normal search
								if strings.Contains(strings.ToLower(meta.Name), strings.ToLower(val)) ||
									strings.Contains(strings.ToLower(meta.Command), strings.ToLower(val)) {
									matches = true
								}
							}

							if matches {
								filtered = append(filtered, item)
							}
						}
					}
				}
				cmds = append(cmds, m.list.SetItems(filtered))
				m.updateListHeight()
				return m, tea.Batch(cmds...)
			}
		}

		// Key bindings when list is focused
		switch msg.String() {
		case "/":
			m.searchInput.Focus()
			return m, textinput.Blink
		}

		// Fallthrough only if not caught above
		if m.state == stateList {
			switch msg.String() {
			case "ctrl+c", "q":
				if !m.searchInput.Focused() {
					return m, tea.Quit
				}
			case "a":
				m.state = stateForm
				m.formData = &formData{mode: "create", itemType: "script"}
				m.form = m.newCreateScriptForm()
				return m, m.form.Init()
			case "i":
				m.state = stateForm
				m.formData = &formData{mode: "import", itemType: "script"}
				m.form = m.newImportScriptForm()
				return m, m.form.Init()
			case "l":
				m.state = stateForm
				m.formData = &formData{mode: "alias", itemType: "alias"}
				m.form = m.newAliasForm()
				return m, m.form.Init()
			case "c":
				m.state = stateForm
				m.formData = &formData{mode: "config", name: m.config.VellumDir}
				m.form = m.newConfigForm()
				return m, m.form.Init()
			case "d":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					m.store.DeleteItem(meta)
					// Remove from allItems as well
					// Re-loading is safer
					return m, loadItems(m.store)
				}
			case "e":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					m.state = stateForm
					m.formData = &formData{
						mode:         "edit",
						name:         meta.Name,
						description:  meta.Desc,
						itemType:     string(meta.Type),
						scriptType:   string(meta.ScriptType),
						command:      meta.Command,
						requiresSudo: meta.RequiresSudo,
						usesParams:   meta.UsesParams,
					}
					if meta.Type == model.TypeScript {
						content, _ := m.store.GetScriptContent(meta)
						m.formData.content = content
						m.form = m.newEditScriptForm()
					} else {
						m.form = m.newAliasForm()
					}
					return m, m.form.Init()
				}

			case "enter":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					// Run logic
					logFile, err := m.store.CreateLogFile(meta)
					if err != nil {
						m.err = err
						return m, nil
					}

					cmdObj, err := runner.PrepareCommand(meta, nil, logFile) // No params for now
					if err != nil {
						m.err = err
						return m, nil
					}

					c := tea.ExecProcess(cmdObj, func(err error) tea.Msg {
						logFile.Close()
						m.store.UpdateLastRun(meta)
						if err != nil {
							return fmt.Errorf("finished with error: %v", err)
						}
						return nil
					})
					return m, c
				}
			}
		}
	}

	// State updates
	if m.state == stateList {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == stateForm {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			// Process form data
			if err := m.processForm(); err != nil {
				m.status = fmt.Sprintf("Error: %v", err)
			} else {
				m.status = ""
			}
			m.state = stateList
			cmds = append(cmds, loadItems(m.store))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateListHeight() {
	if m.width == 0 || m.height == 0 {
		return
	}

	h, v := appStyle.GetFrameSize()
	lh, _ := listStyle.GetFrameSize() // Use only width overhead

	// Estimated overhead:
	// Logo: ~6 lines (with margins)
	// Search: ~4 lines
	// Header: ~2 lines
	// Footer: ~3 lines
	// Padding: ~2 lines
	// Total overhead ~ 17-19 lines.
	// We use 19 to be safe.
	overhead := 19

	// Available height for the list frame (including border)
	// listStyle usually adds border (2 lines).
	// So max content height is roughly available - 2.
	// But SetSize sets the size of the list component including its internal paginator/help/etc?
	// The bubbletea list component SetSize usually sets the size including borders if the list handles borders,
	// but here we wrap list in listStyle.
	// Bubbles list SetSize(width, height) sets the height of the list logic (items + pagination).
	// listStyle is an external wrapper.
	// m.list.SetSize sets the size of the INNER content if we render it like listStyle.Render(m.list.View()).
	// Actually, m.list.SetSize sets the dimensions of the list MODEL.
	// If we wrap it, we subtract wrapper overhead.

	// Let's assume listStyle.GetFrameSize returns horizontal and vertical overhead (borders/padding).
	_, lv := listStyle.GetFrameSize()

	availableHeight := m.height - v - overhead - lv
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Dynamic sizing based on items
	itemCount := len(m.list.VisibleItems())
	// Min height
	targetHeight := itemCount
	if targetHeight < 5 {
		targetHeight = 5
	}

	// Cap at available height
	if targetHeight > availableHeight {
		targetHeight = availableHeight
	}

	m.list.SetSize(m.width-h-lh, targetHeight)
}

func (m *Model) processForm() error {
	if m.formData.mode == "create" || m.formData.mode == "edit" {
		if m.formData.itemType == "script" || m.formData.itemType == string(model.TypeScript) {
			meta := model.Metadata{
				ID:           strings.ToLower(m.formData.name),
				Name:         m.formData.name,
				Desc:         m.formData.description,
				Type:         model.TypeScript,
				ScriptType:   model.ScriptType(m.formData.scriptType),
				Command:      m.formData.command,
				RequiresSudo: m.formData.requiresSudo,
				UsesParams:   m.formData.usesParams,
			}
			return m.store.SaveScript(meta, m.formData.content)
		} else { // Alias
			meta := model.Metadata{
				ID:      strings.ToLower(m.formData.name),
				Name:    m.formData.name,
				Desc:    m.formData.description,
				Type:    model.TypeAlias,
				Command: m.formData.command,
			}
			if err := m.store.SaveAlias(meta); err != nil {
				return err
			}
			return m.store.AddAliasToShell(meta)
		}
	} else if m.formData.mode == "import" {
		content, err := os.ReadFile(m.formData.filePath)
		if err != nil {
			return fmt.Errorf("read file failed: %w", err)
		}
		ext := filepath.Ext(m.formData.filePath)
		meta := model.Metadata{
			ID:           strings.ToLower(m.formData.name),
			Name:         m.formData.name,
			Desc:         m.formData.description,
			Type:         model.TypeScript,
			ScriptType:   model.ScriptType(ext),
			RequiresSudo: m.formData.requiresSudo,
		}
		return m.store.SaveScript(meta, string(content))
	} else if m.formData.mode == "alias" {
		meta := model.Metadata{

			ID:      strings.ToLower(m.formData.name),
			Name:    m.formData.name,
			Desc:    m.formData.description,
			Type:    model.TypeAlias,
			Command: m.formData.command,
		}
		if err := m.store.SaveAlias(meta); err != nil {
			return err
		}
		return m.store.AddAliasToShell(meta)
	} else if m.formData.mode == "config" {
		m.config.VellumDir = m.formData.name
		if err := config.SaveConfig(m.config); err != nil {
			return err
		}
		m.store = store.NewStore(m.config) // Reload store with new config
		if err := m.store.EnsureDirs(); err != nil {
			return err
		}
	}
	return nil
}

// Form Builders

func (m *Model) newCreateScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name),
			huh.NewInput().Title("Description").Value(&m.formData.description),
			huh.NewSelect[string]().
				Title("Type").
				Options(
					huh.NewOption("Python", ".py"),
					huh.NewOption("Bash", ".sh"),
					huh.NewOption("Go", ".go"),
					huh.NewOption("JavaScript", ".js"),
				).Value(&m.formData.scriptType),
			huh.NewText().Title("Script Content").Value(&m.formData.content),
			huh.NewConfirm().Title("Requires Sudo?").Value(&m.formData.requiresSudo),
		),
	).WithTheme(huh.ThemeDracula())
}

func (m *Model) newImportScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name),
			huh.NewInput().Title("File Path").Value(&m.formData.filePath),
			huh.NewInput().Title("Description").Value(&m.formData.description),
		),
	).WithTheme(huh.ThemeDracula())
}

func (m *Model) newAliasForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name),
			huh.NewInput().Title("Description").Value(&m.formData.description),
			huh.NewInput().Title("Command").Value(&m.formData.command),
		),
	).WithTheme(huh.ThemeDracula())
}

func (m *Model) newEditScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name), // Maybe readonly?
			huh.NewText().Title("Script Content").Value(&m.formData.content),
		),
	).WithTheme(huh.ThemeDracula())
}

func (m *Model) newConfigForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Vellum Directory").Value(&m.formData.name),
		),
	).WithTheme(huh.ThemeDracula())
}
