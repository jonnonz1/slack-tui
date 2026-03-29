package markdown

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Render converts Slack mrkdwn to styled terminal text.
// Slack mrkdwn is NOT standard Markdown:
//   - *bold* (not **bold**)
//   - _italic_ (same)
//   - ~strikethrough~
//   - `code` and ```code blocks```
//   - <@U123> user mentions
//   - <#C123> channel mentions
//   - <url|label> links
type Renderer struct {
	boldStyle   lipgloss.Style
	italicStyle lipgloss.Style
	codeStyle   lipgloss.Style
	linkStyle   lipgloss.Style
	mentionStyle lipgloss.Style

	// UserResolver maps user IDs to display names.
	UserResolver func(id string) string
	// ChannelResolver maps channel IDs to names.
	ChannelResolver func(id string) string
}

func NewRenderer() *Renderer {
	return &Renderer{
		boldStyle:    lipgloss.NewStyle().Bold(true),
		italicStyle:  lipgloss.NewStyle().Italic(true),
		codeStyle:    lipgloss.NewStyle().Background(lipgloss.Color("#262a31")).Foreground(lipgloss.Color("#f6afef")),
		linkStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#5edda0")).Underline(true),
		mentionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#f6afef")).Bold(true),
	}
}

var (
	// Slack mrkdwn patterns
	boldRe          = regexp.MustCompile(`\*([^*]+)\*`)
	italicRe        = regexp.MustCompile(`_([^_]+)_`)
	strikeRe        = regexp.MustCompile(`~([^~]+)~`)
	codeRe          = regexp.MustCompile("`([^`]+)`")
	codeBlockRe     = regexp.MustCompile("```([\\s\\S]*?)```")
	userMentionRe   = regexp.MustCompile(`<@([A-Z0-9]+)(?:\|([^>]*))?> `)
	chanMentionRe   = regexp.MustCompile(`<#([A-Z0-9]+)(?:\|([^>]*))?> `)
	linkRe          = regexp.MustCompile(`<(https?://[^|>]+)(?:\|([^>]+))?>`)
	emojiRe         = regexp.MustCompile(`:([a-z0-9_+\-]+):`)
)

func (r *Renderer) Render(text string) string {
	// Code blocks first (preserve contents)
	text = codeBlockRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := codeBlockRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			return "\n" + r.codeStyle.Render(strings.TrimSpace(inner[1])) + "\n"
		}
		return match
	})

	// Inline code
	text = codeRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := codeRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			return r.codeStyle.Render(inner[1])
		}
		return match
	})

	// User mentions: <@U123|name> or <@U123>
	text = userMentionRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := userMentionRe.FindStringSubmatch(match)
		if len(inner) > 2 && inner[2] != "" {
			return r.mentionStyle.Render("@"+inner[2]) + " "
		}
		if len(inner) > 1 && r.UserResolver != nil {
			name := r.UserResolver(inner[1])
			if name != "" {
				return r.mentionStyle.Render("@"+name) + " "
			}
		}
		return match
	})

	// Channel mentions: <#C123|name>
	text = chanMentionRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := chanMentionRe.FindStringSubmatch(match)
		if len(inner) > 2 && inner[2] != "" {
			return r.mentionStyle.Render("#"+inner[2]) + " "
		}
		if len(inner) > 1 && r.ChannelResolver != nil {
			name := r.ChannelResolver(inner[1])
			if name != "" {
				return r.mentionStyle.Render("#"+name) + " "
			}
		}
		return match
	})

	// Links: <url|label> or <url>
	text = linkRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := linkRe.FindStringSubmatch(match)
		if len(inner) > 2 && inner[2] != "" {
			return r.linkStyle.Render(inner[2])
		}
		if len(inner) > 1 {
			return r.linkStyle.Render(inner[1])
		}
		return match
	})

	// Bold
	text = boldRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := boldRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			return r.boldStyle.Render(inner[1])
		}
		return match
	})

	// Italic
	text = italicRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := italicRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			return r.italicStyle.Render(inner[1])
		}
		return match
	})

	// Strikethrough
	text = strikeRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := strikeRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			return r.italicStyle.Render("~" + inner[1] + "~") // no native strikethrough in most terminals
		}
		return match
	})

	// Emoji shortcodes — convert common ones to unicode, leave rest as-is
	text = emojiRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := emojiRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			if emoji, ok := emojiMap[inner[1]]; ok {
				return emoji
			}
		}
		return match
	})

	return text
}

// Common emoji shortcode -> unicode mappings.
var emojiMap = map[string]string{
	"thumbsup":    "👍",
	"+1":          "👍",
	"thumbsdown":  "👎",
	"-1":          "👎",
	"heart":       "❤️",
	"tada":        "🎉",
	"rocket":      "🚀",
	"fire":        "🔥",
	"eyes":        "👀",
	"wave":        "👋",
	"check":       "✅",
	"white_check_mark": "✅",
	"x":           "❌",
	"warning":     "⚠️",
	"bulb":        "💡",
	"memo":        "📝",
	"link":        "🔗",
	"gear":        "⚙️",
	"bug":         "🐛",
	"hammer":      "🔨",
	"construction": "🚧",
	"rotating_light": "🚨",
	"thinking_face": "🤔",
	"raised_hands": "🙌",
	"pray":        "🙏",
	"clap":        "👏",
	"100":         "💯",
	"sparkles":    "✨",
	"zap":         "⚡",
	"lock":        "🔒",
	"unlock":      "🔓",
	"bell":        "🔔",
	"no_bell":     "🔕",
	"speech_balloon": "💬",
	"hourglass":   "⏳",
	"stopwatch":   "⏱️",
	"calendar":    "📅",
	"package":     "📦",
	"chart_with_upwards_trend": "📈",
	"chart_with_downwards_trend": "📉",
}
