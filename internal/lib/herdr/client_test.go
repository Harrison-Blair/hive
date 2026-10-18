package herdr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

type testRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type testServer struct {
	t       *testing.T
	path    string
	ln      net.Listener
	done    chan struct{}
	once    sync.Once
	wg      sync.WaitGroup
	mu      sync.Mutex
	methods []string
	ids     []string
	respond func(testRequest) (reply []byte, delay time.Duration, hold bool)
}

func startServer(t *testing.T, respond func(testRequest) ([]byte, time.Duration, bool)) *testServer {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &testServer{t: t, path: path, ln: ln, done: make(chan struct{}), respond: respond}
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			server.wg.Add(1)
			go server.serve(conn)
		}
	}()
	t.Cleanup(server.Close)
	return server
}

func (s *testServer) serve(conn net.Conn) {
	defer s.wg.Done()
	defer func() { _ = conn.Close() }()
	var request testRequest
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		return
	}
	s.mu.Lock()
	s.methods = append(s.methods, request.Method)
	s.ids = append(s.ids, request.ID)
	s.mu.Unlock()
	reply, delay, hold := s.respond(request)
	if delay > 0 {
		time.Sleep(delay)
	}
	if hold {
		<-s.done
		return
	}
	if reply == nil {
		return
	}
	_, _ = conn.Write(reply)
}

func (s *testServer) Close() {
	s.once.Do(func() {
		close(s.done)
		_ = s.ln.Close()
		s.wg.Wait()
	})
}

func (s *testServer) Methods() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.methods...)
}

func (s *testServer) IDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.ids...)
}

func resultReply(request testRequest, result string) []byte {
	return []byte(fmt.Sprintf(`{"id":%q,"result":%s}`+"\n", request.ID, result))
}

func okResponder(request testRequest) ([]byte, time.Duration, bool) {
	if request.Method == "ping" {
		return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
	}
	return resultReply(request, `{"type":"ok"}`), 0, false
}

func TestCallSuccessEnvelope(t *testing.T) {
	server := startServer(t, okResponder)
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	var result struct {
		Type string `json:"type"`
	}
	if err := Call(context.Background(), "workspace.list", map[string]any{}, &result); err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if result.Type != "ok" {
		t.Fatalf("result type = %q, want ok", result.Type)
	}
	if got := server.Methods(); len(got) != 2 || got[0] != "ping" || got[1] != "workspace.list" {
		t.Fatalf("methods = %v, want [ping workspace.list]", got)
	}
}

func TestCallErrorEnvelope(t *testing.T) {
	server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
		if request.Method == "ping" {
			return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
		}
		return []byte(fmt.Sprintf(`{"id":%q,"error":{"code":"workspace_not_found","message":"workspace w9 not found"}}`+"\n", request.ID)), 0, false
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	sentinel := "before"
	err := Call(context.Background(), "workspace.get", map[string]any{"workspace_id": "w9"}, &sentinel)
	var herdrErr *Error
	if !errors.As(err, &herdrErr) {
		t.Fatalf("Call() error = %v, want *herdr.Error", err)
	}
	if herdrErr.Code != CodeWorkspaceNotFound || herdrErr.Message != "workspace w9 not found" {
		t.Fatalf("error = %+v, want workspace_not_found", herdrErr)
	}
	if sentinel != "before" {
		t.Fatalf("result changed to %q on a failure", sentinel)
	}
}

func TestCallRejectsBadReplies(t *testing.T) {
	cases := []struct {
		name  string
		reply func(testRequest) []byte
	}{
		{
			name: "another id",
			reply: func(testRequest) []byte {
				return []byte(`{"id":"other","result":{"type":"ok"}}` + "\n")
			},
		},
		{
			name: "both result and error",
			reply: func(request testRequest) []byte {
				return []byte(fmt.Sprintf(`{"id":%q,"result":{"type":"ok"},"error":{"code":"x","message":"y"}}`+"\n", request.ID))
			},
		},
		{
			name: "neither result nor error",
			reply: func(request testRequest) []byte {
				return []byte(fmt.Sprintf(`{"id":%q}`+"\n", request.ID))
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
				if request.Method == "ping" {
					return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
				}
				return testCase.reply(request), 0, false
			})
			t.Setenv("HERDR_SOCKET_PATH", server.path)
			sentinel := map[string]any{"keep": true}
			err := Call(context.Background(), "workspace.list", map[string]any{}, &sentinel)
			if err == nil || !strings.Contains(err.Error(), "bad reply") {
				t.Fatalf("Call() error = %v, want bad reply", err)
			}
			if len(sentinel) != 1 || sentinel["keep"] != true {
				t.Fatalf("result changed to %v on a failure", sentinel)
			}
		})
	}
}

