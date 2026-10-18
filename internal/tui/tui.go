// Package tui is the hive terminal interface that bare `hive` opens.
package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/logs"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
	statusRows    = 2
)

var headerStyle = lipgloss.NewStyle().Bold(true)

type mode int

const (
	viewing mode = iota
	editing
)

// Model is the root Bubble Tea model.
type Model struct {
	home   string
	now    func() time.Time
	reader *logs.Reader
	logger *slog.Logger
	closer io.Closer

	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once

	events        []logs.Event
	eventCh       <-chan logs.Event
	followDone    chan struct{}
	followErr     error
	followStarted bool

	viewport   viewport.Model
	input      textinput.Model
	mode       mode
	level      string
	textFilter string
	follow     bool
	errText    string

	supervisor    string
	agentCount    int
	unread        int
	headerErr     string
	headerStopped bool
	inputEnded    <-chan struct{}

	width  int
	height int
}

// New builds a log view for one resolved home and one clock.
func New(home string, now func() time.Time) *Model {
	if now == nil {
		now = time.Now
	}
	m := &Model{
		home:   home,
		now:    now,
		level:  "debug",
		follow: true,
		width:  defaultWidth,
		height: defaultHeight,
	}
	m.reader = logs.NewReaderWithClock(home, now)
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.viewport = viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.bodyHeight()))
	m.viewport.SoftWrap = true
	m.input = textinput.New()
	m.input.Prompt = "filter: "
	m.input.CharLimit = 0
	m.input.Blur()

	writer, err := hivelog.NewWriter(home, "tui", now)
	if err != nil {
		m.setError("log: " + err.Error())
		m.refreshHeader()
		return m
	}
	m.closer = writer
	m.logger = slog.New(writer)

	m.refreshHeader()
	m.loadInitial()
	return m
}

func (m *Model) loadInitial() {
	events, offset, err := m.reader.Read(context.Background(), time.Time{})
	if err != nil {
		m.setError(err.Error())
		return
	}
	m.events = events
	m.rebuild()
	m.startFollow(offset)
}

func (m *Model) startFollow(offset logs.Offset) {
	m.followStarted = true
	events, errs := m.reader.Follow(m.ctx, offset)
	m.eventCh = events
	m.followDone = make(chan struct{})
	go func() {
		err := <-errs
		m.followErr = err
		close(m.followDone)
	}()
}

// Stop cancels the follower and waits for it to settle.
func (m *Model) Stop() {
	m.stopOnce.Do(func() {
		if m.cancel != nil {
			m.cancel()
		}
		if m.followDone != nil {
			select {
			case <-m.followDone:
			case <-time.After(5 * time.Second):
			}
		}
		if m.closer != nil {
			_ = m.closer.Close()
		}
	})
}

// Init starts the refresh tick and the follower error reader. The tick is the
// only consumer of the follower event channel, so events stay in file order.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.tick()}
	if m.followStarted {
		cmds = append(cmds, m.waitFollow())
	}
	return tea.Batch(cmds...)
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil
	case tickMsg:
		m.drainEvents()
		m.refreshHeader()
		if m.follow && m.errText == "" {
			m.viewport.GotoBottom()
		}
		if m.inputEnded != nil {
			select {
			case <-m.inputEnded:
				return m, tea.Quit
			default:
			}
		}
		return m, m.tick()
	case followErrMsg:
		if msg.err != nil {
			m.setError(msg.err.Error())
		}
		return m, nil
	case tea.KeyPressMsg:
		if m.mode == editing {
			return m, m.updateEditing(msg)
		}
		return m, m.updateViewing(msg)
	}
	return m, nil
}

// View renders the header, the body, and the status line.
func (m *Model) View() tea.View {
	parts := []string{headerStyle.Render(m.headerLine()), m.bodyView(), m.statusLine()}
	if m.mode == editing {
		parts = append(parts, m.input.View())
	}
	return tea.NewView(joinLines(parts))
}

func (m *Model) bodyHeight() int {
	height := m.height - statusRows
	if height < 1 {
		return 1
	}
	return height
}

func (m *Model) drainEvents() {
	for m.eventCh != nil {
		select {
		case event, ok := <-m.eventCh:
			if !ok {
				m.eventCh = nil
				return
			}
			m.appendEvent(event)
		default:
			return
		}
	}
}

