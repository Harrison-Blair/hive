package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/logs"
	xterm "github.com/charmbracelet/x/term"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
)

type mode int

const (
	viewing mode = iota
	editing
)

// Model is the root Bubble Tea model for the eight-screen shell.
type Model struct {
	home        string
	now         func() time.Time
	reader      *logs.Reader
	logger      *slog.Logger
	gatedLogger *slog.Logger
	closer      io.Closer
	logs        logGate

	events      []logs.Event
	filterDraft string
	mode        mode

	chatData          chatScreenData
	chatFeed          chatFeed
	chatInput         textarea.Model
	chatInputActive   bool
	chatSending       bool
	chatCallerCwd     string
	chatPaneID        string
	chatCallerBlocked bool
	chatNoise         string
	chatEventsOpen    bool
	chatEventOpen     map[uint64]bool
	chatStatusAge     string
	chatErr           string
	textFilter        string
	follow            bool
	logDate           time.Time
	logDatePinned     bool
	logSelection      int
	logSelectedKey    string
	logOffset         int
	logPendingTime    string
	logErr            string
	logFilterErr      string
	logRaw            bool

	header               headerState
	headerErr            string
	screenErr            string
	attachErr            string
	startupNotice        string
	startupErr           string
	headerLoader         func(string) (headerState, error)
	logLoader            func(*logs.Reader, time.Time) ([]logs.Event, error)
	healthLoader         func(string, *slog.Logger) (doctor.Result, error)
	agentsLoader         func(string, time.Time) (agentsData, error)
	paneReader           func(context.Context, string) (string, error)
	agents               agentsState
	graph                graphState
	health               healthState
	tasksLoader          func(string) (tasksData, error)
	tasks                tasksState
	frictionLoader       func(string) (frictionData, error)
	friction             frictionState
	chatLoader           func(string, time.Time) (chatScreenData, error)
	headerLoading        bool
	screenLoading        bool
	refreshSnapshot      refreshSnapshot
	refreshChecking      bool
	headerRefreshPending bool
	screenRefreshPending bool
	lastHealthRefresh    time.Time

	screenViewCache      string
	screenViewCacheValid bool
	screenViewWidth      int
	screenViewHeight     int
	screenViewIndex      int

	screens     []screen
	screenIndex int

	tab           int
	views         [tabCount]int
	regionFocus   [tabCount]int
	detailsOffset [tabCount]int
	helpOpen      bool
	helpOffset    int
	popupOpen     bool
	popupOffset   int
	quitNotice    bool
	tuiConfig     home.TUIConfig
	chordActions  map[string]string
	// statusCommandLine is the first output line of the status command.
	statusCommandLine string
	// statusCommandAt is the start of the last run of the status command.
	statusCommandAt      time.Time
	statusCommandRunning bool
	// statusCommandFailed is true after a failure was logged, until a run
	// succeeds.
	statusCommandFailed bool
	quitAt              time.Time

	inputEnded <-chan struct{}
	width      int
	height     int

	// The ask cards of Chat. askAnswer, askDismiss, and askDeliver write
	// the ask records and send the prompts.
	askCards      map[string]*askCard
	askFocusID    string
	askBusy       bool
	askDelivering bool
	askRows       []askRow
	askAnswer     func(context.Context, string, string, []ask.Reply) output.Envelope
	askDismiss    func(context.Context, string, string) output.Envelope
	askDeliver    func(context.Context, string) error

	demo *demoState
}

// New builds the TUI for one resolved home and one clock.
func New(home string, now func() time.Time) *Model {
	return newModel(home, now, nil)
}

