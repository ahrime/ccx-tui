package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPaths(t *testing.T) {
	home, _ := os.UserHomeDir()
	p := DefaultPaths()
	if p.Binary != filepath.Join(home, ".local", "bin", "ccx") {
		t.Errorf("Binary path mismatch: %s", p.Binary)
	}
	if p.PidFile == "" {
		t.Error("PidFile should not be empty")
	}
	if p.LogFile == "" {
		t.Error("LogFile should not be empty")
	}
	if p.EnvFile == "" {
		t.Error("EnvFile should not be empty")
	}
	if p.ConfigFile == "" {
		t.Error("ConfigFile should not be empty")
	}
}

func TestBinaryExists(t *testing.T) {
	p := DefaultPaths()
	_ = p.BinaryExists()
}

func TestLoadEnv_NotFound(t *testing.T) {
	_, err := LoadEnv("/nonexistent/.env")
	if err == nil {
		t.Error("should fail for nonexistent file")
	}
}

func TestLoadEnv_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "PORT=8080\nPROXY_ACCESS_KEY=mykey\n# comment\nADMIN_ACCESS_KEY=adminkey\n"
	os.WriteFile(path, []byte(content), 0644)

	env, err := LoadEnv(path)
	if err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}
	if env.Port != "8080" {
		t.Errorf("Port mismatch: got %s", env.Port)
	}
	if env.ProxyAccessKey != "mykey" {
		t.Errorf("ProxyAccessKey mismatch: got %s", env.ProxyAccessKey)
	}
	if env.AdminAccessKey != "adminkey" {
		t.Errorf("AdminAccessKey mismatch: got %s", env.AdminAccessKey)
	}
}

func TestAdminKey_Fallback(t *testing.T) {
	env := &EnvConfig{ProxyAccessKey: "proxy", AdminAccessKey: ""}
	if env.AdminKey() != "proxy" {
		t.Error("AdminKey should fallback to ProxyAccessKey")
	}
	env2 := &EnvConfig{ProxyAccessKey: "proxy", AdminAccessKey: "admin"}
	if env2.AdminKey() != "admin" {
		t.Error("AdminKey should use AdminAccessKey when set")
	}
}
