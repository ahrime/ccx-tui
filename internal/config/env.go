package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type EnvConfig struct {
	Port           string
	ProxyAccessKey string
	AdminAccessKey string
	EnableWebUI    string
	AppUILanguage  string
	LogLevel       string
}

func LoadEnv(path string) (*EnvConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open .env: %w", err)
	}
	defer f.Close()

	env := &EnvConfig{
		Port:          "3000",
		EnableWebUI:   "true",
		AppUILanguage: "en",
		LogLevel:      "info",
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "PORT":
			env.Port = val
		case "PROXY_ACCESS_KEY":
			env.ProxyAccessKey = val
		case "ADMIN_ACCESS_KEY":
			env.AdminAccessKey = val
		case "ENABLE_WEB_UI":
			env.EnableWebUI = val
		case "APP_UI_LANGUAGE":
			env.AppUILanguage = val
		case "LOG_LEVEL":
			env.LogLevel = val
		}
	}

	if env.AdminAccessKey == "" {
		env.AdminAccessKey = env.ProxyAccessKey
	}

	return env, scanner.Err()
}

func (e *EnvConfig) AdminKey() string {
	if e.AdminAccessKey != "" {
		return e.AdminAccessKey
	}
	return e.ProxyAccessKey
}
