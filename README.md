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

## Installation

```bash
go install github.com/jonnonz1/slack-tui@latest
```

Or build from source:

```bash
git clone https://github.com/jonnonz1/slack-tui.git
cd slack-tui
go build -o slack-tui .
```

## Setup

### 1. Create a Slack App

```bash
slack-tui setup
```

This walks you through:
1. Creating an app at [api.slack.com/apps](https://api.slack.com/apps)
2. Adding User Token Scopes (channels, messages, reactions, search, etc.)
3. Enabling Socket Mode with an app-level token
4. Subscribing to message events

### 2. Authenticate

```bash
slack-tui auth
```

Opens your browser for OAuth. The token is stored in your OS keychain.

### 3. Run

```bash
slack-tui
```

For AI features:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
slack-tui
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SLACK_TOKEN` | No | Override user token (skip keychain) |
| `SLACK_APP_TOKEN` | No | Override app-level token (skip keychain) |
| `ANTHROPIC_API_KEY` | For AI hooks | Claude API key |

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

Config lives at `~/.config/monospace-cmd/config.json`:

```json
{
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

## Architecture

```
slack-tui/
  cmd/                     CLI entry points (cobra)
    root.go, auth.go, setup.go, run.go
  internal/
    app/                   Root Bubbletea model, theme, keymap
    ui/
      sidebar/             Channel list
      messages/            Message viewport with AI hook rendering
      input/               Text input with block cursor
      thread/              Thread panel with AI analysis
      modal/               Quick switcher (Ctrl+K), search (Ctrl+F)
      statusbar/           Connection status, current channel
    slack/                 Web API wrapper, Socket Mode bridge, cache
    ai/                    LLM engine — summarizer, drafter, analyzer
    markdown/              Slack mrkdwn -> terminal renderer
    config/                Settings, keychain token storage
  beads/                   HTML design mockups
```

## Tech Stack

- **[Bubbletea v2](https://github.com/charmbracelet/bubbletea)** — TUI framework (Elm Architecture)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling
- **[slack-go/slack](https://github.com/slack-go/slack)** — Slack Web API + Socket Mode
- **[Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go)** — Claude API for AI hooks
- **[go-keyring](https://github.com/zalando/go-keyring)** — Secure token storage

## License

[MIT](LICENSE)
