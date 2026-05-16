//go:build !windows

package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func (m *Manager) IsRunning() bool {
	pid, err := m.readPid()
	if err != nil {
		return findProcess(m.binary) > 0
	}
	if syscall.Kill(pid, 0) == nil {
		return true
	}
	return findProcess(m.binary) > 0
}

func (m *Manager) Stop() error {
	if !m.IsRunning() {
		return fmt.Errorf("not running")
	}
	pid, _ := m.readPid()
	if pid == 0 {
		pid = findProcess(m.binary)
	}
	if pid == 0 {
		return fmt.Errorf("cannot find pid")
	}

	syscall.Kill(pid, syscall.SIGTERM)
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		if syscall.Kill(pid, 0) != nil {
			os.Remove(m.pidFile)
			return nil
		}
	}

	syscall.Kill(pid, syscall.SIGKILL)
	os.Remove(m.pidFile)
	return nil
}

func findProcess(binary string) int {
	data, _ := exec.Command("pgrep", "-f", binary).Output()
	if len(data) == 0 {
		return 0
	}
	pid, _ := strconv.Atoi(string(data[:len(data)-1]))
	return pid
}

func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
