# MONOSPACE_CMD_V1.0 — Design Beads

Visual design reference for the Slack TUI terminal client. Open any `.html` file in a browser to preview.

## Beads

| # | File | View | Description |
|---|------|------|-------------|
| 01 | `01-design-system.html` | Main Workspace | Channel sidebar, message feed, AI hook interjections, thread detail panel, command prompt |
| 02 | `02-config-view.html` | Configuration | Vim-like YAML editor with line numbers, syntax highlighting, terminal overlay |
| 03 | `03-ai-interaction.html` | AI Interaction | Split-pane: channel context buffer (left) + AI analysis with summary & draft replies (right) |
| 04 | `04-thread-view.html` | Thread + AI Analysis | ASCII-boxed original post, threaded replies with tree lines, AI_CONFETTI sidebar with sentiment/takeaways |

## Design Tokens

- **Primary**: `#f6afef` (pink — accents, branding, AI elements)
- **Secondary**: `#5edda0` (green — success, active states, online indicators)
- **Background**: `#10141a` (deep dark)
- **Surface**: `#181c22` / `#1c2026` (containers)
- **Primary Container**: `#4a154b` (Slack purple — buttons, tags)
- **On-Surface**: `#dfe2eb` (body text)
- **Error**: `#ffb4ab` (AI draft text, warnings)
- **Outline**: `#4f434c` (borders, dividers)

## Typography

- **Headlines**: Space Grotesk (italic, tracking-widest)
- **Body/Mono**: JetBrains Mono (everything else — this is a terminal app)
- **Labels**: Inter (when needed for clarity)

## Key Aesthetic Rules

- Zero border-radius (everything is sharp rectangles)
- UPPERCASE labels and section headers
- ASCII art dividers and box-drawing characters
- Block cursor animation (`█` blinking pink)
- IRC-style message format: `[HH:MM:SS] <username> message`
- Bracket-wrapped commands: `[ ACTION_NAME ]`
- Underscore-separated identifiers: `AI_SUMMARIZER_BOT`
