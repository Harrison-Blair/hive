package worktree

import (
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Now returns the current UTC time in RFC 3339 format with nanoseconds.
func Now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// FailedEnvelope returns a failed envelope with no effects.
func FailedEnvelope(phase, message string) output.Envelope {
	return FailedEnvelopeWithEffects(phase, message, nil)
}

// FailedEnvelopeWithEffects returns a failed envelope with its known effects.
func FailedEnvelopeWithEffects(phase, message string, effects []output.Effect) output.Envelope {
	if effects == nil {
		effects = []output.Effect{}
	}
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: effects, Error: OneLine(message), Result: nil}
}

// PartialEnvelope returns a partial envelope with its effects and result.
func PartialEnvelope(phase, message string, effects []output.Effect, result any) output.Envelope {
	if effects == nil {
		effects = []output.Effect{}
	}
	return output.Envelope{Outcome: "partial", Phase: phase, Effects: effects, Error: OneLine(message), Result: result}
}

// OneLine replaces each control character with a space.
func OneLine(value string) string {
	return strings.Map(func(character rune) rune {
		if character < 0x20 || character == 0x7f {
			return ' '
		}
		return character
	}, value)
}
