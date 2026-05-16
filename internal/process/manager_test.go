package process

import (
	"path/filepath"
	"testing"
)

func TestIsRunning_NoPidFile(t *testing.T) {
	m := NewManager("/nonexistent/ccx.pid", "/nonexistent/ccx", "/nonexistent/ccx.log")
	if m.IsRunning() {
		t.Error("should not be running without pid file")
	}
}

func TestStart_BinaryNotFound(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(
		filepath.Join(dir, "ccx.pid"),
		"/nonexistent/ccx-binary",
		filepath.Join(dir, "ccx.log"),
	)
	_, err := m.Start()
	if err == nil {
		t.Error("should fail when binary not found")
	}
}

func TestStop_NotRunning(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(
		filepath.Join(dir, "ccx.pid"),
		"/nonexistent/ccx-binary",
		filepath.Join(dir, "ccx.log"),
	)
	err := m.Stop()
	if err == nil {
		t.Error("should fail when not running")
	}
}

func TestManagerAccessors(t *testing.T) {
	m := NewManager("/tmp/pid", "/tmp/bin", "/tmp/log")
	if m.Binary() != "/tmp/bin" {
		t.Error("Binary accessor mismatch")
	}
	if m.PidFile() != "/tmp/pid" {
		t.Error("PidFile accessor mismatch")
	}
	if m.LogFile() != "/tmp/log" {
		t.Error("LogFile accessor mismatch")
	}
}
