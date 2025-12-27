package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
	state  state
	list   list.Model
	form   *huh.Form
	store  *store.Store
	config *config.Config
	
	// Form data holders
	formData *formData

	// Status/Error
	status string
    err error
}

type formData struct {
    mode        string // "create", "import", "alias", "edit"
    name        string
    description string
    itemType    string // "script" or "alias"
    scriptType  string // extension
    content     string // script content
    command     string // command or alias
    requiresSudo bool
    usesParams  bool
    filePath    string // for import
}

func InitialModel() Model {
	cfg, _ := config.LoadConfig()
	s := store.NewStore(cfg)
    s.EnsureDirs()
    s.CleanLogs()

	items := []list.Item{}
    // Initial load happens in Init()

    // Delegate setup
    delegate := list.NewDefaultDelegate()

	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
    l.SetShowStatusBar(true)
    l.SetFilteringEnabled(true)
    l.KeyMap.ShowFullHelp = key.NewBinding(
        key.WithKeys("h", "?"),
        key.WithHelp("h/?", "toggle help"),
    )

	return Model{
		state:  stateList,
		list:   l,
		store:  s,
		config: cfg,
	}
}
