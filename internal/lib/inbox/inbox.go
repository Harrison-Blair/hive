package inbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// MaxTextBytes is the maximum UTF-8 byte length of one inbox record text.
const MaxTextBytes = 4096

const maxFieldBytes = 512

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,40}$`)

// AppendOptions holds explicit values for one inbox record.
type AppendOptions struct {
	Role  string
	Kind  string
	Task  string
	Agent string
	Text  string
}

// Result is the records that a consume call marked and the records that remain unread.
type Result struct {
	Records   []records.Inbox `json:"records"`
	Remaining []records.Inbox `json:"remaining,omitempty"`
}

// MarkFunc marks one record as read and returns the marked record.
type MarkFunc func(*store.LockedStore, records.Inbox) (records.Inbox, error)

type lockedWriter interface {
	Next(name string) (uint64, error)
	WriteIn(collection []string, id string, value any) error
}

type recordLister interface {
	ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error)
}

// Append writes one inbox record.
func Append(home string, options AppendOptions) output.Envelope {
	if err := validateAppend(options); err != nil {
		envelope := failed("validate", err.Error())
		logAppend(home, options, envelope)
		return envelope
	}
	registry, err := store.Open(home)
	if err != nil {
		envelope := failed("store", err.Error())
		logAppend(home, options, envelope)
		return envelope
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = appendLocked(locked, locked, options, time.Now)
		return nil
	})
	if err != nil {
		envelope = failed("store", err.Error())
	}
	logAppend(home, options, envelope)
	return envelope
}

func appendLocked(locked *store.LockedStore, writer lockedWriter, options AppendOptions, now func() time.Time) output.Envelope {
	_ = locked
	if err := validateAppend(options); err != nil {
		return failed("validate", err.Error())
	}
	kind := options.Kind
	if kind == "" {
		kind = records.InboxMessage
	}
	seq, err := writer.Next("inbox:" + options.Role)
	if err != nil {
		return failed("counter", "next inbox counter: "+err.Error())
	}
	effects := []output.Effect{counterEffect(options.Role)}
	record := records.Inbox{
		Seq:       seq,
		Role:      options.Role,
		Kind:      kind,
		Task:      options.Task,
		Agent:     options.Agent,
		Text:      options.Text,
		CreatedAt: now().UTC().Format(time.RFC3339Nano),
		ReadAt:    "",
		NudgedAt:  "",
	}
	id := strconv.FormatUint(seq, 10)
	if err := writer.WriteIn([]string{"inbox", options.Role}, id, record); err != nil {
		return output.Envelope{Outcome: "partial", Phase: "record", Effects: effects, Error: "write inbox record: " + err.Error(), Result: nil}
	}
	effects = append(effects, output.Effect{Resource: "inbox", Path: recordPath(options.Role, seq), Action: "created"})
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: record}
}

// Consume marks each unread record of key and returns the marked records.
func Consume(home, key string) output.Envelope {
	if err := checkKey(key); err != nil {
		envelope := failed("validate", err.Error())
		logConsume(home, key, envelope)
		return envelope
	}
	registry, err := store.Open(home)
	if err != nil {
		envelope := failed("store", err.Error())
		logConsume(home, key, envelope)
		return envelope
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = consumeLocked(locked, key, markRead, home)
		return nil
	})
	if err != nil {
		envelope = failed("store", err.Error())
	}
	logConsume(home, key, envelope)
	return envelope
}

func consumeLocked(locked *store.LockedStore, key string, mark MarkFunc, home ...string) output.Envelope {
	base := ""
	if len(home) != 0 {
		base = home[0]
	}
	unread, err := listLocked(locked, key, false, base)
	if err != nil {
		return failed("consume", err.Error())
	}
	if len(unread) == 0 {
		return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: Result{Records: []records.Inbox{}}}
	}
	marked := make([]records.Inbox, 0, len(unread))
	effects := make([]output.Effect, 0, len(unread))
	for i, record := range unread {
		updated, err := mark(locked, record)
		if err != nil {
			if len(marked) == 0 {
				return failed("consume", "mark inbox record "+strconv.FormatUint(record.Seq, 10)+": "+err.Error())
			}
			return output.Envelope{
				Outcome: "partial",
				Phase:   "consume",
				Effects: effects,
				Error:   "mark inbox record " + strconv.FormatUint(record.Seq, 10) + ": " + err.Error(),
				Result:  Result{Records: marked, Remaining: append([]records.Inbox(nil), unread[i:]...)},
			}
		}
		marked = append(marked, updated)
		effects = append(effects, output.Effect{Resource: "inbox", Path: recordPath(key, record.Seq), Action: "consumed"})
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: Result{Records: marked}}
}

func markRead(locked *store.LockedStore, record records.Inbox) (records.Inbox, error) {
	var updated records.Inbox
	err := locked.UpdateIn([]string{"inbox", record.Role}, strconv.FormatUint(record.Seq, 10), func(raw json.RawMessage) (json.RawMessage, error) {
		if err := json.Unmarshal(raw, &updated); err != nil {
			return nil, err
		}
		updated.ReadAt = time.Now().UTC().Format(time.RFC3339Nano)
		return json.Marshal(updated)
	})
	return updated, err
}

// Unread returns each unread record of key in sequence order.
func Unread(home, key string) ([]records.Inbox, error) {
	if err := checkKey(key); err != nil {
		logReadFailure(home, key, err.Error())
		return nil, err
	}
	registry, err := store.Open(home)
	if err != nil {
		logReadFailure(home, key, err.Error())
		return nil, err
	}
	found, err := listLocked(registry, key, false, home)
	if err != nil {
		logReadFailure(home, key, err.Error())
		return nil, err
	}
	return found, nil
}

// All returns each record of key in sequence order.
func All(home, key string) ([]records.Inbox, error) {
	if err := checkKey(key); err != nil {
		logReadFailure(home, key, err.Error())
		return nil, err
	}
	registry, err := store.Open(home)
	if err != nil {
		logReadFailure(home, key, err.Error())
		return nil, err
	}
	found, err := listLocked(registry, key, true, home)
	if err != nil {
		logReadFailure(home, key, err.Error())
		return nil, err
	}
	return found, nil
}

func listLocked(lister recordLister, key string, all bool, home ...string) ([]records.Inbox, error) {
	base := ""
	if len(home) != 0 {
		base = home[0]
	}
	names := []string{}
	raw, err := lister.ListIn([]string{"inbox", key}, func(filename string) bool {
		names = append(names, filename)
		return true
	})
	if err != nil {
		return nil, err
	}
	found := make([]records.Inbox, 0, len(raw))
	for i, data := range raw {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		path := recordErrorPath(base, key, name)
		var record records.Inbox
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s: %w", path, err)
		}
		id := strings.TrimSuffix(name, ".json")
		seq, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("record invalid: %s: the sequence name is not numeric", path)
		}
		if id != strconv.FormatUint(seq, 10) {
			return nil, fmt.Errorf("record invalid: %s: the sequence name is not canonical", path)
		}
		if record.Seq != seq {
			return nil, fmt.Errorf("record invalid: %s: sequence %d does not match the record path", path, record.Seq)
		}
		if record.Role != key {
			return nil, fmt.Errorf("record invalid: %s: role %q does not match inbox %q", path, record.Role, key)
		}
		if !all && record.ReadAt != "" {
			continue
		}
		found = append(found, record)
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Seq < found[j].Seq })
	return found, nil
}

func recordErrorPath(home, key, name string) string {
	if home != "" {
		return filepath.ToSlash(filepath.Join(home, ".hive", "state", "inbox", key, name))
	}
	return filepath.ToSlash(filepath.Join(".hive", "state", "inbox", key, name))
}

// UnreadCount returns the number of unread records of key.
func UnreadCount(home, key string) (int, error) {
	found, err := Unread(home, key)
	if err != nil {
		return 0, err
	}
	return len(found), nil
}

// UnreadCountAll returns the unread count of each inbox under the home.
func UnreadCountAll(home string) (map[string]int, error) {
	if _, err := store.Open(home); err != nil {
		logReadFailure(home, "", err.Error())
		return nil, err
	}
	root := filepath.Join(home, ".hive", "state", "inbox")
	info, err := os.Lstat(root)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]int{}, nil
	}
	if err != nil {
		logReadFailure(home, "", err.Error())
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		message := fmt.Sprintf("store path invalid: %s", root)
		logReadFailure(home, "", message)
		return nil, errors.New(message)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		logReadFailure(home, "", err.Error())
		return nil, err
	}
	counts := map[string]int{}
	for _, entry := range entries {
		key := entry.Name()
		if checkKey(key) != nil {
			continue
		}
		rolePath := filepath.Join(root, key)
		roleInfo, err := os.Lstat(rolePath)
		if err != nil {
			logReadFailure(home, key, err.Error())
			return nil, err
		}
		if roleInfo.Mode()&os.ModeSymlink != 0 || !roleInfo.IsDir() {
			message := fmt.Sprintf("store path invalid: %s", rolePath)
			logReadFailure(home, key, message)
			return nil, errors.New(message)
		}
		count, err := UnreadCount(home, key)
		if err != nil {
			return nil, err
		}
		counts[key] = count
	}
	return counts, nil
}

func checkKey(key string) error {
	if !keyPattern.MatchString(key) {
		return fmt.Errorf("invalid inbox key: %s", displayValue(key))
	}
	return nil
}

func boundedText(text string) error {
	if len(text) > MaxTextBytes {
		return fmt.Errorf("inbox text is %d bytes; the maximum is %d", len(text), MaxTextBytes)
	}
	return nil
}

func validateAppend(options AppendOptions) error {
	if err := checkKey(options.Role); err != nil {
		return err
	}
	if err := boundedText(options.Text); err != nil {
		return err
	}
	kind := options.Kind
	if kind == "" {
		kind = records.InboxMessage
	}
	switch kind {
	case records.InboxCompleted, records.InboxVerified, records.InboxBlocked,
		records.InboxExited, records.InboxQuiet, records.InboxUnresolved, records.InboxMessage:
	default:
		return fmt.Errorf("invalid inbox kind: %s", displayValue(kind))
	}
	for _, field := range []struct {
		label string
		value string
	}{
		{"task", options.Task},
		{"agent", options.Agent},
	} {
		if len(field.value) > maxFieldBytes {
			return fmt.Errorf("invalid inbox %s: value is longer than %d bytes", field.label, maxFieldBytes)
		}
		for _, r := range field.value {
			if unicode.IsControl(r) {
				return fmt.Errorf("invalid inbox %s: %q", field.label, field.value)
			}
		}
	}
	return nil
}

func displayValue(value string) string {
	if !utf8.ValidString(value) {
		return strconv.Quote(value)
	}
	if strings.ContainsAny(value, "\n\r\t") {
		return strconv.Quote(value)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return strconv.Quote(value)
		}
	}
	return value
}

func logInboxEvent(home, message string, attrs ...any) {
	writer, err := hivelog.NewWriter(home, "inbox", time.Now)
	if err != nil {
		return
	}
	defer func() { _ = writer.Close() }()
	slog.New(writer).Debug(message, attrs...)
}

func logReadFailure(home, key, message string) {
	attrs := []any{"phase", "read", "error", message}
	if key != "" {
		attrs = append([]any{"role", key}, attrs...)
	}
	logInboxEvent(home, "inbox operation failed", attrs...)
}

func logAppend(home string, options AppendOptions, envelope output.Envelope) {
	kind := options.Kind
	if kind == "" {
		kind = records.InboxMessage
	}
	attrs := []any{"role", options.Role, "agent", options.Agent, "agent_id", options.Agent, "kind", kind}
	sequence := uint64(0)
	for _, effect := range envelope.Effects {
		switch effect.Resource {
		case "counter":
			logInboxEvent(home, "inbox counter updated", append(attrs, "path", effect.Path)...)
		case "inbox":
			sequence = effectSequence(effect.Path)
			logInboxEvent(home, "inbox record appended", append(attrs, "sequence", sequence)...)
		}
	}
	if envelope.Error == nil {
		return
	}
	errorAttrs := append([]any{}, attrs...)
	if sequence != 0 {
		errorAttrs = append(errorAttrs, "sequence", sequence)
	}
	errorAttrs = append(errorAttrs, "phase", envelope.Phase, "error", fmt.Sprint(envelope.Error))
	logInboxEvent(home, "inbox operation failed", errorAttrs...)
}

func logConsume(home, key string, envelope output.Envelope) {
	attrs := []any{}
	if key != "" {
		attrs = append(attrs, "role", key)
	}
	senders := map[uint64]string{}
	if result, ok := envelope.Result.(Result); ok {
		for _, record := range result.Records {
			senders[record.Seq] = record.Agent
		}
	}
	sequence := uint64(0)
	for _, effect := range envelope.Effects {
		if effect.Resource != "inbox" || effect.Action != "consumed" {
			continue
		}
		sequence = effectSequence(effect.Path)
		sender := senders[sequence]
		markAttrs := append([]any{}, attrs...)
		markAttrs = append(markAttrs, "sequence", sequence, "agent", sender, "agent_id", sender)
		logInboxEvent(home, "inbox record consumed", markAttrs...)
	}
	if envelope.Error == nil {
		return
	}
	errorAttrs := append([]any{}, attrs...)
	if sequence != 0 {
		errorAttrs = append(errorAttrs, "sequence", sequence)
	}
	errorAttrs = append(errorAttrs, "phase", envelope.Phase, "error", fmt.Sprint(envelope.Error))
	logInboxEvent(home, "inbox operation failed", errorAttrs...)
}

func effectSequence(path string) uint64 {
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	seq, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0
	}
	return seq
}

func counterEffect(key string) output.Effect {
	return output.Effect{
		Resource: "counter",
		Path:     filepath.ToSlash(filepath.Join(".hive", "state", "counters", "inbox:"+key+".json")),
		Action:   "updated",
	}
}

func recordPath(key string, seq uint64) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", "inbox", key, strconv.FormatUint(seq, 10)+".json"))
}

func failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}

// HeldWriter is the store write surface of one held inbox append.
type HeldWriter interface {
	Next(name string) (uint64, error)
	WriteIn(collection []string, id string, value any) error
}

// ValidateAppend checks one inbox append before a caller writes any record.
func ValidateAppend(options AppendOptions) error {
	return validateAppend(options)
}

// AppendHeld appends one record through a held store handle.
func AppendHeld(writer HeldWriter, options AppendOptions) output.Envelope {
	return appendLocked(nil, writer, options, time.Now)
}

// ConsumeHeld marks the unread records through a held store handle.
func ConsumeHeld(locked *store.LockedStore, key string) output.Envelope {
	if err := checkKey(key); err != nil {
		return failed("validate", err.Error())
	}
	return consumeLocked(locked, key, markRead)
}

// AllHeld returns each record of key through a held store handle.
func AllHeld(locked *store.LockedStore, key string) ([]records.Inbox, error) {
	if err := checkKey(key); err != nil {
		return nil, err
	}
	return listLocked(locked, key, true)
}

// LogAppend writes the log events of one append call.
func LogAppend(home string, options AppendOptions, envelope output.Envelope) {
	logAppend(home, options, envelope)
}

// LogConsume writes the log events of one consume call.
func LogConsume(home, key string, envelope output.Envelope) {
	logConsume(home, key, envelope)
}
