package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/mattn/go-runewidth"
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	meta, ok := listItem.(model.Metadata)
	if !ok {
		return
	}

	width := m.Width()
	if width <= 0 {
		width = 80
	}

	// Dynamic Width Calculation based on shared geometry constants
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

	typeBadge := "[SCR]"
	if meta.Type == model.TypeAlias {
		typeBadge = "[ALS]"
	}

	name := truncateRunes(meta.Name, nameWidth)
	desc := meta.Desc
	if desc == "" {
		desc = meta.Command
	}
	desc = truncateRunes(desc, descWidth)

	lastRun := "Never"
	if !meta.LastRunAt.IsZero() {
		lastRun = timeSince(meta.LastRunAt)
		if meta.LastExitCode != 0 {
			lastRun += fmt.Sprintf(" (!%d)", meta.LastExitCode)
		}
	}
	lastRun = truncateRunes(lastRun, ColLastRunWidth)

	row := fmt.Sprintf("%s%s%s%s%s%s%s",
		padRunes(name, nameWidth),
		strings.Repeat(" ", ColGap),
		padRunes(typeBadge, ColTypeWidth),
		strings.Repeat(" ", ColGap),
		padRunes(desc, descWidth),
		strings.Repeat(" ", ColGap),
		padRunes(lastRun, ColLastRunWidth),
	)

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render(row))
	} else {
		fmt.Fprint(w, itemStyle.Render(row))
	}
}

func truncateRunes(s string, maxCells int) string {
	if maxCells <= 3 {
		return runewidth.Truncate(s, maxCells, "")
	}
	if runewidth.StringWidth(s) > maxCells {
		return runewidth.Truncate(s, maxCells-3, "") + "..."
	}
	return s
}

func padRunes(s string, targetCells int) string {
	sw := runewidth.StringWidth(s)
	if sw > targetCells {
		return truncateRunes(s, targetCells)
	}
	return s + strings.Repeat(" ", targetCells-sw)
}

func timeSince(t time.Time) string {
	d := time.Since(t)
	if d < 1*time.Minute {
		return "Just now"
	}
	if d < 1*time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 48*time.Hour {
		return "Yesterday"
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}

