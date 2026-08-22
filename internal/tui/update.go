package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/jurekzsl/vellum/internal/runner"
)

type execFinishedMsg struct {
	meta       model.Metadata
	exitCode   int
	durationMs int64
	err        error
}

type delayedExecMsg struct {
	cmdObj     *exec.Cmd
	logFile    *os.File
	meta       model.Metadata
	copyOutput bool
}

type copyResultMsg struct {
	status string
	err    error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Key translation for text inputs in Huh forms
	if keyMsg, ok := msg.(tea.KeyMsg); ok && m.state == stateForm {
		if keyMsg.String() == "ctrl+]" {
			msg = tea.KeyMsg{Type: tea.KeyEnter, Alt: true}
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, _ := appStyle.GetFrameSize()
		lh, _ := listStyle.GetFrameSize()

		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = msg.Width - h - lh - 4

		// Resize viewport
		vpWidth := m.width - h - lh
		if vpWidth < 20 {
			vpWidth = 20
		}
		vpHeight := m.height - 10
		if vpHeight < 5 {
			vpHeight = 5
		}
		m.logViewport.Width = vpWidth
		m.logViewport.Height = vpHeight

		m.updateListHeight()

	case itemsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, item := range msg {
			items[i] = item
		}
		m.allItems = items
		cmds = append(cmds, m.list.SetItems(items))
		m.updateListHeight()

	case execFinishedMsg:
		_ = m.store.UpdateExecution(msg.meta, msg.exitCode, msg.durationMs)
		if msg.err != nil {
			m.status = fmt.Sprintf("Finished with error (Exit %d in %dms): %v", msg.exitCode, msg.durationMs, msg.err)
		} else {
			m.status = fmt.Sprintf("Completed successfully (Exit 0 in %dms)", msg.durationMs)
		}
		cmds = append(cmds, loadItems(m.store))

	case delayedExecMsg:
		return m, m.runCommandWithTea(msg.cmdObj, msg.logFile, msg.meta, msg.copyOutput)

	case copyResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Clipboard error: %v", msg.err)
		} else {
			m.status = msg.status
		}

	case tea.KeyMsg:
		// 1. Handle Log View State Keybindings
		if m.state == stateLogView {
			switch msg.String() {
			case "esc", "q":
				m.state = stateList
				m.status = ""
				return m, nil

			case "n", "]", "tab":
				// Next (older) log
				if len(m.logFiles) > 0 && m.currentLogIdx < len(m.logFiles)-1 {
					m.currentLogIdx++
					m.loadLogContent()
				}
				return m, nil

			case "p", "[", "shift+tab":
				// Previous (newer) log
				if len(m.logFiles) > 0 && m.currentLogIdx > 0 {
					m.currentLogIdx--
					m.loadLogContent()
				}
				return m, nil

			case "c":
				if m.logContent != "" {
					return m, copyToClipboardCmd(m.logContent, "Log content copied to clipboard")
				}
				return m, nil

			case "g", "home":
				m.logViewport.GotoTop()
				return m, nil

			case "G", "end":
				m.logViewport.GotoBottom()
				return m, nil

			default:
				m.logViewport, cmd = m.logViewport.Update(msg)
				return m, cmd
			}
		}

		// 2. Handle Form State Keybindings
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

		// 3. Handle Search Input Focused Keybindings
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
				val := strings.TrimSpace(m.searchInput.Value())
				var filtered []list.Item
				if val == "" {
					filtered = m.allItems
				} else {
					lowerVal := strings.ToLower(val)
					isScriptFilter := strings.Contains(lowerVal, "#s")
					isAliasFilter := strings.Contains(lowerVal, "#a")
					cleanVal := strings.ReplaceAll(strings.ReplaceAll(lowerVal, "#s", ""), "#a", "")
					cleanVal = strings.TrimSpace(cleanVal)

					for _, item := range m.allItems {
						if meta, ok := item.(model.Metadata); ok {
							if isScriptFilter && meta.Type != model.TypeScript {
								continue
							}
							if isAliasFilter && meta.Type != model.TypeAlias {
								continue
							}

							if cleanVal == "" ||
								strings.Contains(strings.ToLower(meta.Name), cleanVal) ||
								strings.Contains(strings.ToLower(meta.Desc), cleanVal) ||
								strings.Contains(strings.ToLower(meta.Command), cleanVal) {
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

		// 4. Handle List State Global Keybindings
		if m.state == stateList {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit

			case "/":
				m.searchInput.Focus()
				return m, textinput.Blink

			case "v", "V":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					m.currentLogMeta = meta
					logs, err := m.store.ListLogs(meta)
					if err != nil {
						m.status = fmt.Sprintf("Error reading logs: %v", err)
						return m, nil
					}
					m.logFiles = logs
					m.currentLogIdx = 0
					m.state = stateLogView
					m.loadLogContent()
					return m, nil
				}

			case "x", "X":
				exportMD, err := m.store.ExportMarkdown()
				if err != nil {
					m.status = fmt.Sprintf("Export error: %v", err)
					return m, nil
				}
				exportPath := filepath.Join(m.config.VellumDir, fmt.Sprintf("vellum_export_%s.md", time.Now().Format("2006-01-02_150405")))
				if wErr := os.WriteFile(exportPath, []byte(exportMD), 0600); wErr != nil {
					m.status = fmt.Sprintf("Error writing export: %v", wErr)
					return m, nil
				}
				return m, copyToClipboardCmd(exportMD, fmt.Sprintf("Exported to %s and copied to clipboard", exportPath))

			case "a":
				m.state = stateForm
				m.formData = &formData{mode: "create", itemType: "script"}
				m.activeFormID = "create_meta"
				m.form = m.newCreateScriptMetaForm()
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
					return m, copyToClipboardCmd(content, "Copied to clipboard")
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
						content, err := m.store.GetScriptContent(meta)
						if err != nil {
							m.status = fmt.Sprintf("Error reading content: %v", err)
							return m, nil
						}
						m.state = stateForm
						m.formData = &formData{
							mode:         "edit",
							name:         meta.Name,
							description:  meta.Desc,
							itemType:     "script",
							scriptType:   string(meta.ScriptType),
							content:      content,
							requiresSudo: meta.RequiresSudo,
							usesParams:   meta.UsesParams,
						}
						m.activeFormID = "edit_script"
						m.form = m.newFullScriptForm()
						return m, m.form.Init()
					}
				}

			case "enter":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					logFile, err := m.store.CreateLogFile(meta)
					if err != nil {
						m.status = fmt.Sprintf("Log creation error: %v", err)
						return m, nil
					}

					cmdObj, err := runner.PrepareCommand(meta, nil, logFile, m.config.DefaultShell, "")
					if err != nil {
						_ = logFile.Close()
						m.status = fmt.Sprintf("Prepare command error: %v", err)
						return m, nil
					}

					return m, m.runCommandWithTea(cmdObj, logFile, meta, false)
				}

			case "r":
				if i := m.list.SelectedItem(); i != nil {
					meta := i.(model.Metadata)
					m.state = stateForm
					m.formData = &formData{
						mode:     "exec_advanced",
						name:     meta.Name,
						itemType: string(meta.Type),
						command:  meta.Command,
						advPath:  filepath.Dir(meta.FilePath),
					}
					m.form = m.newAdvancedRunForm()
					return m, m.form.Init()
				}

			case "s":
				switch m.config.SortOrder {
				case "name":
					m.config.SortOrder = "type"
				case "type":
					m.config.SortOrder = "lastrun"
				case "lastrun":
					m.config.SortOrder = "name"
				default:
					m.config.SortOrder = "name"
				}

				if err := config.SaveConfig(m.config); err != nil {
					m.status = fmt.Sprintf("Error saving config: %v", err)
				} else {
					cmds = append(cmds, loadItems(m.store))
				}
				return m, tea.Batch(cmds...)
			}
		}
	}

	// State updates
	if m.state == stateList {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == stateForm && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			if m.activeFormID == "create_meta" {
				if m.formData.content == "" {
					m.formData.content = model.GetDefaultTemplate(model.ScriptType(m.formData.scriptType))
				}
				m.activeFormID = "create_content"
				m.form = m.newCreateScriptContentForm()
				cmds = append(cmds, m.form.Init())
				return m, tea.Batch(cmds...)
			}

			pCmd, err := m.processForm()
			if err != nil {
				m.status = fmt.Sprintf("Error: %v", err)
			} else {
				m.status = ""
			}

			if pCmd != nil {
				cmds = append(cmds, pCmd)
			}

			m.state = stateList
			cmds = append(cmds, loadItems(m.store))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) loadLogContent() {
	if len(m.logFiles) == 0 {
		m.logContent = "No log files found for this item."
		m.logViewport.SetContent(m.logContent)
		return
	}

	currentFile := m.logFiles[m.currentLogIdx]
	data, err := m.store.ReadLog(m.currentLogMeta, currentFile)
	if err != nil {
		m.logContent = fmt.Sprintf("Error reading log '%s': %v", currentFile, err)
	} else if strings.TrimSpace(data) == "" {
		m.logContent = fmt.Sprintf("Log file '%s' is empty.", currentFile)
	} else {
		m.logContent = data
	}
	m.logViewport.SetContent(m.logContent)
	m.logViewport.GotoTop()
}

