package send_test

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
)

const userChatRefusal = "chat send is for the user only: type in the Chat screen of the hive TUI"

const chatImportPath = "github.com/Harrison-Blair/hive/internal/chat"

// TestChatSendRefusesEachCaller proves AC1, AC2, and AC3: each prohibited
// call gets the exact refusal, writes no chat record, sends no prompt, and
// logs the refusal with its pane and working directory.
func TestChatSendRefusesEachCaller(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		args     []string
		pane     string
		wantPane string
	}{
		{name: "no flags", args: []string{"chat", "send", "hello"}},
		{name: "agent flag", args: []string{"chat", "send", "--agent", "coordinator", "hello"}},
		{name: "pane in a new tab", args: []string{"chat", "send", "hello"}, pane: "w1:p2", wantPane: "w1:p2"},
		{name: "agent flag in a user pane", args: []string{"chat", "send", "--agent", "coordinator", "hello"}, pane: "w1:p2", wantPane: "w1:p2"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			home := newChatCommandHome(t)
			fake := newFakeChatHerdr(t, "w1")
			runChatCommandFrom(t, home, home, testCase.pane)
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			var out bytes.Buffer
			if err := cmd.ExecuteWithArgs(testCase.args, strings.NewReader(""), &out); err == nil {
				t.Fatalf("chat send succeeded; want the refusal; output %q", out.String())
			}
			if got, want := out.String(), "hive: "+userChatRefusal+"\n"; got != want {
				t.Fatalf("output = %q; want %q", got, want)
			}
			assertNoChatHistory(t, home)
			if paneGets, prompts := fake.counts(); paneGets != 0 || prompts != 0 {
				t.Fatalf("herdr calls = pane.get %d, agent.prompt %d; want none", paneGets, prompts)
			}
			events := chatSendRefusals(t, home)
			if len(events) != 1 {
				t.Fatalf("chat send refused events = %d; want one", len(events))
			}
			if events[0]["pane_id"] != testCase.wantPane || events[0]["cwd"] != cwd {
				t.Fatalf("refusal event = %#v; want pane_id %q and cwd %q", events[0], testCase.wantPane, cwd)
			}
		})
	}
}

// TestNoCmdFileCallsChatSend proves AC4 over the live command tree.
func TestNoCmdFileCallsChatSend(t *testing.T) {
	offenders, err := filesCallingChatSend("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) != 0 {
		t.Fatalf("cmd files call chat.Send: %v", offenders)
	}
}

// TestScratchFileCallingChatSendIsFound proves that the AC4 check fails on a
// scratch file with a chat.Send call.
func TestScratchFileCallingChatSendIsFound(t *testing.T) {
	dir := t.TempDir()
	source := `package scratch

import (
	"context"

	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

var _ = func() output.Envelope { return chatlib.Send(context.Background(), "", chatlib.SendOptions{}) }
`
	path := filepath.Join(dir, "scratch.go")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	offenders, err := filesCallingChatSend(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) != 1 || !strings.HasPrefix(offenders[0], path+":") {
		t.Fatalf("offenders = %v; want the scratch file call", offenders)
	}
}

// filesCallingChatSend parses each Go file below root and returns the position
// of each call of the Send function of the internal chat package.
func filesCallingChatSend(root string) ([]string, error) {
	fset := token.NewFileSet()
	var offenders []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		alias := ""
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil || importPath != chatImportPath {
				continue
			}
			alias = "chat"
			if spec.Name != nil {
				alias = spec.Name.Name
			}
		}
		if alias == "" {
			return nil
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Send" {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if ok && ident.Name == alias {
				offenders = append(offenders, fset.Position(call.Pos()).String())
			}
			return true
		})
		return nil
	})
	return offenders, err
}

// chatSendRefusals reads the chat send refused events of one home.
func chatSendRefusals(t *testing.T, home string) []map[string]any {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("log line %q: %v", line, err)
			}
			if event["msg"] == "chat send refused" {
				events = append(events, event)
			}
		}
	}
	return events
}