func (m *Model) layout() {
	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(m.bodyHeight())
	if m.follow {
		m.viewport.GotoBottom()
	}
}

func (m *Model) logAction(action string) {
	if m.logger == nil {
		return
	}
	m.logger.Debug("tui action", "action", action)
}

func (m *Model) logError(context string, err error) {
	if m.logger == nil || err == nil {
		return
	}
	m.logger.Debug("tui error", "context", context, "error", boundedError(err.Error()))
}

func joinLines(parts []string) string {
	result := ""
	for index, part := range parts {
		if index != 0 {
			result += "\n"
		}
		result += part
	}
	return result
}

// Run starts the TUI on in and out and resolves the home itself.
func Run(in io.Reader, out io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	resolved, err := home.Resolve("", false, os.Getenv("HIVE_HOME"), cwd)
	if err != nil {
		return err
	}
	return RunHome(resolved, in, out)
}

// RunHome starts the TUI for one resolved home. It checks the input and the
// output at this boundary. A finite input end or a stream fault settles the
// program and the follower, because the installed renderer discards its own
// flush errors and an input end is not an error message.
func RunHome(home string, in io.Reader, out io.Writer) error {
	model := New(home, time.Now)
	defer model.Stop()

	inputEnded := make(chan struct{})
	var endOnce sync.Once
	model.inputEnded = inputEnded

	faults := make(chan struct{}, 1)
	signal := func() {
		select {
		case faults <- struct{}{}:
		default:
		}
	}
	inputGuard := &streamGuard{onFault: func(err error) {
		if errors.Is(err, io.EOF) {
			endOnce.Do(func() { close(inputEnded) })
			return
		}
		signal()
	}}
	outputGuard := &streamGuard{onFault: func(error) { signal() }}
	guardedIn := guardReader(in, inputGuard)
	guardedOut := guardWriter(out, outputGuard)

	program := tea.NewProgram(model,
		tea.WithInput(guardedIn),
		tea.WithOutput(guardedOut),
		tea.WithWindowSize(defaultWidth, defaultHeight),
	)
	watched := make(chan struct{})
	go func() {
		select {
		case <-faults:
			program.Kill()
		case <-watched:
		}
	}()

	_, runErr := program.Run()
	close(watched)

	if err := outputGuard.Err(); err != nil {
		return fmt.Errorf("tui: output: %w", err)
	}
	if err := inputGuard.Err(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("tui: input: %w", err)
	}
	if runErr != nil {
		return fmt.Errorf("tui: %w", runErr)
	}
	return nil
}

type streamGuard struct {
	mu      sync.Mutex
	err     error
	onFault func(error)
}

func (g *streamGuard) fault(err error) {
	if err == nil {
		return
	}
	g.mu.Lock()
	first := g.err == nil
	if first {
		g.err = err
	}
	g.mu.Unlock()
	if first && g.onFault != nil {
		g.onFault(err)
	}
}

func (g *streamGuard) Err() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.err
}

type guardedReader struct {
	in    io.Reader
	guard *streamGuard
}

func (r *guardedReader) Read(data []byte) (int, error) {
	n, err := r.in.Read(data)
	if err != nil {
		r.guard.fault(err)
	}
	return n, err
}

type guardedWriter struct {
	out   io.Writer
	guard *streamGuard
}

func (w *guardedWriter) Write(data []byte) (int, error) {
	n, err := w.out.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		w.guard.fault(err)
	}
	if n < 0 {
		n = 0
	}
	if n > len(data) {
		n = len(data)
	}
	return n, err
}

type fdHolder interface {
	Fd() uintptr
}

// guardReader and guardWriter keep a terminal file unwrapped, so the terminal
// detection and raw mode of Bubble Tea stay unchanged.
func guardReader(in io.Reader, guard *streamGuard) io.Reader {
	if _, ok := in.(fdHolder); ok {
		return in
	}
	return &guardedReader{in: in, guard: guard}
}

func guardWriter(out io.Writer, guard *streamGuard) io.Writer {
	if _, ok := out.(fdHolder); ok {
		return out
	}
	return &guardedWriter{out: out, guard: guard}
}
