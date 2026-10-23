// Package ask stores the questions that an agent asks the user, and brings
// the answer of the user back to the agent.
package ask

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/identity"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/panelock"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// The states of an ask.
const (
	StateOpen      = "open"
	StateAnswered  = "answered"
	StateDismissed = "dismissed"
)

// The limits of an ask (decision 91).
const (
	MaxQuestions = 4
	MinOptions   = 2
	MaxOptions   = 5
	MaxHeader    = 12
)

// OtherLabel is the row that hive adds after the options of each question.
const OtherLabel = "Other: type an answer"

var collection = []string{"asks"}

// Option is one answer that the asker gives.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Preview     string `json:"preview,omitempty"`
}

// Question is one question of an ask.
type Question struct {
	Question    string   `json:"question"`
	Header      string   `json:"header"`
	Options     []Option `json:"options"`
	MultiSelect bool     `json:"multiSelect,omitempty"`
}

// Request is the JSON that hive ask reads.
type Request struct {
	Questions []Question `json:"questions"`
}

// Reply is the answer of the user to one question. Picks are 1-based option
// numbers in ascending order. The number after the last option is the Other
// row, and Other holds its text.
type Reply struct {
	Picks []int  `json:"picks"`
	Other string `json:"other,omitempty"`
	Note  string `json:"note,omitempty"`
}

