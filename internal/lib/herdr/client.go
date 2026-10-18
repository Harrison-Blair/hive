package herdr

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

const (
	// DefaultDeadline is the time limit of one call when the caller gives none.
	DefaultDeadline = 10 * time.Second
	maxRequestLine  = 1 << 20
	maxReplyLine    = 1 << 20
)

var errRequestTooLarge = errors.New("request too large")

var (
	idPrefix  = newRequestPrefix()
	idCounter atomic.Uint64
)

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

func newCallOptions(opts []Option) callOptions {
	options := callOptions{deadline: DefaultDeadline}
	for _, option := range opts {
		if option != nil {
			option(&options)
		}
	}
	if options.deadline <= 0 {
		options.deadline = DefaultDeadline
	}
	return options
}

func newRequestPrefix() string {
	value := make([]byte, 4)
	if _, err := rand.Read(value); err != nil {
		return fmt.Sprintf("%08x", uint32(time.Now().UnixNano())^uint32(os.Getpid())<<8)
	}
	return hex.EncodeToString(value)
}

func nextRequestID() string {
	return fmt.Sprintf("%s-%d", idPrefix, idCounter.Add(1))
}

type requestEnvelope struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

// Call opens one connection, writes one request line, reads one reply line, and
// closes the connection.
func Call(ctx context.Context, method string, params, result any, opts ...Option) error {
	if ctx == nil {
		ctx = context.Background()
	}
	options := newCallOptions(opts)
	callCtx, cancel := context.WithTimeout(ctx, options.deadline)
	defer cancel()
	path, err := socketPath()
	if err != nil {
		return err
	}
	requestID := nextRequestID()
	if params == nil {
		params = struct{}{}
	}
	line, err := json.Marshal(requestEnvelope{ID: requestID, Method: method, Params: params})
	if err != nil {
		return fmt.Errorf("herdr request: %s", err)
	}
	line = append(line, '\n')
	if len(line) > maxRequestLine {
		return errRequestTooLarge
	}
	if method != "ping" {
		if err := ensureProtocol(callCtx, options); err != nil {
			return err
		}
	}
	conn, err := dial(callCtx, path)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if deadline, ok := callCtx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	stopWatch := make(chan struct{})
	defer close(stopWatch)
	go func() {
		select {
		case <-callCtx.Done():
			_ = conn.Close()
		case <-stopWatch:
		}
	}()

	started := time.Now()
	logEvent(options.logger, "herdr request", requestID, method, started, line)
	if err := writeLine(conn, line); err != nil {
		return callFailure(callCtx, fmt.Errorf("herdr %s: write request: %s", method, plainError(err)))
	}
	reply, err := readReply(bufio.NewReader(conn), maxReplyLine)
	if err != nil {
		return callFailure(callCtx, err)
	}
	logEvent(options.logger, "herdr reply", requestID, method, started, reply)
	return settleReply(requestID, reply, result)
}

func socketPath() (string, error) {
	path := os.Getenv("HERDR_SOCKET_PATH")
	if path == "" {
		return "", errors.New("herdr socket: HERDR_SOCKET_PATH is not set")
	}
	if strings.ContainsFunc(path, isControlRune) {
		return "", errors.New("herdr socket: invalid socket path")
	}
	return path, nil
}

func isControlRune(value rune) bool {
	return value < 0x20 || value == 0x7f
}

func dial(ctx context.Context, path string) (net.Conn, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("herdr socket: %s: no such file", path)
		}
		return nil, fmt.Errorf("herdr socket: %s: %s", path, plainError(err))
	}
	if info.Mode()&os.ModeSymlink != 0 || info.Mode()&os.ModeSocket == 0 {
		return nil, fmt.Errorf("herdr socket: %s: not a socket", path)
	}
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unix", path)
	if err != nil {
		return nil, callFailure(ctx, fmt.Errorf("herdr socket: %s: %s", path, socketDialMessage(path, err)))
	}
	return conn, nil
}

func socketDialMessage(path string, err error) string {
	message := plainError(err)
	message = strings.TrimPrefix(message, "dial unix "+path+": ")
	return strings.TrimPrefix(message, "connect: ")
}

func writeLine(conn net.Conn, line []byte) error {
	for len(line) > 0 {
		written, err := conn.Write(line)
		if err != nil {
			return err
		}
		line = line[written:]
	}
	return nil
}

func readReply(reader *bufio.Reader, limit int) ([]byte, error) {
	var line []byte
	for {
		chunk, err := reader.ReadSlice('\n')
		line = append(line, chunk...)
		if len(line) > limit {
			return nil, errors.New("reply too large")
		}
		switch {
		case err == nil:
			return line, nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		case errors.Is(err, io.EOF):
			if len(line) == 0 {
				return nil, errors.New("bad reply: empty reply")
			}
			return nil, errors.New("bad reply: incomplete reply")
		default:
			return nil, fmt.Errorf("bad reply: %s", plainError(err))
		}
	}
}

func settleReply(requestID string, reply []byte, result any) error {
	var envelope struct {
		ID     string           `json:"id"`
		Result *json.RawMessage `json:"result"`
		Error  *Error           `json:"error"`
	}
	if err := json.Unmarshal(reply, &envelope); err != nil {
		return fmt.Errorf("bad reply: %s", err)
	}
	if envelope.ID != requestID {
		return fmt.Errorf("bad reply: id %q, want %q", envelope.ID, requestID)
	}
	switch {
	case envelope.Result != nil && envelope.Error != nil:
		return errors.New("bad reply: has result and error")
	case envelope.Result == nil && envelope.Error == nil:
		return errors.New("bad reply: has no result or error")
	}
	if envelope.Error != nil {
		return envelope.Error
	}
	if result != nil {
		if err := json.Unmarshal(*envelope.Result, result); err != nil {
			return fmt.Errorf("bad reply: %s", err)
		}
	}
	return nil
}

func callFailure(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
		return context.DeadlineExceeded
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return err
}

func logEvent(logger *slog.Logger, message, requestID, method string, started time.Time, body []byte) {
	if logger == nil {
		return
	}
	logger.Debug(message,
		"request_id", requestID,
		"method", method,
		"duration_ms", time.Since(started).Milliseconds(),
		hivelog.Body(string(bytes.TrimSuffix(body, []byte("\n")))),
	)
}

func plainError(err error) string {
	if err == nil {
		return ""
	}
	return strings.ReplaceAll(err.Error(), "\n", " ")
}