// newModel builds the TUI. With demo state, the model reads only the fixture
// and checks no caller and no file.
func newModel(home string, now func() time.Time, demo *demoState) *Model {
	if now == nil {
		now = time.Now
	}
	m := &Model{
		home:         home,
		now:          now,
		textFilter:   defaultLogFilter,
		logDate:      startOfDay(now()),
		follow:       true,
		chatNoise:    chatNoiseNormal,
		chatFeed:     chatFeed{topItem: -1, follow: true},
		chatLoader:   loadChatScreenData,
		askAnswer:    ask.Answer,
		askDismiss:   ask.Dismiss,
		askDeliver:   ask.DeliverPending,
		width:        defaultWidth,
		height:       defaultHeight,
		header:       headerState{supervisor: "stale"},
		agents:       agentsState{},
		graph:        graphState{hideStopped: true},
		headerLoader: readHeader,
		logLoader: func(_ *logs.Reader, day time.Time) ([]logs.Event, error) {
			// The reader names the day file in the local zone, so it gets
			// noon of the same date there.
			date := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, time.Local)
			dayReader := logs.NewReaderWithClock(home, func() time.Time { return date })
			events, _, err := dayReader.Read(context.Background(), time.Time{})
			return events, err
		},
		agentsLoader:   loadAgents,
		healthLoader:   loadHealth,
		paneReader:     readAgentPane,
		tasksLoader:    loadTasks,
		frictionLoader: loadFriction,
		demo:           demo,
	}
	if demo != nil {
		m.useDemoLoaders(demo.fixture)
	} else if callerCwd, err := os.Getwd(); err != nil {
		m.chatCallerBlocked = true
		m.chatErr = boundedError("chat send is for the user only: rule 2 check failed: " + err.Error())
	} else {
		m.chatCallerCwd = callerCwd
		m.chatPaneID = os.Getenv("HERDR_PANE_ID")
		if err := chatlib.ValidateCaller(context.Background(), home, "", m.chatPaneID, callerCwd); err != nil {
			m.chatCallerBlocked = true
			m.chatErr = boundedError(err.Error())
		}
	}
	m.reader = logs.NewReaderWithClock(home, now)
	m.chatInput = newChatInput()
	m.setTUIConfig(defaultTUI())
	// Golden-frames decision 19: there is no normal mode, so the composer
	// has the focus at start. A blocked caller cannot type (rule 2).
	if !m.chatCallerBlocked {
		m.chatInputActive = true
		m.chatInput.Focus()
	}
	m.screens = m.makeScreens()

	logClock := now
	if demo != nil {
		// The fixture clock is for the screens. Log events keep the real time.
		logClock = time.Now
	}
	writer, err := hivelog.NewWriter(home, "tui", logClock)
	if err != nil {
		m.logErr = boundedError("log: " + err.Error())
	} else {
		m.closer = writer
		handler := &gatedHandler{gate: &m.logs, next: writer}
		m.logger = slog.New(writer)
		m.gatedLogger = slog.New(handler)
		if demo != nil {
			m.logger = slog.New(writer.WithAttrs([]slog.Attr{slog.Bool("demo", true)}))
			m.gatedLogger = slog.New(handler.WithAttrs([]slog.Attr{slog.Bool("demo", true)}))
		}
	}
	if demo != nil {
		m.applyDemoUI(demo.fixture.UI)
	} else {
		m.refreshSnapshot = snapshotRefreshState(home, m.logDate, m.tasks.data)
	}
	return m
}

// Stop closes the TUI log writer.
func (m *Model) Stop() {
	m.logs.close()
	if m.closer != nil {
		_ = m.closer.Close()
	}
}

var errLogClosed = errors.New("tui: log writer is closed")

// gatedHandler passes a log event to the writer until the program ends. A late
// event is dropped, so a running loader cannot keep the log file open.
type gatedHandler struct {
	gate *logGate
	next slog.Handler
}

func (h *gatedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *gatedHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.gate.begin() {
		return errLogClosed
	}
	defer h.gate.end()
	return h.next.Handle(ctx, record)
}

func (h *gatedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &gatedHandler{gate: h.gate, next: h.next.WithAttrs(attrs)}
}

