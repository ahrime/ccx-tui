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

	overview  overview.Model
	channels  [4]channels.Model
	morePanel more.Model
	logsPanel logs.Model
	settings  settings.Model
	health    health.Model
	upgrade   upgrade.Model
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
	up := upgrade.New(paths.Binary, ii)

	return App{
		activeTab: TabOverview,
		keys:      km,
		help:      h,
		ctxStack:  NewContextStack(Context{ID: ContextMain}),
		client:    apiClient,
		process:   mgr,
		env:       env,
		paths:     paths,
		i18n:      ii,
		overview:  ov,
		channels:  chs,
		morePanel: mp,
		logsPanel: lp,
		settings:  sp,
		health:    hp,
		upgrade:   up,
	}
}

func (a App) Init() tea.Cmd {
	return a.checkConnection
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
		return a, nil

	case tea.KeyMsg:
		if a.quitting {
			return a, tea.Quit
		}
		switch {
		case key.Matches(msg, a.keys.Quit) && a.ctxStack.Len() == 1:
			a.quitting = true
			return a, tea.Quit
		case key.Matches(msg, a.keys.Back):
			if a.ctxStack.Len() > 1 {
				a.ctxStack.Pop()
			}
			return a, nil
		case key.Matches(msg, a.keys.Tab1):
			a.activeTab = TabOverview
		case key.Matches(msg, a.keys.Tab2):
			a.activeTab = TabMessages
		case key.Matches(msg, a.keys.Tab3):
			a.activeTab = TabChat
		case key.Matches(msg, a.keys.Tab4):
			a.activeTab = TabCodex
		case key.Matches(msg, a.keys.Tab5):
			a.activeTab = TabGemini
		case key.Matches(msg, a.keys.Tab6):
			a.activeTab = TabMore
		case key.Matches(msg, a.keys.NextTab):
			a.activeTab = (a.activeTab + 1) % TabCount
		case key.Matches(msg, a.keys.PrevTab):
			a.activeTab = (a.activeTab - 1 + TabCount) % TabCount
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
				a.morePanel, cmd = a.morePanel.Update(msg)
			}
			return a, cmd
		}
		return a, nil

	case connectionMsg:
		a.connected = msg.connected
		return a, nil
	}

	return a, nil
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
		content = a.morePanel.View()
	}
	return style.Render(content)
}

func (a App) renderHelp() string {
	return a.help.View(a.keys)
}
