package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/BenedictKing/ccx-tui/internal/client"
	"github.com/BenedictKing/ccx-tui/internal/config"
	"github.com/BenedictKing/ccx-tui/internal/i18n"
	"github.com/BenedictKing/ccx-tui/internal/process"
	"github.com/BenedictKing/ccx-tui/internal/ui/channels"
	"github.com/BenedictKing/ccx-tui/internal/ui/health"
	"github.com/BenedictKing/ccx-tui/internal/ui/logs"
	"github.com/BenedictKing/ccx-tui/internal/ui/metrics"
	"github.com/BenedictKing/ccx-tui/internal/ui/more"
	"github.com/BenedictKing/ccx-tui/internal/ui/overview"
	"github.com/BenedictKing/ccx-tui/internal/ui/settings"
	"github.com/BenedictKing/ccx-tui/internal/ui/upgrade"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var Version = "0.1.0-dev"

type App struct {
	width     int
	height    int
	activeTab Tab
	keys      KeyMap
	help      help.Model
	ctxStack  *ContextStack
	client    *client.APIClient
	process   *process.Manager
	env       *config.EnvConfig
	paths     config.Paths
	i18n      *i18n.I18n
	connected bool
	quitting  bool

	overview    overview.Model
	channels    [4]channels.Model
	morePanel   more.Model
	logsPanel   logs.Model
	settings    settings.Model
	health      health.Model
	upgradePanel upgrade.Model
	metricsPanel metrics.Model

	activeSubPanel more.SubPanel
}

var channelTabs = [4]Tab{TabMessages, TabChat, TabCodex, TabGemini}
var channelTypes = [4]client.ChannelType{client.ChannelTypeMessages, client.ChannelTypeChat, client.ChannelTypeResponses, client.ChannelTypeGemini}

func NewApp() App {
	paths := config.DefaultPaths()
	env, _ := config.LoadEnv(paths.EnvFile)
	locale := "en"
	if env != nil && env.AppUILanguage != "" {
		locale = env.AppUILanguage
	}

	var apiClient *client.APIClient
	var mgr *process.Manager
	if env != nil {
		apiClient = client.NewAPIClient("http://127.0.0.1:"+env.Port, env.AdminKey())
		mgr = process.NewManager(paths.PidFile, paths.Binary, paths.LogFile)
	}

	km := DefaultKeyMap()
	h := help.New()
	h.ShowAll = false
	ii := i18n.New(locale)

	ov := overview.New(apiClient, mgr, paths, ii)
	var chs [4]channels.Model
	for i := range chs {
		chs[i] = channels.New(channelTypes[i], apiClient, ii)
	}
	mp := more.New(ii)
	lp := logs.New(paths.LogFile, ii)
	sp := settings.New(apiClient, ii)
	hp := health.New(apiClient, mgr, paths, ii)
	up := upgrade.New(paths.Binary, ii, Version)
	mt := metrics.New(apiClient, ii)

	return App{
		activeTab:     TabOverview,
		keys:          km,
		help:          h,
		ctxStack:      NewContextStack(Context{ID: ContextMain}),
		client:        apiClient,
		process:       mgr,
		env:           env,
		paths:         paths,
		i18n:          ii,
		overview:      ov,
		channels:      chs,
		morePanel:     mp,
		logsPanel:     lp,
		settings:      sp,
		health:        hp,
		upgradePanel:  up,
		metricsPanel:  mt,
		activeSubPanel: more.SubNone,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.checkConnection,
		a.overview.Init(),
		a.channels[0].Init(),
		a.channels[1].Init(),
		a.channels[2].Init(),
		a.channels[3].Init(),
	)
}

type connectionMsg struct {
	connected bool
}

func (a App) checkConnection() tea.Msg {
	if a.client == nil {
		return connectionMsg{connected: false}
	}
	_, err := a.client.HealthCheck(context.Background())
	return connectionMsg{connected: err == nil}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.overview, _ = a.overview.Update(msg)
		for i := range a.channels {
			a.channels[i], _ = a.channels[i].Update(msg)
		}
		a.logsPanel, _ = a.logsPanel.Update(msg)
		a.settings, _ = a.settings.Update(msg)
		a.health, _ = a.health.Update(msg)
		a.upgradePanel, _ = a.upgradePanel.Update(msg)
		a.metricsPanel, _ = a.metricsPanel.Update(msg)
		return a, nil

	case tea.KeyMsg:
		if a.quitting {
			return a, tea.Quit
		}

		if a.isInputFocused() {
			var cmd tea.Cmd
			cmd = a.forwardToActiveTab(msg)
			return a, cmd
		}

		switch {
		case key.Matches(msg, a.keys.Quit) && a.ctxStack.Len() == 1:
			a.quitting = true
			return a, tea.Quit
		case key.Matches(msg, a.keys.Back):
			if a.activeSubPanel != more.SubNone {
				a.activeSubPanel = more.SubNone
				return a, nil
			}
			if a.ctxStack.Len() > 1 {
				a.ctxStack.Pop()
			}
			return a, nil
		case key.Matches(msg, a.keys.Tab1):
			a.activeTab = TabOverview
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.Tab2):
			a.activeTab = TabMessages
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.Tab3):
			a.activeTab = TabChat
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.Tab4):
			a.activeTab = TabCodex
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.Tab5):
			a.activeTab = TabGemini
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.Tab6):
			a.activeTab = TabMore
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.NextTab):
			a.activeTab = (a.activeTab + 1) % TabCount
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.PrevTab):
			a.activeTab = (a.activeTab - 1 + TabCount) % TabCount
			a.activeSubPanel = more.SubNone
		case key.Matches(msg, a.keys.Help):
			a.help.ShowAll = !a.help.ShowAll
		default:
			var cmd tea.Cmd
			switch a.activeTab {
			case TabOverview:
				a.overview, cmd = a.overview.Update(msg)
			case TabMessages, TabChat, TabCodex, TabGemini:
				idx := channelTabIndex(a.activeTab)
				if idx >= 0 {
					a.channels[idx], cmd = a.channels[idx].Update(msg)
				}
			case TabMore:
				if a.activeSubPanel != more.SubNone {
					cmd = a.updateSubPanel(msg)
				} else {
					switch msg.String() {
					case "enter":
						a.activeSubPanel = a.morePanel.SelectedPanel()
						cmd = a.initSubPanel()
					default:
						a.morePanel, cmd = a.morePanel.Update(msg)
					}
				}
			}
			return a, cmd
		}
		return a, nil

	case connectionMsg:
		a.connected = msg.connected
		return a, nil
	}

	var cmd tea.Cmd
	a.overview, _ = a.overview.Update(msg)
	for i := range a.channels {
		a.channels[i], _ = a.channels[i].Update(msg)
	}
	a.logsPanel, _ = a.logsPanel.Update(msg)
	a.settings, _ = a.settings.Update(msg)
	a.health, _ = a.health.Update(msg)
	a.upgradePanel, _ = a.upgradePanel.Update(msg)
	a.metricsPanel, _ = a.metricsPanel.Update(msg)
	return a, cmd
}

