// Package chat stores and transports messages between the user and coordinator.
package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/panelock"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

var chatIDPattern = regexp.MustCompile(`^c[1-9][0-9]*$`)

// Entry is one message in the chat history.
type Entry struct {
	ID       string `json:"id"`
	At       string `json:"at"`
	From     string `json:"from,omitempty"`
	Text     string `json:"text,omitempty"`
	Delivery string `json:"delivery,omitempty"`
	ReplyTo  string `json:"reply_to,omitempty"`
}

// SendOptions holds the caller and text of one user message.
type SendOptions struct {
	Agent      string
	PaneID     string
	WorkingDir string
	Text       string
}

// SendResult reports the ID and delivery state of one user message.
type SendResult struct {
	ID       string `json:"id"`
	Delivery string `json:"delivery"`
}

// ReplyOptions holds the caller and text of one coordinator reply.
type ReplyOptions struct {
	Agent  string
	PaneID string
	To     string
	Text   string
}

// ReplyResult reports one coordinator reply.
type ReplyResult struct {
	Entry Entry `json:"entry"`
}

// LogOptions filters the chat history.
type LogOptions struct {
	Since string
	Limit int
}

// LogResult reports the messages and malformed lines in the history.
type LogResult struct {
	Entries  []Entry `json:"entries"`
	BadLines int     `json:"bad_lines"`
}

// Send appends one user message and prompts the live coordinator once.
func Send(ctx context.Context, home string, options SendOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.Text == "" {
		return failed("validate", "chat text is empty", nil)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error(), nil)
	}
	var coordinator records.Agent
	var entry Entry
	var date string
	var phase, message string
	lockErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
		if err := checkAgentCaller(locked, options.Agent, options.PaneID); err != nil {
			phase, message = "identity", err.Error()
			return nil
		}
		coordinator, err = identity.ResolveHeld(locked, "coordinator", "")
		if err != nil {
			phase, message = "coordinator", "no live coordinator: "+err.Error()
			return nil
		}
		if err := checkCallerContext(ctx, locked, home, options.PaneID, options.WorkingDir); err != nil {
			phase, message = "identity", err.Error()
			return nil
		}
		at := time.Now().UTC().Format(time.RFC3339Nano)
		seq, err := locked.Next("chat")
		if err != nil {
			phase, message = "counter", "next chat ID: "+err.Error()
			return nil
		}
		entry = Entry{ID: fmt.Sprintf("c%d", seq), At: at, From: "user", Text: options.Text, Delivery: "pending"}
		date = at[:10]
		if err := appendLine(home, date, entry); err != nil {
			phase, message = "write", "append chat message: "+err.Error()
		}
		return nil
	})
	if lockErr != nil {
		return failed("store", lockErr.Error(), nil)
	}
	if message != "" {
		return failed(phase, message, nil)
	}

	promptErr := prompt(ctx, home, coordinator.PaneID, marker.Prefix+"user: "+options.Text)
	delivery := "sent"
	if promptErr != nil {
		delivery = "unknown"
	}
	if err := appendDelivery(ctx, home, date, entry.ID, delivery); err != nil {
		return failed("delivery", "write chat delivery state: "+err.Error(), SendResult{ID: entry.ID, Delivery: "pending"})
	}
	if promptErr != nil {
		return failed("delivery", "chat delivery is unknown: "+promptErr.Error(), SendResult{ID: entry.ID, Delivery: delivery})
	}
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{},
		Error:   nil,
		Result:  SendResult{ID: entry.ID, Delivery: delivery},
	}
}

const userOnlyPrefix = "chat send is for the user only"

// ValidateCaller checks whether one caller can send user chat.
func ValidateCaller(ctx context.Context, homePath, agentFlag, paneID, workingDir string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	registry, err := store.Open(homePath)
	if err != nil {
		return err
	}
	var inner error
	if err := registry.WithLock(ctx, func(locked *store.LockedStore) error {
		inner = checkUserCaller(ctx, locked, homePath, agentFlag, paneID, workingDir)
		return nil
	}); err != nil {
		return err
	}
	return inner
}

func checkUserCaller(ctx context.Context, locked *store.LockedStore, homePath, agentFlag, paneID, workingDir string) error {
	if err := checkAgentCaller(locked, agentFlag, paneID); err != nil {
		return err
	}
	return checkCallerContext(ctx, locked, homePath, paneID, workingDir)
}

func checkAgentCaller(locked *store.LockedStore, agentFlag, paneID string) error {
	if agentFlag != "" {
		return fmt.Errorf("%s: only the user can send chat; rule 1 (caller used --agent)", userOnlyPrefix)
	}
	if paneID == "" {
		return nil
	}
	caller, err := identity.ResolveHeld(locked, "", paneID)
	if err == nil {
		return fmt.Errorf("%s: only the user can send chat; rule 1 (caller pane belongs to the live %s agent)", userOnlyPrefix, caller.Role)
	}
	if !errors.Is(err, identity.ErrCallerUnknown) {
		return fmt.Errorf("%s: only the user can send chat; rule 1 check failed: resolve caller: %w", userOnlyPrefix, err)
	}
	return nil
}

