package thread

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/ai"
	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// ThreadLoadedMsg is returned after fetching thread replies.
type ThreadLoadedMsg struct {
	ChannelID string
	ThreadTS  string
	Messages  []slack.Message
	Err       error
}

type Model struct {
	client    *slack.Client
	messages  []slack.Message
	channelID string
	threadTS  string
	cursor    int
	offset    int
	width     int
	height    int
	open      bool
	loading   bool

	analysis *ai.AnalysisResultMsg
}

func New(client *slack.Client) Model {
	return Model{
		client: client,
	}
}

func (m Model) Open(channelID, threadTS string) (Model, tea.Cmd) {
	m.channelID = channelID
	m.threadTS = threadTS
	m.open = true
	m.loading = true
	m.messages = nil
	m.analysis = nil
	m.cursor = 0
	m.offset = 0

	client := m.client
	return m, func() tea.Msg {
		msgs, err := client.GetReplies(channelID, threadTS)
		return ThreadLoadedMsg{
			ChannelID: channelID,
			ThreadTS:  threadTS,
			Messages:  msgs,
			Err:       err,
		}
	}
}

func (m Model) Close() Model {
	m.open = false
	return m
}

func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

func (m Model) SetAnalysis(a ai.AnalysisResultMsg) Model {
	m.analysis = &a
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ThreadLoadedMsg:
		if msg.ThreadTS == m.threadTS && msg.Err == nil {
			m.messages = msg.Messages
			m.loading = false
			m.cursor = len(m.messages) - 1
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.messages)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "escape", "q":
			m.open = false
		}
	}

	return m, nil
}

func (m Model) View(width, height int, focused bool) string {
	if !m.open || width == 0 || height == 0 {
		return ""
	}

	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Foreground(lipgloss.Color("#f6afef")).
		Render("THREAD_DETAILS")
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	if m.loading {
		b.WriteString("  Loading thread...")
		return m.containerStyle(width, height, focused).Render(b.String())
	}

	// Thread messages
	for i, msg := range m.messages {
		isOP := i == 0

		tsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5edda0")).Bold(true)

		if isOP {
			userStyle = userStyle.Foreground(lipgloss.Color("#f6afef"))
		}

		ts := tsStyle.Render(msg.Timestamp.Format("15:04"))
		user := userStyle.Render(msg.Username)

		tag := ""
		if isOP {
			tag = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render(" [OP]")
		}

		b.WriteString(fmt.Sprintf("  %s %s%s\n", ts, user, tag))

		// Tree line for non-OP messages
		prefix := "  │ "
		if isOP {
			prefix = "  "
		}
		b.WriteString(prefix + msg.Text + "\n")

		if len(msg.Reactions) > 0 {
			var rxns []string
			for _, r := range msg.Reactions {
				rxns = append(rxns, fmt.Sprintf("[:%s: %d]", r.Name, r.Count))
			}
			b.WriteString(prefix + strings.Join(rxns, " ") + "\n")
		}

		b.WriteString("\n")
	}

	// AI Analysis sidebar (like bead 04)
	if m.analysis != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5edda0")).
			Bold(true).
			Render("[ AI_CONFETTI_PROCESSOR ]"))
		b.WriteString("\n\n")

		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render("SENTIMENT: "))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5edda0")).
			Render(m.analysis.Sentiment))
		b.WriteString("\n\n")

		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render("KEY_TAKEAWAYS"))
		b.WriteString("\n")
		for _, t := range m.analysis.Takeaways {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#dfe2eb")).
				Render("  > " + t))
			b.WriteString("\n")
		}
	}

	return m.containerStyle(width, height, focused).Render(b.String())
}

func (m Model) containerStyle(width, height int, focused bool) lipgloss.Style {
	borderColor := lipgloss.Color("#4f434c")
	if focused {
		borderColor = lipgloss.Color("#f6afef")
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)
}
