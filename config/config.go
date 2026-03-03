package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	EventName   []string
}

func New() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		EventName:   []string{"orders_inserted"},
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("no DB_URL given")
	}
	fmt.Print(cfg, "\n")
	return cfg, nil
}
