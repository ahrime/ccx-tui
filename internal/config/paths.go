package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Binary     string
	PidFile    string
	LogFile    string
	EnvFile    string
	ConfigDir  string
	ConfigFile string
}

func DefaultPaths() Paths {
	home, _ := os.UserHomeDir()
	ccxDir := filepath.Join(home, ".ccx")
	return Paths{
		Binary:     filepath.Join(home, ".local", "bin", "ccx"),
		PidFile:    filepath.Join(ccxDir, "ccx.pid"),
		LogFile:    filepath.Join(ccxDir, "ccx.log"),
		EnvFile:    filepath.Join(ccxDir, ".env"),
		ConfigDir:  filepath.Join(ccxDir, ".config"),
		ConfigFile: filepath.Join(ccxDir, ".config", "config.json"),
	}
}

func (p Paths) BinaryExists() bool {
	info, err := os.Stat(p.Binary)
	return err == nil && !info.IsDir()
}