func (h *gatedHandler) WithGroup(name string) slog.Handler {
	return &gatedHandler{gate: h.gate, next: h.next.WithGroup(name)}
}

// logGate stops a log write after the program ends. A command that started
// before the close finishes; a command that starts after it does not write.
type logGate struct {
	mu     sync.Mutex
	closed bool
	active sync.WaitGroup
}

func (g *logGate) begin() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false
	}
	g.active.Add(1)
	return true
}

func (g *logGate) end() {
	g.active.Done()
}

// close waits for each started log command, then refuses a new one.
func (g *logGate) close() {
	g.mu.Lock()
	g.closed = true
	g.mu.Unlock()
	g.active.Wait()
}

// trackLogCmd marks a command that can write a log event.
func (m *Model) trackLogCmd(command tea.Cmd) tea.Cmd {
	if command == nil {
		return nil
	}
	return func() tea.Msg {
		if !m.logs.begin() {
			return nil
		}
		defer m.logs.end()
		return command()
	}
}

// Init starts the first data loads and the refresh tick.
func (m *Model) Init() tea.Cmd {
	return batch(m.tick(), m.startHeaderLoad(), m.startCurrentScreenLoad(), m.startStatusCommand())
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.invalidateView()
		m.width = msg.Width
		m.height = msg.Height
		if m.demo != nil {
			m.demoResize(msg.Width, msg.Height)
		}
		return m, nil
	case tickMsg:
		if m.activeTab() == tabChat {
			age := coordinatorStatusAge(m.header.coordinatorChangedAt, m.now())
			if age != m.chatStatusAge {
				m.chatStatusAge = age
				m.invalidateView()
			}
		}
		if m.follow && !m.logDatePinned && !sameDay(m.logDate, m.now()) {
			m.logDate = startOfDay(m.now())
			m.events = nil
			m.logSelectedKey = ""
			m.logSelection = 0
			m.logOffset = 0
			m.invalidateView()
		}
		if m.inputEnded != nil {
			select {
			case <-m.inputEnded:
				return m, tea.Quit
			default:
			}
		}
		return m, batch(m.tick(), m.startRefreshCheck(), m.startAskDelivery(), m.startStatusCommand())
	case statusCommandMsg:
		return m, m.updateStatusCommand(msg)
	case refreshCheckMsg:
		m.refreshChecking = false
		headerChanged := msg.header != m.refreshSnapshot.header
		if headerChanged {
			m.refreshSnapshot.header = msg.header
		}
		screenChanged := msg.screen != m.refreshSnapshot.screens[msg.request]
		m.refreshSnapshot.screens[msg.request] = msg.screen
		cmds := make([]tea.Cmd, 0, 2)
		if headerChanged {
			if m.headerLoading {
				m.headerRefreshPending = true
			} else {
				cmds = append(cmds, m.startHeaderLoad())
			}
		}
		if msg.request == m.currentScreen().DataRequest() {
			healthDue := msg.request == healthDataRequest && (m.lastHealthRefresh.IsZero() || m.now().Sub(m.lastHealthRefresh) >= 4*time.Second)
			if screenChanged || healthDue {
				if m.screenLoading {
					m.screenRefreshPending = true
				} else {
					cmds = append(cmds, m.startCurrentScreenLoad())
				}
			}
		}
		return m, batch(cmds...)
	case headerLoadMsg:
		m.headerLoading = false
		if msg.err != nil {
			m.invalidateView()
			m.headerErr = boundedError(msg.err.Error())
			logCmd := m.logError("header", msg.err)
			if m.headerRefreshPending {
				m.headerRefreshPending = false
				return m, batch(logCmd, m.startHeaderLoad())
			}
			return m, logCmd
		}
		chatStatusChanged := msg.state.coordinator != m.header.coordinator || msg.state.coordinatorChangedAt != m.header.coordinatorChangedAt
		m.header = msg.state
		m.chatStatusAge = coordinatorStatusAge(msg.state.coordinatorChangedAt, m.now())
		if m.activeTab() == tabChat && chatStatusChanged {
			m.invalidateView()
		}
		m.headerErr = ""
		if m.headerRefreshPending {
			m.headerRefreshPending = false
			return m, m.startHeaderLoad()
		}
		return m, nil
	case screenLoadMsg:
		m.invalidateView()
		m.screenLoading = false
		if msg.request == healthDataRequest {
			// health.md rule 7: checking ends with the result or the error.
			// In a demo frame of the checking state, hive doctor never ends.
			m.health.checking = m.demo != nil && m.demo.fixture.UI.Checking
		}
		if msg.request != m.currentScreen().DataRequest() {
			return m, m.startCurrentScreenLoad()
		}
		if msg.err != nil {
			m.screenErr = boundedError(msg.err.Error())
			logCmd := m.logError("screen", msg.err)
			if m.screenRefreshPending {
				m.screenRefreshPending = false
				return m, batch(logCmd, m.startCurrentScreenLoad())
			}
			return m, logCmd
		}
		m.screenErr = ""
		switch msg.request {
		case agentsDataRequest:
			m.agents.accept(msg.agents)
			if tick := m.litTick(); tick != nil {
				return m, batch(tick, m.finishScreenRefresh())
			}
		case healthDataRequest:
			m.health.accept(msg.health)
			m.health.acceptAgents(msg.agents)
			m.health.checkedAt = m.now()
			m.lastHealthRefresh = m.now()
		case chatDataRequest:
			m.chatData = msg.chat
			m.rebuildChat()
		case tasksDataRequest:
			m.tasks.accept(msg.tasks)
		case frictionDataRequest:
			m.friction.accept(msg.friction)
		case logsDataRequest:
			if !msg.day.IsZero() && !sameDay(msg.day, m.logDate) {
				return m, m.startCurrentScreenLoad()
			}
			m.acceptLogEvents(msg.events)
		default:
			return m, nil
		}
		return m, m.finishScreenRefresh()
	case agentsLitMsg:
		m.graph.litAt = time.Time{}
		m.invalidateView()
		return m, m.litTick()
	case taskFileMsg:
		m.invalidateView()
		m.acceptTaskFile(msg)
		if msg.err != nil {
			return m, m.logError("task file read", msg.err)
		}
		return m, nil
	case chatSendMsg:
		m.invalidateView()
		return m, m.updateChatSend(msg)
	case askDoneMsg:
		return m, m.updateAskDone(msg)
	case askDeliveredMsg:
		m.askDelivering = false
		if msg.err != nil {
			return m, m.logError("ask delivery", msg.err)
		}
		return m, nil
	case quitNoticeExpiredMsg:
		if m.quitNotice && msg.at.Equal(m.quitAt) {
			m.quitNotice = false
			m.invalidateView()
		}
		return m, nil
	case coordinatorRestartCheckMsg:
		return m, m.acceptCoordinatorRestartCheck(msg)
	case agentRestartMsg:
		m.invalidateView()
		m.agents.restartRunning = false
		if msg.err != nil {
			m.agents.restartNotice = "restart failed: " + boundedError(msg.err.Error())
			return m, batch(
				m.logAction("agent restart result", slog.String("agent_id", msg.agentKey), slog.String("result", "failed"), slog.String("error", boundedError(msg.err.Error()))),
				m.logError("agent restart", msg.err),
			)
		}
		m.agents.restartNotice = "restarted " + msg.result.Tab.Tab.Label
		if msg.result.Warning != "" {
			m.agents.restartNotice += "; " + msg.result.Warning
		}
		return m, m.logAction("agent restart result", slog.String("agent_id", msg.agentKey), slog.String("result", "started"), slog.String("tab_id", msg.result.Tab.Tab.TabID), slog.String("warning", msg.result.Warning))
	case agentsPeekMsg:
		m.invalidateView()
		m.agents.peekOpen = true
		m.agents.peekPane = msg.paneID
		m.agents.peekText = msg.text
		if msg.err != nil {
			m.agents.peekText = "pane read: " + boundedError(msg.err.Error())
			return m, m.logError("pane read", msg.err)
		}
		return m, nil
	case attachMsg:
		m.invalidateView()
		if msg.err != nil {
			m.attachErr = boundedError(msg.err.Error())
			return m, m.logError("attach", msg.err)
		}
		m.attachErr = ""
		return m, nil
	case logWriteErrMsg:
		if msg.err != nil && m.logErr == "" {
			m.logErr = boundedError(msg.err.Error())
			m.invalidateView()
		}
		return m, nil
	case tea.PasteMsg:
		m.pasteLogFilter(msg.Content)
		return m, nil
	case tea.KeyPressMsg:
		m.invalidateView()
		return m, m.updateKey(msg)
	case tea.MouseClickMsg, tea.MouseWheelMsg:
		m.invalidateView()
		return m, m.updateMouse(msg.(tea.MouseMsg))
	}
	return m, nil
}

