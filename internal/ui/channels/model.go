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
	viewAPIKeys
)

type formField struct {
	label    string
	input    textinput.Model
	isSelect bool
	options  []string
	selIdx   int
	isHidden bool
}

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
	keyCursor  int

	fields    []formField
	fieldIdx  int
	isEdit    bool
}

var serviceTypes = []string{"openai", "anthropic", "azure", "gemini", "cohere"}

type channelsLoadedMsg struct {
	chType   client.ChannelType
	channels []client.UpstreamConfig
	err      error
}

type channelActionMsg struct {
	action string
	err    error
}

type formResultMsg struct {
	err error
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
		return channelsLoadedMsg{chType: m.chType, err: fmt.Errorf("not connected")}
	}
	chs, err := m.client.ListChannels(context.Background(), m.chType)
	return channelsLoadedMsg{chType: m.chType, channels: chs, err: err}
}

func newTextInput(placeholder string, hidden bool) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 200
	ti.Width = 40
	if hidden {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}
	return ti
}

func (m *Model) initAddFields() {
	m.isEdit = false
	m.fieldIdx = 0
	m.fields = []formField{
		{label: "Name", input: newTextInput("e.g. my-openai", false)},
		{label: "Base URL", input: newTextInput("https://api.openai.com/v1", false)},
		{label: "Service Type", isSelect: true, options: serviceTypes, selIdx: 0},
		{label: "API Key", input: newTextInput("sk-...", true), isHidden: true},
		{label: "Description", input: newTextInput("optional description", false)},
	}
	m.fields[0].input.Focus()
}

func (m *Model) initEditFields() {
	ch := m.selectedChannel()
	if ch == nil {
		return
	}
	m.isEdit = true
	m.fieldIdx = 0
	nameInput := newTextInput("channel name", false)
	nameInput.SetValue(ch.Name)
	urlInput := newTextInput("https://...", false)
	urlInput.SetValue(ch.BaseURL)
	descInput := newTextInput("optional", false)
	descInput.SetValue(ch.Description)
	proxyInput := newTextInput("optional", false)
	proxyInput.SetValue(ch.ProxyURL)
	selIdx := 0
	for i, st := range serviceTypes {
		if st == ch.ServiceType {
			selIdx = i
			break
		}
	}
	m.fields = []formField{
		{label: "Name", input: nameInput},
		{label: "Base URL", input: urlInput},
		{label: "Service Type", isSelect: true, options: serviceTypes, selIdx: selIdx},
		{label: "Description", input: descInput},
		{label: "Proxy URL", input: proxyInput},
	}
	m.fields[0].input.Focus()
}

