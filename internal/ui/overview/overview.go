package overview

import (
	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/BenedictKing/ccx-tui/internal/process"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width     int
	height    int
	client    *client.APIClient
	process   *process.Manager
	i18n      *i18n.I18n
	connected bool
	version   string
	uptime    string
	memory    string
	port      string
	loading   bool
}

func New(c *client.APIClient, p *process.Manager, i *i18n.I18n) Model {
	return Model{client: c, process: p, i18n: i}
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
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 2).
		Width(24)

	labelStyle := lipgloss.NewStyle().Faint(true).MarginBottom(1)
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	connIcon := "●"
	connColor := lipgloss.Color("42")
	connText := m.i18n.T("overview.connected")
	if !m.connected {
		connIcon = "○"
		connColor = lipgloss.Color("9")
		connText = m.i18n.T("overview.offline")
	}

	running := false
	if m.process != nil {
		running = m.process.IsRunning()
	}

	statusCard := cardStyle.Render(
		labelStyle.Render(m.i18n.T("overview.version"))+"\n"+
			valueStyle.Render(m.version)+"\n\n"+
			labelStyle.Render("Status")+"\n"+
			lipgloss.NewStyle().Foreground(connColor).Render(connIcon+" ")+
			valueStyle.Render(connText),
	)

	procCard := cardStyle.Render(
		labelStyle.Render("Process")+"\n"+
			valueStyle.Render(func() string {
				if running {
					return "Running"
				}
				return "Stopped"
			}()),
	)

	row := lipgloss.JoinHorizontal(lipgloss.Top, statusCard, "  ", procCard)

	if !running {
		row += "\n\n  [s] Start CCX"
	} else {
		row += "\n\n  [x] Stop  [r] Restart  [d] Doctor"
	}

	return row
}

func (m Model) Bindings() []key.Binding { return nil }
