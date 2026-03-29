package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Slack via OAuth",
	Long:  "Opens your browser to authorize the app with your Slack workspace and stores the token securely.",
	RunE:  runAuth,
}

func runAuth(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return fmt.Errorf("client_id and client_secret must be set in config — run 'slack-tui setup' first")
	}

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
		fmt.Fprint(w, `
			<html><body style="background:#10141a;color:#f6afef;font-family:monospace;padding:40px;text-align:center">
			<h1>MONOSPACE_CMD</h1>
			<p style="color:#5edda0">>>> AUTH_SUCCESS — you can close this tab.</p>
			</body></html>
		`)
	})

	server := &http.Server{
		Addr:    ":9876",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	scopes := "channels:read,channels:history,groups:read,groups:history," +
		"im:read,im:history,mpim:read,mpim:history," +
		"chat:write,reactions:read,reactions:write," +
		"users:read,files:read,files:write," +
		"pins:read,pins:write,search:read," +
		"team:read,usergroups:read"

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&user_scope=%s&redirect_uri=%s",
		cfg.ClientID,
		scopes,
		"http://localhost:9876/callback",
	)

	fmt.Println("[MONOSPACE_CMD] >>> Opening browser for Slack authorization...")
	fmt.Println("[MONOSPACE_CMD] >>> If the browser doesn't open, visit:")
	fmt.Printf("    %s\n", authURL)

	_ = browser.OpenURL(authURL)

	select {
	case code := <-codeCh:
		token, err := config.ExchangeCode(cfg.ClientID, cfg.ClientSecret, code)
		if err != nil {
			return fmt.Errorf("token exchange failed: %w", err)
		}
		if err := config.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}
		fmt.Println("[MONOSPACE_CMD] >>> AUTH_COMPLETE — token saved to keyring.")

	case err := <-errCh:
		return err

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("auth timed out after 5 minutes")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	return nil
}
