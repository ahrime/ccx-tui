package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int

const (
	viewList viewMode = iota
	viewDetail
	viewAdd
	viewEdit
)

type Model struct {
	width      int
	height     int
	chType     client.ChannelType
	client     *client.APIClient
	i18n       *i18n.I18n
	channels   []client.UpstreamConfig
	cursor     int
	loading    bool
	view       viewMode
	selectedID string
	filter     textinput.Model
	filtering  bool
	statusMsg  string
}

type channelsLoadedMsg struct {
	channels []client.UpstreamConfig
	err      error
}

type channelActionMsg struct {
	action string
	err    error
}

func New(chType client.ChannelType, c *client.APIClient, i *i18n.I18n) Model {
	fi := textinput.New()
	fi.Placeholder = "filter..."
	fi.CharLimit = 50
	fi.Width = 20
	return Model{
		chType: chType,
		client: c,
		i18n:   i,
		filter: fi,
		view:   viewList,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadChannels
}

func (m Model) loadChannels() tea.Msg {
	if m.client == nil {
		return channelsLoadedMsg{err: fmt.Errorf("not connected")}
	}
	chs, err := m.client.ListChannels(context.Background(), m.chType)
	return channelsLoadedMsg{channels: chs, err: err}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case channelsLoadedMsg:
		if msg.err == nil {
			m.channels = msg.channels
		} else {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		}
		m.loading = false
		return m, nil

	case channelActionMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("%s failed: %v", msg.action, msg.err)
		} else {
			m.statusMsg = fmt.Sprintf("%s success", msg.action)
		}
		return m, m.loadChannels

	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterInput(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleFilterInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filtering = false
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	visible := m.visibleChannels()
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(visible)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if len(visible) > 0 && m.cursor < len(visible) {
			m.selectedID = visible[m.cursor].ID
			m.view = viewDetail
		}
	case "esc":
		if m.view != viewList {
			m.view = viewList
			m.statusMsg = ""
		}
	case "a":
		m.view = viewAdd
		m.statusMsg = ""
	case "e":
		if m.view == viewDetail {
			m.view = viewEdit
		}
	case "d":
		if m.view == viewDetail && m.selectedID != "" {
			return m, m.deleteChannel()
		}
	case "s":
		if m.view == viewDetail && m.selectedID != "" {
			return m, m.toggleChannelStatus()
		}
	case "p":
		if m.view == viewDetail && m.selectedID != "" {
			return m, m.pingChannel()
		}
	case "/":
		m.filtering = true
		m.filter.Focus()
	}
	return m, nil
}

func (m Model) visibleChannels() []client.UpstreamConfig {
	f := strings.ToLower(m.filter.Value())
	if f == "" {
		return m.activeChannels()
	}
	var filtered []client.UpstreamConfig
	for _, ch := range m.activeChannels() {
		if strings.Contains(strings.ToLower(ch.Name), f) {
			filtered = append(filtered, ch)
		}
	}
	return filtered
}

func (m Model) activeChannels() []client.UpstreamConfig {
	var active []client.UpstreamConfig
	for _, ch := range m.channels {
		if ch.Status != "disabled" {
			active = append(active, ch)
		}
	}
	return active
}

func (m Model) disabledChannels() []client.UpstreamConfig {
	var disabled []client.UpstreamConfig
	for _, ch := range m.channels {
		if ch.Status == "disabled" {
			disabled = append(disabled, ch)
		}
	}
	return disabled
}

func (m Model) selectedChannel() *client.UpstreamConfig {
	for i := range m.channels {
		if m.channels[i].ID == m.selectedID {
			return &m.channels[i]
		}
	}
	return nil
}

func (m Model) deleteChannel() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.DeleteChannel(context.Background(), m.chType, m.selectedID)
		return channelActionMsg{action: "delete", err: err}
	}
}

func (m Model) toggleChannelStatus() tea.Cmd {
	return func() tea.Msg {
		ch := m.selectedChannel()
		if ch == nil {
			return channelActionMsg{action: "toggle", err: fmt.Errorf("channel not found")}
		}
		newStatus := "disabled"
		if ch.Status == "active" {
			newStatus = "suspended"
		} else if ch.Status == "suspended" {
			newStatus = "active"
		}
		_, err := m.client.SetChannelStatus(context.Background(), m.chType, m.selectedID, newStatus)
		return channelActionMsg{action: "toggle", err: err}
	}
}

func (m Model) pingChannel() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.PingChannel(context.Background(), m.chType, m.selectedID)
		return channelActionMsg{action: "ping", err: err}
	}
}

