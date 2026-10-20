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
	xterm "github.com/charmbracelet/x/term"
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
	logErr     string

	supervisor    string
	agentCount    int
	startingCount int
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

// logAction writes one action event through the checked logging path.
func (m *Model) logAction(action string) {
	m.recordLog("tui action", slog.String("action", action))
}

// logError writes one error event through the checked logging path.
func (m *Model) logError(context string, err error) {
	if err == nil {
		return
	}
	m.recordLog("tui error", slog.String("context", context), slog.String("error", boundedError(err.Error())))
}

// recordLog writes one event through the handler and keeps the first refusal in
// a bounded view. The writer checks the static log path on each new daily file.
// slog discards the handler error, so this path calls the handler and shows the
// error itself. It never logs the refusal, so it cannot recurse.
func (m *Model) recordLog(message string, attrs ...slog.Attr) {
	if m.logger == nil {
		return
	}
	record := slog.NewRecord(m.now(), slog.LevelDebug, message, 0)
	record.AddAttrs(attrs...)
	if err := m.logger.Handler().Handle(context.Background(), record); err != nil {
		if m.logErr == "" {
			m.logErr = boundedError(err.Error())
		}
	}
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
	resolved, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
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
			program.Quit()
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

type guardedOutput struct {
	out   io.Writer
	guard *streamGuard
}

func (w *guardedOutput) Write(data []byte) (int, error) {
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

// Read and Close keep the term.File shape for the installed renderer. The
// renderer uses Fd to detect a terminal and to get its size.
func (w *guardedOutput) Read(data []byte) (int, error) {
	if reader, ok := w.out.(io.Reader); ok {
		return reader.Read(data)
	}
	return 0, errors.New("tui: output is not readable")
}

func (w *guardedOutput) Close() error {
	if closer, ok := w.out.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (w *guardedOutput) Fd() uintptr {
	if file, ok := w.out.(fdHolder); ok {
		return file.Fd()
	}
	return ^uintptr(0)
}

type fdHolder interface {
	Fd() uintptr
}

// guardReader keeps a real terminal unwrapped, so the input raw mode of Bubble
// Tea stays unchanged. Any other reader gets the guard: Bubble Tea then uses
// its fallback reader, and the guard turns a finite end of input into a program
// quit. A raw Fd on a regular file fails in epoll, and a pipe end otherwise
// leaves the program active.
func guardReader(in io.Reader, guard *streamGuard) io.Reader {
	if file, ok := in.(fdHolder); ok && xterm.IsTerminal(file.Fd()) {
		return in
	}
	return &guardedReader{in: in, guard: guard}
}

func guardWriter(out io.Writer, guard *streamGuard) io.Writer {
	return &guardedOutput{out: out, guard: guard}
}
