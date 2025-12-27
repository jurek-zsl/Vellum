package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jurekzsl/vellum/internal/model"
)

type itemDelegate struct{}

func (d itemDelegate) Height() int { return 1 }

func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	str, ok := listItem.(model.Metadata)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render(strings.Join(s, " "))
		}
	}

	var typeIcon string
	if str.Type == model.TypeAlias {
		typeIcon = "A" // Alias
	} else {
		typeIcon = "S" // Script
	}

	// Calculate widths
	width := m.Width()
	if width <= 0 {
		width = 80 // Fallback
	}

	// subtract padding (itemStyle has Left 4) - verify if this is included or excluded from m.Width().
	// m.Width() is usually the content width if SetSize set the frame size correctly.
	// However, let's keep it safe.

	// Fixed columns
	typeWidth := 5
	lastRunWidth := 15
	gap := 2
	totalGap := gap * 3

	remaining := width - typeWidth - lastRunWidth - totalGap - 6 // -4 for padding + 2 extra for safety
	if remaining < 10 {
		remaining = 10
	}

	nameWidth := int(float64(remaining) * 0.3)
	cmdWidth := remaining - nameWidth

	name := truncate(str.Name, nameWidth)
	details := truncate(str.Command, cmdWidth)
	if str.Desc != "" {
		details = truncate(str.Desc, cmdWidth)
	}

	lastRun := "Never"
	if !str.LastRunAt.IsZero() {
		lastRun = timeSince(str.LastRunAt)
	}

	// Helper to ensure fixed width
	pad := func(s string, w int) string {
		if len(s) > w {
			return s[:w]
		}
		return s + strings.Repeat(" ", w-len(s))
	}

	row := fmt.Sprintf("%s  %s  %s  %s",
		pad(name, nameWidth),
		pad(typeIcon, typeWidth),
		pad(details, cmdWidth),
		pad(lastRun, lastRunWidth),
	)

	fmt.Fprint(w, fn(row))
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func timeSince(t time.Time) string {
	d := time.Since(t)
	if d < 24*time.Hour {
		return "Today"
	}
	if d < 48*time.Hour {
		return "Yesterday"
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%d days ago", days)
}
