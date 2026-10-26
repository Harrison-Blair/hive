package harness

import "testing"

// TestSystemPromptDelivery gives the delivery of each harness. Claude Code
// reads the profile and the guide from the file that
// --append-system-prompt-file names. Herdr start arguments cannot carry
// multiline text. Each other
// harness takes them at the start of the first prompt (D4).
func TestSystemPromptDelivery(t *testing.T) {
	cases := map[string]string{
		"claude":   "--append-system-prompt-file",
		"pi":       "",
		"codex":    "",
		"opencode": "",
		"cursor":   "",
	}
	for kind, want := range cases {
		t.Run(kind, func(t *testing.T) {
			profile, err := Lookup(kind)
			if err != nil {
				t.Fatalf("Lookup(%s) error = %v", kind, err)
			}
			if got := profile.SystemPromptArg(); got != want {
				t.Fatalf("SystemPromptArg() = %q; want %q", got, want)
			}
		})
	}
}