func checkCallerContext(ctx context.Context, locked *store.LockedStore, homePath, paneID, workingDir string) error {
	if workingDir != "" {
		below, err := callerBelowProjects(homePath, workingDir)
		if err != nil {
			return fmt.Errorf("%s: only the user can send chat; rule 2 check failed: %w", userOnlyPrefix, err)
		}
		if below {
			return fmt.Errorf("%s: only the user can send chat; rule 2 (working directory is below projects/)", userOnlyPrefix)
		}
	}
	if paneID == "" {
		return nil
	}
	pane, err := herdr.PaneGet(ctx, paneID)
	if err != nil {
		return fmt.Errorf("%s: only the user can send chat; rule 3 check failed: read caller pane: %w", userOnlyPrefix, err)
	}
	coordinator, err := identity.ResolveHeld(locked, "coordinator", "")
	if err == nil && pane.WorkspaceID == coordinator.WorkspaceID {
		return nil
	}
	agents, err := locked.ListIn([]string{"agents"}, nil)
	if err != nil {
		return fmt.Errorf("%s: only the user can send chat; rule 3 check failed: read agent records: %w", userOnlyPrefix, err)
	}
	for _, data := range agents {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return fmt.Errorf("%s: only the user can send chat; rule 3 check failed: decode agent record: %w", userOnlyPrefix, err)
		}
		if record.Status == records.AgentLive && record.WorkspaceID == pane.WorkspaceID &&
			(record.Role == "manager" || record.Role == "worker" || record.Role == "verifier") {
			return fmt.Errorf("%s: only the user can send chat; rule 3 (caller pane shares a live %s workspace)", userOnlyPrefix, record.Role)
		}
	}
	return nil
}

func callerBelowProjects(homePath, workingDir string) (bool, error) {
	absoluteDir, err := filepath.Abs(workingDir)
	if err != nil {
		return false, err
	}
	canonicalDir, err := filepath.EvalSymlinks(absoluteDir)
	if err != nil {
		return false, err
	}
	projectHome := homePath
	if markedHome, ok := home.FindMarkerAbove(canonicalDir); ok {
		projectHome = markedHome
	}
	canonicalHome, err := filepath.EvalSymlinks(projectHome)
	if err != nil {
		return false, err
	}
	projectsDir := filepath.Join(canonicalHome, "projects")
	if resolved, err := filepath.EvalSymlinks(projectsDir); err == nil {
		projectsDir = resolved
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	relative, err := filepath.Rel(projectsDir, canonicalDir)
	if err != nil {
		return false, err
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relative)), nil
}

func prompt(ctx context.Context, home, paneID, text string) error {
	lock, err := panelock.Lock(ctx, home, paneID)
	if err != nil {
		return fmt.Errorf("pane lock: %w", err)
	}
	defer func() { _ = lock.Unlock() }()
	_, err = herdr.AgentPrompt(ctx, herdr.AgentPromptParams{Target: paneID, Text: text})
	return err
}

func appendDelivery(ctx context.Context, home, date, id, delivery string) error {
	registry, err := store.Open(home)
	if err != nil {
		return err
	}
	return registry.WithLock(ctx, func(locked *store.LockedStore) error {
		update := Entry{ID: id, At: time.Now().UTC().Format(time.RFC3339Nano), Delivery: delivery}
		return appendLine(home, date, update)
	})
}

// Reply appends one message from the live coordinator.
func Reply(home string, options ReplyOptions) output.Envelope {
	if options.Text == "" {
		return failed("validate", "chat reply text is empty", nil)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error(), nil)
	}
	var entry Entry
	var phase, message string
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		caller, err := identity.ResolveHeld(locked, options.Agent, options.PaneID)
		if err != nil || caller.Role != "coordinator" {
			phase = "identity"
			message = "only the coordinator can run hive chat reply"
			if err != nil {
				message += ": " + err.Error()
			}
			return nil
		}
		if options.To != "" {
			if !chatIDPattern.MatchString(options.To) {
				phase, message = "validate", "chat user message not found: "+options.To
				return nil
			}
			existing, _, err := readEntriesLocked(home, "")
			if err != nil {
				phase, message = "read", "read chat history: "+err.Error()
				return nil
			}
			found := false
			for _, candidate := range existing {
				if candidate.ID == options.To && candidate.From == "user" {
					found = true
					break
				}
			}
			if !found {
				phase, message = "validate", "chat user message not found: "+options.To
				return nil
			}
		}
		at := time.Now().UTC().Format(time.RFC3339Nano)
		seq, err := locked.Next("chat")
		if err != nil {
			phase, message = "counter", "next chat ID: "+err.Error()
			return nil
		}
		entry = Entry{ID: fmt.Sprintf("c%d", seq), At: at, From: "coordinator", Text: options.Text, ReplyTo: options.To}
		if err := appendLine(home, at[:10], entry); err != nil {
			phase, message = "write", "append chat reply: "+err.Error()
		}
		return nil
	})
	if lockErr != nil {
		return failed("store", lockErr.Error(), nil)
	}
	if message != "" {
		return failed(phase, message, nil)
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: ReplyResult{Entry: entry}}
}

