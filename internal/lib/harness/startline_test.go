package harness

import "testing"

func TestStartLineLengthQuotesEachArgument(t *testing.T) {
	// claude '--a' 'it'\''s'
	if got, want := StartLineLength("claude", []string{"--a", "it's"}), 22; got != want {
		t.Fatalf("StartLineLength() = %d; want %d", got, want)
	}
	if got, want := StartLineLength("pi", nil), 2; got != want {
		t.Fatalf("StartLineLength(no args) = %d; want %d", got, want)
	}
	if MaxStartLine != 4000 {
		t.Fatalf("MaxStartLine = %d; want 4000", MaxStartLine)
	}
}
