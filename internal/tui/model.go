package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/store"
)

type state int

const (
	stateList state = iota
	stateForm
)

type Model struct {
	state       state
	list        list.Model
	searchInput textinput.Model
	allItems    []list.Item
	form        *huh.Form
	store       *store.Store
	config      *config.Config

	// Form data holders
	formData *formData

	// Status/Error
	status string
	err    error

	// Window Size (for dynamic resizing)
	width  int
	height int
}

type formData struct {
	mode         string // "create", "import", "alias", "edit"
	name         string
	description  string
	itemType     string // "script" or "alias"
	scriptType   string // extension
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
	cfg, _ := config.LoadConfig()
	UpdateStyles(cfg.Theme) // Apply theme
	s := store.NewStore(cfg)
	s.EnsureDirs()
	s.CleanLogs()

	items := []list.Item{}
	// Initial load happens in Init()

	// Initial load happens in Init()

	// Delegate setup
	delegate := itemDelegate{} // Use our new custom delegate

	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)    // We'll render our own if needed, or keep it minimal
	l.SetFilteringEnabled(false) // We handle filtering manually now
	l.SetShowHelp(false)         // We render our own footer

	// Search input setup
	ti := textinput.New()
	ti.Placeholder = "Search hosts or tags..."
	// ti.Focus() // Start unfocused as requested
	ti.CharLimit = 156
	ti.Width = 50
	l.KeyMap.ShowFullHelp = key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h/?", "toggle help"),
	)

	return Model{
		state:       stateList,
		list:        l,
		searchInput: ti,
		store:       s,
		config:      cfg,
	}
}
