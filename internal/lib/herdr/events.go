package herdr

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// Event is one pushed line of an events.subscribe stream.
type Event struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// Subscription is one entry of an events.subscribe request.
type Subscription struct {
	Type        string `json:"type"`
	PaneID      string `json:"pane_id,omitempty"`
	AgentStatus string `json:"agent_status,omitempty"`
}

// SubscribeParams is one events.subscribe request.
type SubscribeParams struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// Stream is one open events.subscribe connection.
type Stream struct {
	conn   net.Conn
	events chan Event
	errs   chan error
	done   chan struct{}
	once   sync.Once
}

// Events returns the channel of pushed events. The channel closes when the
// stream ends.
func (s *Stream) Events() <-chan Event {
	return s.events
}

// Errors returns the channel that carries the reason the stream ended.
func (s *Stream) Errors() <-chan error {
	return s.errs
}

// Close closes the connection. It is safe to call more than once.
func (s *Stream) Close() error {
	s.once.Do(func() {
		close(s.done)
		_ = s.conn.Close()
	})
	return nil
}

// Subscribe opens one events.subscribe connection, waits for the ack, and
// returns a stream of pushed events. The connection has no deadline. The
// caller closes the stream.
func Subscribe(ctx context.Context, params SubscribeParams, opts ...Option) (*Stream, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	options := newCallOptions(opts)
	path, err := socketPath()
	if err != nil {
		return nil, err
	}
	if err := ensureProtocol(ctx, options); err != nil {
		return nil, err
	}
	requestID := nextRequestID()
	started := time.Now()
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("herdr request: %s", plainError(err))
	}
	line, err := json.Marshal(requestEnvelope{ID: requestID, Method: "events.subscribe", Params: json.RawMessage(paramsJSON)})
	if err != nil {
		return nil, fmt.Errorf("herdr request: %s", plainError(err))
	}
	line = append(line, '\n')
	if len(line) > maxRequestLine {
		return nil, errRequestTooLarge
	}
	conn, err := dial(ctx, path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	if err := writeLine(conn, line); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("herdr events.subscribe: write request: %s", plainError(err))
	}
	reply, err := readReply(reader, maxReplyLine)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	logEvent(options.logger, "herdr reply", requestID, "events.subscribe", started, reply)
	if err := settleAck(reply, requestID); err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	stream := &Stream{
		conn:   conn,
		events: make(chan Event, 16),
		errs:   make(chan error, 1),
		done:   make(chan struct{}),
	}
	logSubscribe(options.logger, requestID, params)
	go stream.read(reader, options.logger)
	return stream, nil
}

func settleAck(reply []byte, requestID string) error {
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
	if envelope.Error != nil {
		return settleReply(requestID, reply, nil)
	}
	var ack struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(envelope.Result, &ack); err != nil {
		return fmt.Errorf("bad reply: %s", plainError(err))
	}
	if ack.Type != "subscription_started" {
		return fmt.Errorf("bad reply: result type %q, want subscription_started", ack.Type)
	}
	return nil
}

func (s *Stream) read(reader *bufio.Reader, logger *slog.Logger) {
	defer close(s.events)
	for {
		line, err := readReply(reader, maxReplyLine)
		if err != nil {
			s.fail(streamEndError(err))
			return
		}
		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			s.fail(fmt.Errorf("bad event: %s", plainError(err)))
			return
		}
		if event.Event == "" {
			s.fail(errors.New("bad event: missing event name"))
			return
		}
		if logger != nil {
			logger.Debug("herdr event", "method", event.Event, hivelog.Body(string(bytes.TrimSpace(line))))
		}
		select {
		case s.events <- event:
		case <-s.done:
			return
		}
	}
}

func (s *Stream) fail(err error) {
	select {
	case s.errs <- err:
	default:
	}
}

func streamEndError(err error) error {
	if errors.Is(err, io.EOF) {
		return errors.New("subscription closed")
	}
	var netErr net.Error
	if errors.As(err, &netErr) && errors.Is(err, os.ErrDeadlineExceeded) {
		return errors.New("subscription closed")
	}
	return fmt.Errorf("subscription: %s", plainError(err))
}

func logSubscribe(logger *slog.Logger, requestID string, params SubscribeParams) {
	if logger == nil {
		return
	}
	logger.Debug("herdr subscribe", "request_id", requestID, "method", "events.subscribe", "subscriptions", len(params.Subscriptions))
}
