package agent

import (
	"errors"
	"log/slog"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// CallerName returns the name of the record whose pane is pane, or the empty
// string when pane is empty or no record matches.
func CallerName(home, pane string) (string, error) {
	if pane == "" {
		return "", nil
	}
	record, err := FindByPane(home, pane)
	if errors.Is(err, ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return record.Name, nil
}

// IsCode reports whether err is a herdr server error with code.
func IsCode(err error, code string) bool {
	var serverError *herdr.Error
	return errors.As(err, &serverError) && serverError.Code == code
}

// Failed returns a failed envelope with no effects.
func Failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}

// Logger returns logger, or a logger that drops each event when logger is nil.
func Logger(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return logger
}

// SessionID returns the native session ID of the harness, or the empty string
// when herdr reports no agent session.
func SessionID(session *herdr.AgentSessionInfo) string {
	if session == nil {
		return ""
	}
	return session.Value
}
