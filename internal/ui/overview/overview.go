package overview

import (
	"context"
	"fmt"
	"time"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/config"
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
	paths     config.Paths
	i18n      *i18n.I18n
	connected bool
	version   string
	uptime    string
	memory    string
	port      string
	loading   bool
	statusMsg string
}

type healthResultMsg struct {
	connected bool
	version   string
}

type processActionMsg struct {
	action string
	pid    int
	err    error
}

const HEALTH_INTERVAL = 5 * time.Second

type tickHealthMsg struct{}

func New(c *client.APIClient, p *process.Manager, paths config.Paths, i *i18n.I18n) Model {
	return Model{client: c, process: p, paths: paths, i18n: i, loading: true}
}

func (m Model) Init() tea.Cmd {
	return m.checkHealth
}

func (m Model) checkHealth() tea.Msg {
	if m.client == nil {
		return healthResultMsg{connected: false}
	}
	resp, err := m.client.HealthCheck(context.Background())
	if err != nil {
		return healthResultMsg{connected: false}
	}
	ver := ""
	if v, ok := resp["version"]; ok {
		if vm, ok := v.(map[string]interface{}); ok {
			if vs, ok := vm["version"]; ok {
				ver = fmt.Sprintf("%v", vs)
			}
		}
	}
	return healthResultMsg{connected: true, version: ver}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case healthResultMsg:
		m.connected = msg.connected
		m.version = msg.version
		m.loading = false
		return m, tea.Tick(HEALTH_INTERVAL, func(t time.Time) tea.Msg {
			return tickHealthMsg{}
		})

	case tickHealthMsg:
		return m, m.checkHealth
	case processActionMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("%s failed: %v", msg.action, msg.err)
		} else {
			m.statusMsg = fmt.Sprintf("%s success (PID: %d)", msg.action, msg.pid)
		}
		return m, m.checkHealth
	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			return m, m.startProcess()
		case "x":
			return m, m.stopProcess()
		case "r":
			return m, m.restartProcess()
		}
	}
	return m, nil
}

func (m Model) startProcess() tea.Cmd {
	return func() tea.Msg {
		if m.process == nil {
			return processActionMsg{action: "start", err: fmt.Errorf("no process manager")}
		}
		pid, err := m.process.Start()
		return processActionMsg{action: "start", pid: pid, err: err}
	}
}

func (m Model) stopProcess() tea.Cmd {
	return func() tea.Msg {
		if m.process == nil {
			return processActionMsg{action: "stop", err: fmt.Errorf("no process manager")}
		}
		err := m.process.Stop()
		return processActionMsg{action: "stop", err: err}
	}
}

func (m Model) restartProcess() tea.Cmd {
	return func() tea.Msg {
		if m.process == nil {
			return processActionMsg{action: "restart", err: fmt.Errorf("no process manager")}
		}
		pid, err := m.process.Restart()
		return processActionMsg{action: "restart", pid: pid, err: err}
	}
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

	procStatus := "Stopped"
	procIcon := "✗"
	procColor := lipgloss.Color("9")
	if running {
		procStatus = "Running"
		procIcon = "●"
		procColor = lipgloss.Color("42")
	}

	procCard := cardStyle.Render(
		labelStyle.Render("Process")+"\n"+
			lipgloss.NewStyle().Foreground(procColor).Render(procIcon+" ")+
			valueStyle.Render(procStatus),
	)

	binaryCard := cardStyle.Render(
		labelStyle.Render("Binary")+"\n"+
			valueStyle.Render(func() string {
				if m.paths.BinaryExists() {
					return "✓ Found"
				}
				return "✗ Not Found"
			}()),
	)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, statusCard, "  ", procCard, "  ", binaryCard)

	helpStyle := lipgloss.NewStyle().Faint(true)
	if !running {
		row1 += "\n\n" + helpStyle.Render("  [s] Start CCX")
	} else {
		row1 += "\n\n" + helpStyle.Render("  [x] Stop  [r] Restart  [d] Doctor")
	}

	if m.statusMsg != "" {
		row1 += "\n  " + m.statusMsg
	}
	return row1
}

func (m Model) Bindings() []key.Binding { return nil }
