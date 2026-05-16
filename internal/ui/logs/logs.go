package logs

import (
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	logFile string
	i18n    *i18n.I18n
	follow  bool
	lines   []string
}

func New(logFile string, i *i18n.I18n) Model {
	return Model{logFile: logFile, i18n: i}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	s := "  " + m.i18n.T("logs.title") + "\n"
	s += "  ─────────────────────\n"
	s += "  [f] " + m.i18n.T("logs.follow") + "  [/] " + m.i18n.T("logs.search") + "  [p] " + m.i18n.T("logs.prune") + "\n"
	if m.follow {
		s += "  Following: " + m.logFile + "\n"
	} else {
		s += "  Log: " + m.logFile + "\n"
	}
	return s
}
