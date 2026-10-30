package agenttest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/project"
)

// Request is one request that the fake received.
type Request struct {
	Method string
	Params map[string]any
}

// Workspace is one fake workspace.
type Workspace struct {
	ID       string
	Label    string
	Checkout string
	Linked   bool
	// NoMetadata hides the worktree metadata, as herdr 0.9.5 does for a
	// workspace that workspace.create made.
	NoMetadata bool
	tabs       int
}

// Pane is one fake pane. Each fake tab holds one pane.
type Pane struct {
	ID          string
	WorkspaceID string
	TabID       string
	// TabLabel is the explicit label of the tab of the pane, or empty.
	TabLabel string
	Cwd      string
	Status   string
	Name     string
	Kind     string
	Args     []string
	Env      map[string]string
	// StartEnv is a copy of Env when agent.start ran in the pane.
	StartEnv map[string]string
	Prompts  []string
	// Session is the native session ID that herdr reports as agent_session.
	Session string
	// ready is true when herdr reports interactive_ready for the agent.
	ready bool
	// pending counts the agent.get calls that still see the launch.
	pending int
}

type failure struct{ code, message string }

// Herdr is the fake herdr server.
type Herdr struct {
	t          testing.TB
	listener   net.Listener
	path       string
	mu         sync.Mutex
	workspaces []*Workspace
	panes      []*Pane
	requests   []Request
	failures   map[string][]failure
	nextWS     int
	nextPane   int
	// StartStatus is the status that agent.wait gives after agent.start.
	StartStatus string
	// WaitStatus is the status that a prompt with a wait gives.
	WaitStatus string
	// StartSession is the session ID that an agent gets when it is ready.
	StartSession string
	// LaunchSession is the session ID that only the agent.start reply gives.
	LaunchSession string
	// PendingSession is the session ID that only the agent.get replies before
	// the agent is ready give.
	PendingSession string
	// PromptSession is the session ID that only the agent.prompt reply gives.
	PromptSession string
}

// NewHerdr starts a fake herdr server and sets HERDR_SOCKET_PATH.
func NewHerdr(t testing.TB) *Herdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	fake := &Herdr{t: t, listener: listener, path: path, failures: map[string][]failure{}, StartStatus: "idle", WaitStatus: "idle"}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() { _ = listener.Close() })
	return fake
}

// AddWorkspace adds a workspace with one tab and one pane. A non-empty
// checkout gives the workspace worktree metadata.
func (h *Herdr) AddWorkspace(label, checkout string, linked bool) (*Workspace, *Pane) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.addWorkspace(label, checkout, linked)
}

// AddPane adds a tab with one pane to a workspace.
func (h *Herdr) AddPane(workspaceID, cwd string) *Pane {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.addPane(h.workspace(workspaceID), cwd, nil)
}

// SetStatus sets the agent status of one pane.
func (h *Herdr) SetStatus(paneID, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pane(paneID).Status = status
}

// StartAgent puts a named agent of kind with status in one pane.
func (h *Herdr) StartAgent(paneID, name, kind, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	pane := h.pane(paneID)
	pane.Name, pane.Kind, pane.Status = name, kind, status
	pane.ready = status != "blocked"
}

// SetSession sets the agent session ID of one pane.
func (h *Herdr) SetSession(paneID, session string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pane(paneID).Session = session
}

// ClosePane removes one pane, as a user who closes it by hand.
func (h *Herdr) ClosePane(paneID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closePane(paneID)
}

// Fail makes the next call of method fail with code and message.
func (h *Herdr) Fail(method, code, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures[method] = append(h.failures[method], failure{code, message})
}

// Workspaces returns a copy of each workspace.
func (h *Herdr) Workspaces() []Workspace {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Workspace, 0, len(h.workspaces))
	for _, ws := range h.workspaces {
		out = append(out, *ws)
	}
	return out
}

// Pane returns a copy of one pane and whether it exists.
func (h *Herdr) Pane(paneID string) (Pane, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	pane := h.pane(paneID)
	if pane == nil {
		return Pane{}, false
	}
	return *pane, true
}

// Methods returns the method of each request in the order of arrival.
func (h *Herdr) Methods() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, 0, len(h.requests))
	for _, request := range h.requests {
		out = append(out, request.Method)
	}
	return out
}

// Requests returns each request of one method.
func (h *Herdr) Requests(method string) []Request {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []Request
	for _, request := range h.requests {
		if request.Method == method {
			out = append(out, request)
		}
	}
	return out
}

