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

type inputResult struct {
	data []byte
	err  error
}

// ReadInput reads input up to limit bytes before deadline.
func ReadInput(reader io.Reader, limit int64, deadline time.Duration) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("input: nil reader")
	}
	if limit < 0 {
		return nil, ErrInputTooLarge
	}
	resultCh := make(chan inputResult, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(reader, limit+1))
		resultCh <- inputResult{data: data, err: err}
	}()
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	select {
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}
		if int64(len(result.data)) > limit {
			return nil, ErrInputTooLarge
		}
		return result.data, nil
	case <-timer.C:
		return nil, ErrInputTimeout
	}
}
