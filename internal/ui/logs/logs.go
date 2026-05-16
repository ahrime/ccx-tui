package logs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	logFile   string
	i18n      *i18n.I18n
	follow    bool
	searching bool
	search    textinput.Model
	viewport  viewport.Model
	ready     bool
	width     int
	height    int
	allLines  []string
	cancelFn  context.CancelFunc
}

type logLineMsg struct {
	line string
}

type logReadDoneMsg struct{}

func New(logFile string, i *i18n.I18n) Model {
	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 100
	si.Width = 30
	return Model{
		logFile: logFile,
		i18n:    i,
		search:  si,
	}
}

func (m Model) Init() tea.Cmd {
	return m.readInitialLogs
}

func (m Model) readInitialLogs() tea.Msg {
	f, err := os.Open(m.logFile)
	if err != nil {
		return logReadDoneMsg{}
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > 1000 {
		lines = lines[len(lines)-1000:]
	}
	return logBatchMsg{lines: lines}
}

type logBatchMsg struct {
	lines []string
}

func (m Model) tailLogs() tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(m.logFile)
		if err != nil {
			return logReadDoneMsg{}
		}
		defer f.Close()

		f.Seek(0, 2)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			return logLineMsg{line: scanner.Text()}
		}
		return logReadDoneMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-6)
			m.viewport.Style = lipgloss.NewStyle().Padding(0, 1)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 6
		}

	case logBatchMsg:
		m.allLines = append(m.allLines, msg.lines...)
		m.refreshViewport()
		if m.follow {
			return m, m.tailLogs()
		}

	case logLineMsg:
		m.allLines = append(m.allLines, msg.line)
		if len(m.allLines) > 2000 {
			m.allLines = m.allLines[len(m.allLines)-2000:]
		}
		m.refreshViewport()
		if m.follow {
			return m, m.tailLogs()
		}

	case tea.KeyMsg:
		if m.searching {
			return m.handleSearchInput(msg)
		}
		return m.handleKey(msg)
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleSearchInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searching = false
		m.refreshViewport()
		return m, nil
	case "esc":
		m.searching = false
		m.search.SetValue("")
		m.refreshViewport()
		return m, nil
	default:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		m.follow = !m.follow
		if m.follow {
			return m, m.tailLogs()
		}
		if m.cancelFn != nil {
			m.cancelFn()
			m.cancelFn = nil
		}
	case "/":
		m.searching = true
		m.search.Focus()
	case "p":
		os.Truncate(m.logFile, 0)
		m.allLines = nil
		m.refreshViewport()
	case "r":
		return m, m.readInitialLogs
	}
	return m, nil
}

func (m *Model) refreshViewport() {
	q := strings.ToLower(m.search.Value())
	var visible []string
	for _, l := range m.allLines {
		if q == "" || strings.Contains(strings.ToLower(l), q) {
			visible = append(visible, l)
		}
	}
	if m.ready {
		m.viewport.SetContent(strings.Join(visible, "\n"))
		if m.follow {
			m.viewport.GotoBottom()
		}
	}
}

func (m Model) View() string {
	s := "  " + m.i18n.T("logs.title")
	if m.follow {
		s += " [FOLLOWING]"
	}
	s += "\n  ─────────────────────────────────────────────\n"

	if m.ready {
		s += m.viewport.View()
	} else {
		s += "\n  Log: " + m.logFile + "\n"
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  f:Follow  /:Search  p:Prune  r:Reload  ↑↓:Scroll")

	if m.searching {
		s += "\n  Search: " + m.search.View()
	} else if m.search.Value() != "" {
		s += fmt.Sprintf("\n  Filter: \"%s\"", m.search.Value())
	}
	return s
}