// exportLine matches a shell line that exports one variable with a single
// quoted value.
var exportLine = regexp.MustCompile(`^export ([A-Za-z_][A-Za-z0-9_]*)='([^']*)'$`)

func envParam(params map[string]any) map[string]string {
	env := map[string]string{}
	if raw, ok := params["env"].(map[string]any); ok {
		for key, value := range raw {
			env[key], _ = value.(string)
		}
	}
	return env
}

func (h *Herdr) addWorkspace(label, checkout string, linked bool) (*Workspace, *Pane) {
	h.nextWS++
	ws := &Workspace{ID: fmt.Sprintf("w%d", h.nextWS), Label: label, Checkout: checkout, Linked: linked}
	if ws.Label == "" {
		ws.Label = filepath.Base(checkout)
	}
	h.workspaces = append(h.workspaces, ws)
	return ws, h.addPane(ws, checkout, nil)
}

func (h *Herdr) addPane(ws *Workspace, cwd string, env map[string]string) *Pane {
	h.nextPane++
	ws.tabs++
	pane := &Pane{
		ID:          fmt.Sprintf("%s:p%d", ws.ID, h.nextPane),
		WorkspaceID: ws.ID,
		TabID:       fmt.Sprintf("%s:t%d", ws.ID, ws.tabs),
		Cwd:         cwd,
		Status:      "unknown",
		Env:         env,
	}
	h.panes = append(h.panes, pane)
	return pane
}

func (h *Herdr) workspace(id string) *Workspace {
	for _, ws := range h.workspaces {
		if ws.ID == id {
			return ws
		}
	}
	return nil
}

func (h *Herdr) pane(id string) *Pane {
	for _, pane := range h.panes {
		if pane.ID == id || (pane.Name != "" && pane.Name == id) {
			return pane
		}
	}
	return nil
}

func (h *Herdr) closePane(id string) {
	for index, pane := range h.panes {
		if pane.ID != id {
			continue
		}
		h.panes = append(h.panes[:index], h.panes[index+1:]...)
		for _, other := range h.panes {
			if other.WorkspaceID == pane.WorkspaceID {
				return
			}
		}
		h.closeWorkspace(pane.WorkspaceID)
		return
	}
}

func (h *Herdr) closeWorkspace(id string) {
	kept := h.panes[:0]
	for _, pane := range h.panes {
		if pane.WorkspaceID != id {
			kept = append(kept, pane)
		}
	}
	h.panes = kept
	for index, ws := range h.workspaces {
		if ws.ID == id {
			h.workspaces = append(h.workspaces[:index], h.workspaces[index+1:]...)
			return
		}
	}
}

func (h *Herdr) serve() {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			return
		}
		go h.handle(conn)
	}
}

type serverError struct{ code, message string }

func (h *Herdr) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var request struct {
		ID     string         `json:"id"`
		Method string         `json:"method"`
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal(line, &request); err != nil {
		return
	}
	result, failed := h.dispatch(request.Method, request.Params)
	reply := map[string]any{"id": request.ID}
	if failed != nil {
		reply["error"] = map[string]string{"code": failed.code, "message": failed.message}
	} else {
		reply["result"] = result
	}
	data, _ := json.Marshal(reply)
	_, _ = conn.Write(append(data, '\n'))
}