func (m *Model) updateListHeight() {
	if m.width == 0 || m.height == 0 {
		return
	}

	h, v := appStyle.GetFrameSize()
	lh, _ := listStyle.GetFrameSize()
	_, lv := listStyle.GetFrameSize()

	overhead := 18
	availableHeight := m.height - v - overhead - lv
	if availableHeight < 1 {
		availableHeight = 1
	}

	itemCount := len(m.list.VisibleItems())
	targetHeight := itemCount
	if targetHeight < 5 {
		targetHeight = 5
	}
	if targetHeight > availableHeight {
		targetHeight = availableHeight
	}

	m.list.SetSize(m.width-h-lh, targetHeight)
}

func (m *Model) processForm() (tea.Cmd, error) {
	if m.formData.mode == "create" || m.formData.mode == "edit" {
		if strings.TrimSpace(m.formData.name) == "" {
			return nil, fmt.Errorf("name is required")
		}

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
			return nil, m.store.SaveScript(meta, m.formData.content)
		} else {
			meta := model.Metadata{
				ID:           strings.ToLower(m.formData.name),
				Name:         m.formData.name,
				Desc:         m.formData.description,
				Type:         model.TypeAlias,
				Command:      m.formData.command,
				RequiresSudo: m.formData.requiresSudo,
			}
			if err := m.store.SaveAlias(meta); err != nil {
				return nil, err
			}
			return nil, m.store.AddAliasToShell(meta)
		}
	} else if m.formData.mode == "import" {
		if strings.TrimSpace(m.formData.name) == "" {
			return nil, fmt.Errorf("name is required")
		}
		filePath := m.formData.filePath
		if strings.HasPrefix(filePath, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				filePath = filepath.Join(home, filePath[2:])
			}
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file failed: %w", err)
		}
		ext := strings.ToLower(filepath.Ext(filePath))
		scriptType := model.ScriptType(ext)
		if scriptType == "" {
			scriptType = model.ScriptTypeShell
		}
		meta := model.Metadata{
			ID:           strings.ToLower(m.formData.name),
			Name:         m.formData.name,
			Desc:         m.formData.description,
			Type:         model.TypeScript,
			ScriptType:   scriptType,
			RequiresSudo: m.formData.requiresSudo,
		}
		return nil, m.store.SaveScript(meta, string(content))
	} else if m.formData.mode == "alias" {
		if strings.TrimSpace(m.formData.name) == "" {
			return nil, fmt.Errorf("name is required")
		}
		meta := model.Metadata{
			ID:           strings.ToLower(m.formData.name),
			Name:         m.formData.name,
			Desc:         m.formData.description,
			Type:         model.TypeAlias,
			Command:      m.formData.command,
			RequiresSudo: m.formData.requiresSudo,
		}
		if err := m.store.SaveAlias(meta); err != nil {
			return nil, err
		}
		return nil, m.store.AddAliasToShell(meta)
	} else if m.formData.mode == "exec_advanced" {
		var meta model.Metadata
		found := false
		for _, item := range m.allItems {
			if mItem, ok := item.(model.Metadata); ok {
				if strings.EqualFold(mItem.Name, m.formData.name) {
					meta = mItem
					found = true
					break
				}
			}
		}

		if !found {
			return nil, fmt.Errorf("item not found")
		}

		logFile, err := m.store.CreateAdvLogFile(meta)
		if err != nil {
			return nil, err
		}

		header := fmt.Sprintf("Advanced Run: %s\nParams: %s\nUser: %s\nPath: %s\n---\n",
			time.Now().Format(time.RFC3339),
			m.formData.advParams,
			m.formData.advUser,
			m.formData.advPath,
		)
		_, _ = logFile.WriteString(header)

		var params []string
		if m.formData.advParams != "" {
			params = strings.Fields(m.formData.advParams)
		}

		cmdObj, err := runner.PrepareCommand(meta, params, logFile, m.config.DefaultShell, m.formData.advUser)
		if err != nil {
			_ = logFile.Close()
			return nil, err
		}

		if m.formData.advPath != "" {
			cmdObj.Dir = m.formData.advPath
		}

		var delayDuration time.Duration
		if m.formData.advDelay != "" {
			d, err := time.ParseDuration(m.formData.advDelay)
			if err == nil {
				delayDuration = d
			}
		}

		if delayDuration > 0 {
			copyOutput := m.formData.advCopy
			return tea.Tick(delayDuration, func(t time.Time) tea.Msg {
				return delayedExecMsg{
					cmdObj:     cmdObj,
					logFile:    logFile,
					meta:       meta,
					copyOutput: copyOutput,
				}
			}), nil
		}

		return m.runCommandWithTea(cmdObj, logFile, meta, m.formData.advCopy), nil
	} else if m.formData.mode == "delete" {
		if m.formData.confirm {
			var itemType model.ItemType
			if m.formData.itemType == "script" || m.formData.itemType == string(model.TypeScript) {
				itemType = model.TypeScript
			} else {
				itemType = model.TypeAlias
			}

			meta := model.Metadata{
				ID:      strings.ToLower(m.formData.name),
				Name:    m.formData.name,
				Desc:    m.formData.description,
				Type:    itemType,
				Command: m.formData.command,
			}
			if err := m.store.DeleteItem(meta); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

// Form Builders

func getAllScriptOptions() []huh.Option[string] {
	return []huh.Option[string]{
		// Coding
		huh.NewOption("Python (.py)", string(model.ScriptTypePython)),
		huh.NewOption("JavaScript (.js)", string(model.ScriptTypeJavascript)),
		huh.NewOption("TypeScript (.ts)", string(model.ScriptTypeTypescript)),
		huh.NewOption("React JSX (.jsx)", string(model.ScriptTypeJSX)),
		huh.NewOption("React TSX (.tsx)", string(model.ScriptTypeTSX)),
		huh.NewOption("Ruby (.rb)", string(model.ScriptTypeRuby)),
		huh.NewOption("Perl (.pl)", string(model.ScriptTypePerl)),
		huh.NewOption("PHP (.php)", string(model.ScriptTypePHP)),
		huh.NewOption("Lua (.lua)", string(model.ScriptTypeLua)),
		huh.NewOption("Tcl (.tcl)", string(model.ScriptTypeTcl)),

		// Shell
		huh.NewOption("Bash / Shell (.sh)", string(model.ScriptTypeShell)),
		huh.NewOption("Bash (.bash)", string(model.ScriptTypeBash)),
		huh.NewOption("Zsh (.zsh)", string(model.ScriptTypeZsh)),
		huh.NewOption("Fish (.fish)", string(model.ScriptTypeFish)),
		huh.NewOption("PowerShell (.ps1)", string(model.ScriptTypePowershell)),

		// Compiled
		huh.NewOption("Go (.go)", string(model.ScriptTypeGo)),
		huh.NewOption("Rust (.rs)", string(model.ScriptTypeRust)),
		huh.NewOption("Java (.java)", string(model.ScriptTypeJava)),
		huh.NewOption("Kotlin (.kt)", string(model.ScriptTypeKotlin)),
		huh.NewOption("Swift (.swift)", string(model.ScriptTypeSwift)),

		// System
		huh.NewOption("AppleScript (.applescript)", string(model.ScriptTypeAppleScript)),
		huh.NewOption("Windows Batch (.bat)", string(model.ScriptTypeBat)),
		huh.NewOption("Windows Cmd (.cmd)", string(model.ScriptTypeCmd)),
		huh.NewOption("VBScript (.vbs)", string(model.ScriptTypeVBS)),
	}
}

func (m *Model) newCreateScriptMetaForm() *huh.Form {
	if m.formData.scriptType == "" {
		m.formData.scriptType = string(model.ScriptTypeShell)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Script Name").
				Description("Unique identifier for this script (alphanumeric, dash, underscore)").
				Placeholder("e.g. backup-database").
				Value(&m.formData.name).
				Validate(func(s string) error {
					return m.store.ValidateName(s)
				}),
			huh.NewInput().
				Title("Description").
				Description("Optional description of the script's purpose").
				Placeholder("e.g. Dumps database and compresses archive").
				Value(&m.formData.description),
			huh.NewSelect[string]().
				Title("Script Language / Interpreter").
				Description("Select language (scroll or press '/' to search)").
				Options(getAllScriptOptions()...).
				Height(6).
				Filtering(true).
				Value(&m.formData.scriptType),
			huh.NewConfirm().
				Title("Requires Root / Sudo?").
				Description("Execute command with elevated sudo privileges").
				Value(&m.formData.requiresSudo),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newFullScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Script Name").
				Description("Unique identifier for this script").
				Value(&m.formData.name).
				Validate(func(s string) error {
					return m.store.ValidateName(s)
				}),
			huh.NewInput().
				Title("Description").
				Description("Summary of script behavior").
				Value(&m.formData.description),
			huh.NewSelect[string]().
				Title("Script Language / Interpreter").
				Description("Select language (scroll or press '/' to search)").
				Options(getAllScriptOptions()...).
				Height(6).
				Filtering(true).
				Value(&m.formData.scriptType),
			huh.NewConfirm().
				Title("Requires Root / Sudo?").
				Value(&m.formData.requiresSudo),
			huh.NewText().
				Title("Script Content").
				Description("Code executed when script runs (ctrl+e for external editor)").
				Lines(10).
				ShowLineNumbers(true).
				EditorExtension(m.formData.scriptType).
				Value(&m.formData.content),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newCreateScriptContentForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title(fmt.Sprintf("Script Content (%s)", m.formData.scriptType)).
				Description("Write or paste your script code. Use ctrl+] for new line, ctrl+e for editor, tab to finish.").
				Lines(12).
				ShowLineNumbers(true).
				EditorExtension(m.formData.scriptType).
				Value(&m.formData.content),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newImportScriptForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Script Name").
				Description("Unique identifier for the imported script").
				Placeholder("e.g. build-deploy").
				Value(&m.formData.name).
				Validate(func(s string) error {
					return m.store.ValidateName(s)
				}),
			huh.NewInput().
				Title("Description").
				Description("Optional description of the script").
				Placeholder("e.g. Imported deployment script").
				Value(&m.formData.description),
			huh.NewInput().
				Title("File Path").
				Description("Source file location on disk").
				Placeholder("e.g. ~/scripts/deploy.sh").
				Value(&m.formData.filePath).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("file path cannot be empty")
					}
					clean := s
					if strings.HasPrefix(clean, "~/") {
						if home, err := os.UserHomeDir(); err == nil {
							clean = filepath.Join(home, clean[2:])
						}
					}
					info, err := os.Stat(clean)
					if err != nil {
						return fmt.Errorf("file does not exist: %w", err)
					}
					if info.IsDir() {
						return fmt.Errorf("specified path is a directory, not a file")
					}
					return nil
				}),
			huh.NewConfirm().
				Title("Requires Root / Sudo?").
				Value(&m.formData.requiresSudo),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newAliasForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Alias Name").
				Description("Shell shortcut name (e.g. gco, myip, k)").
				Placeholder("e.g. gco").
				Value(&m.formData.name).
				Validate(func(s string) error {
					return m.store.ValidateName(s)
				}),
			huh.NewInput().
				Title("Description").
				Description("Short summary of what this alias does").
				Placeholder("e.g. Quick git checkout branch").
				Value(&m.formData.description),
			huh.NewInput().
				Title("Shell Command").
				Description("The command or pipe sequence executed by this alias").
				Placeholder("e.g. git checkout").
				Value(&m.formData.command).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("command cannot be empty")
					}
					return nil
				}),
			huh.NewConfirm().
				Title("Requires Root / Sudo?").
				Value(&m.formData.requiresSudo),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newDeleteConfirmForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete '%s'?", m.formData.name)).
				Description("This will permanently remove the script, metadata, and all execution logs.").
				Value(&m.formData.confirm),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) newAdvancedRunForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Working Directory").
				Description("Directory from which the script will execute (blank for default)").
				Placeholder("e.g. /var/www or ~/projects").
				Value(&m.formData.advPath),
			huh.NewInput().
				Title("Command Arguments").
				Description("Parameters passed to the script ($1, $2, or sys.argv)").
				Placeholder("e.g. --port 8080 --env prod").
				Value(&m.formData.advParams),
			huh.NewInput().
				Title("Run As User (sudo)").
				Description("Target system user (leave empty for current user)").
				Placeholder("e.g. root, postgres, www-data").
				Value(&m.formData.advUser),
			huh.NewInput().
				Title("Execution Delay").
				Description("Pause before starting execution (e.g. 5s, 1m)").
				Placeholder("Leave blank for immediate run").
				Value(&m.formData.advDelay),
			huh.NewConfirm().
				Title("Copy Output to Clipboard?").
				Description("Automatically copy execution output to system clipboard").
				Value(&m.formData.advCopy),
		),
	).WithTheme(MakeFormTheme(m.config.Theme)).WithShowHelp(false)
}