func (m Model) formValues() (name, baseURL, serviceType, apiKey, description, proxyURL string) {
	for _, f := range m.fields {
		switch f.label {
		case "Name":
			name = f.input.Value()
		case "Base URL":
			baseURL = f.input.Value()
		case "Service Type":
			if f.isSelect && f.selIdx < len(f.options) {
				serviceType = f.options[f.selIdx]
			}
		case "API Key":
			apiKey = f.input.Value()
		case "Description":
			description = f.input.Value()
		case "Proxy URL":
			proxyURL = f.input.Value()
		}
	}
	return
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case channelsLoadedMsg:
		if msg.chType != m.chType {
			return m, nil
		}
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

	case formResultMsg:
		m.fields = nil
		m.view = viewList
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("✗ form error: %v", msg.err)
		} else {
			m.statusMsg = "✓ Channel saved"
		}
		return m, m.loadChannels

	case tea.KeyMsg:
		if m.view == viewAdd || m.view == viewEdit {
			return m.handleFormKey(msg)
		}
		if m.filtering {
			return m.handleFilterInput(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleFormKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	curField := m.fields[m.fieldIdx]

	switch msg.String() {
	case "esc":
		m.view = viewList
		m.fields = nil
		return m, nil
	case "up":
		if curField.isSelect {
			f := &m.fields[m.fieldIdx]
			if f.selIdx > 0 {
				f.selIdx--
			}
		} else {
			m.focusField(m.fieldIdx - 1)
		}
		return m, nil
	case "down":
		if curField.isSelect {
			f := &m.fields[m.fieldIdx]
			if f.selIdx < len(f.options)-1 {
				f.selIdx++
			}
		} else {
			m.focusField(m.fieldIdx + 1)
		}
		return m, nil
	case "tab":
		m.focusField(m.fieldIdx + 1)
		return m, nil
	case "shift+tab":
		m.focusField(m.fieldIdx - 1)
		return m, nil
	case "enter":
		if curField.isSelect {
			m.focusField(m.fieldIdx + 1)
			return m, nil
		}
		if m.fieldIdx == len(m.fields)-1 {
			return m, m.submitForm()
		}
		m.focusField(m.fieldIdx + 1)
		return m, nil
	}

	if curField.isSelect {
		switch msg.String() {
		case "j", "k":
			return m, nil
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.fields[m.fieldIdx].input, cmd = m.fields[m.fieldIdx].input.Update(msg)
	return m, cmd
}

func (m *Model) focusField(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.fields) {
		idx = len(m.fields) - 1
	}
	if !m.fields[m.fieldIdx].isSelect {
		m.fields[m.fieldIdx].input.Blur()
	}
	m.fieldIdx = idx
	if !m.fields[m.fieldIdx].isSelect {
		m.fields[m.fieldIdx].input.Focus()
	}
}

func (m Model) submitForm() tea.Cmd {
	name, baseURL, serviceType, apiKey, desc, proxyURL := m.formValues()
	isEdit := m.isEdit
	selectedID := m.selectedID
	chType := m.chType
	cli := m.client
	return func() tea.Msg {
		if cli == nil {
			return formResultMsg{err: fmt.Errorf("not connected")}
		}
		ch := client.UpstreamConfig{
			Name:        name,
			BaseURL:     baseURL,
			ServiceType: serviceType,
			Description: desc,
			ProxyURL:    proxyURL,
		}
		if apiKey != "" {
			ch.APIKeys = []string{apiKey}
		}
		if !isEdit {
			_, err := cli.AddChannel(context.Background(), chType, ch)
			return formResultMsg{err: err}
		}
		ch.ID = selectedID
		_, err := cli.UpdateChannel(context.Background(), chType, selectedID, ch)
		return formResultMsg{err: err}
	}
}

func (m Model) handleFilterInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filtering = false
		return m, nil
	default:
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	visible := m.visibleChannels()
	switch m.view {
	case viewAPIKeys:
		return m.handleAPIKeysKey(msg)
	case viewDetail:
		return m.handleDetailKey(msg)
	default:
	}

	switch msg.String() {
	case "j", "down":
		if m.cursor < len(visible)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "J":
		if m.cursor < len(visible)-1 && len(visible) > 1 {
			i := m.cursor
			j := i + 1
			visible[i], visible[j] = visible[j], visible[i]
			m.cursor = j
			return m, m.reorderChannels(visible)
		}
	case "K":
		if m.cursor > 0 && len(visible) > 1 {
			i := m.cursor
			j := i - 1
			visible[i], visible[j] = visible[j], visible[i]
			m.cursor = j
			return m, m.reorderChannels(visible)
		}
	case "enter":
		if len(visible) > 0 && m.cursor < len(visible) {
			m.selectedID = visible[m.cursor].ID
			m.view = viewDetail
		}
	case "a":
		m.view = viewAdd
		m.statusMsg = ""
		m.initAddFields()
	case "/":
		m.filtering = true
		m.filter.Focus()
	}
	return m, nil
}

func (m Model) handleDetailKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewList
		m.statusMsg = ""
	case "e":
		m.view = viewEdit
		m.initEditFields()
	case "d":
		if m.selectedID != "" {
			return m, m.deleteChannel()
		}
	case "s":
		if m.selectedID != "" {
			return m, m.toggleChannelStatus()
		}
	case "p":
		if m.selectedID != "" {
			return m, m.pingChannel()
		}
	case "k":
		m.view = viewAPIKeys
		m.keyCursor = 0
	}
	return m, nil
}

func (m Model) handleAPIKeysKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	ch := m.selectedChannel()
	if ch == nil {
		m.view = viewDetail
		return m, nil
	}
	keys := ch.APIKeys
	switch msg.String() {
	case "esc":
		m.view = viewDetail
	case "j", "down":
		if m.keyCursor < len(keys)-1 {
			m.keyCursor++
		}
	case "k", "up":
		if m.keyCursor > 0 {
			m.keyCursor--
		}
	case "a":
		m.statusMsg = "Enter API key in status bar (placeholder)"
	case "d":
		if m.keyCursor < len(keys) {
			return m, m.deleteAPIKey(keys[m.keyCursor])
		}
	case "t":
		if m.keyCursor < len(keys) {
			return m, m.moveKeyTop(keys[m.keyCursor])
		}
	case "b":
		if m.keyCursor < len(keys) {
			return m, m.moveKeyBottom(keys[m.keyCursor])
		}
	}
	return m, nil
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

func (m Model) reorderChannels(visible []client.UpstreamConfig) tea.Cmd {
	order := make([]string, len(visible))
	for i, ch := range visible {
		order[i] = ch.ID
	}
	return func() tea.Msg {
		_, err := m.client.ReorderChannels(context.Background(), m.chType, order)
		return channelActionMsg{action: "reorder", err: err}
	}
}