func (h *Herdr) dispatch(method string, params map[string]any) (any, *serverError) {
	if method == "ping" {
		return map[string]any{"type": "pong", "version": "0.9.5", "protocol": 22}, nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requests = append(h.requests, Request{Method: method, Params: params})
	if list := h.failures[method]; len(list) > 0 {
		h.failures[method] = list[1:]
		return nil, &serverError{list[0].code, list[0].message}
	}
	text := func(key string) string { value, _ := params[key].(string); return value }
	paneOf := func() (*Pane, *serverError) {
		pane := h.pane(text("pane_id"))
		if pane == nil {
			return nil, &serverError{"pane_not_found", "pane " + text("pane_id") + " not found"}
		}
		return pane, nil
	}
	switch method {
	case "workspace.create":
		ws, pane := h.addWorkspace(text("label"), text("cwd"), false)
		ws.NoMetadata = true
		pane.Env = envParam(params)
		return map[string]any{"type": "workspace_created", "workspace": h.workspaceInfo(ws), "tab": tabInfo(pane), "root_pane": paneInfo(pane)}, nil
	case "workspace.close":
		if h.workspace(text("workspace_id")) == nil {
			return nil, &serverError{"workspace_not_found", "workspace not found"}
		}
		h.closeWorkspace(text("workspace_id"))
		return map[string]any{"type": "ok"}, nil
	case "tab.create":
		ws := h.workspace(text("workspace_id"))
		if ws == nil {
			return nil, &serverError{"workspace_not_found", "workspace not found"}
		}
		pane := h.addPane(ws, text("cwd"), envParam(params))
		pane.TabLabel = text("label")
		return map[string]any{"type": "tab_created", "tab": tabInfo(pane), "root_pane": paneInfo(pane)}, nil
	case "tab.rename":
		for _, pane := range h.panes {
			if pane.TabID == text("tab_id") {
				pane.TabLabel = text("label")
				return map[string]any{"type": "tab_info", "tab": tabInfo(pane)}, nil
			}
		}
		return nil, &serverError{"tab_not_found", "tab " + text("tab_id") + " not found"}
	case "worktree.create":
		source := h.workspace(text("workspace_id"))
		if source == nil {
			return nil, &serverError{"workspace_not_found", "workspace not found"}
		}
		if source.Linked {
			return nil, &serverError{"linked_worktree_source", "linked source"}
		}
		out, err := exec.Command("git", "-C", source.Checkout, "worktree", "add", "-b", text("branch"), text("path"), text("base")).CombinedOutput()
		if err != nil {
			return nil, &serverError{"worktree_create_failed", string(out)}
		}
		ws, pane := h.addWorkspace(text("label"), text("path"), true)
		return map[string]any{
			"type": "worktree_created", "workspace": h.workspaceInfo(ws), "tab": tabInfo(pane), "root_pane": paneInfo(pane),
			"worktree": map[string]any{"path": text("path"), "label": ws.Label, "branch": text("branch"), "is_bare": false, "is_detached": false, "is_prunable": false, "is_linked_worktree": true, "open_workspace_id": ws.ID},
		}, nil
	case "worktree.list":
		cwd := text("cwd")
		entry := map[string]any{"path": cwd, "label": filepath.Base(cwd), "branch": "main", "is_bare": false, "is_detached": false, "is_prunable": false, "is_linked_worktree": false}
		source := map[string]any{"repo_key": cwd + "/.git", "repo_name": filepath.Base(cwd), "repo_root": cwd, "source_checkout_path": cwd}
		for _, ws := range h.workspaces {
			if ws.Checkout == cwd && !ws.Linked {
				entry["open_workspace_id"] = ws.ID
				source["source_workspace_id"] = ws.ID
				break
			}
		}
		return map[string]any{"type": "worktree_list", "source": source, "worktrees": []any{entry}}, nil
	case "pane.get":
		pane, failed := paneOf()
		if failed != nil {
			return nil, failed
		}
		return map[string]any{"type": "pane_info", "pane": paneInfo(pane)}, nil
	case "pane.close":
		if _, failed := paneOf(); failed != nil {
			return nil, failed
		}
		h.closePane(text("pane_id"))
		return map[string]any{"type": "ok"}, nil
	case "pane.send_input":
		pane, failed := paneOf()
		if failed != nil {
			return nil, failed
		}
		// The shell of the pane runs a submitted export line.
		keys, _ := params["keys"].([]any)
		if match := exportLine.FindStringSubmatch(text("text")); match != nil && len(keys) == 1 && keys[0] == "enter" {
			if pane.Env == nil {
				pane.Env = map[string]string{}
			}
			pane.Env[match[1]] = match[2]
		}
		return map[string]any{"type": "ok"}, nil
	case "agent.start":
		pane, failed := paneOf()
		if failed != nil {
			return nil, &serverError{"agent_pane_not_found", failed.message}
		}
		pane.StartEnv = maps.Clone(pane.Env)
		if pane.Kind != "" {
			return nil, &serverError{"agent_pane_busy", "agent target pane " + pane.ID + " is not an available shell"}
		}
		pane.Name, pane.Kind = text("name"), text("kind")
		// As herdr 0.9.5 does, the launch keeps the shell status until the
		// agent settles, and agent.get sees the launch for a short time.
		pane.Status, pane.ready, pane.pending = "idle", false, 1
		pane.Args = nil
		if raw, ok := params["args"].([]any); ok {
			for _, arg := range raw {
				value, _ := arg.(string)
				pane.Args = append(pane.Args, value)
			}
		}
		argv := append([]string{pane.Kind}, pane.Args...)
		return map[string]any{"type": "agent_started", "agent": withSession(agentInfo(pane), pane.Kind, h.LaunchSession), "argv": argv}, nil
	case "agent.get":
		pane := h.pane(text("target"))
		if pane == nil || pane.Kind == "" {
			return nil, &serverError{"agent_not_found", "agent " + text("target") + " not found"}
		}
		if !pane.ready && pane.pending > 0 {
			pane.pending--
			return map[string]any{"type": "agent_info", "agent": withSession(agentInfo(pane), pane.Kind, h.PendingSession)}, nil
		} else if !pane.ready && pane.Status != "blocked" {
			pane.Status = h.StartStatus
			pane.ready = h.StartStatus != "blocked"
			if pane.ready {
				pane.Session = h.StartSession
			}
		}
		return map[string]any{"type": "agent_info", "agent": agentInfo(pane)}, nil
	case "agent.prompt":
		pane := h.pane(text("target"))
		if pane == nil || pane.Kind == "" {
			return nil, &serverError{"agent_not_found", "agent " + text("target") + " not found"}
		}
		if pane.Status == "blocked" {
			return nil, &serverError{"agent_blocked", "agent " + text("target") + " is blocked and requires interactive input"}
		}
		if !pane.ready {
			return nil, &serverError{"agent_not_ready", "agent " + text("target") + " is not an active named agent"}
		}
		pane.Prompts = append(pane.Prompts, text("text"))
		if _, ok := params["wait"]; ok {
			pane.Status = h.WaitStatus
		} else {
			pane.Status = "working"
		}
		return map[string]any{"type": "agent_prompted", "agent": withSession(agentInfo(pane), pane.Kind, h.PromptSession)}, nil
	}
	h.t.Errorf("fake herdr: unexpected method %q", method)
	return nil, &serverError{"unknown_method", method}
}

func (h *Herdr) workspaceInfo(ws *Workspace) map[string]any {
	count := 0
	for _, pane := range h.panes {
		if pane.WorkspaceID == ws.ID {
			count++
		}
	}
	info := map[string]any{
		"workspace_id": ws.ID, "number": 1, "label": ws.Label, "focused": false, "pane_count": count,
		"tab_count": ws.tabs, "active_tab_id": ws.ID + ":t1", "agent_status": "unknown",
	}
	if ws.Checkout != "" && !ws.NoMetadata {
		info["worktree"] = map[string]any{
			"repo_key": ws.Checkout + "/.git", "repo_name": filepath.Base(ws.Checkout), "repo_root": ws.Checkout,
			"checkout_path": ws.Checkout, "is_linked_worktree": ws.Linked,
		}
	}
	return info
}

func tabInfo(pane *Pane) map[string]any {
	label := pane.TabLabel
	if label == "" {
		label = "1"
	}
	return map[string]any{"tab_id": pane.TabID, "workspace_id": pane.WorkspaceID, "number": 1, "label": label, "focused": false, "pane_count": 1, "agent_status": "unknown"}
}

func paneInfo(pane *Pane) map[string]any {
	info := map[string]any{
		"pane_id": pane.ID, "terminal_id": "term_" + pane.ID, "workspace_id": pane.WorkspaceID, "tab_id": pane.TabID,
		"focused": false, "agent_status": pane.Status, "revision": 0, "cwd": pane.Cwd,
	}
	if pane.Kind != "" {
		info["agent"] = pane.Kind
	}
	return withSession(info, pane.Kind, pane.Session)
}

// withSession adds an agent_session with ID session to info, when session is
// not empty.
func withSession(info map[string]any, kind, session string) map[string]any {
	if session != "" {
		info["agent_session"] = map[string]any{"source": "herdr:" + kind, "agent": kind, "kind": "id", "value": session}
	}
	return info
}

func agentInfo(pane *Pane) map[string]any {
	info := paneInfo(pane)
	info["name"] = pane.Name
	if pane.ready {
		info["interactive_ready"] = true
	}
	return info
}

// GitRepo makes a Git repository with one commit on branch main.
func GitRepo(t testing.TB) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

// Home returns the canonical path of a new empty hive home.
func Home(t testing.TB) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return home
}

// RegisterProject writes one project record in the home.
func RegisterProject(t testing.TB, home, id, path, base string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := project.Write(registry, records.Project{ID: id, Path: path, Base: base}); err != nil {
		t.Fatal(err)
	}
}

// RequestCount returns the number of requests that the fake received.
func (h *Herdr) RequestCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.requests)
}
