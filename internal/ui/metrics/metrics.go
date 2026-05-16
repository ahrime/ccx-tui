package metrics

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	client      *client.APIClient
	i18n        *i18n.I18n
	chType      client.ChannelType
	stats       []channelStat
	globalStats *globalStat
	history     []historyPoint
	loading     bool
	cursor      int
}

type channelStat struct {
	name         string
	totalReq     int64
	successCount int64
	failCount    int64
	successRate  float64
	avgLatency   float64
}

type globalStat struct {
	totalReq    int64
	totalOK     int64
	totalFail   int64
	activeCh    int
	totalCh     int
}

type historyPoint struct {
	time    time.Time
	success int64
	fail    int64
	latency float64
}

type metricsLoadedMsg struct {
	stats   []channelStat
	global  *globalStat
	history []historyPoint
	err     error
}

func New(c *client.APIClient, i *i18n.I18n) Model {
	return Model{client: c, i18n: i, chType: client.ChannelTypeMessages}
}

func (m Model) Init() tea.Cmd {
	return m.loadMetrics
}

func (m Model) loadMetrics() tea.Msg {
	if m.client == nil {
		return metricsLoadedMsg{err: fmt.Errorf("not connected")}
	}

	var stats []channelStat
	resp, err := m.client.GetChannelMetrics(context.Background(), m.chType)
	if err != nil {
		return metricsLoadedMsg{err: err}
	}

	if channelsRaw, ok := resp["channels"]; ok {
		if chArr, ok := channelsRaw.([]interface{}); ok {
			for _, ch := range chArr {
				if m, ok := ch.(map[string]interface{}); ok {
					s := channelStat{
						name:         strVal(m, "name"),
						totalReq:     int64Val(m, "totalRequests"),
						successCount: int64Val(m, "successCount"),
						failCount:    int64Val(m, "failCount"),
						successRate:  float64Val(m, "successRate"),
						avgLatency:   float64Val(m, "avgLatencyMs"),
					}
					stats = append(stats, s)
				}
			}
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].totalReq > stats[j].totalReq
	})

	histResp, _ := m.client.GetMetricsHistory(context.Background(), m.chType)
	var history []historyPoint
	if histResp != nil {
		if points, ok := histResp["points"]; ok {
			if arr, ok := points.([]interface{}); ok {
				for _, p := range arr {
					if pm, ok := p.(map[string]interface{}); ok {
						history = append(history, historyPoint{
							success: int64Val(pm, "success"),
							fail:    int64Val(pm, "fail"),
							latency: float64Val(pm, "avgLatencyMs"),
						})
					}
				}
			}
		}
	}

	return metricsLoadedMsg{stats: stats, history: history}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case metricsLoadedMsg:
		if msg.err == nil {
			m.stats = msg.stats
			m.globalStats = msg.global
			m.history = msg.history
		}
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.stats)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "r":
			m.loading = true
			return m, m.loadMetrics
		case "1":
			m.chType = client.ChannelTypeMessages
			return m, m.loadMetrics
		case "2":
			m.chType = client.ChannelTypeChat
			return m, m.loadMetrics
		case "3":
			m.chType = client.ChannelTypeResponses
			return m, m.loadMetrics
		case "4":
			m.chType = client.ChannelTypeGemini
			return m, m.loadMetrics
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := lipgloss.NewStyle().Bold(true).Render("  Metrics — " + string(m.chType))
	s += "\n  ─────────────────────────────────────────────\n"

	if m.loading {
		s += "\n  Loading..."
		return s
	}

	if len(m.stats) == 0 {
		s += "\n  No metrics data. Press 'r' to refresh.\n"
	} else {
		headerStyle := lipgloss.NewStyle().Bold(true).Faint(true)
		s += fmt.Sprintf("  %-28s %10s %8s %8s %10s\n",
			headerStyle.Render("Channel"),
			headerStyle.Render("Requests"),
			headerStyle.Render("Success"),
			headerStyle.Render("Fail"),
			headerStyle.Render("Latency"))
		s += "  " + strings.Repeat("─", 68) + "\n"

		for i, st := range m.stats {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}
			rateColor := lipgloss.Color("42")
			if st.successRate < 0.9 {
				rateColor = lipgloss.Color("202")
			}
			if st.successRate < 0.5 {
				rateColor = lipgloss.Color("9")
			}
			s += fmt.Sprintf("  %s%-28s %10d %8d %8d %s\n",
				cursor, st.name, st.totalReq, st.successCount, st.failCount,
				lipgloss.NewStyle().Foreground(rateColor).Render(fmt.Sprintf("%8.1fms", st.avgLatency)))
		}
	}

	if len(m.history) > 0 {
		s += "\n  Success/Fail Trend\n"
		s += "  " + renderSparkline(m.history, 60) + "\n"
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  1-4:Protocol  r:Refresh  j/k:Nav  ↑↓:Scroll")

	return s
}

func renderSparkline(points []historyPoint, width int) string {
	if len(points) == 0 {
		return ""
	}
	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var maxVal int64
	for _, p := range points {
		total := p.success + p.fail
		if total > maxVal {
			maxVal = total
		}
	}
	if maxVal == 0 {
		return strings.Repeat(string(bars[0]), width)
	}

	step := float64(len(points)) / float64(width)
	var sb strings.Builder
	for i := 0; i < width; i++ {
		idx := int(float64(i) * step)
		if idx >= len(points) {
			idx = len(points) - 1
		}
		total := points[idx].success + points[idx].fail
		bucket := int(float64(total) / float64(maxVal) * float64(len(bars)-1))
		if bucket >= len(bars) {
			bucket = len(bars) - 1
		}
		sb.WriteRune(bars[bucket])
	}
	return sb.String()
}

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func int64Val(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int64:
			return n
		}
	}
	return 0
}

func float64Val(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		}
	}
	return 0
}
