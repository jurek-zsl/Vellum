package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	// ... (imports need to be handled carefully, I will add os/exec to imports in a separate small edit if ReplaceFileContent doesn't support adding it easily, but here I can try replacing the import block or just assume it is there? wait I can't assume. Update.go already has "os", not "os/exec". I need to add it.)
	// Let's replace the import block first to be safe, then the handler.
	// Actually, I'll do the handler logic here and rely on the fact that I can edit imports separately or if I include imports in replacement it might work if I match enough context.
	// I will edit imports first.

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/jurekzsl/vellum/internal/runner"
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
			if msg.String() == "esc" {
				m.state = stateList
				m.form = nil
				m.err = nil
				m.status = ""
				return m, nil
			}
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
				// ... (This replace is inefficient for just handling Esc. I should insert the Esc handling in the Update loop first).
				// Let's modify the KeyMsg handler for stateForm.

				m.form = m.newCreateScriptForm()
				return m, m.form.Init()
			case "i":
				m.state = stateForm
				m.formData = &formData{mode: "import", itemType: "script"}
				m.form = m.newImportScriptForm()
				return m, m.form.Init()
			case "c":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					var content string
					if meta.Type == model.TypeScript {
						c, err := m.store.GetScriptContent(meta)
						if err != nil {
							m.status = fmt.Sprintf("Error reading content: %v", err)
							return m, nil
						}
						content = c
					} else {
						content = meta.Command
					}

					// Try pbcopy (mac) first, then xclip/wl-copy
					// Since we know user is on mac, pbcopy is primary.
					// We'll try a generic approach if possible or just hardcode for Mac as user requested Mac support primarily.
					// Implementation: execute command and write to stdin.

					var copyCmd *exec.Cmd
					if _, err := exec.LookPath("pbcopy"); err == nil {
						copyCmd = exec.Command("pbcopy")
					} else if _, err := exec.LookPath("wl-copy"); err == nil {
						copyCmd = exec.Command("wl-copy")
					} else if _, err := exec.LookPath("xclip"); err == nil {
						copyCmd = exec.Command("xclip", "-selection", "clipboard")
					} else {
						m.status = "No clipboard tool found (pbcopy/wl-copy/xclip)"
						return m, nil
					}

					in, err := copyCmd.StdinPipe()
					if err != nil {
						m.status = fmt.Sprintf("Error creating stdin pipe: %v", err)
						return m, nil
					}

					if err := copyCmd.Start(); err != nil {
						m.status = fmt.Sprintf("Error starting clipboard cmd: %v", err)
						return m, nil
					}

					go func() {
						defer in.Close()
						in.Write([]byte(content))
					}()

					if err := copyCmd.Wait(); err != nil {
						m.status = fmt.Sprintf("Clipboard error: %v", err)
					} else {
						m.status = "Copied to clipboard"
					}
					return m, nil
				}
			case "l":
				m.state = stateForm
				m.formData = &formData{mode: "alias", itemType: "alias"}
				m.form = m.newAliasForm()
				return m, m.form.Init()
			case "o":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					dir := m.store.GetItemDir(meta)

					cmd := exec.Command(m.config.FileManager, dir)
					if err := cmd.Start(); err != nil {
						m.status = fmt.Sprintf("Error opening directory: %v", err)
					} else {
						m.status = fmt.Sprintf("Opened %s", dir)
					}
					return m, nil
				}
			case "d":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					m.state = stateForm
					m.formData = &formData{
						mode:        "delete",
						name:        meta.Name,
						description: meta.Desc,
						itemType:    string(meta.Type),
						command:     meta.Command,
					}
					m.form = m.newDeleteConfirmForm()
					return m, m.form.Init()
				}
			case "e":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)

					if meta.Type == model.TypeAlias {
						// Aliases use the internal form
						m.state = stateForm
						m.formData = &formData{
							mode:        "edit",
							name:        meta.Name,
							description: meta.Desc,
							itemType:    string(meta.Type),
							command:     meta.Command,
						}
						m.form = m.newAliasForm()
						return m, m.form.Init()
					} else {
						// Scripts use external editor
						// Using m.config.DefaultEditor
						editor := m.config.DefaultEditor
						if editor == "" {
							editor = "nano" // Fallback
						}

						if meta.FilePath == "" {
							m.status = "Error: File path missing for script"
							return m, nil
						}

						c := tea.ExecProcess(exec.Command(editor, meta.FilePath), func(err error) tea.Msg {
							// Reload items after edit? Content might change but metadata?
							// Metadata might change if they edit metadata.json manually but usually just script.
							// We can just return nil or reload.
							if err != nil {
								return fmt.Errorf("editor finished with error: %v", err)
							}
							return nil
						})
						return m, c
					}
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

					cmdObj, err := runner.PrepareCommand(meta, nil, logFile, m.config.DefaultShell) // No params for now
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
	// We use 16 to be safe (reduced from 19).
	overhead := 16

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
	} else if m.formData.mode == "delete" {
		if m.formData.confirm {
			// Reconstruct metadata from formData to delete
			var itemType model.ItemType
			if m.formData.itemType == "script" || m.formData.itemType == string(model.TypeScript) {
				itemType = model.TypeScript
			} else {
				itemType = model.TypeAlias
			}

			meta := model.Metadata{
				ID:      strings.ToLower(m.formData.name), // This name is the one to be deleted
				Name:    m.formData.name,
				Desc:    m.formData.description,
				Type:    itemType,
				Command: m.formData.command,
			}
			if err := m.store.DeleteItem(meta); err != nil {
				return err
			}
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
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newImportScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name),
			huh.NewInput().Title("Description").Value(&m.formData.description),
			huh.NewInput().Title("File Path").Value(&m.formData.filePath),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newAliasForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name),
			huh.NewInput().Title("Description").Value(&m.formData.description),
			huh.NewInput().Title("Command").Value(&m.formData.command),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newEditScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&m.formData.name), // Maybe readonly?
			huh.NewText().Title("Script Content").Value(&m.formData.content),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newConfigsForm() *huh.Form {
	// Placeholder if we ever need it, but 'c' is now copy
	return nil
}

func (m *Model) newDeleteConfirmForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Are you sure you want to delete '%s'?", m.formData.name)).
				Value(&m.formData.confirm),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}
