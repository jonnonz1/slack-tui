package slack

import (
	"testing"
	"time"
)

func TestParseSlackTimestamp(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect time.Time
	}{
		{
			name:   "standard Slack timestamp",
			input:  "1234567890.123456",
			expect: time.Unix(1234567890, 123456000),
		},
		{
			name:   "empty string",
			input:  "",
			expect: time.Time{},
		},
		{
			name:   "seconds only",
			input:  "1700000000.000000",
			expect: time.Unix(1700000000, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSlackTimestamp(tt.input)
			if !got.Equal(tt.expect) {
				t.Errorf("parseSlackTimestamp(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}
