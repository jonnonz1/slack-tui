```
  __  __  ___  _   _  ___  ____  ____   _    ____ _____
 |  \/  |/ _ \| \ | |/ _ \/ ___||  _ \ / \  / ___| ____|
 | |\/| | | | |  \| | | | \___ \| |_) / _ \| |   |  _|
 | |  | | |_| | |\  | |_| |___) |  __/ ___ \ |___| |___
 |_|  |_|\___/|_| \_|\___/|____/|_| /_/   \_\____|_____|
                   ___ __  __ ____
                  / __|  \/  |  _ \
                 | (__|  |\/| | | | |
                  \___|_|  |_|_| |_|  v1.0
```

# MONOSPACE_CMD

A retro terminal Slack client with integrated AI hooks. Built with Go + [Bubbletea](https://github.com/charmbracelet/bubbletea). Single workspace, keyboard-first, zero border-radius.

## Features

- **Channel browsing** — public, private, DMs, group DMs
- **Real-time messaging** — Socket Mode for live updates
- **Threaded conversations** — open, read, and reply to threads
- **Reactions** — add/remove emoji reactions
- **Search** — workspace-wide message search (Ctrl+F)
- **Quick switch** — fuzzy channel picker (Ctrl+K)

### AI Hooks

LLM-powered features via the Anthropic Claude API:

| Hook | Trigger | What it does |
|------|---------|-------------|
| **AI-SUMMARIZER** | `Ctrl+S` | Summarizes the last 20 messages in the active channel |
| **DRAFT-BOT** | `Ctrl+D` | Generates 3 reply drafts with different tones (assertive, clarification, status update) |
| **AI_CONFETTI** | Auto on thread open | Sentiment analysis, key takeaways, participant mapping |

AI hooks require an `ANTHROPIC_API_KEY` environment variable. They use Claude Sonnet by default (configurable).

## Screenshots

Design beads are in [`beads/`](beads/) — open any `.html` file in a browser to preview the target aesthetic.

## Quick Start

```bash
go install github.com/jonnonz1/slack-tui@latest
slack-tui
```

That's it. If no Slack workspace is configured, the interactive setup wizard launches automatically and walks you through everything step by step.

### What the setup wizard does

The wizard runs in 5 steps — it opens the right pages in your browser and tells you exactly what to click:

1. **Create Slack App** — opens [api.slack.com/apps](https://api.slack.com/apps), you create a new app
2. **Add Scopes & Install** — lists the exact scopes to add, then you install to your workspace and copy the User OAuth Token
3. **Enable Socket Mode** — guides you to create an app-level token
4. **Subscribe to Events** — lists the exact events to subscribe to
5. **Enter Tokens** — paste your User OAuth Token (`xoxp-...`) and App-Level Token (`xapp-...`)

Config is saved to `~/.config/slack-tui/config.json`. Tokens are stored securely in your OS keychain.

### Build from source

```bash
git clone https://github.com/jonnonz1/slack-tui.git
cd slack-tui
go build -o slack-tui .
./slack-tui
```

### For AI features

```bash
export ANTHROPIC_API_KEY=sk-ant-...
slack-tui
```

## Commands

| Command | Description |
|---------|-------------|
| `slack-tui` | Launch the TUI (runs setup wizard if not configured) |
| `slack-tui setup` | Re-run the interactive setup wizard |
| `slack-tui auth` | Update your Slack token (e.g. if it expired) |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SLACK_TOKEN` | No | Override user token (skip keychain) |
| `SLACK_APP_TOKEN` | No | Override app-level token (skip keychain) |
| `ANTHROPIC_API_KEY` | For AI hooks | Claude API key for summarizer, drafter, analyzer |

## Key Bindings

```
Navigation
  Tab / Shift+Tab     Cycle focus: sidebar > messages > input
  j/k                 Move up/down in lists
  Enter               Select channel / open thread
  Esc                 Close panel / back
  Ctrl+K              Quick channel switcher
  Ctrl+F              Search messages
  Ctrl+N / Ctrl+P     Next/prev unread channel

Messages
  i                   Focus input
  Enter               Send message
  t                   Open thread
  r                   Add reaction

AI Hooks
  Ctrl+S              Summarize channel
  Ctrl+D              Generate draft replies
  Ctrl+A              Toggle AI panel

Global
  Ctrl+C              Quit
  ?                   Help
```

## Configuration

Config lives at `~/.config/slack-tui/config.json` and is created automatically by the setup wizard:

```json
{
  "client_id": "...",
  "client_secret": "...",
  "theme": {
    "primary": "#f6afef",
    "secondary": "#5edda0",
    "surface": "#10141a"
  },
  "ai": {
    "summarizer": {
      "enabled": true,
      "provider": "anthropic",
      "model": "claude-sonnet-4-6"
    },
    "drafter": {
      "enabled": true,
      "provider": "anthropic",
      "model": "claude-sonnet-4-6"
    },
    "analyzer": {
      "enabled": true,
      "provider": "anthropic",
      "model": "claude-haiku-4-5-20251001"
    }
  },
  "sidebar_width": 25,
  "time_format": "15:04:05"
}
```

Tokens (`xoxp-` user token and `xapp-` app-level token) are stored in your OS keychain via [go-keyring](https://github.com/zalando/go-keyring), not in the config file.

## Architecture

```
slack-tui/
  cmd/                     CLI entry points (cobra)
    root.go                  command registration
    setup.go                 interactive 6-step setup wizard + OAuth
    auth.go                  re-authentication
    run.go                   TUI launch (auto-detects missing config)
  internal/
    app/                   Root Bubbletea model, theme, keymap
    ui/
      sidebar/             Channel list with unread indicators
      messages/            Message viewport with AI hook rendering
      input/               Text input with block cursor
      thread/              Thread panel with AI analysis sidebar
      modal/               Quick switcher (Ctrl+K), search (Ctrl+F)
      statusbar/           Connection status, current channel, user
    slack/                 Web API wrapper, Socket Mode bridge, cache
    ai/                    LLM engine — summarizer, drafter, analyzer
    markdown/              Slack mrkdwn -> terminal renderer
    config/                Settings, keychain token storage
  beads/                   HTML design mockups (4 views)
```

## Tech Stack

- **[Bubbletea v2](https://github.com/charmbracelet/bubbletea)** — TUI framework (Elm Architecture)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling
- **[slack-go/slack](https://github.com/slack-go/slack)** — Slack Web API + Socket Mode
- **[Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go)** — Claude API for AI hooks
- **[go-keyring](https://github.com/zalando/go-keyring)** — Secure OS keychain token storage
- **[Cobra](https://github.com/spf13/cobra)** — CLI framework

## License

[MIT](LICENSE)
