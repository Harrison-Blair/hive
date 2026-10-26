// Package report sends the result of a general agent for the agent report command.
package report

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	identitylib "github.com/Harrison-Blair/hive/internal/lib/identity"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Options holds the inputs of one hive agent report.
type Options struct {
	// Agent is the value of --agent.
	Agent string
	// Pane is the value of --pane. It names the caller.
	Pane string
	// HerdrPane is HERDR_PANE_ID.
	HerdrPane string
	// Input gives the result text when FileSet is false.
	Input io.Reader
	// File is the value of --file.
	File    string
	FileSet bool
}

// Result is the inbox record of one report.
type Result struct {
	Inbox  string        `json:"inbox"`
	Record records.Inbox `json:"record"`
}

// Run sends the result of a general agent to the inbox of its parent
// (decisions 18 to 22). It writes the inbox record, then sets reported_at.
func Run(home string, options Options) output.Envelope {
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		return agentlib.Failed("log", err.Error())
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)
	logger.Debug("agent report request", "command", "agent report", "agent", options.Agent, "pane_id", options.Pane, "file", options.File)

	envelope := report(home, options, logger)
	attrs := []any{"command", "agent report", "outcome", envelope.Outcome, "phase", envelope.Phase}
	if result, ok := envelope.Result.(Result); ok {
		attrs = append(attrs, "agent_id", result.Record.Agent, "inbox", result.Inbox)
	}
	if envelope.Error != nil {
		attrs = append(attrs, "error", envelope.Error)
		logger.Error("agent report end", attrs...)
	} else {
		logger.Debug("agent report end", attrs...)
	}
	return envelope
}

func report(home string, options Options, logger *slog.Logger) output.Envelope {
	text, path, message := readReport(options)
	if message != "" {
		return agentlib.Failed("validate", message)
	}
	logger.Debug("agent report text", "command", "agent report", "path", path, hivelog.Body(text))
	registry, err := store.Open(home)
	if err != nil {
		return agentlib.Failed("store", err.Error())
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = reportHeld(locked, options, text, path)
		return nil
	})
	if err != nil {
		return agentlib.Failed("store", err.Error())
	}
	return envelope
}

// readReport reads the result text from the input (decision 19) or from
// --file (decision 20). It returns the text, the absolute file path, and a
// refusal message.
func readReport(options Options) (string, string, string) {
	if !options.FileSet {
		if options.Input == nil {
			return "", "", "report text is empty"
		}
		data, err := io.ReadAll(io.LimitReader(options.Input, inboxlib.MaxTextBytes+1))
		if err != nil {
			return "", "", "read standard input: " + err.Error()
		}
		if len(data) > inboxlib.MaxTextBytes {
			return "", "", fmt.Sprintf("report text is longer than %d bytes: use --file", inboxlib.MaxTextBytes)
		}
		if !utf8.Valid(data) {
			return "", "", "report text is not valid UTF-8"
		}
		return string(data), "", ""
	}
	path, err := filepath.Abs(options.File)
	if err != nil {
		return "", "", "report file: " + err.Error()
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", "report file: " + err.Error()
	}
	if !info.Mode().IsRegular() {
		return "", "", "report file " + path + " is not a regular file"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", "report file: " + err.Error()
	}
	if !utf8.Valid(data) {
		return "", "", "report file " + path + " is not valid UTF-8"
	}
	cut := len(data)
	if cut > inboxlib.MaxTextBytes {
		cut = inboxlib.MaxTextBytes
		for cut > 0 && !utf8.RuneStart(data[cut]) {
			cut--
		}
	}
	return string(data[:cut]), path, ""
}

// reportHeld finds the caller, writes the inbox record of its parent, then
// sets reported_at (decision 21). When the second write fails, it lists the
// first write as an effect.
func reportHeld(locked agentlib.LockedWriter, options Options, text, path string) output.Envelope {
	caller, err := agentlib.Caller(locked, agentlib.CallerOptions{Agent: options.Agent, Pane: options.Pane, HerdrPane: options.HerdrPane})
	if err != nil {
		if errors.Is(err, identitylib.ErrCallerUnknown) {
			return agentlib.Failed("identity", "caller unknown: pass --agent")
		}
		return agentlib.Failed("identity", err.Error())
	}
	if caller.Parent == "" {
		return agentlib.Failed("validate", "agent report needs a parent: "+agentlib.DisplayValue(caller.Name)+" has none")
	}
	found, err := agentlib.LoadAgents(locked)
	if err != nil {
		return agentlib.Failed("store", err.Error())
	}
	agents := make(map[string]records.Agent, len(found))
	for _, record := range found {
		agents[record.Key] = record
	}
	key := agentlib.ParentInbox(agents, caller)
	appended := inboxlib.AppendHeld(locked, inboxlib.AppendOptions{Inbox: key, Kind: records.InboxReport, Agent: caller.Name, Text: text, Path: path})
	if appended.Outcome != "ok" {
		return appended
	}
	inbox, _ := appended.Result.(records.Inbox)
	result := Result{Inbox: key, Record: inbox}
	effects := appended.Effects
	now := time.Now().UTC().Format(time.RFC3339Nano)
	caller.ReportedAt = now
	caller.UpdatedAt = now
	if err := locked.WriteIn([]string{"agents"}, caller.Key, caller); err != nil {
		return output.Envelope{Outcome: "partial", Phase: "agent", Effects: effects, Error: "set reported_at: " + err.Error(), Result: result}
	}
	effects = append(effects, output.Effect{Resource: "agent", Path: agentlib.StatePath("agents", caller.Key), Action: "reported"})
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: result}
}

// Text renders the inbox of one report.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok || envelope.Error != nil {
		return ""
	}
	return fmt.Sprintf("reported to %s as record %d\n", result.Inbox, result.Record.Seq)
}
