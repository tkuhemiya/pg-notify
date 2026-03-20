package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string   `json:"database_url"`
	EventNames  []string `json:"event_names"`
	Port        int      `json:"port"`
}

func NewFromFile(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse %s as json: %w", path, err)
	}

	port := cfg.Port
	if port <= 0 {
		port = 8080
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("no DATABASE_URL given in pg-notify.cfg")
	}
	return cfg, nil
}
