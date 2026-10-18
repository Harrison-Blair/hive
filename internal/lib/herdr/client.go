package herdr

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// DefaultDeadline is the time limit of one call when the caller gives none.
const DefaultDeadline = 10 * time.Second

type callOptions struct {
	deadline time.Duration
	logger   *slog.Logger
}

// Option changes the behavior of one call.
type Option func(*callOptions)

// WithDeadline sets the deadline of one call.
func WithDeadline(deadline time.Duration) Option {
	return func(options *callOptions) {
		options.deadline = deadline
	}
}

// WithLogger sets the logger of one call.
func WithLogger(logger *slog.Logger) Option {
	return func(options *callOptions) {
		options.logger = logger
	}
}

// Call sends one request and reads one reply.
func Call(ctx context.Context, method string, params, result any, opts ...Option) error {
	return nil
}

var idCounter atomic.Uint64

// nextRequestID returns the next request ID of this process.
func nextRequestID() string {
	return ""
}
