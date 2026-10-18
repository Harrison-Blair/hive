package herdr

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPingChecksProtocol(t *testing.T) {
	server := startServer(t, okResponder)
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	protocol, err := Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if protocol != Protocol {
		t.Fatalf("Ping() protocol = %d, want %d", protocol, Protocol)
	}
}

func TestPingRejectsOtherProtocol(t *testing.T) {
	server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
		return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":21}`), 0, false
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	protocol, err := Ping(context.Background())
	if err == nil || err.Error() != "herdr protocol 21, hive needs 22" {
		t.Fatalf("Ping() error = %v, want the protocol message", err)
	}
	if protocol != 21 {
		t.Fatalf("Ping() protocol = %d, want 21", protocol)
	}
}

func TestProtocolGateRunsPingOnce(t *testing.T) {
	server := startServer(t, okResponder)
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	for range 2 {
		if err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{}); err != nil {
			t.Fatalf("Call() error = %v", err)
		}
	}
	got := server.Methods()
	want := []string{"ping", "workspace.list", "workspace.list"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("methods = %v, want %v", got, want)
	}
}

func TestProtocolGateFailureBlocksLaterCalls(t *testing.T) {
	server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
		if request.Method == "ping" {
			return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":21}`), 0, false
		}
		return resultReply(request, `{"type":"ok"}`), 0, false
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	var first error
	for range 2 {
		err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
		if err == nil || !strings.Contains(err.Error(), "herdr protocol 21") {
			t.Fatalf("Call() error = %v, want the protocol error", err)
		}
		if first == nil {
			first = err
		} else if err.Error() != first.Error() {
			t.Fatalf("later error = %v, want the same error %v", err, first)
		}
	}
	got := server.Methods()
	if len(got) != 1 || got[0] != "ping" {
		t.Fatalf("methods = %v, want only ping", got)
	}
}

func TestProtocolGateIsPerSocketPath(t *testing.T) {
	first := startServer(t, okResponder)
	second := startServer(t, okResponder)
	for _, path := range []string{first.path, second.path} {
		t.Setenv("HERDR_SOCKET_PATH", path)
		if err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{}); err != nil {
			t.Fatalf("Call() error = %v", err)
		}
	}
	for index, server := range []*testServer{first, second} {
		if got := server.Methods(); len(got) != 2 || got[0] != "ping" {
			t.Fatalf("server %d methods = %v, want ping then workspace.list", index, got)
		}
	}
}

func TestCallWithMethodPingDoesNotGate(t *testing.T) {
	server := startServer(t, okResponder)
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	var result struct {
		Type string `json:"type"`
	}
	if err := Call(context.Background(), "ping", map[string]any{}, &result); err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if result.Type != "pong" {
		t.Fatalf("result type = %q, want pong", result.Type)
	}
	if got := server.Methods(); len(got) != 1 || got[0] != "ping" {
		t.Fatalf("methods = %v, want one ping", got)
	}
}
