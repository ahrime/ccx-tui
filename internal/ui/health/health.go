package health

import (
	"context"
	"fmt"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/config"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/BenedictKing/ccx-tui/internal/process"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	client  *client.APIClient
	process *process.Manager
	paths   config.Paths
	i18n    *i18n.I18n
	results []string

	runningDoctor bool
	doctorResults []string
	cursor        int
	channels      []channelInfo
	testJobs      map[string]testJob
}

type channelInfo struct {
	id   string
	name string
}

type testJob struct {
	channelID string
	jobID     string
	status    string
	progress  float64
}

type doctorDoneMsg struct {
	results []string
}

type healthCheckMsg struct {
	results []string
}

type capabilityTestMsg struct {
	channelID string
	jobID     string
	status    string
	err       error
}

type pingAllMsg struct {
	results map[string]string
	err     error
}

func New(c *client.APIClient, p *process.Manager, paths config.Paths, i *i18n.I18n) Model {
	return Model{
		client:   c,
		process:  p,
		paths:    paths,
		i18n:     i,
		testJobs: make(map[string]testJob),
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadChannelList
}

func (m Model) loadChannelList() tea.Msg {
	if m.client == nil {
		return healthCheckMsg{}
	}
	var chs []channelInfo
	for _, chType := range client.AllChannelTypes {
		list, err := m.client.ListChannels(context.Background(), chType)
		if err != nil {
			continue
		}
		for _, ch := range list {
			chs = append(chs, channelInfo{id: ch.ID, name: fmt.Sprintf("%s/%s", chType, ch.Name)})
		}
	}
	return healthCheckMsg{results: nil}
}

func (m Model) runDoctor() tea.Cmd {
	return func() tea.Msg {
		var results []string
		results = append(results, "─ Binary Check")
		if m.paths.BinaryExists() {
			results = append(results, "  ✓ Binary found: "+m.paths.Binary)
		} else {
			results = append(results, "  ✗ Binary not found")
		}

		results = append(results, "─ Process Check")
		if m.process != nil && m.process.IsRunning() {
			pid, _ := m.process.Pid()
			results = append(results, fmt.Sprintf("  ✓ Running (PID: %d)", pid))
		} else {
			results = append(results, "  ✗ Not running")
		}

		results = append(results, "─ API Check")
		if m.client != nil {
			_, err := m.client.HealthCheck(context.Background())
			if err == nil {
				results = append(results, "  ✓ API responding")
			} else {
				results = append(results, fmt.Sprintf("  ✗ API error: %v", err))
			}
		} else {
			results = append(results, "  ✗ No API client")
		}

		results = append(results, "─ Config Check")
		if _, err := config.LoadEnv(m.paths.EnvFile); err != nil {
			results = append(results, fmt.Sprintf("  ✗ Config error: %v", err))
		} else {
			results = append(results, "  ✓ Config valid")
		}

		return doctorDoneMsg{results: results}
	}
}

func (m Model) pingAll() tea.Cmd {
	return func() tea.Msg {
		results := make(map[string]string)
		for _, chType := range client.AllChannelTypes {
			resp, err := m.client.PingAll(context.Background(), chType)
			if err != nil {
				results[string(chType)] = fmt.Sprintf("error: %v", err)
			} else {
				results[string(chType)] = fmt.Sprintf("%v", resp)
			}
		}
		return pingAllMsg{results: results}
	}
}

func (m Model) startTest(chID string) tea.Cmd {
	return func() tea.Msg {
		for _, chType := range client.AllChannelTypes {
			resp, err := m.client.StartCapabilityTest(context.Background(), chType, chID)
			if err != nil {
				continue
			}
			if jobID, ok := resp["jobId"]; ok {
				return capabilityTestMsg{channelID: chID, jobID: fmt.Sprintf("%v", jobID), status: "running"}
			}
		}
		return capabilityTestMsg{channelID: chID, status: "error", err: fmt.Errorf("failed to start test")}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doctorDoneMsg:
		m.runningDoctor = false
		m.doctorResults = msg.results
		return m, nil

	case healthCheckMsg:
		return m, nil

	case capabilityTestMsg:
		if msg.err != nil {
			m.results = append(m.results, fmt.Sprintf("✗ %s: %v", msg.channelID, msg.err))
		} else {
			job := testJob{channelID: msg.channelID, jobID: msg.jobID, status: msg.status}
			m.testJobs[msg.channelID] = job
			m.results = append(m.results, fmt.Sprintf("● %s: test %s (job: %s)", msg.channelID, msg.status, msg.jobID))
		}
		return m, nil

	case pingAllMsg:
		for k, v := range msg.results {
			m.results = append(m.results, fmt.Sprintf("ping %s: %s", k, v))
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			m.runningDoctor = true
			m.doctorResults = nil
			return m, m.runDoctor()
		case "p":
			return m, m.pingAll()
		case "j", "down":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "c":
			if len(m.results) > 0 {
				m.results = nil
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := lipgloss.NewStyle().Bold(true).Render("  " + m.i18n.T("health.title") + " / " + m.i18n.T("health.doctor"))
	s += "\n  ─────────────────────────────────────────────\n"

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

	if m.runningDoctor {
		s += "\n  Running doctor diagnostics...\n"
	}

	if len(m.doctorResults) > 0 {
		s += "\n"
		for _, r := range m.doctorResults {
			s += "  " + r + "\n"
		}
	}

	if len(m.results) > 0 {
		s += "\n  Results\n"
		visible := m.results
		maxVisible := 10
		start := 0
		if len(visible) > maxVisible {
			start = len(visible) - maxVisible
		}
		for i := start; i < len(visible); i++ {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}
			s += fmt.Sprintf("  %s%s\n", cursor, visible[i])
		}
	}

	if len(m.testJobs) > 0 {
		s += "\n  Active Tests\n"
		for chID, job := range m.testJobs {
			s += fmt.Sprintf("  %s: %s (%s)\n", chID, job.status, job.jobID)
		}
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  d:Doctor  p:PingAll  c:Clear  j/k:Scroll")

	return s
}