func TestCallReplyTooLarge(t *testing.T) {
	for _, size := range []int{1048575, 1048576, 1048577} {
		t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
			server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
				if request.Method == "ping" {
					return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
				}
				base := fmt.Sprintf(`{"id":%q,"result":{"type":"ok"}}`, request.ID)
				pad := size - len(base) - 1
				if pad < 0 {
					t.Fatalf("reply base is too large: %d", len(base))
				}
				return []byte(base + strings.Repeat(" ", pad) + "\n"), 0, false
			})
			t.Setenv("HERDR_SOCKET_PATH", server.path)
			err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
			switch size {
			case 1048577:
				if err == nil || !strings.Contains(err.Error(), "reply too large") {
					t.Fatalf("Call() error = %v, want reply too large", err)
				}
			default:
				if err != nil {
					t.Fatalf("Call() error = %v, want a reply of %d bytes", err, size)
				}
			}
		})
	}
}

func TestCallReplyTooLargeTenMegabytes(t *testing.T) {
	server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
		if request.Method == "ping" {
			return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
		}
		return []byte(fmt.Sprintf(`{"id":%q,"result":{"type":"ok","pad":%q}}`, request.ID, strings.Repeat("a", 10<<20))), 0, false
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	started := time.Now()
	err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "reply too large") {
		t.Fatalf("Call() error = %v, want reply too large", err)
	}
	if elapsed := time.Since(started); elapsed > 3*time.Second {
		t.Fatalf("oversized reply took %v, want a bounded failure", elapsed)
	}
}

func TestCallDeadlineWithSilentServer(t *testing.T) {
	server := startServer(t, func(testRequest) ([]byte, time.Duration, bool) {
		return nil, 0, true
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	started := time.Now()
	err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{}, WithDeadline(200*time.Millisecond))
	if err == nil {
		t.Fatal("Call() error = nil, want a deadline error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Call() error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("deadline call took %v, want a bounded failure", elapsed)
	}
}

func TestCallSlowReply(t *testing.T) {
	server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
		if request.Method == "ping" {
			return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
		}
		return resultReply(request, `{"type":"ok"}`), 1500 * time.Millisecond, false
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)

	err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{}, WithDeadline(time.Second))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Call() error = %v, want context.DeadlineExceeded", err)
	}

	var result struct {
		Type string `json:"type"`
	}
	if err := Call(context.Background(), "workspace.list", map[string]any{}, &result, WithDeadline(3*time.Second)); err != nil {
		t.Fatalf("Call() with a longer deadline error = %v", err)
	}
	if result.Type != "ok" {
		t.Fatalf("result type = %q, want ok", result.Type)
	}
}

func TestCallClosedReply(t *testing.T) {
	server := startServer(t, func(request testRequest) ([]byte, time.Duration, bool) {
		if request.Method == "ping" {
			return resultReply(request, `{"type":"pong","version":"0.9.5","protocol":22}`), 0, false
		}
		return nil, 0, false
	})
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "bad reply") {
		t.Fatalf("Call() error = %v, want bad reply", err)
	}
}

func TestCallWithoutSocketPath(t *testing.T) {
	t.Setenv("HERDR_SOCKET_PATH", "")
	err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
	if err == nil || err.Error() != "herdr socket: HERDR_SOCKET_PATH is not set" {
		t.Fatalf("Call() error = %v, want the unset socket message", err)
	}
}

func TestCallSocketPathTypes(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "none.sock")
		t.Setenv("HERDR_SOCKET_PATH", path)
		err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
		if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "no such file") {
			t.Fatalf("Call() error = %v, want the path and no such file", err)
		}
	})
	t.Run("directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "dir.sock")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("HERDR_SOCKET_PATH", path)
		err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
		if err == nil || !strings.Contains(err.Error(), "not a socket") {
			t.Fatalf("Call() error = %v, want not a socket", err)
		}
	})
	t.Run("control character", func(t *testing.T) {
		t.Setenv("HERDR_SOCKET_PATH", "/tmp/bad\npath.sock")
		err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{})
		if err == nil || strings.Contains(err.Error(), "\n") {
			t.Fatalf("Call() error = %q, want one plain line", err)
		}
	})
}

func requestLineSize(t *testing.T, method string, params any) int {
	t.Helper()
	payload := struct {
		ID     string `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params"`
	}{ID: "12345678-1", Method: method, Params: params}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return len(data) + 1
}

func nonASCIIText(size int) string {
	text := strings.Repeat("é", size/2)
	if size%2 != 0 {
		text += "a"
	}
	return text
}

