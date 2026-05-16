package more

import (
	"fmt"

	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	i18n   *i18n.I18n
	cursor int
	items  []string
}

func New(i *i18n.I18n) Model {
	return Model{
		i18n: i,
		items: []string{
			i.T("more.logs"),
			i.T("more.settings"),
			i.T("more.health"),
			i.T("more.upgrade"),
		},
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

func (m Model) View() string {
	s := "  More\n  ─────────────────────\n"
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}
		s += fmt.Sprintf("  %s%s\n", cursor, item)
	}
	s += "\n  Enter: Open  Esc: Back"
	return s
}

func (m Model) Bindings() []key.Binding { return nil }