func (m *Model) runCommandWithTea(cmd *exec.Cmd, logFile *os.File, meta model.Metadata, copyOutput bool) tea.Cmd {
	startTime := time.Now()
	logPath := ""
	if logFile != nil {
		logPath = logFile.Name()
	}

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		durationMs := time.Since(startTime).Milliseconds()
		if logFile != nil {
			_ = logFile.Close()
		}

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}

		if copyOutput && logPath != "" {
			if content, rErr := os.ReadFile(logPath); rErr == nil {
				_ = copyToClipboard(string(content))
			}
		}

		return execFinishedMsg{
			meta:       meta,
			exitCode:   exitCode,
			durationMs: durationMs,
			err:        err,
		}
	})
}

func copyToClipboard(content string) error {
	var copyCmd *exec.Cmd
	if _, err := exec.LookPath("pbcopy"); err == nil {
		copyCmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("wl-copy"); err == nil {
		copyCmd = exec.Command("wl-copy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		copyCmd = exec.Command("xclip", "-selection", "clipboard")
	} else {
		return fmt.Errorf("no clipboard utility found (pbcopy/wl-copy/xclip)")
	}

	copyCmd.Stdin = strings.NewReader(content)
	return copyCmd.Run()
}

func copyToClipboardCmd(content, successMsg string) tea.Cmd {
	return func() tea.Msg {
		err := copyToClipboard(content)
		return copyResultMsg{
			status: successMsg,
			err:    err,
		}
	}
}

