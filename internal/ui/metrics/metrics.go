package metrics

import (
	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	client *client.APIClient
	i18n   *i18n.I18n
}

func New(c *client.APIClient, i *i18n.I18n) Model {
	return Model{client: c, i18n: i}
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	return "  Metrics (coming soon)\n"
}
