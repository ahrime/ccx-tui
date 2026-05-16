package more

import (
	"fmt"

	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubPanel int

const (
	SubNone SubPanel = iota
	SubLogs
	SubSettings
	SubHealth
	SubUpgrade
	SubMetrics
)

type Model struct {
	i18n    *i18n.I18n
	cursor  int
	items   []string
	panels  []SubPanel
}

func New(i *i18n.I18n) Model {
	panels := []SubPanel{SubLogs, SubSettings, SubHealth, SubUpgrade, SubMetrics}
	items := []string{
		i.T("more.logs"),
		i.T("more.settings"),
		i.T("more.health"),
		i.T("more.upgrade"),
		i.T("more.metrics"),
	}
	return Model{
		i18n:   i,
		items:  items,
		panels: panels,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m Model) SelectedPanel() SubPanel {
	if m.cursor < len(m.panels) {
		return m.panels[m.cursor]
	}
	return SubNone
}

func (m Model) View() string {
	s := lipgloss.NewStyle().Bold(true).Render("  More")
	s += "\n  ─────────────────────────────────────────────\n"
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}
		s += fmt.Sprintf("  %s%s\n", cursor, item)
	}
	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  Enter:Open  Esc:Back  j/k:Nav")
	return s
}

func (m Model) Bindings() []key.Binding { return nil }
