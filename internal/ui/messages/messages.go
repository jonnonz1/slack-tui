package messages

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/ai"
	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// MessagesLoadedMsg is returned after fetching channel history.
type MessagesLoadedMsg struct {
	ChannelID string
	Messages  []slack.Message
	Err       error
}

type Model struct {
	client    *slack.Client
	messages  []slack.Message
	channelID string
	cursor    int
	offset    int
	width     int
	height    int
	loading   bool

	// AI hook display state
	aiSummary *ai.SummaryResultMsg
	aiDrafts  *ai.DraftResultMsg

	// Style colors
	usernameColor lipgloss.Color
	timestampColor lipgloss.Color
	textColor     lipgloss.Color
	aiColor       lipgloss.Color
	borderColor   lipgloss.Color
}

func New(client *slack.Client, username, timestamp, text, aiClr, border lipgloss.Color) Model {
	return Model{
		client:        client,
		usernameColor: username,
		timestampColor: timestamp,
		textColor:     text,
		aiColor:       aiClr,
		borderColor:   border,
	}
}

func (m Model) LoadChannel(channelID string) tea.Cmd {
	m.channelID = channelID
	client := m.client
	return func() tea.Msg {
		msgs, err := client.GetHistory(channelID, 50)
		return MessagesLoadedMsg{
			ChannelID: channelID,
			Messages:  msgs,
			Err:       err,
		}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case MessagesLoadedMsg:
		if msg.Err == nil && msg.ChannelID == m.channelID {
			m.messages = msg.Messages
			m.loading = false
			m.cursor = len(m.messages) - 1
			m.scrollToBottom()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.messages)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "G":
			m.cursor = len(m.messages) - 1
			m.scrollToBottom()
		case "g":
			m.cursor = 0
			m.offset = 0
		}
	}

	return m, nil
}

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

func (m Model) AppendMessage(msg slack.Message) Model {
	if msg.ChannelID == m.channelID {
		m.messages = append(m.messages, msg)
		// Auto-scroll if at bottom
		if m.cursor >= len(m.messages)-2 {
			m.cursor = len(m.messages) - 1
			m.scrollToBottom()
		}
	}
	return m
}

func (m Model) AddReaction(messageTS, reaction, userID string) Model {
	for i := range m.messages {
		if m.messages[i].Timestamp.String() == messageTS {
			found := false
			for j := range m.messages[i].Reactions {
				if m.messages[i].Reactions[j].Name == reaction {
					m.messages[i].Reactions[j].Count++
					m.messages[i].Reactions[j].Users = append(m.messages[i].Reactions[j].Users, userID)
					found = true
					break
				}
			}
			if !found {
				m.messages[i].Reactions = append(m.messages[i].Reactions, slack.Reaction{
					Name:  reaction,
					Count: 1,
					Users: []string{userID},
				})
			}
			break
		}
	}
	return m
}

func (m Model) RemoveReaction(messageTS, reaction, userID string) Model {
	for i := range m.messages {
		if m.messages[i].Timestamp.String() == messageTS {
			for j := range m.messages[i].Reactions {
				if m.messages[i].Reactions[j].Name == reaction {
					m.messages[i].Reactions[j].Count--
					if m.messages[i].Reactions[j].Count <= 0 {
						m.messages[i].Reactions = append(m.messages[i].Reactions[:j], m.messages[i].Reactions[j+1:]...)
					}
					break
				}
			}
			break
		}
	}
	return m
}

func (m Model) AppendAIHook(summary ai.SummaryResultMsg) Model {
	m.aiSummary = &summary
	return m
}

func (m Model) ShowDrafts(drafts ai.DraftResultMsg) Model {
	m.aiDrafts = &drafts
	return m
}

func (m Model) RecentTexts(n int) []string {
	start := 0
	if len(m.messages) > n {
		start = len(m.messages) - n
	}
	texts := make([]string, 0, n)
	for _, msg := range m.messages[start:] {
		texts = append(texts, fmt.Sprintf("<%s> %s", msg.Username, msg.Text))
	}
	return texts
}

