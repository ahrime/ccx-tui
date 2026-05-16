package upgrade

import (
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	binary      string
	i18n        *i18n.I18n
	current     string
	latest      string
	downloading bool
}

func New(binary string, i *i18n.I18n) Model {
	return Model{binary: binary, i18n: i}
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	s := "  " + m.i18n.T("upgrade.title") + "\n"
	s += "  ─────────────────────\n"
	s += "  " + m.i18n.T("upgrade.current") + ": " + m.current + "\n"
	s += "  " + m.i18n.T("upgrade.latest") + ":  " + m.latest + "\n"
	if m.downloading {
		s += "\n  " + m.i18n.T("upgrade.downloading") + "\n"
	} else {
		s += "\n  [u] Check & Upgrade\n"
	}
	return s
}