func (m Model) View() string {
	if m.loading {
		return "  Loading channels..."
	}

	switch m.view {
	case viewDetail:
		return m.viewDetail()
	case viewAdd:
		return m.viewAddForm()
	case viewEdit:
		return m.viewEditForm()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	visible := m.visibleChannels()
	disabled := m.disabledChannels()

	s := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("  %s Channels", strings.Title(string(m.chType))))
	s += "\n  ─────────────────────────────────────────────\n"

	if len(visible) == 0 && len(disabled) == 0 {
		s += "\n  No channels configured. Press 'a' to add.\n"
		return s
	}

	if len(visible) > 0 {
		s += "  Failover Sequence\n"
		for i, ch := range visible {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}
			icon := statusIcon(ch.Status)
			s += fmt.Sprintf("  %s%s %-28s  P:%-3d  %s\n", cursor, icon, ch.Name, ch.Priority, ch.Status)
		}
	}

	if len(disabled) > 0 {
		s += "\n  Disabled Pool\n"
		for _, ch := range disabled {
			icon := statusIcon(ch.Status)
			s += fmt.Sprintf("    %s %-28s  %s\n", icon, ch.Name, ch.Status)
		}
	}

	if m.filtering {
		s += fmt.Sprintf("\n  Filter: %s", m.filter.View())
	} else if m.filter.Value() != "" {
		s += fmt.Sprintf("\n  Filter: %s  (press / to edit)", m.filter.Value())
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n\n" + helpStyle.Render("  a:Add  enter:Detail  /:Filter  ?:Help")

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg
	}
	return s
}

func (m Model) viewDetail() string {
	ch := m.selectedChannel()
	if ch == nil {
		return "  Channel not found. Press esc to go back."
	}

	s := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("  %s — %s", ch.Name, ch.Status))
	s += "\n  ─────────────────────────────────────────────\n"

	labelStyle := lipgloss.NewStyle().Faint(true).Width(16)
	valueStyle := lipgloss.NewStyle().Bold(true)

	s += fmt.Sprintf("  %s %s\n", labelStyle.Render(m.i18n.T("channel.name")), valueStyle.Render(ch.Name))
	s += fmt.Sprintf("  %s %s\n", labelStyle.Render(m.i18n.T("detail.baseurl")), ch.BaseURL)
	if len(ch.BaseURLs) > 0 {
		s += fmt.Sprintf("  %s %s\n", labelStyle.Render("Alt URLs"), strings.Join(ch.BaseURLs, ", "))
	}
	s += fmt.Sprintf("  %s %s\n", labelStyle.Render(m.i18n.T("detail.service_type")), ch.ServiceType)
	s += fmt.Sprintf("  %s %s\n", labelStyle.Render(m.i18n.T("channel.status")), statusIcon(ch.Status)+" "+ch.Status)
	s += fmt.Sprintf("  %s %d\n", labelStyle.Render(m.i18n.T("channel.priority")), ch.Priority)

	if ch.PromotionUntil != nil {
		s += fmt.Sprintf("  %s until %s\n", labelStyle.Render(m.i18n.T("detail.promotion")), ch.PromotionUntil.Format("2006-01-02 15:04"))
	}

	s += fmt.Sprintf("\n  %s (%d)\n", m.i18n.T("detail.api_keys"), len(ch.APIKeys))
	for i, k := range ch.APIKeys {
		masked := maskKey(k)
		s += fmt.Sprintf("    %d. %s\n", i+1, masked)
	}

	if len(ch.ModelMapping) > 0 {
		s += fmt.Sprintf("\n  %s\n", m.i18n.T("detail.model_mapping"))
		for k, v := range ch.ModelMapping {
			s += fmt.Sprintf("    %s → %s\n", k, v)
		}
	}

	if len(ch.SupportedModels) > 0 {
		s += fmt.Sprintf("\n  %s (%d)\n", m.i18n.T("detail.supported_models"), len(ch.SupportedModels))
		for _, mod := range ch.SupportedModels {
			s += fmt.Sprintf("    %s\n", mod)
		}
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  e:Edit  d:Delete  s:Toggle  p:Ping  t:Test  m:Promote  esc:Back")

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg
	}
	return s
}

func (m Model) viewAddForm() string {
	s := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("  Add %s Channel", strings.Title(string(m.chType))))
	s += "\n  ─────────────────────────────────────────────\n"
	s += "\n  (Form integration with huh coming soon)"
	s += "\n  Press esc to cancel"
	return s
}

func (m Model) viewEditForm() string {
	ch := m.selectedChannel()
	if ch == nil {
		return "  Channel not found. Press esc to go back."
	}
	s := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("  Edit: %s", ch.Name))
	s += "\n  ─────────────────────────────────────────────\n"
	s += "\n  (Form integration with huh coming soon)"
	s += "\n  Press esc to cancel"
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

func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func (m Model) Bindings() []key.Binding { return nil }
