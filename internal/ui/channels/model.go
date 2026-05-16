package channels

import (
	"fmt"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	width    int
	height   int
	chType   client.ChannelType
	client   *client.APIClient
	i18n     *i18n.I18n
	channels []client.UpstreamConfig
	cursor   int
	loading  bool
}

func New(chType client.ChannelType, c *client.APIClient, i *i18n.I18n) Model {
	return Model{chType: chType, client: c, i18n: i}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return "  Loading channels..."
	}
	if len(m.channels) == 0 {
		return fmt.Sprintf("  No %s channels configured. Press 'a' to add.", m.chType)
	}

	var active, disabled []client.UpstreamConfig
	for _, ch := range m.channels {
		if ch.Status == "disabled" {
			disabled = append(disabled, ch)
		} else {
			active = append(active, ch)
		}
	}

	s := "  Failover Sequence\n"
	s += "  ─────────────────────────────────────────────\n"
	for i, ch := range active {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}
		icon := statusIcon(ch.Status)
		s += fmt.Sprintf("  %s%s %-28s  P:%-3d  %s\n", cursor, icon, ch.Name, ch.Priority, ch.Status)
	}

	if len(disabled) > 0 {
		s += "\n  Disabled Pool\n"
		s += "  ─────────────────────────────────────────────\n"
		for _, ch := range disabled {
			icon := statusIcon(ch.Status)
			s += fmt.Sprintf("    %s %-28s  %s\n", icon, ch.Name, ch.Status)
		}
	}

	s += "\n  a:Add  e:Edit  d:Delete  p:Ping  t:Test  s:Toggle  /:Filter"
	return s
}

func statusIcon(status string) string {
	switch status {
	case "active":
		return "●"
	case "suspended":
		return "⏸"
	default:
		return "✗"
	}
}

func (m Model) Bindings() []key.Binding { return nil }
