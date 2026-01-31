package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIToken string `json:"api_token"`
}

func Load() (*Config, error) {
	// First try environment variable
	if token := os.Getenv("FASTMAIL_API_TOKEN"); token != "" {
		return &Config{APIToken: token}, nil
	}

	// Fallback to config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".config", "fastmail-agent", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