// Record is one ask in .hive/state/asks/ID.json.
type Record struct {
	ID          string     `json:"id"`
	Asker       string     `json:"asker"`
	AskerKey    string     `json:"asker_key"`
	Parent      string     `json:"parent,omitempty"`
	ParentKey   string     `json:"parent_key,omitempty"`
	CreatedAt   string     `json:"created_at"`
	Questions   []Question `json:"questions"`
	State       string     `json:"state"`
	Replies     []Reply    `json:"replies,omitempty"`
	ClosedAt    string     `json:"closed_at,omitempty"`
	DeliveredAt string     `json:"delivered_at,omitempty"`
	DeliveredTo string     `json:"delivered_to,omitempty"`
	Attempts    int        `json:"attempts,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
}

// CreateOptions holds the caller and the JSON of one ask.
type CreateOptions struct {
	Agent  string
	PaneID string
	Data   []byte
}

// Parse decodes and checks the JSON of one ask.
func Parse(data []byte) (Request, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var request Request
	if err := decoder.Decode(&request); err != nil {
		return Request{}, fmt.Errorf("ask JSON: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return Request{}, errors.New("ask JSON: more than one value")
	}
	if n := len(request.Questions); n < 1 || n > MaxQuestions {
		return Request{}, fmt.Errorf("an ask has 1 to %d questions, not %d", MaxQuestions, n)
	}
	for index, question := range request.Questions {
		at := fmt.Sprintf("question %d: ", index+1)
		if strings.TrimSpace(question.Question) == "" {
			return Request{}, errors.New(at + "the question text is empty")
		}
		if n := utf8.RuneCountInString(question.Header); strings.TrimSpace(question.Header) == "" || n > MaxHeader {
			return Request{}, fmt.Errorf("%sgive a header of %d characters at most, not %q", at, MaxHeader, question.Header)
		}
		if n := len(question.Options); n < MinOptions || n > MaxOptions {
			return Request{}, fmt.Errorf("%sa question has %d to %d options, not %d", at, MinOptions, MaxOptions, n)
		}
		for number, option := range question.Options {
			if strings.TrimSpace(option.Label) == "" {
				return Request{}, fmt.Errorf("%soption %d: the label is empty", at, number+1)
			}
			if strings.TrimSpace(option.Description) == "" {
				return Request{}, fmt.Errorf("%soption %d: the description is empty", at, number+1)
			}
		}
	}
	return request, nil
}

// Create checks one ask, writes its open record, and returns at once.
func Create(ctx context.Context, home string, options CreateOptions) output.Envelope {
	request, err := Parse(options.Data)
	if err != nil {
		return failed("validate", err.Error(), nil)
	}
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error(), nil)
	}
	var record Record
	var phase, message string
	lockErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
		asker, err := identity.ResolveHeld(locked, options.Agent, options.PaneID)
		if err != nil {
			phase, message = "identity", "hive ask needs a live agent: "+err.Error()
			return nil
		}
		parent := ""
		if asker.Parent != "" {
			if found, ok, err := identity.FindHeld(locked, asker.Parent); err == nil && ok {
				parent = found.Name
			}
		}
		seq, err := locked.Next("ask")
		if err != nil {
			phase, message = "counter", "next ask ID: "+err.Error()
			return nil
		}
		record = Record{
			ID:        "a" + strconv.FormatUint(seq, 10),
			Asker:     asker.Name,
			AskerKey:  asker.Key,
			Parent:    parent,
			ParentKey: asker.Parent,
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
			Questions: request.Questions,
			State:     StateOpen,
		}
		if err := locked.WriteIn(collection, record.ID, record); err != nil {
			phase, message = "write", "write ask record: "+err.Error()
		}
		return nil
	})
	if lockErr != nil {
		return failed("store", lockErr.Error(), nil)
	}
	if message != "" {
		return failed(phase, message, nil)
	}
	logTransition(home, "ask", record, "", StateOpen)
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: record}
}

// Text renders the result of hive ask.
func Text(envelope output.Envelope) string {
	record, ok := envelope.Result.(Record)
	if !ok {
		return ""
	}
	return fmt.Sprintf("ask %s %s\n", record.ID, record.State)
}

// List gives each ask record, oldest first.
func List(home string) ([]Record, error) {
	registry, err := store.Open(home)
	if err != nil {
		return nil, err
	}
	var listed []Record
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		listed, err = listLocked(locked)
		return err
	})
	return listed, err
}

func listLocked(locked *store.LockedStore) ([]Record, error) {
	raw, err := locked.ListIn(collection, nil)
	if err != nil {
		return nil, err
	}
	listed := make([]Record, 0, len(raw))
	for _, data := range raw {
		var record Record
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("decode ask record: %w", err)
		}
		listed = append(listed, record)
	}
	sort.Slice(listed, func(i, j int) bool { return sequence(listed[i].ID) < sequence(listed[j].ID) })
	return listed, nil
}

func sequence(id string) uint64 {
	n, _ := strconv.ParseUint(strings.TrimPrefix(id, "a"), 10, 64)
	return n
}

// Route gives the agent that gets the answer: the live asker, or else the
// live parent of the asker at ask time. stopped is true when the asker is
// not live. ok is false when no live agent can get the answer.
func Route(record Record, agents []records.Agent) (target records.Agent, stopped, ok bool) {
	for _, candidate := range agents {
		if candidate.Key == record.AskerKey && candidate.Status == records.AgentLive {
			return candidate, false, true
		}
	}
	for _, candidate := range agents {
		if candidate.Key == record.ParentKey && candidate.Status == records.AgentLive {
			return candidate, true, true
		}
	}
	if record.Parent != "" {
		for _, candidate := range agents {
			if candidate.Name == record.Parent && candidate.Status == records.AgentLive {
				return candidate, true, true
			}
		}
	}
	return records.Agent{}, true, false
}

// Headers gives the headers of the questions, joined with " · ".
func Headers(record Record) string {
	headers := make([]string, len(record.Questions))
	for index, question := range record.Questions {
		headers[index] = question.Header
	}
	return strings.Join(headers, " · ")
}

// Labels gives the picked labels of one reply, joined with ", ".
func Labels(question Question, reply Reply) string {
	labels := make([]string, 0, len(reply.Picks))
	for _, pick := range reply.Picks {
		if pick >= 1 && pick <= len(question.Options) {
			labels = append(labels, question.Options[pick-1].Label)
		} else if pick == len(question.Options)+1 {
			labels = append(labels, reply.Other)
		}
	}
	return strings.Join(labels, ", ")
}

// AnswerText is the prompt that brings the answer to the agent.
func AnswerText(record Record, replies []Reply, stopped bool) string {
	from := record.Asker
	if stopped {
		from += " (asker stopped)"
	}
	parts := make([]string, 0, len(record.Questions))
	for index, question := range record.Questions {
		if index >= len(replies) {
			break
		}
		part := question.Header + " → " + Labels(question, replies[index])
		if replies[index].Note != "" {
			part += " · note: " + replies[index].Note
		}
		parts = append(parts, part)
	}
	return fmt.Sprintf("%suser: answer to ask %s from %s · %s", marker.Prefix, record.ID, from, strings.Join(parts, "; "))
}

// DismissText is the prompt that tells the agent that no answer comes.
func DismissText(record Record) string {
	return fmt.Sprintf("%shive: the user dismissed ask %s from %s · %s. No answer comes.", marker.Prefix, record.ID, record.Asker, Headers(record))
}

// checkReplies checks that each question has a valid answer.
func checkReplies(record Record, replies []Reply) error {
	if len(replies) != len(record.Questions) {
		return fmt.Errorf("give %d answers, not %d", len(record.Questions), len(replies))
	}
	for index, question := range record.Questions {
		reply := replies[index]
		at := fmt.Sprintf("question %d: ", index+1)
		if len(reply.Picks) == 0 {
			return errors.New(at + "pick an option")
		}
		if !question.MultiSelect && len(reply.Picks) > 1 {
			return errors.New(at + "pick one option")
		}
		for _, pick := range reply.Picks {
			if pick < 1 || pick > len(question.Options)+1 {
				return fmt.Errorf("%soption %d does not exist", at, pick)
			}
			if pick == len(question.Options)+1 && strings.TrimSpace(reply.Other) == "" {
				return errors.New(at + "type the Other answer")
			}
		}
	}
	return nil
}

// Answer records the answer of the user and sends it to the agent. When the
// send fails, the record keeps the answer, and DeliverPending tries again.
func Answer(ctx context.Context, home, id string, replies []Reply) output.Envelope {
	return closeAsk(ctx, home, id, StateAnswered, replies)
}

// Dismiss records that the user dismissed the ask and tells the agent.
func Dismiss(ctx context.Context, home, id string) output.Envelope {
	return closeAsk(ctx, home, id, StateDismissed, nil)
}

func closeAsk(ctx context.Context, home, id, state string, replies []Reply) output.Envelope {
	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error(), nil)
	}
	var record Record
	var phase, message string
	lockErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
		if err := locked.ReadIn(collection, id, &record); err != nil {
			phase, message = "read", "read ask "+id+": "+err.Error()
			return nil
		}
		if record.State != StateOpen {
			phase, message = "validate", "ask "+id+" is not open: "+record.State
			return nil
		}
		if state == StateAnswered {
			if err := checkReplies(record, replies); err != nil {
				phase, message = "validate", err.Error()
				return nil
			}
			record.Replies = replies
		}
		record.State = state
		record.ClosedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := locked.WriteIn(collection, record.ID, record); err != nil {
			phase, message = "write", "write ask record: "+err.Error()
		}
		return nil
	})
	if lockErr != nil {
		return failed("store", lockErr.Error(), nil)
	}
	if message != "" {
		return failed(phase, message, nil)
	}
	logTransition(home, state, record, StateOpen, state)
	record, err = deliver(ctx, home, record.ID, false)
	if err != nil {
		return output.Envelope{Outcome: "partial", Phase: "delivery", Effects: []output.Effect{}, Error: "the answer waits for delivery: " + err.Error(), Result: record}
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: record}
}

// DeliverPending tries again to send each closed ask that hive did not
// deliver. As with other prompts, it sends only to an idle or done agent.
func DeliverPending(ctx context.Context, home string) error {
	listed, err := List(home)
	if err != nil {
		return err
	}
	for _, record := range listed {
		if record.State == StateOpen || record.DeliveredAt != "" {
			continue
		}
		_, _ = deliver(ctx, home, record.ID, true)
	}
	return nil
}

// deliver sends the prompt of one closed ask. The record keeps each try.
// When waitIdle is true, deliver sends only to an idle or done agent.
func deliver(ctx context.Context, home, id string, waitIdle bool) (Record, error) {
	registry, err := store.Open(home)
	if err != nil {
		return Record{}, err
	}
	var record Record
	var target records.Agent
	var stopped, skip bool
	var routeErr error
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		if err := locked.ReadIn(collection, id, &record); err != nil {
			return err
		}
		if record.State == StateOpen || record.DeliveredAt != "" {
			skip = true
			return nil
		}
		agents, err := listAgents(locked)
		if err != nil {
			return err
		}
		var ok bool
		target, stopped, ok = Route(record, agents)
		if !ok {
			routeErr = errors.New("no live asker or parent")
			skip = waitIdle
			return nil
		}
		if waitIdle && target.AgentStatus != "idle" && target.AgentStatus != "done" {
			skip = true
		}
		return nil
	})
	if err != nil || skip {
		return record, err
	}
	text := DismissText(record)
	if record.State == StateAnswered {
		text = AnswerText(record, record.Replies, stopped)
	}
	sendErr := routeErr
	if sendErr == nil {
		sendErr = prompt(ctx, home, target.PaneID, text)
	}
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current Record
		if err := locked.ReadIn(collection, id, &current); err != nil {
			return err
		}
		current.Attempts++
		current.LastError = ""
		if sendErr != nil {
			current.LastError = sendErr.Error()
		} else {
			current.DeliveredAt = time.Now().UTC().Format(time.RFC3339Nano)
			current.DeliveredTo = target.Name
		}
		record = current
		return locked.WriteIn(collection, id, current)
	})
	logDelivery(home, record, target.Name, sendErr)
	if err != nil {
		return record, err
	}
	return record, sendErr
}

func listAgents(locked *store.LockedStore) ([]records.Agent, error) {
	raw, err := locked.ListIn([]string{"agents"}, nil)
	if err != nil {
		return nil, err
	}
	agents := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("decode agent record: %w", err)
		}
		agents = append(agents, record)
	}
	return agents, nil
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

// withLog runs fn with the ask log writer. A log that does not open writes
// nothing.
func withLog(home string, fn func(*slog.Logger)) {
	writer, err := hivelog.NewWriter(home, "ask", time.Now)
	if err != nil {
		return
	}
	defer func() { _ = writer.Close() }()
	fn(slog.New(writer))
}

func logTransition(home, command string, record Record, from, to string) {
	withLog(home, func(logger *slog.Logger) {
		fields := []any{"command", "ask " + command, "ask_id", record.ID, "agent_id", record.AskerKey, "to_status", to}
		if from != "" {
			fields = append(fields, "from_status", from)
		}
		logger.Debug("ask transition", fields...)
	})
}

func logDelivery(home string, record Record, target string, sendErr error) {
	withLog(home, func(logger *slog.Logger) {
		fields := []any{"command", "ask deliver", "ask_id", record.ID, "agent_id", record.AskerKey, "state", record.State, "target", target, "attempt", record.Attempts}
		if sendErr != nil {
			logger.Debug("ask delivery failed", append(fields, "error", sendErr.Error())...)
			return
		}
		logger.Debug("ask delivered", fields...)
	})
}

func failed(phase, message string, result any) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: result}
}
