package style

import "testing"

func TestColorWrapsTextInOneSGRRun(t *testing.T) {
	if got := Color("38;2;1;2;3", "hi"); got != "\x1b[38;2;1;2;3mhi\x1b[0m" {
		t.Fatalf("Color = %q", got)
	}
}

func TestColorWithNoStyleKeepsTheText(t *testing.T) {
	if got := Color("", "pi"); got != "pi" {
		t.Fatalf("Color(\"\", pi) = %q; want the plain text", got)
	}
}

func TestHarnessColours(t *testing.T) {
	for name, want := range map[string]string{
		"claude": "38;2;217;119;87",
		"codex":  "38;2;166;227;201",
		"pi":     "",
		"other":  "",
	} {
		if got := Harness(name); got != want {
			t.Errorf("Harness(%q) = %q; want %q", name, got, want)
		}
	}
}
