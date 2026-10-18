package scope

import (
	"errors"
	"io"
	"time"
)

// ErrInputTooLarge reports input above the byte limit.
var ErrInputTooLarge = errors.New("input too large")

// ErrInputTimeout reports input that did not end before the deadline.
var ErrInputTimeout = errors.New("input timeout")

// ReadInput reads input up to limit bytes before deadline.
func ReadInput(reader io.Reader, limit int64, deadline time.Duration) ([]byte, error) {
	return nil, nil
}
