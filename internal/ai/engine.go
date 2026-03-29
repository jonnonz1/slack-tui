package ai

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anthropics/anthropic-sdk-go"
)

// SummaryResultMsg is the result of an AI summarization.
type SummaryResultMsg struct {
	ChannelID string
	Points    []string
	Err       error
}

// DraftResultMsg is the result of AI draft generation.
type DraftResultMsg struct {
	ChannelID string
	Drafts    []Draft
	Err       error
}

// Draft is a single AI-generated reply option.
type Draft struct {
	Tone       string // e.g., "ASSERTIVE", "CLARIFICATION", "STATUS_UPDATE"
	Text       string
	Confidence int
}

// AnalysisResultMsg is the result of thread analysis.
type AnalysisResultMsg struct {
	ThreadTS   string
	Sentiment  string
	Score      float64
	Takeaways  []string
	Err        error
}

// Engine manages LLM interactions for all AI hooks.
type Engine struct {
	provider string
	model    string
	client   *anthropic.Client
}

func NewEngine(provider, model string) *Engine {
	// The Anthropic SDK reads ANTHROPIC_API_KEY from env automatically
	client := anthropic.NewClient()

	return &Engine{
		provider: provider,
		model:    model,
		client:   &client,
	}
}

// Summarize generates a bullet-point summary of recent channel messages.
func (e *Engine) Summarize(channelID string, recentMessages []string) tea.Cmd {
	return func() tea.Msg {
		if len(recentMessages) == 0 {
			return SummaryResultMsg{ChannelID: channelID, Points: []string{"No messages to summarize."}}
		}

		context := strings.Join(recentMessages, "\n")
		prompt := fmt.Sprintf(`You are an AI assistant embedded in a terminal Slack client called MONOSPACE_CMD.
Analyze the following channel messages and provide a concise summary as 3-5 bullet points.
Focus on: key decisions, action items, blockers, and important context.
Be terse and technical — this is for engineers in a terminal.

Messages:
%s

Respond with ONLY bullet points, one per line, starting with "• ".`, context)

		result, err := e.complete(prompt)
		if err != nil {
			return SummaryResultMsg{ChannelID: channelID, Err: err}
		}

		points := parsePoints(result)
		return SummaryResultMsg{ChannelID: channelID, Points: points}
	}
}

// Draft generates multiple reply options with different tones.
func (e *Engine) Draft(channelID string, recentMessages []string) tea.Cmd {
	return func() tea.Msg {
		if len(recentMessages) == 0 {
			return DraftResultMsg{ChannelID: channelID, Err: fmt.Errorf("no context")}
		}

		context := strings.Join(recentMessages, "\n")
		prompt := fmt.Sprintf(`You are an AI assistant embedded in a terminal Slack client called MONOSPACE_CMD.
Based on the conversation below, generate exactly 3 reply drafts that the user could send.

Each draft should have a different tone:
1. ASSERTIVE — confident, action-oriented
2. CLARIFICATION — asks a clarifying question
3. STATUS_UPDATE — provides a status or acknowledgment

Format each draft as:
TONE: <tone_name>
CONFIDENCE: <0-100>
TEXT: <the draft message>
---

Keep each draft under 2 sentences. Be natural and technical.

Conversation:
%s`, context)

		result, err := e.complete(prompt)
		if err != nil {
			return DraftResultMsg{ChannelID: channelID, Err: err}
		}

		drafts := parseDrafts(result)
		return DraftResultMsg{ChannelID: channelID, Drafts: drafts}
	}
}

// Analyze performs thread sentiment analysis and takeaway extraction.
func (e *Engine) Analyze(threadTS string, messages []string) tea.Cmd {
	return func() tea.Msg {
		if len(messages) == 0 {
			return AnalysisResultMsg{ThreadTS: threadTS, Err: fmt.Errorf("no messages")}
		}

		context := strings.Join(messages, "\n")
		prompt := fmt.Sprintf(`Analyze this thread conversation. Respond in this exact format:

SENTIMENT: <POSITIVE|NEGATIVE|NEUTRAL|MIXED>
SCORE: <0.0-1.0>
TAKEAWAYS:
• <takeaway 1>
• <takeaway 2>
• <takeaway 3>

Thread:
%s`, context)

		result, err := e.complete(prompt)
		if err != nil {
			return AnalysisResultMsg{ThreadTS: threadTS, Err: err}
		}

		analysis := parseAnalysis(result, threadTS)
		return analysis
	}
}

func (e *Engine) complete(prompt string) (string, error) {
	resp, err := e.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     e.model,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("AI completion failed: %w", err)
	}

	var result strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}
	return result.String(), nil
}

func parsePoints(text string) []string {
	var points []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip leading bullet markers
		line = strings.TrimPrefix(line, "• ")
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		if line != "" {
			points = append(points, line)
		}
	}
	return points
}

func parseDrafts(text string) []Draft {
	var drafts []Draft
	blocks := strings.Split(text, "---")

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		draft := Draft{Confidence: 80}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "TONE:") {
				draft.Tone = strings.TrimSpace(strings.TrimPrefix(line, "TONE:"))
			} else if strings.HasPrefix(line, "CONFIDENCE:") {
				fmt.Sscanf(strings.TrimPrefix(line, "CONFIDENCE:"), "%d", &draft.Confidence)
			} else if strings.HasPrefix(line, "TEXT:") {
				draft.Text = strings.TrimSpace(strings.TrimPrefix(line, "TEXT:"))
			}
		}

		if draft.Text != "" {
			drafts = append(drafts, draft)
		}
	}

	return drafts
}

func parseAnalysis(text, threadTS string) AnalysisResultMsg {
	result := AnalysisResultMsg{ThreadTS: threadTS}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SENTIMENT:") {
			result.Sentiment = strings.TrimSpace(strings.TrimPrefix(line, "SENTIMENT:"))
		} else if strings.HasPrefix(line, "SCORE:") {
			fmt.Sscanf(strings.TrimPrefix(line, "SCORE:"), "%f", &result.Score)
		} else if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") {
			takeaway := strings.TrimPrefix(strings.TrimPrefix(line, "•"), "-")
			takeaway = strings.TrimSpace(takeaway)
			if takeaway != "" {
				result.Takeaways = append(result.Takeaways, takeaway)
			}
		}
	}

	return result
}
