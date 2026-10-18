package herdr

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// eventsServer answers the ping gate and one events.subscribe request.
type eventsServer struct {
	t        *testing.T
	listener net.Listener
	path     string
	frames   []string
}

func newEventsServer(t *testing.T, frames ...string) *eventsServer {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Setenv("HERDR_SOCKET_PATH", path)
	server := &eventsServer{t: t, listener: listener, path: path, frames: frames}
	go server.serve()
	t.Cleanup(func() {
		_ = listener.Close()
		_ = os.Remove(path)
	})
	return server
}

func (s *eventsServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *eventsServer) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}
	var request struct {
		ID     string `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(line, &request); err != nil {
		return
	}
	if request.Method == "ping" {
		_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"pong","version":"0.9.5","protocol":22}}`+"\n", request.ID)
		return
	}
	if request.Method != "events.subscribe" {
		_, _ = fmt.Fprintf(conn, `{"id":%q,"error":{"code":"unknown_method","message":"%s"}}`+"\n", request.ID, request.Method)
		return
	}
	_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"subscription_started"}}`+"\n", request.ID)
	for _, frame := range s.frames {
		_, _ = fmt.Fprint(conn, frame+"\n")
	}
	time.Sleep(50 * time.Millisecond)
}

func TestSubscribeReadsLiteralFrames(t *testing.T) {
	statusFrame := `{"data":{"pane_id":"w1:p2","workspace_id":"w1","agent_status":"working"},"event":"pane.agent_status_changed"}`
	plainFrame := `{"data":{"type":"tab_created","tab":{"tab_id":"w1:t1"}},"event":"tab_created"}`
	server := newEventsServer(t, statusFrame, plainFrame)
	_ = server

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stream, err := Subscribe(ctx, SubscribeParams{Subscriptions: []Subscription{
		{Type: "pane.agent_status_changed", PaneID: "w1:p2"},
		{Type: "tab.closed"},
	}})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = stream.Close() }()

	first := receiveEvent(t, stream)
	if first.Event != "pane.agent_status_changed" {
		t.Fatalf("first event = %q; want pane.agent_status_changed", first.Event)
	}
	var status struct {
		PaneID      string `json:"pane_id"`
		WorkspaceID string `json:"workspace_id"`
		AgentStatus string `json:"agent_status"`
	}
	if err := json.Unmarshal(first.Data, &status); err != nil {
		t.Fatal(err)
	}
	if status.PaneID != "w1:p2" || status.WorkspaceID != "w1" || status.AgentStatus != "working" {
		t.Fatalf("status frame = %+v; want the literal values", status)
	}
	if _, hasType := rawType(first.Data); hasType {
		t.Fatalf("dotted status frame must not carry a type field")
	}

	second := receiveEvent(t, stream)
	if second.Event != "tab_created" {
		t.Fatalf("second event = %q; want tab_created", second.Event)
	}
	if _, hasType := rawType(second.Data); !hasType {
		t.Fatalf("plain frame must carry data.type")
	}
}

func TestSubscribeEndsWhenConnectionCloses(t *testing.T) {
	server := newEventsServer(t)
	_ = server
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stream, err := Subscribe(ctx, SubscribeParams{Subscriptions: []Subscription{{Type: "tab.closed"}}})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = stream.Close() }()
	select {
	case _, ok := <-stream.Events():
		if ok {
			t.Fatalf("expected the events channel to close")
		}
	case err := <-stream.Errors():
		if err == nil {
			t.Fatalf("expected a close error")
		}
	case <-time.After(time.Second):
		t.Fatalf("the stream did not end")
	}
}

func receiveEvent(t *testing.T, stream *Stream) Event {
	t.Helper()
	select {
	case event, ok := <-stream.Events():
		if !ok {
			t.Fatalf("the events channel closed early")
		}
		return event
	case err := <-stream.Errors():
		t.Fatalf("stream error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatalf("no event arrived")
	}
	return Event{}
}

func rawType(data json.RawMessage) (string, bool) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return "", false
	}
	raw, ok := object["type"]
	if !ok {
		return "", false
	}
	var value string
	_ = json.Unmarshal(raw, &value)
	return value, true
}
