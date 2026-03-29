package cmd

import (
	"fmt"

	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Re-authenticate with Slack (run setup first if you haven't)",
	RunE:  runAuth,
}

func runAuth(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		fmt.Println()
		fmt.Println("  No Slack app configured yet. Running full setup...")
		fmt.Println()
		return runSetup(cmd, args)
	}

	fmt.Println()
	fmt.Printf("  %sRe-authenticating with existing app config...%s\n", colorDim, colorReset)

	token, err := runOAuth(cfg)
	if err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	if err := config.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	fmt.Printf("  %s>>> AUTH_COMPLETE — run 'slack-tui' to start.%s\n", colorGreen, colorReset)
	fmt.Println()
	return nil
}
