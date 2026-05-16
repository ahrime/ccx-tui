package settings

import (
	"fmt"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	client        *client.APIClient
	i18n          *i18n.I18n
	fuzzyMode     bool
	stripBilling  bool
}

func New(c *client.APIClient, i *i18n.I18n) Model {
	return Model{client: c, i18n: i}
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	s := "  " + m.i18n.T("settings.title") + "\n"
	s += "  ─────────────────────\n"
	fuzzyIcon := "○"
	if m.fuzzyMode {
		fuzzyIcon = "●"
	}
	s += fmt.Sprintf("  %s %s  [f] toggle\n", fuzzyIcon, m.i18n.T("settings.fuzzy_mode"))
	stripIcon := "○"
	if m.stripBilling {
		stripIcon = "●"
	}
	s += fmt.Sprintf("  %s %s  [b] toggle\n", stripIcon, m.i18n.T("settings.strip_billing"))
	return s
}