// Log reads messages in order and applies the requested filters.
func Log(home string, options LogOptions) output.Envelope {
	if options.Since != "" {
		parsed, err := time.Parse("2006-01-02", options.Since)
		if err != nil || parsed.Format("2006-01-02") != options.Since {
			return failed("validate", "--since must use YYYY-MM-DD", nil)
		}
	}
	if options.Limit < 0 {
		return failed("validate", "--limit must be zero or greater", nil)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error(), nil)
	}
	var result LogResult
	if err := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var err error
		result.Entries, result.BadLines, err = readEntriesLocked(home, options.Since)
		return err
	}); err != nil {
		return failed("read", "read chat history: "+err.Error(), nil)
	}
	if options.Limit > 0 && len(result.Entries) > options.Limit {
		result.Entries = result.Entries[len(result.Entries)-options.Limit:]
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: result}
}

func readEntriesLocked(home, since string) ([]Entry, int, error) {
	dir := filepath.Join(home, ".hive", "state", "chat")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Entry{}, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	files := make([]string, 0, len(entries))
	for _, item := range entries {
		name := item.Name()
		if item.IsDir() || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		date := strings.TrimSuffix(name, ".jsonl")
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil || parsed.Format("2006-01-02") != date || (since != "" && date < since) {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	messages := make([]Entry, 0)
	messageIndex := make(map[string]int)
	badLines := 0
	for _, path := range files {
		info, err := os.Lstat(path)
		if err != nil {
			return nil, badLines, err
		}
		if !info.Mode().IsRegular() {
			return nil, badLines, fmt.Errorf("chat file is not regular: %s", path)
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, badLines, err
		}
		reader := bufio.NewReader(file)
		for {
			line, readErr := reader.ReadString('\n')
			if len(line) != 0 {
				var entry Entry
				if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &entry); err != nil || !validTimestamp(entry.At) || !chatIDPattern.MatchString(entry.ID) {
					badLines++
				} else if entry.From == "" && entry.Text == "" && entry.ReplyTo == "" {
					index, found := messageIndex[entry.ID]
					if !found || messages[index].From != "user" || (entry.Delivery != "sent" && entry.Delivery != "unknown") {
						badLines++
					} else {
						messages[index].Delivery = entry.Delivery
					}
				} else if !validMessage(entry) {
					badLines++
				} else if _, exists := messageIndex[entry.ID]; exists {
					badLines++
				} else {
					messageIndex[entry.ID] = len(messages)
					messages = append(messages, entry)
				}
			}
			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				_ = file.Close()
				return nil, badLines, readErr
			}
		}
		if err := file.Close(); err != nil {
			return nil, badLines, err
		}
	}
	return messages, badLines, nil
}

func validTimestamp(value string) bool {
	if value == "" {
		return false
	}
	_, err := time.Parse(time.RFC3339Nano, value)
	return err == nil
}

func validMessage(entry Entry) bool {
	if entry.Text == "" {
		return false
	}
	switch entry.From {
	case "user":
		return entry.Delivery == "pending" && entry.ReplyTo == ""
	case "coordinator":
		return entry.Delivery == "" && (entry.ReplyTo == "" || chatIDPattern.MatchString(entry.ReplyTo))
	default:
		return false
	}
}

func appendLine(home, date string, value any) error {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil || parsed.Format("2006-01-02") != date {
		return fmt.Errorf("invalid chat date %q", date)
	}
	dir := filepath.Join(home, ".hive", "state", "chat")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("chat path is not a directory")
	}
	path := filepath.Join(dir, date+".jsonl")
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("chat file is not regular")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func failed(phase, message string, result any) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: result}
}

// SendText renders the ID and delivery state of one user message.
func SendText(envelope output.Envelope) string {
	result, ok := envelope.Result.(SendResult)
	if !ok {
		return ""
	}
	return fmt.Sprintf("chat %s %s\n", result.ID, result.Delivery)
}

// ReplyText renders the ID of one coordinator reply.
func ReplyText(envelope output.Envelope) string {
	result, ok := envelope.Result.(ReplyResult)
	if !ok {
		return ""
	}
	return fmt.Sprintf("chat %s reply saved\n", result.Entry.ID)
}

// LogText renders each message and the count of malformed lines.
func LogText(envelope output.Envelope) string {
	result, ok := envelope.Result.(LogResult)
	if !ok {
		return ""
	}
	var text strings.Builder
	for _, entry := range result.Entries {
		fmt.Fprintf(&text, "%s %s", entry.ID, entry.From)
		if entry.Delivery != "" {
			fmt.Fprintf(&text, " [%s]", entry.Delivery)
		}
		if entry.ReplyTo != "" {
			fmt.Fprintf(&text, " -> %s", entry.ReplyTo)
		}
		fmt.Fprintf(&text, ": %s\n", entry.Text)
	}
	if result.BadLines != 0 {
		fmt.Fprintf(&text, "skipped %d bad chat lines\n", result.BadLines)
	}
	return text.String()
}
