package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "slack-tui",
	Short: "MONOSPACE_CMD — a retro terminal Slack client",
	Long: `MONOSPACE_CMD_V1.0
A retro terminal UI for Slack with integrated AI hooks.
Single workspace, keyboard-first, Claude Code aesthetic.`,
	RunE: runApp,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(setupCmd)
}
