// Package output renders command results as text or JSON envelopes.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Effect describes one resource that a command created or found.
type Effect struct {
	Resource string `json:"resource"`
	Path     string `json:"path"`
	Action   string `json:"action"`
}

// Envelope is the result contract for a command.
type Envelope struct {
	Outcome string   `json:"outcome"`
	Phase   string   `json:"phase"`
	Effects []Effect `json:"effects"`
	Error   any      `json:"error"`
	Result  any      `json:"result"`
}

// Render writes one JSON object or the command's text and error lines.
func Render(stdout, stderr io.Writer, jsonOutput bool, envelope Envelope, text string) error {
	if jsonOutput {
		data, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, string(data))
		return err
	}
	if text != "" {
		if _, err := io.WriteString(stdout, text); err != nil {
			return err
		}
	}
	if envelope.Error != nil {
		_, err := fmt.Fprintf(stderr, "hive: %v\n", envelope.Error)
		return err
	}
	return nil
}

// Status returns the process status for an outcome.
func Status(outcome string) int {
	if outcome == "ok" {
		return 0
	}
	return 1
}
