package logs

import (
	"strings"
	"testing"
)

func TestTextLineEscapesControlsInFieldsAndKeys(t *testing.T) {
	event := Event{Fields: map[string]any{
		"time":               "2026-10-18T10:00:00-04:00",
		"level":              "info",
		"component":          "cli\ninjected",
		"msg":                "first\nsecond",
		"key\nwith\nnewline": "value\twith\x01control",
		"other":              "carriage\rreturn",
	}}
	line := TextLine(event)

	if strings.ContainsAny(line, "\n\r\t") {
		t.Fatalf("rendered line has a physical control: %q", line)
	}
	if strings.Count(line, "\n") != 0 || strings.Count(line, "\r") != 0 {
		t.Fatalf("rendered line splits a physical line: %q", line)
	}
	for _, want := range []string{
		`cli\ninjected`,
		`first\nsecond`,
		`key\nwith\nnewline=`,
		`\t`,
		`\u0001`,
		`carriage\rreturn`,
	} {
		if !strings.Contains(line, want) {
			t.Errorf("rendered line %q has no %q", line, want)
		}
	}
	keyIndex := strings.Index(line, `key\nwith\nnewline=`)
	otherIndex := strings.Index(line, `other=`)
	if keyIndex < 0 || otherIndex < 0 || keyIndex > otherIndex {
		t.Fatalf("rendered line %q changes key order", line)
	}
}
