package settings

import (
	"context"
	"fmt"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	client       *client.APIClient
	i18n         *i18n.I18n
	fuzzyMode    bool
	stripBilling bool
	loaded       bool
	statusMsg    string
}

type settingsLoadedMsg struct {
	fuzzyMode    bool
	stripBilling bool
	err          error
}

type settingToggledMsg struct {
	setting string
	value   bool
	err     error
}

func New(c *client.APIClient, i *i18n.I18n) Model {
	return Model{client: c, i18n: i}
}

func (m Model) Init() tea.Cmd {
	return m.loadSettings
}

func (m Model) loadSettings() tea.Msg {
	if m.client == nil {
		return settingsLoadedMsg{err: fmt.Errorf("not connected")}
	}
	var fuzzy, strip bool
	resp, err := m.client.GetFuzzyMode(context.Background())
	if err == nil {
		if v, ok := resp["enabled"]; ok {
			fuzzy = fmt.Sprintf("%v", v) == "true"
		}
	}
	resp2, err2 := m.client.GetStripBillingHeader(context.Background())
	if err2 == nil {
		if v, ok := resp2["enabled"]; ok {
			strip = fmt.Sprintf("%v", v) == "true"
		}
	}
	return settingsLoadedMsg{fuzzyMode: fuzzy, stripBilling: strip, err: err}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case settingsLoadedMsg:
		if msg.err == nil {
			m.fuzzyMode = msg.fuzzyMode
			m.stripBilling = msg.stripBilling
			m.loaded = true
		} else {
			m.statusMsg = fmt.Sprintf("load error: %v", msg.err)
		}
		return m, nil

	case settingToggledMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("%s toggle failed: %v", msg.setting, msg.err)
		} else {
			if msg.setting == "fuzzy" {
				m.fuzzyMode = msg.value
			} else {
				m.stripBilling = msg.value
			}
			m.statusMsg = fmt.Sprintf("%s → %v", msg.setting, msg.value)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			newVal := !m.fuzzyMode
			return m, m.toggleFuzzy(newVal)
		case "b":
			newVal := !m.stripBilling
			return m, m.toggleStripBilling(newVal)
		case "r":
			return m, m.loadSettings
		}
	}
	return m, nil
}

func (m Model) toggleFuzzy(val bool) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.SetFuzzyMode(context.Background(), val)
		return settingToggledMsg{setting: "fuzzy", value: val, err: err}
	}
}

func (m Model) toggleStripBilling(val bool) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.SetStripBillingHeader(context.Background(), val)
		return settingToggledMsg{setting: "strip-billing", value: val, err: err}
	}
}

func (m Model) View() string {
	s := lipgloss.NewStyle().Bold(true).Render("  " + m.i18n.T("settings.title"))
	s += "\n  ─────────────────────────────────────────────\n"

	labelStyle := lipgloss.NewStyle().Width(24)

	fuzzyIcon := "○"
	fuzzyColor := lipgloss.Color("9")
	if m.fuzzyMode {
		fuzzyIcon = "●"
		fuzzyColor = lipgloss.Color("42")
	}
	s += fmt.Sprintf("  %s %s  %s  [f] toggle\n",
		lipgloss.NewStyle().Foreground(fuzzyColor).Render(fuzzyIcon),
		labelStyle.Render(m.i18n.T("settings.fuzzy_mode")),
		fmt.Sprintf("%v", m.fuzzyMode))

	stripIcon := "○"
	stripColor := lipgloss.Color("9")
	if m.stripBilling {
		stripIcon = "●"
		stripColor = lipgloss.Color("42")
	}
	s += fmt.Sprintf("  %s %s  %s  [b] toggle\n",
		lipgloss.NewStyle().Foreground(stripColor).Render(stripIcon),
		labelStyle.Render(m.i18n.T("settings.strip_billing")),
		fmt.Sprintf("%v", m.stripBilling))

	helpStyle := lipgloss.NewStyle().Faint(true)
	s += "\n" + helpStyle.Render("  f:Fuzzy  b:StripBilling  r:Reload")

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg
	}
	return s
}