func (a *App) updateSubPanel(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch a.activeSubPanel {
	case more.SubLogs:
		a.logsPanel, cmd = a.logsPanel.Update(msg)
	case more.SubSettings:
		a.settings, cmd = a.settings.Update(msg)
	case more.SubHealth:
		a.health, cmd = a.health.Update(msg)
	case more.SubUpgrade:
		a.upgradePanel, cmd = a.upgradePanel.Update(msg)
	case more.SubMetrics:
		a.metricsPanel, cmd = a.metricsPanel.Update(msg)
	}
	return cmd
}

func (a *App) initSubPanel() tea.Cmd {
	switch a.activeSubPanel {
	case more.SubLogs:
		return a.logsPanel.Init()
	case more.SubSettings:
		return a.settings.Init()
	case more.SubHealth:
		return a.health.Init()
	case more.SubMetrics:
		return a.metricsPanel.Init()
	}
	return nil
}

func channelTabIndex(tab Tab) int {
	for i, t := range channelTabs {
		if t == tab {
			return i
		}
	}
	return -1
}

func (a App) View() string {
	if a.quitting {
		return ""
	}
	tabBar := a.renderTabBar()
	content := a.renderContent()
	helpBar := a.renderHelp()
	return fmt.Sprintf("%s\n%s\n%s", tabBar, content, helpBar)
}

func (a App) renderTabBar() string {
	base := lipgloss.NewStyle().Padding(0, 1)
	active := base.Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62"))
	inactive := base.Faint(true)

	var tabs []string
	tabKeys := [TabCount]string{"overview", "messages", "chat", "codex", "gemini", "more"}
	for i := Tab(0); i < TabCount; i++ {
		label := a.i18n.T("tab." + tabKeys[i])
		if i == a.activeTab {
			tabs = append(tabs, active.Render(label))
		} else {
			tabs = append(tabs, inactive.Render(label))
		}
	}

	title := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("ccx-tui v%s", Version))
	connIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("●")
	if !a.connected {
		connIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("○")
	}
	return title + "  " + connIcon + "  " + strings.Join(tabs, " ")
}

func (a App) renderContent() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(max(a.height-4, 1)).
		Padding(1, 2)

	var content string
	switch a.activeTab {
	case TabOverview:
		content = a.overview.View()
	case TabMessages, TabChat, TabCodex, TabGemini:
		idx := channelTabIndex(a.activeTab)
		if idx >= 0 {
			content = a.channels[idx].View()
		}
	case TabMore:
		switch a.activeSubPanel {
		case more.SubLogs:
			content = a.logsPanel.View()
		case more.SubSettings:
			content = a.settings.View()
		case more.SubHealth:
			content = a.health.View()
		case more.SubUpgrade:
			content = a.upgradePanel.View()
		case more.SubMetrics:
			content = a.metricsPanel.View()
		default:
			content = a.morePanel.View()
		}
	}
	return style.Render(content)
}

func (a App) renderHelp() string {
	return a.help.View(a.keys)
}

func (a App) isInputFocused() bool {
	switch a.activeTab {
	case TabMessages, TabChat, TabCodex, TabGemini:
		idx := channelTabIndex(a.activeTab)
		if idx >= 0 {
			return a.channels[idx].IsFormActive()
		}
	}
	return false
}

func (a *App) forwardToActiveTab(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch a.activeTab {
	case TabOverview:
		a.overview, cmd = a.overview.Update(msg)
	case TabMessages, TabChat, TabCodex, TabGemini:
		idx := channelTabIndex(a.activeTab)
		if idx >= 0 {
			a.channels[idx], cmd = a.channels[idx].Update(msg)
		}
	case TabMore:
		if a.activeSubPanel != more.SubNone {
			cmd = a.updateSubPanel(msg)
		} else {
			a.morePanel, cmd = a.morePanel.Update(msg)
		}
	}
	return cmd
}
