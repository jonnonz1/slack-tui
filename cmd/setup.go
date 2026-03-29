package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for MONOSPACE_CMD",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════╗")
	fmt.Println("  ║  MONOSPACE_CMD_V1.0 — SETUP_WIZARD      ║")
	fmt.Println("  ╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  Follow these steps to connect to your Slack workspace:")
	fmt.Println()
	fmt.Println("  1. Go to https://api.slack.com/apps")
	fmt.Println("  2. Click 'Create New App' > 'From scratch'")
	fmt.Println("  3. Name it 'MONOSPACE_CMD' and select your workspace")
	fmt.Println("  4. Under 'OAuth & Permissions', add these User Token Scopes:")
	fmt.Println("     channels:read, channels:history, groups:read, groups:history,")
	fmt.Println("     im:read, im:history, mpim:read, mpim:history, chat:write,")
	fmt.Println("     reactions:read, reactions:write, users:read, files:read,")
	fmt.Println("     files:write, pins:read, search:read, team:read")
	fmt.Println("  5. Under 'Socket Mode', enable it and create an app-level token")
	fmt.Println("     with the 'connections:write' scope")
	fmt.Println("  6. Under 'Event Subscriptions', subscribe to bot events:")
	fmt.Println("     message.channels, message.groups, message.im, message.mpim,")
	fmt.Println("     reaction_added, reaction_removed")
	fmt.Println()

	fmt.Print("  Client ID: ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	fmt.Print("  Client Secret: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	fmt.Print("  App-Level Token (xapp-...): ")
	appToken, _ := reader.ReadString('\n')
	appToken = strings.TrimSpace(appToken)

	if clientID == "" || clientSecret == "" || appToken == "" {
		return fmt.Errorf("all fields are required")
	}

	cfg := &config.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if err := config.SaveAppToken(appToken); err != nil {
		return fmt.Errorf("failed to save app token: %w", err)
	}

	fmt.Println()
	fmt.Println("  >>> CONFIG_SAVED")
	fmt.Println("  >>> Now run 'slack-tui auth' to complete OAuth flow.")
	fmt.Println()

	return nil
}
