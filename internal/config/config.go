package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName = "config.json"
	tokensFileName = "tokens.json"
)

type AIHookConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key,omitempty"`
	Model    string `json:"model"`
}

type AIConfig struct {
	Summarizer AIHookConfig `json:"summarizer"`
	Drafter    AIHookConfig `json:"drafter"`
	Analyzer   AIHookConfig `json:"analyzer"`
}

type ThemeConfig struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Surface   string `json:"surface"`
}

type Config struct {
	ClientID     string      `json:"client_id"`
	ClientSecret string      `json:"client_secret"`
	Theme        ThemeConfig `json:"theme"`
	AI           AIConfig    `json:"ai"`
	SidebarWidth int         `json:"sidebar_width"`
	TimeFormat   string      `json:"time_format"`
}

// tokens is stored in a separate file with restrictive permissions.
type tokens struct {
	UserToken string `json:"user_token,omitempty"`
	AppToken  string `json:"app_token,omitempty"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "slack-tui")
	return dir, os.MkdirAll(dir, 0700)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func tokensPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokensFileName), nil
}

func DefaultConfig() *Config {
	return &Config{
		Theme: ThemeConfig{
			Primary:   "#f6afef",
			Secondary: "#5edda0",
			Surface:   "#10141a",
		},
		AI: AIConfig{
			Summarizer: AIHookConfig{
				Enabled:  true,
				Provider: "anthropic",
				Model:    "claude-sonnet-4-6",
			},
			Drafter: AIHookConfig{
				Enabled:  true,
				Provider: "anthropic",
				Model:    "claude-sonnet-4-6",
			},
			Analyzer: AIHookConfig{
				Enabled:  true,
				Provider: "anthropic",
				Model:    "claude-haiku-4-5-20251001",
			},
		},
		SidebarWidth: 25,
		TimeFormat:   "15:04:05",
	}
}

// IsConfigured returns true if the essential Slack credentials are set.
func (c *Config) IsConfigured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadTokens() (*tokens, error) {
	path, err := tokensPath()
	if err != nil {
		return &tokens{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &tokens{}, nil
		}
		return nil, err
	}

	var t tokens
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func saveTokens(t *tokens) error {
	path, err := tokensPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func GetToken() (string, error) {
	if t := os.Getenv("SLACK_TOKEN"); t != "" {
		return t, nil
	}
	tok, err := loadTokens()
	if err != nil {
		return "", err
	}
	if tok.UserToken == "" {
		return "", fmt.Errorf("no user token configured")
	}
	return tok.UserToken, nil
}

func SaveToken(token string) error {
	tok, err := loadTokens()
	if err != nil {
		tok = &tokens{}
	}
	tok.UserToken = token
	return saveTokens(tok)
}

func GetAppToken() (string, error) {
	if t := os.Getenv("SLACK_APP_TOKEN"); t != "" {
		return t, nil
	}
	tok, err := loadTokens()
	if err != nil {
		return "", err
	}
	if tok.AppToken == "" {
		return "", fmt.Errorf("no app token configured")
	}
	return tok.AppToken, nil
}

func SaveAppToken(token string) error {
	tok, err := loadTokens()
	if err != nil {
		tok = &tokens{}
	}
	tok.AppToken = token
	return saveTokens(tok)
}
