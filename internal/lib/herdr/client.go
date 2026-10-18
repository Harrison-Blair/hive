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
	"reflect"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

const (
	// DefaultDeadline is the time limit of one call when the caller gives none.
	DefaultDeadline = 10 * time.Second
	maxRequestLine  = 1 << 20
	maxReplyLine    = 1 << 20
	maxErrorText    = 4096
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
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// Call opens one connection, writes one request line, reads one reply line, and
// closes the connection.
func Call(ctx context.Context, method string, params, result any, opts ...Option) error {
	return call(ctx, method, params, result, nil, opts...)
}

// call is Call with a private result check that runs after a successful reply.
func call(ctx context.Context, method string, params, result any, check func() error, opts ...Option) error {
	if ctx == nil {
		ctx = context.Background()
	}
	options := newCallOptions(opts)
	callCtx, cancel := context.WithTimeout(ctx, options.deadline)
	defer cancel()
	requestID := nextRequestID()
	started := time.Now()
	fail := func(err error) error {
		logCallError(options.logger, requestID, method, started, err)
		return err
	}
	path, err := socketPath()
	if err != nil {
		return fail(err)
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fail(fmt.Errorf("herdr request: %s", plainError(err)))
	}
	if bytes.Equal(bytes.TrimSpace(paramsJSON), []byte("null")) {
		paramsJSON = []byte("{}")
	}
	line, err := json.Marshal(requestEnvelope{ID: requestID, Method: method, Params: json.RawMessage(paramsJSON)})
	if err != nil {
		return fail(fmt.Errorf("herdr request: %s", plainError(err)))
	}
	line = append(line, '\n')
	if len(line) > maxRequestLine {
		return fail(errRequestTooLarge)
	}
	if method != "ping" {
		if err := ensureProtocol(callCtx, options); err != nil {
			return fail(err)
		}
	}
	conn, err := dial(callCtx, path)
	if err != nil {
		return fail(err)
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

	logEvent(options.logger, "herdr request", requestID, method, started, line)
	if err := writeLine(conn, line); err != nil {
		return fail(callFailure(callCtx, fmt.Errorf("herdr %s: write request: %s", method, plainError(err))))
	}
	reply, err := readReply(bufio.NewReader(conn), maxReplyLine)
	if err != nil {
		return fail(callFailure(callCtx, err))
	}
	logEvent(options.logger, "herdr reply", requestID, method, started, reply)
	if err := settleReply(requestID, reply, result); err != nil {
		return fail(err)
	}
	if check != nil {
		if err := check(); err != nil {
			return fail(err)
		}
	}
	return nil
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
		ID     string          `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(reply, &envelope); err != nil {
		return fmt.Errorf("bad reply: %s", plainError(err))
	}
	if envelope.ID != requestID {
		return fmt.Errorf("bad reply: id %q, want %q", envelope.ID, requestID)
	}
	hasResult := envelope.Result != nil
	hasError := envelope.Error != nil
	switch {
	case hasResult && hasError:
		return errors.New("bad reply: has result and error")
	case !hasResult && !hasError:
		return errors.New("bad reply: has no result or error")
	}
	if hasError {
		if isJSONNull(envelope.Error) {
			return errors.New("bad reply: error is null")
		}
		var serverError Error
		if err := json.Unmarshal(envelope.Error, &serverError); err != nil {
			return fmt.Errorf("bad reply: %s", plainError(err))
		}
		return &serverError
	}
	if result == nil {
		return nil
	}
	return decodeResult(envelope.Result, result)
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func decodeResult(raw json.RawMessage, result any) error {
	value := reflect.ValueOf(result)
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Map {
		if value.IsNil() {
			return errors.New("bad reply: nil map result destination")
		}
		temporary := reflect.MakeMap(value.Type())
		if err := json.Unmarshal(raw, temporary.Interface()); err != nil {
			return fmt.Errorf("bad reply: %s", plainError(err))
		}
		value.Clear()
		iter := temporary.MapRange()
		for iter.Next() {
			value.SetMapIndex(iter.Key(), iter.Value())
		}
		return nil
	}
	if value.Kind() != reflect.Pointer {
		// The encoding package reports an invalid destination and changes nothing.
		if err := json.Unmarshal(raw, result); err != nil {
			return fmt.Errorf("bad reply: %s", plainError(err))
		}
		return nil
	}
	if value.IsNil() {
		return errors.New("bad reply: nil result destination")
	}
	temporary := reflect.New(value.Elem().Type())
	if isJSONNull(raw) {
		// A null decode keeps a plain struct unchanged and sets a map or slice to nil.
		temporary.Elem().Set(value.Elem())
	}
	if err := json.Unmarshal(raw, temporary.Interface()); err != nil {
		return fmt.Errorf("bad reply: %s", plainError(err))
	}
	value.Elem().Set(temporary.Elem())
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

func logCallError(logger *slog.Logger, requestID, method string, started time.Time, err error) {
	if logger == nil || err == nil {
		return
	}
	text := plainError(err)
	truncated := false
	if len(text) > maxErrorText {
		text = text[:maxErrorText]
		for len(text) > 0 && !utf8.ValidString(text) {
			text = text[:len(text)-1]
		}
		truncated = true
	}
	attributes := []any{
		"request_id", requestID,
		"method", method,
		"duration_ms", time.Since(started).Milliseconds(),
		"error", text,
	}
	if truncated {
		attributes = append(attributes, "truncated", true)
	}
	logger.Debug("herdr error", attributes...)
}

func plainError(err error) string {
	if err == nil {
		return ""
	}
	return strings.Map(func(value rune) rune {
		if value < 0x20 || value == 0x7f {
			return ' '
		}
		return value
	}, err.Error())
}
