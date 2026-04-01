package app

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/ai"
	"github.com/jonnonz1/slack-tui/internal/config"
	"github.com/jonnonz1/slack-tui/internal/slack"
	"github.com/jonnonz1/slack-tui/internal/ui/input"
	"github.com/jonnonz1/slack-tui/internal/ui/messages"
	"github.com/jonnonz1/slack-tui/internal/ui/sidebar"
	"github.com/jonnonz1/slack-tui/internal/ui/statusbar"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// Focus tracks which panel has keyboard focus.
type Focus int

const (
	FocusSidebar Focus = iota
	FocusMessages
	FocusInput
	FocusThread
)

// Model is the root bubbletea model for the TUI.
type Model struct {
	client    *slack.Client
	cfg       *config.Config
	theme     Theme
	keymap    KeyMap
	aiEngine  *ai.Engine
	focus     Focus
	width     int
	height    int
	connected bool
	err       string

	sidebar   sidebar.Model
	messages  messages.Model
	input     input.Model
	statusbar statusbar.Model

	activeChannel string
}

func New(client *slack.Client, cfg *config.Config) Model {
	t := DefaultTheme
	km := DefaultKeyMap
	engine := ai.NewEngine(cfg.AI.Summarizer.Provider, cfg.AI.Summarizer.Model)

	return Model{
		client:   client,
		cfg:      cfg,
		theme:    t,
		keymap:   km,
		aiEngine: engine,
		focus:    FocusSidebar,
		sidebar:  sidebar.New(client),
		messages: messages.New(client, t.Username, t.Timestamp, t.OnSurface, t.Primary, t.Outline),
		input:    input.New(),
		statusbar: statusbar.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.sidebar.Init(),
		m.statusbar.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resize()

	case tea.KeyMsg:
		cmd := m.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case slack.ConnectionStatusEvent:
		m.connected = msg.Connected
		m.statusbar = m.statusbar.SetConnected(msg.Connected)

	case slack.NewMessageEvent:
		m.messages = m.messages.AppendMessage(msg.Message)
		m.sidebar = m.sidebar.IncrementUnread(msg.Message.ChannelID)

	case slack.ReactionAddedEvent:
		m.messages = m.messages.AddReaction(msg.MessageTS, msg.Reaction, msg.UserID)

	case slack.ReactionRemovedEvent:
		m.messages = m.messages.RemoveReaction(msg.MessageTS, msg.Reaction, msg.UserID)

	case sidebar.ChannelSelectedMsg:
		m.activeChannel = msg.ChannelID
		m.focus = FocusMessages
		m.statusbar = m.statusbar.SetChannel(msg.ChannelName)
		cmd := m.messages.LoadChannel(msg.ChannelID)
		cmds = append(cmds, cmd)

	case input.SendMessageMsg:
		if m.activeChannel != "" && msg.Text != "" {
			cmds = append(cmds, m.sendMessage(msg.Text))
		}

	case ai.SummaryResultMsg:
		m.messages = m.messages.AppendAIHook(msg)

	case ai.DraftResultMsg:
		m.messages = m.messages.ShowDrafts(msg)
	}

	// Always delegate to sidebar so it receives ChannelsLoadedMsg
	{
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Delegate to focused child
	switch m.focus {
	case FocusMessages:
		var cmd tea.Cmd
		m.messages, cmd = m.messages.Update(msg)
		cmds = append(cmds, cmd)
	case FocusInput:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	v := tea.View{AltScreen: true}

	if m.width == 0 || m.height == 0 {
		v.Content = "  Loading..."
		return v
	}

	sidebarWidth := m.cfg.SidebarWidth
	if sidebarWidth == 0 {
		sidebarWidth = 25
	}
	mainWidth := m.width - sidebarWidth - 1 // -1 for border
	if mainWidth < 10 {
		mainWidth = 10
	}

	inputHeight := 3
	statusHeight := 1
	contentHeight := m.height - statusHeight - inputHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Build sidebar
	sidebarContent := m.sidebar.View(sidebarWidth, contentHeight+inputHeight, m.focus == FocusSidebar)
	sidebarBox := boxed(sidebarContent, sidebarWidth, contentHeight+inputHeight)

	// Build messages
	msgContent := m.messages.View(mainWidth, contentHeight, m.focus == FocusMessages)
	msgBox := boxed(msgContent, mainWidth, contentHeight)

	// Build input
	inputContent := m.input.View(mainWidth, inputHeight, m.focus == FocusInput, m.activeChannel)
	inputBox := boxed(inputContent, mainWidth, inputHeight)

	// Right side: messages + input stacked
	rightPane := msgBox + "\n" + inputBox

	// Main layout: sidebar | right pane (line by line)
	sidebarLines := strings.Split(sidebarBox, "\n")
	rightLines := strings.Split(rightPane, "\n")

	var screen strings.Builder
	totalLines := contentHeight + inputHeight
	for i := 0; i < totalLines; i++ {
		sl := ""
		if i < len(sidebarLines) {
			sl = sidebarLines[i]
		}
		rl := ""
		if i < len(rightLines) {
			rl = rightLines[i]
		}
		// Pad sidebar to exact width
		sl = padRight(sl, sidebarWidth)
		sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#4f434c")).Render("│")
		screen.WriteString(sl + sep + rl + "\n")
	}

	// Status bar
	statusLine := m.statusbar.View(m.width)

	v.Content = screen.String() + statusLine
	return v
}

// boxed constrains content to exactly width x height with truncation and padding.
func boxed(content string, width, height int) string {
	lines := strings.Split(content, "\n")

	var out strings.Builder
	for i := 0; i < height; i++ {
		if i > 0 {
			out.WriteString("\n")
		}
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		out.WriteString(padRight(truncate(line, width), width))
	}
	return out.String()
}

// truncate cuts a string to max visible width (ANSI-aware approximation).
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	visible := 0
	inEsc := false
	cutIdx := len(s)
	for i, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		visible++
		if visible > max {
			cutIdx = i
			break
		}
	}
	return s[:cutIdx]
}

