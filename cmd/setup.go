package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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
	printStep(1, 6, "CREATE SLACK APP")
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
	printStep(2, 6, "ADD USER TOKEN SCOPES")
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

	// Also add the redirect URL
	fmt.Printf("  %sThen scroll down to%s Redirect URLs %sand add:%s\n", colorDim, colorReset, colorBold, colorReset)
	fmt.Printf("  %shttp://localhost:9876/callback%s\n", colorGreen, colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when scopes and redirect URL are saved...")

	// Step 3: Enable Socket Mode
	printStep(3, 6, "ENABLE SOCKET MODE")
	fmt.Println()
	fmt.Printf("  %sGo to:%s %sSocket Mode%s (left sidebar) > toggle %sON%s\n", colorDim, colorReset, colorBold, colorReset, colorGreen, colorReset)
	fmt.Printf("  %sCreate an app-level token with scope:%s %sconnections:write%s\n", colorDim, colorReset, colorGreen, colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when Socket Mode is enabled...")

	// Step 4: Event Subscriptions
	printStep(4, 6, "SUBSCRIBE TO EVENTS")
	fmt.Println()
	fmt.Printf("  %sGo to:%s %sEvent Subscriptions%s > toggle %sON%s\n", colorDim, colorReset, colorBold, colorReset, colorGreen, colorReset)
	fmt.Printf("  %sUnder%s Subscribe to bot events%s, add:%s\n", colorDim, colorBold, colorReset, colorReset)
	fmt.Println()
	events := []string{"message.channels", "message.groups", "message.im", "message.mpim", "reaction_added", "reaction_removed"}
	fmt.Printf("  %s%s%s\n", colorGreen, strings.Join(events, "  "), colorReset)
	fmt.Println()

	waitForEnter(reader, "Press ENTER when events are subscribed...")

	// Step 5: Collect credentials
	printStep(5, 6, "ENTER CREDENTIALS")
	fmt.Println()
	fmt.Printf("  %sGo to%s Basic Information %sto find Client ID and Client Secret.%s\n", colorDim, colorBold, colorDim, colorReset)
	fmt.Printf("  %sYour app-level token (xapp-...) is under%s Basic Information %s>%s App-Level Tokens%s.%s\n", colorDim, colorReset, colorDim, colorReset, colorBold, colorReset)
	fmt.Println()

	clientID := prompt(reader, "Client ID")
	clientSecret := prompt(reader, "Client Secret")
	appToken := prompt(reader, "App-Level Token (xapp-...)")

	if clientID == "" || clientSecret == "" || appToken == "" {
		return fmt.Errorf("all fields are required")
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
	cfg.ClientID = clientID
	cfg.ClientSecret = clientSecret

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	if err := config.SaveAppToken(appToken); err != nil {
		return fmt.Errorf("failed to save app token: %w", err)
	}

	fmt.Printf("\n  %s>>> CONFIG_SAVED%s\n", colorGreen, colorReset)

	// Step 6: OAuth
	printStep(6, 6, "AUTHENTICATE")
	fmt.Println()
	fmt.Printf("  %sOpening browser for Slack OAuth...%s\n", colorDim, colorReset)
	fmt.Println()

	token, err := runOAuth(cfg)
	if err != nil {
		fmt.Printf("  %s>>> AUTH_FAILED: %s%s\n", colorPink, err, colorReset)
		fmt.Printf("  %sYou can retry later with:%s slack-tui auth\n", colorDim, colorReset)
		return nil
	}

	if err := config.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
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

func runOAuth(cfg *config.Config) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			fmt.Fprint(w, "ERROR: No authorization code received.")
			return
		}
		codeCh <- code
		fmt.Fprintf(w, `<!DOCTYPE html>
<html><body style="background:#10141a;color:#f6afef;font-family:'JetBrains Mono',monospace;padding:60px;text-align:center">
<pre style="font-size:10px;line-height:1">
  __  __  ___  _   _  ___  ____  ____   _    ____ _____
 |  \/  |/ _ \| \ | |/ _ \/ ___||  _ \ / \  / ___| ____|
 | |\/| | | | |  \| | | | \___ \| |_) / _ \| |   |  _|
 | |  | | |_| | |\  | |_| |___) |  __/ ___ \ |___| |___
 |_|  |_|\___/|_| \_|\___/|____/|_| /_/   \_\____|_____|
</pre>
<p style="color:#5edda0;font-size:18px">>>> AUTH_SUCCESS</p>
<p style="color:#666">You can close this tab and return to your terminal.</p>
</body></html>`)
	})

	server := &http.Server{Addr: ":9876", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	userScopes := "channels:read,channels:history,groups:read,groups:history," +
		"im:read,im:history,mpim:read,mpim:history," +
		"chat:write,reactions:read,reactions:write," +
		"users:read,files:read,files:write," +
		"pins:read,pins:write,search:read," +
		"team:read,usergroups:read"

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&user_scope=%s&redirect_uri=%s",
		cfg.ClientID, userScopes, "http://localhost:9876/callback",
	)

	fmt.Printf("  %s>>> If the browser doesn't open, visit:%s\n", colorDim, colorReset)
	fmt.Printf("  %s%s%s\n", colorPink, authURL, colorReset)

	_ = browser.OpenURL(authURL)

	fmt.Printf("\n  %sWaiting for authorization...%s\n", colorDim, colorReset)

	var token string
	select {
	case code := <-codeCh:
		var err error
		token, err = config.ExchangeCode(cfg.ClientID, cfg.ClientSecret, code)
		if err != nil {
			return "", err
		}
		fmt.Printf("  %s>>> AUTH_COMPLETE — token saved to keyring.%s\n", colorGreen, colorReset)

	case err := <-errCh:
		return "", err

	case <-time.After(5 * time.Minute):
		return "", fmt.Errorf("timed out after 5 minutes")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	return token, nil
}

func printBanner() {
	fmt.Println()
	fmt.Printf("  %s╔══════════════════════════════════════════╗%s\n", colorPink, colorReset)
	fmt.Printf("  %s║  MONOSPACE_CMD_V1.0 — SETUP_WIZARD      ║%s\n", colorPink, colorReset)
	fmt.Printf("  %s╚══════════════════════════════════════════╝%s\n", colorPink, colorReset)
	fmt.Println()
	fmt.Printf("  %sThis wizard will walk you through connecting\n  to your Slack workspace in 6 steps.%s\n", colorDim, colorReset)
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
