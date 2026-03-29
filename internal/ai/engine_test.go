package ai

import (
	"testing"
)

func TestParsePoints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "bullet points",
			input:    "• Point one\n• Point two\n• Point three",
			expected: []string{"Point one", "Point two", "Point three"},
		},
		{
			name:     "dash points",
			input:    "- First\n- Second",
			expected: []string{"First", "Second"},
		},
		{
			name:     "asterisk points",
			input:    "* Alpha\n* Beta",
			expected: []string{"Alpha", "Beta"},
		},
		{
			name:     "mixed with empty lines",
			input:    "• One\n\n• Two\n\n• Three",
			expected: []string{"One", "Two", "Three"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   \n  \n   ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePoints(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("parsePoints() returned %d points, want %d", len(got), len(tt.expected))
				return
			}
			for i, p := range got {
				if p != tt.expected[i] {
					t.Errorf("point[%d] = %q, want %q", i, p, tt.expected[i])
				}
			}
		})
	}
}

func TestParseDrafts(t *testing.T) {
	input := `TONE: ASSERTIVE
CONFIDENCE: 95
TEXT: I'll handle this now.
---
TONE: CLARIFICATION
CONFIDENCE: 82
TEXT: Should I check the logs first?
---
TONE: STATUS_UPDATE
CONFIDENCE: 90
TEXT: Starting the update now.`

	drafts := parseDrafts(input)

	if len(drafts) != 3 {
		t.Fatalf("expected 3 drafts, got %d", len(drafts))
	}

	tests := []struct {
		tone       string
		confidence int
		textPrefix string
	}{
		{"ASSERTIVE", 95, "I'll handle"},
		{"CLARIFICATION", 82, "Should I check"},
		{"STATUS_UPDATE", 90, "Starting the"},
	}

	for i, tt := range tests {
		if drafts[i].Tone != tt.tone {
			t.Errorf("draft[%d].Tone = %q, want %q", i, drafts[i].Tone, tt.tone)
		}
		if drafts[i].Confidence != tt.confidence {
			t.Errorf("draft[%d].Confidence = %d, want %d", i, drafts[i].Confidence, tt.confidence)
		}
		if len(drafts[i].Text) == 0 {
			t.Errorf("draft[%d].Text is empty", i)
		}
	}
}

func TestParseDrafts_Empty(t *testing.T) {
	drafts := parseDrafts("")
	if len(drafts) != 0 {
		t.Errorf("expected 0 drafts from empty input, got %d", len(drafts))
	}
}

func TestParseDrafts_Partial(t *testing.T) {
	// Missing TEXT field should be skipped
	input := `TONE: ASSERTIVE
CONFIDENCE: 95
---
TONE: CLARIFICATION
CONFIDENCE: 82
TEXT: Valid draft here.`

	drafts := parseDrafts(input)
	if len(drafts) != 1 {
		t.Fatalf("expected 1 valid draft, got %d", len(drafts))
	}
	if drafts[0].Tone != "CLARIFICATION" {
		t.Errorf("expected CLARIFICATION, got %s", drafts[0].Tone)
	}
}

func TestParseAnalysis(t *testing.T) {
	input := `SENTIMENT: POSITIVE
SCORE: 0.85
TAKEAWAYS:
• Team agreed on the approach
• Deadline is Friday
• Need to update docs`

	result := parseAnalysis(input, "12345.6789")

	if result.ThreadTS != "12345.6789" {
		t.Errorf("ThreadTS = %q, want %q", result.ThreadTS, "12345.6789")
	}
	if result.Sentiment != "POSITIVE" {
		t.Errorf("Sentiment = %q, want POSITIVE", result.Sentiment)
	}
	if result.Score != 0.85 {
		t.Errorf("Score = %f, want 0.85", result.Score)
	}
	if len(result.Takeaways) != 3 {
		t.Fatalf("expected 3 takeaways, got %d", len(result.Takeaways))
	}
	if result.Takeaways[0] != "Team agreed on the approach" {
		t.Errorf("takeaway[0] = %q", result.Takeaways[0])
	}
}

func TestParseAnalysis_DashFormat(t *testing.T) {
	input := `SENTIMENT: MIXED
SCORE: 0.5
TAKEAWAYS:
- Some agreed
- Others disagreed`

	result := parseAnalysis(input, "ts")
	if len(result.Takeaways) != 2 {
		t.Errorf("expected 2 takeaways with dash format, got %d", len(result.Takeaways))
	}
}

func TestParseAnalysis_Empty(t *testing.T) {
	result := parseAnalysis("", "ts")
	if result.Sentiment != "" {
		t.Errorf("expected empty sentiment, got %q", result.Sentiment)
	}
	if len(result.Takeaways) != 0 {
		t.Errorf("expected 0 takeaways, got %d", len(result.Takeaways))
	}
}
