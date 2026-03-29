package cmd

import (
	"fmt"

	"github.com/jonnonz1/slack-tui/internal/app"
	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/jonnonz1/slack-tui/internal/slack"
	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
)

func runApp(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("no auth token found — run 'slack-tui auth' first: %w", err)
	}

	appToken, err := config.GetAppToken()
	if err != nil {
		return fmt.Errorf("no app-level token found — run 'slack-tui setup' first: %w", err)
	}

	client, err := slack.NewClient(token, appToken)
	if err != nil {
		return fmt.Errorf("failed to create Slack client: %w", err)
	}

	model := app.New(client, cfg)

	p := tea.NewProgram(model)

	// Start socket mode event relay in background
	go client.StartSocketMode(p)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