func (m *Model) updateKey(press tea.KeyPressMsg) tea.Cmd {
	if press.String() == "ctrl+c" {
		now := m.now()
		if m.quitNotice && !now.Before(m.quitAt) && now.Sub(m.quitAt) <= 3*time.Second {
			m.quitNotice = false
			return tea.Quit
		}
		m.quitNotice = true
		m.quitAt = now
		expires := tea.Tick(3*time.Second, func(time.Time) tea.Msg { return quitNoticeExpiredMsg{at: now} })
		return batch(m.logAction("quit request"), expires)
	}
	m.quitNotice = false
	if m.demo != nil {
		m.demo.notice = ""
	}
	if cmd, ok := m.updateChord(press); ok {
		return cmd
	}
	return m.updateFrameKey(press)
}

// View renders the frame of the current screen, or the too-small notice.
func (m *Model) View() tea.View {
	if pane, need, small := m.tooSmall(); small {
		view := tea.NewView(tooSmallView(pane, need))
		view.AltScreen = true
		return view
	}
	content := m.frameView()
	if m.demo != nil {
		content = clipView(content, m.width, m.height)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.Cursor = m.chatCursor(content)
	return view
}

// finishScreenRefresh starts the screen load that a refresh asked for while
// a load ran.
func (m *Model) finishScreenRefresh() tea.Cmd {
	if !m.screenRefreshPending {
		return nil
	}
	m.screenRefreshPending = false
	return m.startCurrentScreenLoad()
}

func (m *Model) invalidateView() {
	m.screenViewCacheValid = false
}

func (m *Model) startHeaderLoad() tea.Cmd {
	if m.headerLoading {
		return nil
	}
	m.headerLoading = true
	loader, home := m.headerLoader, m.home
	return func() tea.Msg {
		state, err := loader(home)
		return headerLoadMsg{state: state, err: err}
	}
}

func (m *Model) startCurrentScreenLoad() tea.Cmd {
	if m.screenLoading {
		return nil
	}
	request := m.currentScreen().DataRequest()
	m.screenLoading = true
	switch request {
	case agentsDataRequest:
		loader, home, now := m.agentsLoader, m.home, m.now
		return func() tea.Msg {
			data, err := loader(home, now())
			return screenLoadMsg{request: agentsDataRequest, agents: data, err: err}
		}
	case chatDataRequest:
		loader, home, now := m.chatLoader, m.home, m.now
		return func() tea.Msg {
			data, err := loader(home, now())
			return screenLoadMsg{request: chatDataRequest, chat: data, err: err}
		}
	case healthDataRequest:
		// health.md rule 4: the agents table is beside the checks.
		loader, agentsLoader, home, logger, now := m.healthLoader, m.agentsLoader, m.home, m.gatedLogger, m.now
		return func() tea.Msg {
			agents, err := agentsLoader(home, now())
			result, doctorErr := loader(home, logger)
			if doctorErr != nil {
				err = fmt.Errorf("hive doctor: %w", doctorErr)
			}
			return screenLoadMsg{request: healthDataRequest, health: result, agents: agents, err: err}
		}
	case tasksDataRequest:
		loader, home := m.tasksLoader, m.home
		return func() tea.Msg {
			data, err := loader(home)
			return screenLoadMsg{request: tasksDataRequest, tasks: data, err: err}
		}
	case frictionDataRequest:
		loader, home := m.frictionLoader, m.home
		return func() tea.Msg {
			data, err := loader(home)
			return screenLoadMsg{request: frictionDataRequest, friction: data, err: err}
		}
	default:
		loader, reader, day := m.logLoader, m.reader, startOfDay(m.logDate)
		return func() tea.Msg {
			events, err := loader(reader, day)
			if err == nil {
				events = eventsOnDay(events, day)
			}
			return screenLoadMsg{request: logsDataRequest, day: day, events: events, err: err}
		}
	}
}

func (m *Model) attachCmd(paneID string) tea.Cmd {
	if m.demo != nil {
		return m.demoBlocked("attach")
	}
	return func() tea.Msg {
		_, err := herdr.AgentFocus(context.Background(), paneID)
		return attachMsg{err: err}
	}
}

func (m *Model) startAgentsPeek(paneID string) tea.Cmd {
	if m.demo != nil {
		return m.demoBlocked("peek")
	}
	reader := m.paneReader
	return func() tea.Msg {
		text, err := reader(context.Background(), paneID)
		return agentsPeekMsg{paneID: paneID, text: text, err: err}
	}
}

func readAgentPane(ctx context.Context, paneID string) (string, error) {
	output, err := exec.CommandContext(ctx, "herdr", "pane", "read", paneID, "--lines", "20", "--format", "text").Output()
	if err != nil {
		return "", fmt.Errorf("herdr pane read %s: %w", paneID, err)
	}
	return string(output), nil
}

func (m *Model) logAction(action string, attrs ...slog.Attr) tea.Cmd {
	attrs = append(attrs, slog.String("action", action))
	return m.recordLogCmd("tui action", attrs...)
}

func (m *Model) logError(context string, err error) tea.Cmd {
	if err == nil {
		return nil
	}
	return m.recordLogCmd("tui error", slog.String("context", context), slog.String("error", boundedError(err.Error())))
}

func (m *Model) recordLogCmd(message string, attrs ...slog.Attr) tea.Cmd {
	if m.logger == nil {
		return nil
	}
	logger, now := m.logger, m.now
	return m.trackLogCmd(func() tea.Msg {
		record := slog.NewRecord(now(), slog.LevelDebug, message, 0)
		record.AddAttrs(attrs...)
		return logWriteErrMsg{err: logger.Handler().Handle(context.Background(), record)}
	})
}

func batch(cmds ...tea.Cmd) tea.Cmd {
	filtered := make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return tea.Batch(filtered...)
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
// program and the log writer.
func RunHome(home string, in io.Reader, out io.Writer) error {
	cfg, err := readTUISettings()
	if err != nil {
		return err
	}
	model := New(home, time.Now)
	model.setTUIConfig(cfg)
	defer model.Stop()
	notice, startupErr := setupSession(home, model.logger)
	model.startupNotice = notice
	if startupErr != nil {
		model.startupErr = boundedError(startupErr.Error())
	}
	return runProgram(model, in, out, defaultWidth, defaultHeight)
}

// runProgram runs the model on in and out until it quits.
func runProgram(model *Model, in io.Reader, out io.Writer, width, height int) error {
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
		tea.WithWindowSize(width, height),
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
	model.logs.close()

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
