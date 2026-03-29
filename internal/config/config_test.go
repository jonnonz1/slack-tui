package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Theme.Primary != "#f6afef" {
		t.Errorf("expected primary #f6afef, got %s", cfg.Theme.Primary)
	}
	if cfg.Theme.Secondary != "#5edda0" {
		t.Errorf("expected secondary #5edda0, got %s", cfg.Theme.Secondary)
	}
	if cfg.Theme.Surface != "#10141a" {
		t.Errorf("expected surface #10141a, got %s", cfg.Theme.Surface)
	}
	if cfg.SidebarWidth != 25 {
		t.Errorf("expected sidebar width 25, got %d", cfg.SidebarWidth)
	}
	if cfg.TimeFormat != "15:04:05" {
		t.Errorf("expected time format 15:04:05, got %s", cfg.TimeFormat)
	}
	if !cfg.AI.Summarizer.Enabled {
		t.Error("summarizer should be enabled by default")
	}
	if !cfg.AI.Drafter.Enabled {
		t.Error("drafter should be enabled by default")
	}
	if !cfg.AI.Analyzer.Enabled {
		t.Error("analyzer should be enabled by default")
	}
	if cfg.AI.Summarizer.Provider != "anthropic" {
		t.Errorf("expected anthropic provider, got %s", cfg.AI.Summarizer.Provider)
	}
}

func TestConfig_JSONRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ClientID = "test-client-id"
	cfg.ClientSecret = "test-secret"
	cfg.SidebarWidth = 30

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if loaded.ClientID != "test-client-id" {
		t.Errorf("ClientID = %q, want %q", loaded.ClientID, "test-client-id")
	}
	if loaded.SidebarWidth != 30 {
		t.Errorf("SidebarWidth = %d, want 30", loaded.SidebarWidth)
	}
	if loaded.AI.Summarizer.Model != cfg.AI.Summarizer.Model {
		t.Errorf("Summarizer.Model = %q, want %q", loaded.AI.Summarizer.Model, cfg.AI.Summarizer.Model)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Use a temp dir to avoid touching real config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	cfg := DefaultConfig()
	cfg.ClientID = "roundtrip-id"
	cfg.SidebarWidth = 35

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.ClientID != "roundtrip-id" {
		t.Errorf("ClientID = %q, want roundtrip-id", loaded.ClientID)
	}
	if loaded.SidebarWidth != 35 {
		t.Errorf("SidebarWidth = %d, want 35", loaded.SidebarWidth)
	}
}

func TestGetToken_EnvOverride(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "xoxp-test-env-token")

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if token != "xoxp-test-env-token" {
		t.Errorf("expected env token, got %q", token)
	}
}

func TestGetAppToken_EnvOverride(t *testing.T) {
	t.Setenv("SLACK_APP_TOKEN", "xapp-test-env-token")

	token, err := GetAppToken()
	if err != nil {
		t.Fatalf("GetAppToken: %v", err)
	}
	if token != "xapp-test-env-token" {
		t.Errorf("expected env token, got %q", token)
	}
}
