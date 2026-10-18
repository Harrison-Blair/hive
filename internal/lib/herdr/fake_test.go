package herdr

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type recordedRequest struct {
	ID     string
	Method string
	Params string
}

type scriptedReply struct {
	result string
	error  string
}

type fakeServer struct {
	t        *testing.T
	listener net.Listener
	path     string
	mu       sync.Mutex
	replies  map[string][]scriptedReply
	requests []recordedRequest
	hook     func(method string, params json.RawMessage)
	closed   bool
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &fakeServer{t: t, listener: listener, path: path, replies: map[string][]scriptedReply{}}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go server.serve()
	t.Cleanup(server.close)
	return server
}

func (s *fakeServer) close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()
	_ = s.listener.Close()
	_ = os.Remove(s.path)
}

func (s *fakeServer) script(method string, replies ...scriptedReply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replies[method] = append(s.replies[method], replies...)
}

func (s *fakeServer) scriptResult(method, result string) {
	s.script(method, scriptedReply{result: result})
}

func (s *fakeServer) scriptError(method, code, message string) {
	s.script(method, scriptedReply{error: `{"code":` + quote(code) + `,"message":` + quote(message) + `}`})
}

func (s *fakeServer) setHook(hook func(method string, params json.RawMessage)) {
	s.mu.Lock()
	s.hook = hook
	s.mu.Unlock()
}

func (s *fakeServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *fakeServer) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}
	var request struct {
		ID     string          `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(line, &request); err != nil {
		return
	}
	s.mu.Lock()
	s.requests = append(s.requests, recordedRequest{ID: request.ID, Method: request.Method, Params: string(request.Params)})
	hook := s.hook
	var reply scriptedReply
	found := false
	if request.Method == "ping" {
		reply = scriptedReply{result: `{"type":"pong","version":"0.9.5","protocol":22}`}
		found = true
	} else if list := s.replies[request.Method]; len(list) > 0 {
		reply = list[0]
		s.replies[request.Method] = list[1:]
		found = true
	}
	s.mu.Unlock()
	if hook != nil {
		hook(request.Method, request.Params)
	}
	if !found {
		s.t.Errorf("unexpected method %q", request.Method)
		return
	}
	body := `{"id":` + quote(request.ID)
	if reply.error != "" {
		body += `,"error":` + reply.error + "}\n"
	} else {
		body += `,"result":` + reply.result + "}\n"
	}
	_, _ = conn.Write([]byte(body))
}

func (s *fakeServer) recorded(method string) []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	found := []recordedRequest{}
	for _, request := range s.requests {
		if request.Method == method {
			found = append(found, request)
		}
	}
	return found
}

func quote(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}
