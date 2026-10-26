package worktree

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFix02LogBoundKeepsValidUTF8(t *testing.T) {
	asciiExactly := strings.Repeat("a", 4096)
	if got, truncated := boundLogText(asciiExactly); got != asciiExactly || truncated {
		t.Fatalf("4096 ASCII bytes = %d truncated=%v, want the same text without truncation", len(got), truncated)
	}
	asciiOver := strings.Repeat("a", 4097)
	if got, truncated := boundLogText(asciiOver); got != strings.Repeat("a", 4096) || !truncated {
		t.Fatalf("4097 ASCII bytes = %d truncated=%v, want 4096 bytes and truncation", len(got), truncated)
	}
	multibyteExactly := strings.Repeat("\u20ac", 1365) + "A"
	if got, truncated := boundLogText(multibyteExactly); got != multibyteExactly || truncated || !utf8.ValidString(got) {
		t.Fatalf("4096-byte multibyte boundary = %d truncated=%v valid=%v, want unchanged valid text", len(got), truncated, utf8.ValidString(got))
	}
	multibyteOver := strings.Repeat("\u20ac", 1365) + "AB"
	if got, truncated := boundLogText(multibyteOver); got != strings.Repeat("\u20ac", 1365)+"A" || !truncated || !utf8.ValidString(got) {
		t.Fatalf("4097-byte multibyte boundary = %d truncated=%v valid=%v, want 4096 valid bytes", len(got), truncated, utf8.ValidString(got))
	}
	multibyteCut := strings.Repeat("\u20ac", 1366)
	if got, truncated := boundLogText(multibyteCut); got != strings.Repeat("\u20ac", 1365) || len(got) != 4095 || !truncated || !utf8.ValidString(got) {
		t.Fatalf("4098-byte multibyte input = %d truncated=%v valid=%v, want 4095 valid bytes", len(got), truncated, utf8.ValidString(got))
	}
}
