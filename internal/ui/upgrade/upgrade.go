package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BenedictKing/ccx-tui/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const GITHUB_REPO = "BenedictKing/ccx"
const GITHUB_API = "https://api.github.com/repos/" + GITHUB_REPO + "/releases/latest"

type Model struct {
	binary      string
	i18n        *i18n.I18n
	current     string
	latest      string
	downloading bool
	confirming  bool
	downloadURL string
	statusMsg   string
	progress    float64
}

type versionCheckMsg struct {
	latest      string
	downloadURL string
	err         error
}

type downloadProgressMsg struct {
	progress float64
}

type downloadDoneMsg struct {
	err error
}

func New(binary string, i *i18n.I18n) Model {
	return Model{binary: binary, i18n: i}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) checkVersion() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(GITHUB_API)
		if err != nil {
			return versionCheckMsg{err: err}
		}
		defer resp.Body.Close()

		var release struct {
			TagName string `json:"tag_name"`
			Assets  []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return versionCheckMsg{err: err}
		}

		targetSuffix := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		if runtime.GOOS == "windows" {
			targetSuffix += ".exe"
		}
		var dlURL string
		for _, a := range release.Assets {
			if strings.Contains(a.Name, targetSuffix) {
				dlURL = a.BrowserDownloadURL
				break
			}
		}

		return versionCheckMsg{
			latest:      strings.TrimPrefix(release.TagName, "v"),
			downloadURL: dlURL,
		}
	}
}

func (m Model) downloadAndInstall() tea.Cmd {
	return func() tea.Msg {
		if m.downloadURL == "" {
			return downloadDoneMsg{err: fmt.Errorf("no download URL for this platform")}
		}

		resp, err := http.Get(m.downloadURL)
		if err != nil {
			return downloadDoneMsg{err: err}
		}
		defer resp.Body.Close()

		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, "ccx-tui-upgrade")
		f, err := os.Create(tmpFile)
		if err != nil {
			return downloadDoneMsg{err: err}
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			os.Remove(tmpFile)
			return downloadDoneMsg{err: err}
		}
		f.Close()

		if err := os.Chmod(tmpFile, 0755); err != nil {
			return downloadDoneMsg{err: err}
		}

		if err := os.Rename(tmpFile, m.binary); err != nil {
			return downloadDoneMsg{err: fmt.Errorf("install: %w", err)}
		}

		return downloadDoneMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case versionCheckMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("check failed: %v", msg.err)
		} else {
			m.latest = msg.latest
			m.downloadURL = msg.downloadURL
			if m.latest != "" && m.latest != m.current {
				m.confirming = true
			} else {
				m.statusMsg = m.i18n.T("upgrade.uptodate")
			}
		}
		return m, nil

	case downloadProgressMsg:
		m.progress = msg.progress
		return m, nil

	case downloadDoneMsg:
		m.downloading = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("upgrade failed: %v", msg.err)
		} else {
			m.statusMsg = m.i18n.T("upgrade.complete")
			m.current = m.latest
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "u":
			if !m.downloading {
				return m, m.checkVersion()
			}
		case "y":
			if m.confirming {
				m.confirming = false
				m.downloading = true
				return m, m.downloadAndInstall()
			}
		case "n":
			m.confirming = false
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := lipgloss.NewStyle().Bold(true).Render("  " + m.i18n.T("upgrade.title"))
	s += "\n  ─────────────────────────────────────────────\n"

	labelStyle := lipgloss.NewStyle().Faint(true).Width(16)
	valueStyle := lipgloss.NewStyle().Bold(true)

	s += fmt.Sprintf("  %s %s\n", labelStyle.Render(m.i18n.T("upgrade.current")), valueStyle.Render(m.current))
	s += fmt.Sprintf("  %s %s\n", labelStyle.Render(m.i18n.T("upgrade.latest")), valueStyle.Render(m.latest))

	if m.downloading {
		barWidth := 30
		filled := int(m.progress * float64(barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		s += fmt.Sprintf("\n  [%s] %.0f%%", bar, m.progress*100)
		s += "\n  " + m.i18n.T("upgrade.downloading")
	} else if m.confirming {
		s += fmt.Sprintf("\n\n  New version available: %s → %s", m.current, m.latest)
		s += "\n  Confirm upgrade? [y] Yes  [n] No"
	} else {
		s += "\n\n  [u] Check & Upgrade"
	}

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg
	}
	return s
}
