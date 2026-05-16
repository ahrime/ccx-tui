package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Manager struct {
	pidFile string
	binary  string
	logFile string
}

func NewManager(pidFile, binary, logFile string) *Manager {
	return &Manager{pidFile: pidFile, binary: binary, logFile: logFile}
}

func (m *Manager) IsRunning() bool {
	pid, err := m.readPid()
	if err != nil {
		return m.findProcess() > 0
	}
	if syscall.Kill(pid, 0) == nil {
		return true
	}
	pid = m.findProcess()
	return pid > 0
}

func (m *Manager) findProcess() int {
	data, _ := exec.Command("pgrep", "-f", m.binary).Output()
	if len(data) == 0 {
		return 0
	}
	pid, _ := strconv.Atoi(string(data[:len(data)-1]))
	return pid
}

func (m *Manager) readPid() (int, error) {
	data, err := os.ReadFile(m.pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

func (m *Manager) writePid(pid int) error {
	return os.WriteFile(m.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (m *Manager) Start() (int, error) {
	if m.IsRunning() {
		pid, _ := m.readPid()
		if pid == 0 {
			pid = m.findProcess()
		}
		return 0, fmt.Errorf("already running, PID: %d", pid)
	}

	if _, err := os.Stat(m.binary); os.IsNotExist(err) {
		return 0, fmt.Errorf("binary not found: %s", m.binary)
	}

	logDir := filepath.Dir(m.logFile)
	os.MkdirAll(logDir, 0755)

	logF, err := os.OpenFile(m.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, fmt.Errorf("open log file: %w", err)
	}

	cmd := exec.Command(m.binary)
	cmd.Stdout = logF
	cmd.Stderr = logF
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start process: %w", err)
	}

	pid := cmd.Process.Pid
	m.writePid(pid)
	go cmd.Wait()

	return pid, nil
}

func (m *Manager) Stop() error {
	if !m.IsRunning() {
		return fmt.Errorf("not running")
	}
	pid, _ := m.readPid()
	if pid == 0 {
		pid = m.findProcess()
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

func (m *Manager) Restart() (int, error) {
	m.Stop()
	time.Sleep(time.Second)
	return m.Start()
}

func (m *Manager) Pid() (int, error) {
	return m.readPid()
}

func (m *Manager) Binary() string    { return m.binary }
func (m *Manager) PidFile() string   { return m.pidFile }
func (m *Manager) LogFile() string   { return m.logFile }
