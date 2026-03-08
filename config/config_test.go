package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pg-notify.cfg")
	content := `{
  "database_url": "postgres://u:p@localhost:5432/db?sslmode=disable",
  "port": 9090,
  "event_names": ["orders_inserted", "orders_updated"]
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	cfg, err := NewFromFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.DatabaseURL == "" {
		t.Fatalf("expected database url")
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected port=9090 got %d", cfg.Port)
	}
	if len(cfg.EventName) != 2 || cfg.EventName[0] != "orders_inserted" || cfg.EventName[1] != "orders_updated" {
		t.Fatalf("unexpected event names: %#v", cfg.EventName)
	}
}

func TestNewFromFileDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pg-notify.cfg")
	content := `{
  "database_url": "postgres://u:p@localhost:5432/db?sslmode=disable"
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	cfg, err := NewFromFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080 got %d", cfg.Port)
	}
	if len(cfg.EventName) != 1 || cfg.EventName[0] != "orders_inserted" {
		t.Fatalf("unexpected default events: %#v", cfg.EventName)
	}
}

func TestNewFromFileRequiresDatabaseURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pg-notify.cfg")
	if err := os.WriteFile(path, []byte(`{"port":8080}`), 0o644); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	_, err := NewFromFile(path)
	if err == nil {
		t.Fatal("expected error when DATABASE_URL missing")
	}
}
