//go:build windows

package process

import (
	"fmt"
	"os"
	"os/exec"
)

func (m *Manager) IsRunning() bool {
	pid, err := m.readPid()
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(nil) == nil
}

func (m *Manager) Stop() error {
	if !m.IsRunning() {
		return fmt.Errorf("not running")
	}
	pid, _ := m.readPid()
	if pid == 0 {
		return fmt.Errorf("cannot find pid")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("kill: %w", err)
	}
	os.Remove(m.pidFile)
	return nil
}

func findProcess(binary string) int {
	return 0
}

func setProcessGroup(cmd *exec.Cmd) {
}