func (m Model) View(width, height int, focused bool) string {
	if width == 0 || height == 0 {
		return ""
	}

	var b strings.Builder

	if m.loading {
		b.WriteString("\n  Loading messages...")
		return m.containerStyle(width, height, focused).Render(b.String())
	}

	if len(m.messages) == 0 {
		b.WriteString("\n  No messages yet.")
		return m.containerStyle(width, height, focused).Render(b.String())
	}

	tsStyle := lipgloss.NewStyle().Foreground(m.timestampColor)
	userStyle := lipgloss.NewStyle().Foreground(m.usernameColor).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.textColor)
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("#181c22"))

	visibleEnd := m.offset + height
	if visibleEnd > len(m.messages) {
		visibleEnd = len(m.messages)
	}
	visibleStart := m.offset
	if visibleStart < 0 {
		visibleStart = 0
	}

	for i := visibleStart; i < visibleEnd; i++ {
		msg := m.messages[i]

		ts := tsStyle.Render(fmt.Sprintf("[%s]", msg.Timestamp.Format("15:04:05")))
		user := userStyle.Render(fmt.Sprintf("<%s>", msg.Username))
		text := textStyle.Render(msg.Text)

		line := fmt.Sprintf("%s %s %s", ts, user, text)

		// Truncate to width
		if len(line) > width-2 {
			line = line[:width-5] + "..."
		}

		if i == m.cursor && focused {
			line = selectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Reactions
		if len(msg.Reactions) > 0 {
			var rxns []string
			for _, r := range msg.Reactions {
				rxns = append(rxns, fmt.Sprintf("[:%s: %d]", r.Name, r.Count))
			}
			rxnLine := "  " + strings.Join(rxns, " ")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#9b8d97")).Render(rxnLine))
			b.WriteString("\n")
		}

		// Thread indicator
		if msg.ReplyCount > 0 {
			thread := fmt.Sprintf("  [%d replies]", msg.ReplyCount)
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#f6afef")).Render(thread))
			b.WriteString("\n")
		}
	}

	// AI Summary hook (inline, like bead 01)
	if m.aiSummary != nil {
		b.WriteString("\n")
		aiHeader := lipgloss.NewStyle().
			Foreground(m.aiColor).
			Bold(true).
			Render("⚡ AI-SUMMARIZER_BOT [AUTO_SCAN]")
		b.WriteString(aiHeader)
		b.WriteString("\n")
		for _, point := range m.aiSummary.Points {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#dfe2eb")).
				Italic(true).
				Render("  • " + point))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// AI Draft options (like bead 03)
	if m.aiDrafts != nil {
		b.WriteString("\n")
		draftHeader := lipgloss.NewStyle().
			Foreground(m.aiColor).
			Bold(true).
			Render("⚡ DRAFT_REPLIES_GENERATED")
		b.WriteString(draftHeader)
		b.WriteString("\n")
		for i, draft := range m.aiDrafts.Drafts {
			label := lipgloss.NewStyle().
				Foreground(m.aiColor).
				Render(fmt.Sprintf("  [%d: %s]", i+1, draft.Tone))
			conf := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9b8d97")).
				Render(fmt.Sprintf(" CONFIDENCE: %d%%", draft.Confidence))
			b.WriteString(label)
			b.WriteString(conf)
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#dfe2eb")).
				Render("  " + draft.Text))
			b.WriteString("\n")
		}
	}

	return m.containerStyle(width, height, focused).Render(b.String())
}

func (m Model) containerStyle(width, height int, focused bool) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(lipgloss.Color("#0a0e14"))
}

func (m *Model) scrollToBottom() {
	if len(m.messages) > m.height {
		m.offset = len(m.messages) - m.height
	} else {
		m.offset = 0
	}
}

func (m *Model) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}