// padRight pads a string with spaces to reach the target visible width.
func padRight(s string, target int) string {
	visible := visibleLen(s)
	if visible >= target {
		return s
	}
	return s + strings.Repeat(" ", target-visible)
}

// visibleLen counts non-ANSI characters.
func visibleLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}

func (m Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.Key()

	if key == m.keymap.Quit {
		return tea.Quit
	}

	if key == m.keymap.FocusNext {
		m.cycleFocus(1)
		return nil
	}

	if key == m.keymap.FocusPrev {
		m.cycleFocus(-1)
		return nil
	}

	if key == m.keymap.ClosePanel && m.focus != FocusSidebar {
		m.focus = FocusMessages
		return nil
	}

	if key == m.keymap.FocusInput {
		m.focus = FocusInput
		return nil
	}

	if key == m.keymap.AISummarize && m.activeChannel != "" {
		return m.aiEngine.Summarize(m.activeChannel, m.messages.RecentTexts(20))
	}

	if key == m.keymap.AIDraft && m.activeChannel != "" {
		return m.aiEngine.Draft(m.activeChannel, m.messages.RecentTexts(20))
	}

	return nil
}

func (m *Model) cycleFocus(dir int) {
	focuses := []Focus{FocusSidebar, FocusMessages, FocusInput}
	for i, f := range focuses {
		if f == m.focus {
			m.focus = focuses[(i+dir+len(focuses))%len(focuses)]
			return
		}
	}
}

func (m Model) resize() Model {
	sidebarWidth := m.cfg.SidebarWidth
	if sidebarWidth == 0 {
		sidebarWidth = 25
	}
	mainWidth := m.width - sidebarWidth - 1
	if mainWidth < 10 {
		mainWidth = 10
	}
	inputHeight := 3
	statusHeight := 1
	contentHeight := m.height - statusHeight - inputHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	m.sidebar = m.sidebar.SetSize(sidebarWidth, contentHeight+inputHeight)
	m.messages = m.messages.SetSize(mainWidth, contentHeight)
	m.input = m.input.SetSize(mainWidth, inputHeight)
	return m
}

func (m Model) sendMessage(text string) tea.Cmd {
	channelID := m.activeChannel
	client := m.client
	return func() tea.Msg {
		err := client.SendMessage(channelID, text)
		if err != nil {
			return input.SendErrorMsg{Err: err}
		}
		return nil
	}
}

// Ensure fmt is used (for future error display)
var _ = fmt.Sprintf
