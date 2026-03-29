package markdown

import (
	"strings"
	"testing"
)

func TestRenderer_Bold(t *testing.T) {
	r := NewRenderer()
	result := r.Render("this is *bold text* here")
	// The rendered output should contain "bold text" but not the asterisks
	if strings.Contains(result, "*bold text*") {
		t.Error("bold markers should be stripped")
	}
	if !strings.Contains(result, "bold text") {
		t.Error("bold text content should be present")
	}
}

func TestRenderer_Italic(t *testing.T) {
	r := NewRenderer()
	result := r.Render("this is _italic text_ here")
	if strings.Contains(result, "_italic text_") {
		t.Error("italic markers should be stripped")
	}
	if !strings.Contains(result, "italic text") {
		t.Error("italic text content should be present")
	}
}

func TestRenderer_InlineCode(t *testing.T) {
	r := NewRenderer()
	result := r.Render("run `go build` now")
	if strings.Contains(result, "`go build`") {
		t.Error("code backticks should be stripped")
	}
	if !strings.Contains(result, "go build") {
		t.Error("code content should be present")
	}
}

func TestRenderer_Strikethrough(t *testing.T) {
	r := NewRenderer()
	result := r.Render("this is ~deleted~ text")
	// Should not have the tildes as raw markers
	if !strings.Contains(result, "deleted") {
		t.Error("strikethrough content should be present")
	}
}

func TestRenderer_LinkWithLabel(t *testing.T) {
	r := NewRenderer()
	result := r.Render("check <https://example.com|this link> out")
	if strings.Contains(result, "<https://") {
		t.Error("link angle brackets should be stripped")
	}
	if !strings.Contains(result, "this link") {
		t.Error("link label should be present")
	}
}

func TestRenderer_LinkWithoutLabel(t *testing.T) {
	r := NewRenderer()
	result := r.Render("visit <https://example.com> please")
	if strings.Contains(result, "<https://") {
		t.Error("link angle brackets should be stripped")
	}
	if !strings.Contains(result, "https://example.com") {
		t.Error("bare URL should be present")
	}
}

func TestRenderer_UserMention(t *testing.T) {
	r := NewRenderer()
	r.UserResolver = func(id string) string {
		if id == "U12345" {
			return "alice"
		}
		return ""
	}
	result := r.Render("hey <@U12345> check this")
	if !strings.Contains(result, "@alice") {
		t.Errorf("expected @alice in result, got: %s", result)
	}
}

func TestRenderer_UserMentionWithDisplayName(t *testing.T) {
	r := NewRenderer()
	result := r.Render("hey <@U12345|bob> check this")
	if !strings.Contains(result, "@bob") {
		t.Errorf("expected @bob in result, got: %s", result)
	}
}

func TestRenderer_ChannelMention(t *testing.T) {
	r := NewRenderer()
	result := r.Render("post in <#C98765|general> please")
	if !strings.Contains(result, "#general") {
		t.Errorf("expected #general in result, got: %s", result)
	}
}

func TestRenderer_EmojiKnown(t *testing.T) {
	r := NewRenderer()
	result := r.Render("nice work :tada: :rocket:")
	if strings.Contains(result, ":tada:") {
		t.Error(":tada: should be converted to unicode")
	}
	if !strings.Contains(result, "\U0001f389") { // 🎉
		t.Errorf("expected party popper emoji, got: %s", result)
	}
}

func TestRenderer_EmojiUnknown(t *testing.T) {
	r := NewRenderer()
	result := r.Render("custom :partyparrot: here")
	// Unknown emoji should be left as-is
	if !strings.Contains(result, ":partyparrot:") {
		t.Errorf("unknown emoji should remain as shortcode, got: %s", result)
	}
}

func TestRenderer_CodeBlock(t *testing.T) {
	r := NewRenderer()
	result := r.Render("here:\n```\nfunc main() {}\n```\ndone")
	if !strings.Contains(result, "func main()") {
		t.Error("code block content should be present")
	}
}

func TestRenderer_PlainText(t *testing.T) {
	r := NewRenderer()
	input := "just a normal message with no formatting"
	result := r.Render(input)
	if result != input {
		t.Errorf("plain text should pass through unchanged, got: %s", result)
	}
}

func TestRenderer_MixedFormatting(t *testing.T) {
	r := NewRenderer()
	result := r.Render("*bold* and _italic_ and `code`")
	if !strings.Contains(result, "bold") {
		t.Error("bold should be present")
	}
	if !strings.Contains(result, "italic") {
		t.Error("italic should be present")
	}
	if !strings.Contains(result, "code") {
		t.Error("code should be present")
	}
}
