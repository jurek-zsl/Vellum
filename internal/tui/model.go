package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/huh"
	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/jurekzsl/vellum/internal/store"
)

type state int

const (
	stateList state = iota
	stateForm
	stateLogView
)

type Model struct {
	state       state
	list        list.Model
	searchInput textinput.Model
	allItems    []list.Item
	form        *huh.Form
	store       *store.Store
	config      *config.Config

	// Log Viewport
	logViewport   viewport.Model
	currentLogMeta model.Metadata
	logFiles      []string
	currentLogIdx int
	logContent    string

	// Form data holders
	formData     *formData
	activeFormID string // tracks current form step (e.g., "create_meta", "create_content")

	// Status/Error
	status string
	err    error

	// Window Size (for dynamic resizing)
	width  int
	height int
}

type formData struct {
	mode         string // "create", "import", "alias", "edit", "exec_advanced", "delete"
	name         string
	description  string
	itemType     string // "script" or "alias"
	scriptType   string // extension (.py, .sh, etc.)
	content      string // script content
	command      string // command or alias
	requiresSudo bool
	usesParams   bool
	filePath     string // for import
	confirm      bool   // for deletion confirmation
	sourceName   string // for copy operation

	// Advanced Run fields
	advPath    string
	advParams  string
	advUser    string
	advConfirm bool
	advDelay   string
	advCopy    bool
}

func InitialModel() Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{
			VellumDir:        config.GetDefaultDataDir(),
			Theme:            "#BD93F9",
			LogRetentionDays: 7,
			SortOrder:        "name",
			DefaultShell:     "/bin/bash",
			FileManager:      config.DetectDefaultFileManager(),
			MaxLogFiles:      50,
		}
	}

	UpdateStyles(cfg.Theme)
	s := store.NewStore(cfg)
	_ = s.EnsureDirs()

	items := []list.Item{}
	delegate := itemDelegate{}

	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // We handle custom filtering
	l.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "Search scripts & aliases (#s, #a)..."
	ti.CharLimit = 156
	ti.Width = 50

	l.KeyMap.ShowFullHelp = key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h/?", "toggle help"),
	)

	vp := viewport.New(80, 20)

	return Model{
		state:       stateList,
		list:        l,
		searchInput: ti,
		logViewport: vp,
		store:       s,
		config:      cfg,
	}
}

