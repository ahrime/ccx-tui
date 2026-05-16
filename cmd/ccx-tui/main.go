package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/BenedictKing/ccx-tui/internal/app"
	"github.com/BenedictKing/ccx-tui/internal/config"
	"github.com/BenedictKing/ccx-tui/internal/process"
)

var version = "0.1.0-dev"

func main() {
	app.Version = version

	if len(os.Args) > 1 {
		runCLI(os.Args[1:])
		return
	}

	p := tea.NewProgram(app.NewApp(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(args []string) {
	paths := config.DefaultPaths()
	env, _ := config.LoadEnv(paths.EnvFile)
	port := "3000"
	var accessKey string
	if env != nil {
		port = env.Port
		accessKey = env.AdminKey()
	}

	mgr := process.NewManager(paths.PidFile, paths.Binary, paths.LogFile)
	cmd := args[0]

	switch cmd {
	case "version", "-v", "--version":
		fmt.Printf("ccx-tui v%s\n", version)

	case "start":
		pid, err := mgr.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CCX started, PID: %d\n", pid)
		fmt.Printf("Web UI: http://localhost:%s\n", port)

	case "stop":
		if err := mgr.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("CCX stopped")

	case "restart":
		pid, err := mgr.Restart()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CCX restarted, PID: %d\n", pid)

	case "status":
		if !mgr.IsRunning() {
			fmt.Println("CCX not running")
			return
		}
		pid, _ := mgr.Pid()
		fmt.Printf("CCX running, PID: %d\n", pid)
		out, err := exec.Command("curl", "-sf", fmt.Sprintf("http://127.0.0.1:%s/health", port), "--max-time", "3").Output()
		if err == nil {
			var h map[string]interface{}
			if json.Unmarshal(out, &h) == nil {
				fmt.Printf("Web UI: http://localhost:%s\n", port)
				if v, ok := h["version"]; ok {
					if vm, ok := v.(map[string]interface{}); ok {
						fmt.Printf("Version: %v\n", vm["version"])
					}
				}
				if c, ok := h["config"]; ok {
					if cm, ok := c.(map[string]interface{}); ok {
						fmt.Printf("Channels: %v\n", cm["upstreamCount"])
					}
				}
			}
		}

	case "health":
		out, err := exec.Command("curl", "-sf", fmt.Sprintf("http://127.0.0.1:%s/health", port), "--max-time", "3").Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "CCX not responding (http://127.0.0.1:%s/health)\n", port)
			os.Exit(1)
		}
		var pretty interface{}
		raw := json.RawMessage(out)
		pretty = raw
		indented, err2 := json.MarshalIndent(&raw, "", "  ")
		if err2 == nil {
			fmt.Println(string(indented))
		} else {
			fmt.Println(string(out))
		}
		_ = pretty

	case "doctor":
		fmt.Println("━━━ CCX Doctor ━━━")
		fmt.Println()
		if paths.BinaryExists() {
			fmt.Printf("  [Binary]  ✓ %s\n", paths.Binary)
		} else {
			fmt.Printf("  [Binary]  ✗ Not found: %s\n", paths.Binary)
		}
		if _, err := os.Stat(paths.EnvFile); err == nil {
			fmt.Printf("  [Config]  ✓ .env exists (port=%s)\n", port)
		} else {
			fmt.Println("  [Config]  ✗ .env not found")
		}
		if mgr.IsRunning() {
			pid, _ := mgr.Pid()
			fmt.Printf("  [Process] ✓ Running, PID: %d\n", pid)
		} else {
			fmt.Println("  [Process] ✗ Not running")
		}
		out, err := exec.Command("curl", "-sf", "-o", "/dev/null", "-w", "%{http_code}", fmt.Sprintf("http://127.0.0.1:%s/health", port), "--max-time", "3").Output()
		if err == nil && string(out) == "200" {
			fmt.Println("  [Health]  ✓ HTTP 200")
		} else {
			code := "timeout"
			if err == nil {
				code = string(out)
			}
			fmt.Printf("  [Health]  ✗ HTTP %s\n", code)
		}
		if info, err := os.Stat(paths.LogFile); err == nil {
			fmt.Printf("  [Log]     ✓ %d bytes\n", info.Size())
		}

	case "logs":
		n := 50
		follow := false
		for _, a := range args[1:] {
			if a == "-f" || a == "--follow" {
				follow = true
			}
		}
		if follow {
			c := exec.Command("tail", "-f", paths.LogFile)
			c.Stdout = os.Stdout
			c.Run()
		} else {
			c := exec.Command("tail", fmt.Sprintf("-%d", n), paths.LogFile)
			c.Stdout = os.Stdout
			c.Run()
		}

	case "web":
		fmt.Printf("Web UI: http://localhost:%s\n", port)
		if accessKey != "" {
			masked := accessKey
			if len(accessKey) > 8 {
				masked = accessKey[:4] + strings.Repeat("*", len(accessKey)-8) + accessKey[len(accessKey)-4:]
			} else {
				masked = strings.Repeat("*", len(accessKey))
			}
			fmt.Printf("Access Key: %s\n", masked)
		}

	case "help", "-h", "--help":
		fmt.Println("Usage: ccx-tui [command]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  start           Start CCX")
		fmt.Println("  stop            Stop CCX")
		fmt.Println("  restart         Restart CCX")
		fmt.Println("  status          Show status")
		fmt.Println("  health          Health check (JSON)")
		fmt.Println("  doctor          Run diagnostics")
		fmt.Println("  logs [-f]       View logs (add -f to follow)")
		fmt.Println("  web             Show Web UI URL and key")
		fmt.Println("  version         Show version")
		fmt.Println("  help            Show this help")
		fmt.Println()
		fmt.Println("Run without arguments to start TUI.")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintf(os.Stderr, "Run 'ccx-tui help' for usage.\n")
		os.Exit(1)
	}
}

func init() {
	_ = context.Background
	_ = time.Now
}
