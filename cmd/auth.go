package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Update your Slack token (e.g. if it expired)",
	RunE:  runAuth,
}

func runAuth(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil || !cfg.IsConfigured() {
		fmt.Println()
		fmt.Println("  No Slack app configured yet. Running full setup...")
		fmt.Println()
		return runSetup(cmd, args)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("  %s━━━ UPDATE TOKEN ━━━%s\n", colorPink, colorReset)
	fmt.Println()
	fmt.Printf("  %sGo to your app at api.slack.com/apps >%s OAuth & Permissions%s\n", colorDim, colorBold, colorReset)
	fmt.Printf("  %sCopy the%s User OAuth Token %s(starts with xoxp-)%s\n", colorDim, colorBold, colorDim, colorReset)
	fmt.Println()

	token := prompt(reader, "User OAuth Token (xoxp-...)")
	if token == "" {
		return fmt.Errorf("token is required")
	}

	if !strings.HasPrefix(token, "xoxp-") {
		fmt.Printf("\n  %s⚠ That doesn't look like a user token (should start with xoxp-)%s\n", colorPink, colorReset)
		confirm := prompt(reader, "Continue anyway? (y/n)")
		if strings.ToLower(confirm) != "y" {
			return fmt.Errorf("cancelled")
		}
	}

	if err := config.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Printf("\n  %s>>> TOKEN_SAVED — run 'slack-tui' to start.%s\n\n", colorGreen, colorReset)
	return nil
}