func (m Model) deleteAPIKey(key string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.DeleteAPIKey(context.Background(), m.chType, m.selectedID, key)
		return channelActionMsg{action: "delete key", err: err}
	}
}

func (m Model) moveKeyTop(key string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.MoveKeyToTop(context.Background(), m.chType, m.selectedID, key)
		return channelActionMsg{action: "key top", err: err}
	}
}

func (m Model) moveKeyBottom(key string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.MoveKeyToBottom(context.Background(), m.chType, m.selectedID, key)
		return channelActionMsg{action: "key bottom", err: err}
	}
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

func (m Model) View() string {
	if m.loading {
		return "  Loading channels..."
	}
	switch m.view {
	case viewDetail:
		return m.viewDetail()
	case viewAdd, viewEdit:
		return m.viewForm()
	case viewAPIKeys:
		return m.viewAPIKeys()
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
	s += "\n\n" + helpStyle.Render("  a:Add  enter:Detail  /:Filter  J/K:Reorder  ?:Help")

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

	s += fmt.Sprintf("\n  %s (%d)  [k] Manage\n", m.i18n.T("detail.api_keys"), len(ch.APIKeys))

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
	s += "\n" + helpStyle.Render("  e:Edit  d:Delete  s:Toggle  p:Ping  k:Keys  esc:Back")

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg
	}
	return s
}

func (m Model) viewForm() string {
	if len(m.fields) == 0 {
		return "  Form error. Press esc."
	}

	title := "Add Channel"
	if m.isEdit {
		title = "Edit Channel"
	}
	title = fmt.Sprintf("%s %s", title, strings.Title(string(m.chType)))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	labelW := 14
	inputW := 40

	var rows []string
	for i, f := range m.fields {
		isActive := i == m.fieldIdx

		labelStyle := lipgloss.NewStyle().Width(labelW).Bold(isActive)
		if isActive {
			labelStyle = labelStyle.Foreground(lipgloss.Color("86"))
		} else {
			labelStyle = labelStyle.Faint(true)
		}

		var value string
		if f.isSelect {
			value = f.options[f.selIdx]
		} else {
			value = f.input.View()
		}

		prefix := "  "
		suffix := ""
		if isActive {
			prefix = "▸ "
			suffix = " ◂"
		}

		if f.isSelect && isActive {
			row := fmt.Sprintf("%s%s %s", prefix, labelStyle.Render(f.label), value)
			rows = append(rows, row)
			selectStyle := lipgloss.NewStyle().PaddingLeft(labelW + 4)
			for oi, opt := range f.options {
				marker := "○"
				if oi == f.selIdx {
					marker = "●"
				}
				rows = append(rows, selectStyle.Render(fmt.Sprintf("  %s %s", marker, opt)))
			}
		} else {
			inputStyle := lipgloss.NewStyle().Width(inputW)
			row := fmt.Sprintf("%s%s %s%s", prefix, labelStyle.Render(f.label), inputStyle.Render(value), suffix)
			rows = append(rows, row)
		}
	}

	content := strings.Join(rows, "\n")

	helpStyle := lipgloss.NewStyle().Faint(true)
	var help string
	if m.fieldIdx == len(m.fields)-1 {
		help = helpStyle.Render("  ↑↓: Navigate  Tab: Next  Enter: Submit  Esc: Cancel")
	} else {
		help = helpStyle.Render("  ↑↓: Navigate  Tab: Next  Enter: Next  Esc: Cancel")
	}

	box := borderStyle.Render(lipgloss.NewStyle().Bold(true).Render("  "+title) + "\n\n" + content)

	if m.statusMsg != "" {
		return box + "\n  " + m.statusMsg + "\n" + help
	}
	return box + "\n" + help
}

func (m Model) viewAPIKeys() string {
	ch := m.selectedChannel()
	if ch == nil {
		return "  Channel not found. Press esc to go back."
	}

	title := fmt.Sprintf("  API Keys — %s", ch.Name)
	s := lipgloss.NewStyle().Bold(true).Render(title)
	s += "\n  ─────────────────────────────────────────────\n"

	if len(ch.APIKeys) == 0 {
		s += "\n  No API keys. Press 'a' to add.\n"
	} else {
		for i, k := range ch.APIKeys {
			cursor := "  "
			if i == m.keyCursor {
				cursor = "▸ "
			}
			s += fmt.Sprintf("  %s%d. %s\n", cursor, i+1, maskKey(k))
		}
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  a:Add  d:Delete  t:Top  b:Bottom  esc:Back")

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg
	}
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

func (m Model) IsFormActive() bool {
	return (m.view == viewAdd || m.view == viewEdit) && len(m.fields) > 0
}

func (m Model) IsSubViewActive() bool {
	return m.view != viewList
}
