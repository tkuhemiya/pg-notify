package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	EventName   []string
	Port        int
}

type fileConfig struct {
	DatabaseURL string   `json:"database_url"`
	EventNames  []string `json:"event_names"`
	Port        int      `json:"port"`
}

func New() (Config, error) {
	return NewFromFile("pg-notify.cfg")
}

func NewFromFile(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open %s: %w", path, err)
	}

	var fc fileConfig
	if err := json.Unmarshal(raw, &fc); err != nil {
		return Config{}, fmt.Errorf("failed to parse %s as json: %w", path, err)
	}

	port := fc.Port
	if port <= 0 {
		port = 8080
	}
	eventNames := []string{"orders_inserted"}
	if len(fc.EventNames) > 0 {
		eventNames = fc.EventNames
	}

	cfg := Config{
		DatabaseURL: fc.DatabaseURL,
		EventName:   eventNames,
		Port:        port,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("no DATABASE_URL given in pg-notify.cfg")
	}
	return cfg, nil
}
