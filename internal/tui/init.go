package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/jurekzsl/vellum/internal/store"
)

type itemsLoadedMsg []model.Metadata

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadItems(m.store),
		cleanLogsCmd(m.store),
	)
}

func loadItems(s *store.Store) tea.Cmd {
	return func() tea.Msg {
		items, err := s.ListItems()
		if err != nil {
			return itemsLoadedMsg{}
		}
		return itemsLoadedMsg(items)
	}
}

func cleanLogsCmd(s *store.Store) tea.Cmd {
	return func() tea.Msg {
		s.CleanLogs()
		return nil
	}
}

