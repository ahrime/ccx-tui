package health

import (
	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/config"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/BenedictKing/ccx-tui/internal/process"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	client  *client.APIClient
	process *process.Manager
	paths   config.Paths
	i18n    *i18n.I18n
	results []string
}

func New(c *client.APIClient, p *process.Manager, paths config.Paths, i *i18n.I18n) Model {
	return Model{client: c, process: p, paths: paths, i18n: i}
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	s := "  " + m.i18n.T("health.title") + " / " + m.i18n.T("health.doctor") + "\n"
	s += "  ─────────────────────\n"

	if m.process != nil && m.process.IsRunning() {
		s += "  ● Process: Running\n"
	} else {
		s += "  ✗ Process: Stopped\n"
	}

	if m.paths.BinaryExists() {
		s += "  ● Binary: Found\n"
	} else {
		s += "  ✗ Binary: Not Found\n"
	}

	s += "\n  [d] Run Doctor  [p] Ping All"
	return s
}
