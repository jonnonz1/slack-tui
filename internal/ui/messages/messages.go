package messages

import (
	"fmt"
	"strings"

	"github.com/jonnonz1/slack-tui/internal/ai"
	"github.com/jonnonz1/slack-tui/internal/slack"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type MessagesLoadedMsg struct {
	ChannelID string
	Messages  []slack.Message
	Err       error
}

type Model struct {
	client         *slack.Client
	messages       []slack.Message
	channelID      string
	cursor         int
	offset         int
	width          int
	height         int
	loading        bool
	aiSummary      *ai.SummaryResultMsg
	aiDrafts       *ai.DraftResultMsg
	usernameColor  lipgloss.Color
	timestampColor lipgloss.Color
	textColor      lipgloss.Color
	aiColor        lipgloss.Color
	borderColor    lipgloss.Color
}

func New(client *slack.Client, username, timestamp, text, aiClr, border lipgloss.Color) Model {
	return Model{
		client:         client,
		usernameColor:  username,
		timestampColor: timestamp,
		textColor:      text,
		aiColor:        aiClr,
		borderColor:    border,
	}
}

func (m Model) LoadChannel(channelID string) tea.Cmd {
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
		if msg.Err == nil {
			m.channelID = msg.ChannelID
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
					Name: reaction, Count: 1, Users: []string{userID},
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

func (m Model) SelectedMessage() (slack.Message, bool) {
	if m.cursor >= 0 && m.cursor < len(m.messages) {
		return m.messages[m.cursor], true
	}
	return slack.Message{}, false
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

	if m.loading {
		return "  Loading messages..."
	}

	if len(m.messages) == 0 && m.channelID == "" {
		return "  Select a channel (j/k, Enter)"
	}

	if len(m.messages) == 0 {
		return "  No messages yet."
	}

	tsStyle := lipgloss.NewStyle().Foreground(m.timestampColor)
	userStyle := lipgloss.NewStyle().Foreground(m.usernameColor).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.textColor)
	rxnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9b8d97"))
	threadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f6afef"))

	var lines []string

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

		ts := tsStyle.Render(fmt.Sprintf("[%s]", msg.Timestamp.Format("15:04")))
		user := userStyle.Render(fmt.Sprintf("<%s>", msg.Username))
		text := textStyle.Render(msg.Text)

		prefix := " "
		if i == m.cursor && focused {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %s %s %s", prefix, ts, user, text))

		if len(msg.Reactions) > 0 {
			var rxns []string
			for _, r := range msg.Reactions {
				rxns = append(rxns, fmt.Sprintf("[:%s: %d]", r.Name, r.Count))
			}
			lines = append(lines, rxnStyle.Render("    "+strings.Join(rxns, " ")))
		}

		if msg.ReplyCount > 0 {
			lines = append(lines, threadStyle.Render(fmt.Sprintf("    [%d replies]", msg.ReplyCount)))
		}
	}

	return strings.Join(lines, "\n")
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
