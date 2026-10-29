// Package output renders command results as text or JSON envelopes.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// The outcomes of a command.
const (
	OutcomeOK      = "ok"
	OutcomePartial = "partial"
	OutcomeFailed  = "failed"
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

// Failed returns a failed envelope with no effects and no result.
func Failed(phase, message string) Envelope {
	return Envelope{Outcome: OutcomeFailed, Phase: phase, Effects: []Effect{}, Error: message, Result: nil}
}

// Render writes one JSON object or the command's text and error lines. JSON
// prints nil effects as an empty array.
func Render(stdout, stderr io.Writer, jsonOutput bool, envelope Envelope, text string) error {
	if jsonOutput {
		if envelope.Effects == nil {
			envelope.Effects = []Effect{}
		}
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
	if outcome == OutcomeOK {
		return 0
	}
	return 1
}
