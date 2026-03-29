package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName    = "monospace-cmd"
	tokenKey       = "slack-user-token"
	appTokenKey    = "slack-app-token"
	configFileName = "config.json"
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

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "monospace-cmd")
	return dir, os.MkdirAll(dir, 0700)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
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

func GetToken() (string, error) {
	// Check env var first for easy dev/CI usage
	if t := os.Getenv("SLACK_TOKEN"); t != "" {
		return t, nil
	}
	return keyring.Get(serviceName, tokenKey)
}

func SaveToken(token string) error {
	return keyring.Set(serviceName, tokenKey, token)
}

func GetAppToken() (string, error) {
	if t := os.Getenv("SLACK_APP_TOKEN"); t != "" {
		return t, nil
	}
	return keyring.Get(serviceName, appTokenKey)
}

func SaveAppToken(token string) error {
	return keyring.Set(serviceName, appTokenKey, token)
}

type oauthResponse struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error,omitempty"`
	AuthedUser  struct {
		AccessToken string `json:"access_token"`
	} `json:"authed_user"`
}

func ExchangeCode(clientID, clientSecret, code string) (string, error) {
	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {"http://localhost:9876/callback"},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result oauthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.OK {
		return "", fmt.Errorf("oauth error: %s", result.Error)
	}

	return result.AuthedUser.AccessToken, nil
}
