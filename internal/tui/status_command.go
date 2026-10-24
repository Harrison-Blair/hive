package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// The status command of golden-frames decision 12 runs at most once each
// statusCommandInterval and stops after statusCommandTimeout.
const (
	statusCommandInterval = 300 * time.Millisecond
	statusCommandTimeout  = 2 * time.Second
)

// statusCommandMsg is the result of one run of the status command.
type statusCommandMsg struct {
	line     string
	err      error
	duration time.Duration
}

// statusCommandState is the state of the coordinator or the supervisor in
// the standard input of the status command.
type statusCommandState struct {
	State     string `json:"state"`
	ChangedAt string `json:"changed_at"`
}

// statusCommandInput is the JSON object of decision 12.
type statusCommandInput struct {
	Home        string             `json:"home"`
	Screen      string             `json:"screen"`
	Coordinator statusCommandState `json:"coordinator"`
	Supervisor  statusCommandState `json:"supervisor"`
	Asks        int                `json:"asks"`
	NeedsYou    int                `json:"needs_you"`
	Project     string             `json:"project"`
	Clock       string             `json:"clock"`
}

// startStatusCommand runs the status command off the input loop when a
// status row has the command segment (decisions 12 and 13). It runs no
// process in demo mode, while a run is open, or within 300 ms of the last
// start.
func (m *Model) startStatusCommand() tea.Cmd {
	command := m.tuiConfig.StatusCommand
	used := slices.ContainsFunc(m.tuiConfig.StatusLines, func(names []string) bool { return slices.Contains(names, "command") })
	if command == "" || !used || m.demo != nil || m.statusCommandRunning {
		return nil
	}
	now := m.now()
	if !m.statusCommandAt.IsZero() && now.Sub(m.statusCommandAt) < statusCommandInterval {
		return nil
	}
	input, err := json.Marshal(m.statusCommandInput(now))
	if err != nil {
		return m.logError("status command", err)
	}
	m.statusCommandRunning, m.statusCommandAt = true, now
	dir := m.home
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), statusCommandTimeout)
		defer cancel()
		started := time.Now()
		process := exec.CommandContext(ctx, "sh", "-c", command)
		process.Dir = dir
		process.Stdin = bytes.NewReader(input)
		out, err := process.Output()
		first, _, _ := strings.Cut(string(out), "\n")
		return statusCommandMsg{line: first, err: err, duration: time.Since(started)}
	}
}

// updateStatusCommand shows the line of a run. A failure empties the segment
// and logs once until a run succeeds.
func (m *Model) updateStatusCommand(msg statusCommandMsg) tea.Cmd {
	m.statusCommandRunning = false
	line := ""
	if msg.err == nil {
		line = statusCommandText(msg.line)
	}
	if line != m.statusCommandLine {
		m.statusCommandLine = line
		m.invalidateView()
	}
	if msg.err == nil {
		m.statusCommandFailed = false
		return nil
	}
	if m.statusCommandFailed {
		return nil
	}
	m.statusCommandFailed = true
	return m.recordLogCmd("tui error", slog.String("context", "status command"),
		slog.String("command", m.tuiConfig.StatusCommand), slog.Int64("duration_ms", msg.duration.Milliseconds()),
		slog.String("error", boundedError(msg.err.Error())))
}

func (m *Model) statusCommandInput(now time.Time) statusCommandInput {
	coordinator := m.header.coordinator
	switch coordinator {
	case "working", "idle", "blocked":
	case "":
		coordinator = "none"
	default:
		coordinator = "starting"
	}
	supervisor := m.header.supervisor
	if m.header.supervisorOld {
		supervisor = "old"
	}
	return statusCommandInput{
		Home:        m.home,
		Screen:      strings.ToLower(screenRegistry[m.activeTab()].label),
		Coordinator: statusCommandState{State: coordinator, ChangedAt: m.header.coordinatorChangedAt},
		Supervisor:  statusCommandState{State: supervisor, ChangedAt: m.header.supervisorChangedAt},
		Asks:        m.header.asks,
		NeedsYou:    m.header.needsUser,
		Project:     m.selectedProjectID(),
		Clock:       now.UTC().Format(time.RFC3339),
	}
}

// statusEscape matches one escape sequence: a CSI sequence, an OSC string,
// or a two-byte escape.
var statusEscape = regexp.MustCompile(`\x1b(\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(\x07|\x1b\\)?|.?)`)

// statusCommandText keeps the SGR colors of line and drops each other escape
// sequence and control character, so the segment stays on its row.
func statusCommandText(line string) string {
	var out strings.Builder
	rest := line
	for rest != "" {
		loc := statusEscape.FindStringIndex(rest)
		plain := rest
		if loc != nil {
			plain = rest[:loc[0]]
		}
		for _, r := range plain {
			if r >= 0x20 && r != 0x7f {
				out.WriteRune(r)
			}
		}
		if loc == nil {
			break
		}
		if seq := rest[loc[0]:loc[1]]; strings.HasPrefix(seq, "\x1b[") && strings.HasSuffix(seq, "m") {
			out.WriteString(seq)
		}
		rest = rest[loc[1]:]
	}
	return out.String()
}