func TestCallRequestSize(t *testing.T) {
	for _, size := range []int{1048575, 1048576, 1048577} {
		t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
			server := startServer(t, okResponder)
			t.Setenv("HERDR_SOCKET_PATH", server.path)
			idCounter.Store(0)
			overhead := requestLineSize(t, "workspace.list", map[string]any{"pad": ""})
			params := map[string]any{"pad": nonASCIIText(size - overhead)}
			if got := requestLineSize(t, "workspace.list", params); got != size {
				t.Fatalf("test request line = %d bytes, want %d", got, size)
			}
			err := Call(context.Background(), "workspace.list", params, &struct{}{})
			if size == 1048577 {
				if err == nil || !strings.Contains(err.Error(), "request too large") {
					t.Fatalf("Call() error = %v, want request too large", err)
				}
				if methods := server.Methods(); len(methods) != 0 {
					t.Fatalf("server received %v, want no request", methods)
				}
				return
			}
			if err != nil {
				t.Fatalf("Call() error = %v, want a request of %d bytes", err, size)
			}
		})
	}
}

func TestRequestIDsShareOnePrefix(t *testing.T) {
	server := startServer(t, okResponder)
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{}); err != nil {
				t.Errorf("Call() error = %v", err)
			}
		}()
	}
	wg.Wait()
	ids := server.IDs()
	var callIDs []string
	for index, method := range server.Methods() {
		if method == "workspace.list" {
			callIDs = append(callIDs, ids[index])
		}
	}
	if len(callIDs) != 2 {
		t.Fatalf("call IDs = %v, want two", callIDs)
	}
	prefixes := map[string]bool{}
	complete := map[string]bool{}
	for _, id := range callIDs {
		cut := strings.LastIndex(id, "-")
		if cut < 0 {
			t.Fatalf("request ID %q has no prefix", id)
		}
		prefixes[id[:cut]] = true
		complete[id] = true
	}
	if len(prefixes) != 1 {
		t.Fatalf("prefixes = %v, want one shared prefix", prefixes)
	}
	if len(complete) != 2 {
		t.Fatalf("complete IDs = %v, want two unique IDs", complete)
	}
}

func TestHelperProcessRequestID(t *testing.T) {
	if os.Getenv("HIVE_HELPER_REQUEST_ID") != "1" {
		return
	}
	fmt.Printf("ID=%s\n", nextRequestID())
}

func TestRequestIDsDifferBetweenProcesses(t *testing.T) {
	ids := make([]string, 0, 2)
	for range 2 {
		command := exec.Command(os.Args[0], "-test.run=TestHelperProcessRequestID")
		command.Env = append(os.Environ(), "HIVE_HELPER_REQUEST_ID=1")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("helper: %v: %s", err, output)
		}
		var id string
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "ID=") {
				id = strings.TrimPrefix(line, "ID=")
			}
		}
		if id == "" {
			t.Fatalf("helper output has no request ID: %s", output)
		}
		ids = append(ids, id)
	}
	prefix := func(id string) string {
		cut := strings.LastIndex(id, "-")
		if cut < 0 {
			t.Fatalf("request ID %q has no prefix", id)
		}
		return id[:cut]
	}
	if prefix(ids[0]) == prefix(ids[1]) {
		t.Fatalf("child prefixes = %q and %q, want distinct prefixes", prefix(ids[0]), prefix(ids[1]))
	}
}

func TestCallLogsRequestAndReply(t *testing.T) {
	server := startServer(t, okResponder)
	t.Setenv("HERDR_SOCKET_PATH", server.path)
	home := t.TempDir()
	writer, err := hivelog.NewWriter(home, "herdr", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	if err := Call(context.Background(), "workspace.list", map[string]any{}, &struct{}{}, WithLogger(logger)); err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "logs"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("log files = %d, want one", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(home, ".hive", "logs", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	events := map[string]map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Fatalf("log line %q: %v", line, err)
		}
		event = decoded
		message, _ := event["msg"].(string)
		events[message] = event
	}
	requestEvent, ok := events["herdr request"]
	if !ok {
		t.Fatalf("log events = %v, want a herdr request", events)
	}
	replyEvent, ok := events["herdr reply"]
	if !ok {
		t.Fatalf("log events = %v, want a herdr reply", events)
	}
	if requestEvent["request_id"] != replyEvent["request_id"] {
		t.Fatalf("request IDs differ: %v and %v", requestEvent["request_id"], replyEvent["request_id"])
	}
	if requestEvent["method"] != "workspace.list" || replyEvent["method"] != "workspace.list" {
		t.Fatalf("methods = %v and %v, want workspace.list", requestEvent["method"], replyEvent["method"])
	}
	if _, ok := replyEvent["duration_ms"]; !ok {
		t.Fatalf("reply event has no duration_ms: %v", replyEvent)
	}
	if body, _ := requestEvent["body"].(string); !strings.Contains(body, `"method":"workspace.list"`) {
		t.Fatalf("request body = %q, want the method", body)
	}
}
