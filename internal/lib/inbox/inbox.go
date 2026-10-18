package inbox

import (
	"regexp"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// MaxTextBytes is the maximum UTF-8 byte length of one inbox record text.
const MaxTextBytes = 4096

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

// Append writes one inbox record.
func Append(home string, options AppendOptions) output.Envelope {
	return failed("stub", "append not implemented")
}

// Consume marks each unread record of key and returns the marked records.
func Consume(home, key string) output.Envelope {
	return failed("stub", "consume not implemented")
}

// Unread returns each unread record of key in sequence order.
func Unread(home, key string) ([]records.Inbox, error) {
	return nil, nil
}

// All returns each record of key in sequence order.
func All(home, key string) ([]records.Inbox, error) {
	return nil, nil
}

// UnreadCount returns the number of unread records of key.
func UnreadCount(home, key string) (int, error) {
	return 0, nil
}

// UnreadCountAll returns the unread count of each inbox under the home.
func UnreadCountAll(home string) (map[string]int, error) {
	return map[string]int{}, nil
}

func appendLocked(locked *store.LockedStore, writer lockedWriter, options AppendOptions, now func() time.Time) output.Envelope {
	_ = locked
	_ = writer
	_ = options
	_ = now
	return failed("stub", "append not implemented")
}

func consumeLocked(locked *store.LockedStore, key string, mark MarkFunc) output.Envelope {
	_ = locked
	_ = key
	_ = mark
	return failed("stub", "consume not implemented")
}

func checkKey(key string) error {
	return nil
}

func boundedText(text string) error {
	return nil
}

func failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
