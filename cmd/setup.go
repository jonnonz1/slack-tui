package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const (
	colorPink  = "\033[38;2;246;175;239m"
	colorGreen = "\033[38;2;94;221;160m"
	colorDim   = "\033[38;2;102;102;102m"
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard — creates Slack app and authenticates",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	printBanner()

	// Step 1: Create Slack App
	printStep(1, 5, "CREATE SLACK APP")
	fmt.Println()
	fmt.Printf("  %sI'll open api.slack.com/apps in your browser.%s\n", colorDim, colorReset)
	fmt.Println()
	fmt.Printf("  %s1.%s Click %sCreate New App%s > %sFrom scratch%s\n", colorGreen, colorReset, colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %s2.%s Name it anything (e.g. %sslack-tui%s)\n", colorGreen, colorReset, colorPink, colorReset)
	fmt.Printf("  %s3.%s Select your workspace\n", colorGreen, colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when you've created the app...")
	_ = browser.OpenURL("https://api.slack.com/apps")

	// Step 2: User Token Scopes
	printStep(2, 5, "ADD SCOPES & INSTALL")
	fmt.Println()
	fmt.Printf("  %sIn your new app, go to:%s\n", colorDim, colorReset)
	fmt.Printf("  %sOAuth & Permissions%s > %sUser Token Scopes%s > add all of these:\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Println()
	scopes := []string{
		"channels:read", "channels:history", "groups:read", "groups:history",
		"im:read", "im:history", "mpim:read", "mpim:history",
		"chat:write", "reactions:read", "reactions:write", "users:read",
		"files:read", "files:write", "pins:read", "search:read",
		"team:read", "usergroups:read",
	}
	for i := 0; i < len(scopes); i += 4 {
		end := i + 4
		if end > len(scopes) {
			end = len(scopes)
		}
		fmt.Printf("  %s%s%s\n", colorGreen, strings.Join(scopes[i:end], "  "), colorReset)
	}
	fmt.Println()
	fmt.Printf("  %sThen scroll up and click%s Install to Workspace%s > %sAllow%s\n", colorDim, colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sCopy the%s User OAuth Token %s(starts with %sxoxp-%s)%s\n", colorDim, colorBold, colorDim, colorGreen, colorDim, colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when you've installed the app and copied the token...")

	// Step 3: Enable Socket Mode
	printStep(3, 5, "ENABLE SOCKET MODE")
	fmt.Println()
	fmt.Printf("  %sGo to:%s %sSocket Mode%s (left sidebar) > toggle %sON%s\n", colorDim, colorReset, colorBold, colorReset, colorGreen, colorReset)
	fmt.Printf("  %sCreate an app-level token with scope:%s %sconnections:write%s\n", colorDim, colorReset, colorGreen, colorReset)
	fmt.Printf("  %sCopy the token (starts with %sxapp-%s)%s\n", colorDim, colorGreen, colorDim, colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when Socket Mode is enabled...")

	// Step 4: Event Subscriptions
	printStep(4, 5, "SUBSCRIBE TO EVENTS")
	fmt.Println()
	fmt.Printf("  %sGo to:%s %sEvent Subscriptions%s > toggle %sON%s\n", colorDim, colorReset, colorBold, colorReset, colorGreen, colorReset)
	fmt.Printf("  %sUnder%s Subscribe to bot events%s, add:%s\n", colorDim, colorBold, colorReset, colorReset)
	fmt.Println()
	events := []string{"message.channels", "message.groups", "message.im", "message.mpim", "reaction_added", "reaction_removed"}
	fmt.Printf("  %s%s%s\n", colorGreen, strings.Join(events, "  "), colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when events are subscribed...")

	// Step 5: Collect tokens
	printStep(5, 5, "ENTER TOKENS")
	fmt.Println()
	fmt.Printf("  %sPaste the tokens you copied earlier:%s\n", colorDim, colorReset)
	fmt.Println()

	userToken := prompt(reader, "User OAuth Token (xoxp-...)")
	appToken := prompt(reader, "App-Level Token (xapp-...)")

	if userToken == "" || appToken == "" {
		return fmt.Errorf("both tokens are required")
	}

	// Validate token prefixes
	if !strings.HasPrefix(userToken, "xoxp-") {
		fmt.Printf("\n  %s⚠ That doesn't look like a user token (should start with xoxp-)%s\n", colorPink, colorReset)
		confirm := prompt(reader, "Continue anyway? (y/n)")
		if strings.ToLower(confirm) != "y" {
			return fmt.Errorf("setup cancelled")
		}
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		fmt.Printf("\n  %s⚠ That doesn't look like an app-level token (should start with xapp-)%s\n", colorPink, colorReset)
		confirm := prompt(reader, "Continue anyway? (y/n)")
		if strings.ToLower(confirm) != "y" {
			return fmt.Errorf("setup cancelled")
		}
	}

	// Save config
	cfg := config.DefaultConfig()
	cfg.ClientID = "installed"
	cfg.ClientSecret = "installed"

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	if err := config.SaveToken(userToken); err != nil {
		return fmt.Errorf("failed to save user token: %w", err)
	}
	if err := config.SaveAppToken(appToken); err != nil {
		return fmt.Errorf("failed to save app token: %w", err)
	}

	// Done!
	fmt.Println()
	fmt.Printf("  %s╔══════════════════════════════════════════╗%s\n", colorGreen, colorReset)
	fmt.Printf("  %s║          SETUP_COMPLETE                  ║%s\n", colorGreen, colorReset)
	fmt.Printf("  %s╚══════════════════════════════════════════╝%s\n", colorGreen, colorReset)
	fmt.Println()
	fmt.Printf("  Run %sslack-tui%s to start.\n", colorPink, colorReset)
	fmt.Println()
	fmt.Printf("  %sFor AI features (summarize, draft replies):%s\n", colorDim, colorReset)
	fmt.Printf("  %sexport ANTHROPIC_API_KEY=sk-ant-...%s\n", colorGreen, colorReset)
	fmt.Println()

	return nil
}

func printBanner() {
	fmt.Println()
	fmt.Printf("  %s╔══════════════════════════════════════════╗%s\n", colorPink, colorReset)
	fmt.Printf("  %s║       SLACK-TUI — SETUP_WIZARD          ║%s\n", colorPink, colorReset)
	fmt.Printf("  %s╚══════════════════════════════════════════╝%s\n", colorPink, colorReset)
	fmt.Println()
	fmt.Printf("  %sThis wizard will walk you through connecting\n  to your Slack workspace in 5 steps.%s\n", colorDim, colorReset)
	fmt.Println()
}

func printStep(n, total int, title string) {
	fmt.Printf("\n  %s━━━ STEP %d/%d: %s ━━━%s\n", colorPink, n, total, title, colorReset)
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Printf("  %s%s >%s ", colorGreen, label, colorReset)
	val, _ := reader.ReadString('\n')
	return strings.TrimSpace(val)
}

func waitForEnter(reader *bufio.Reader, msg string) {
	fmt.Printf("  %s%s%s", colorDim, msg, colorReset)
	_, _ = reader.ReadString('\n')
}
