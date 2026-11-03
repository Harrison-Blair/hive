package coordinator

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/canvas"
	"github.com/Harrison-Blair/hive/internal/lib/style"
)

// Harnesses are the rows of step 1, in order.
var Harnesses = []string{"claude", "codex", "pi"}

const (
	otherRow     = "Other…"
	lastMark     = "✓ last"
	spinnerDelay = 100 * time.Millisecond
	boxWidth     = 62
	boxHeight    = 15
	minBoxWidth  = 24
	minBoxHeight = 7
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type step int

const (
	stepHarness step = iota
	stepLoading
	stepModels
	stepFailure
	stepOther
)

// modelsMsg carries the discovery result of one harness.
type modelsMsg struct {
	harness string
	models  []string
	err     error
}

// spinnerMsg advances the loading spinner.
type spinnerMsg struct{}

// Picker is the Bubble Tea model of the harness and model picker. The package
// doc describes its steps.
type Picker struct {
	last     Choice
	hasLast  bool
	discover func(string) ([]string, error)

	step    step
	harness int
	models  []string
	filter  string
	cursor  int
	typed   string
	failure string
	frame   int
	width   int
	height  int

	choice Choice
	done   bool
}

// NewPicker returns a picker on step 1 with the cursor on the last harness.
func NewPicker(last Choice, hasLast bool, discover func(string) ([]string, error)) Picker {
	p := Picker{last: last, hasLast: hasLast, discover: discover}
	if hasLast {
		if i := slices.Index(Harnesses, last.Harness); i >= 0 {
			p.harness = i
		}
	}
	return p
}

// Init starts no command.
func (p Picker) Init() tea.Cmd { return nil }

// Result returns the choice. ok is false until the user starts a model.
func (p Picker) Result() (Choice, bool) { return p.choice, p.done }

// Update applies one message.
func (p Picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
	case spinnerMsg:
		if p.step == stepLoading {
			p.frame = (p.frame + 1) % len(spinnerFrames)
			return p, spinnerTick()
		}
	case modelsMsg:
		if p.step == stepLoading && msg.harness == Harnesses[p.harness] {
			p.openModels(msg)
		}
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return p, tea.Quit
		}
		return p.key(msg)
	}
	return p, nil
}

func (p *Picker) openModels(msg modelsMsg) {
	p.filter, p.typed, p.cursor = "", "", 0
	if msg.err == nil && len(msg.models) == 0 {
		msg.err = fmt.Errorf("no models for %s", msg.harness)
	}
	if msg.err != nil {
		p.step, p.failure, p.models = stepFailure, msg.err.Error(), nil
		return
	}
	p.step, p.models = stepModels, msg.models
	if i := slices.Index(p.models, p.lastModel(msg.harness)); i >= 0 {
		p.cursor = i
	}
}

// lastModel gives the remembered model of harness, or "" when it has none.
func (p Picker) lastModel(harness string) string {
	if !p.hasLast {
		return ""
	}
	if model, ok := p.last.Models[harness]; ok {
		return model
	}
	if harness == p.last.Harness {
		return p.last.Model
	}
	return ""
}

func (p Picker) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch p.step {
	case stepHarness:
		switch msg.String() {
		case "esc":
			return p, tea.Quit
		case "up":
			p.harness = max(0, p.harness-1)
		case "down":
			p.harness = min(len(Harnesses)-1, p.harness+1)
		case "enter":
			p.step, p.frame = stepLoading, 0
			harness, discover := Harnesses[p.harness], p.discover
			return p, tea.Batch(spinnerTick(), func() tea.Msg {
				models, err := discover(harness)
				return modelsMsg{harness: harness, models: models, err: err}
			})
		}
	case stepLoading:
		if msg.String() == "esc" {
			p.step = stepHarness
		}
	case stepModels:
		rows := p.visible()
		switch msg.String() {
		case "esc":
			if p.filter != "" {
				p.filter, p.cursor = "", 0
			} else {
				p.step = stepHarness
			}
		case "up":
			p.cursor = max(0, p.cursor-1)
		case "down":
			p.cursor = min(len(rows)-1, p.cursor+1)
		case "backspace":
			p.filter, p.cursor = trimLast(p.filter), 0
		case "enter":
			if rows[p.cursor] == otherRow {
				p.step, p.typed = stepOther, ""
				return p, nil
			}
			return p.start(rows[p.cursor])
		default:
			if msg.Text != "" {
				p.filter, p.cursor = p.filter+msg.Text, 0
			}
		}
	case stepFailure, stepOther:
		switch msg.String() {
		case "esc":
			if p.step == stepOther {
				p.step = stepModels
			} else {
				p.step = stepHarness
			}
		case "backspace":
			p.typed = trimLast(p.typed)
		case "enter":
			if name := strings.TrimSpace(p.typed); name != "" {
				return p.start(name)
			}
		default:
			p.typed += msg.Text
		}
	}
	return p, nil
}

func (p Picker) start(model string) (tea.Model, tea.Cmd) {
	p.choice, p.done = Choice{Harness: Harnesses[p.harness], Model: model}, true
	return p, tea.Quit
}

// visible returns the model rows that match the filter, then the Other… row.
func (p Picker) visible() []string {
	needle := strings.ToLower(p.filter)
	var rows []string
	for _, model := range p.models {
		if strings.Contains(strings.ToLower(model), needle) {
			rows = append(rows, model)
		}
	}
	return append(rows, otherRow)
}

// View draws the box of the current step and the key line under it.
func (p Picker) View() tea.View {
	w, h := p.size()
	c := canvas.New(w, h+1)
	c.Box(0, 0, w, h, []rune("╭╮╰╯─│"), style.Border)
	harness := Harnesses[p.harness]
	title := []span{{"hive", style.Accent}, {harness, style.Harness(harness)}, {"model", ""}}
	keys := "[model] enter start · esc back"
	switch p.step {
	case stepHarness:
		title = []span{{"hive", style.Accent}, {"harness", ""}}
		keys = "[harness] ↑↓ move · enter next · esc quit"
		p.drawHarnesses(c, w)
	case stepLoading:
		keys = "[model] esc back"
		put(c, 2, 2, spinnerFrames[p.frame], style.Working, 1)
		put(c, 4, 2, "listing models…", style.Dim, w-6)
	case stepModels:
		keys = "[model] ↑↓ move · type filter · enter start · esc back"
		p.drawModels(c, w, h)
	case stepFailure:
		put(c, 2, 2, "! model list unavailable", style.Error, w-4)
		put(c, 4, 3, p.failure, style.Dim, w-6)
		drawInput(c, 5, 2, "model", p.typed, w-2)
	case stepOther:
		drawInput(c, 2, 2, "model", p.typed, w-2)
	}
	drawTitle(c, title)
	put(c, 1, h, keys, style.Dim, w-1)
	return tea.NewView(strings.Join(c.Lines(), "\n"))
}

// size gives the box size: boxWidth × boxHeight, cut to fit the pane with one
// row left for the shell prompt and one for the key line.
func (p Picker) size() (int, int) {
	w, h := boxWidth, boxHeight
	if p.width > 0 {
		w = max(minBoxWidth, min(w, p.width))
	}
	if p.height > 0 {
		h = max(minBoxHeight, min(h, p.height-2))
	}
	return w, h
}

// drawHarnesses draws one row per harness with the last model of each.
func (p Picker) drawHarnesses(c *canvas.Canvas, w int) {
	for i, name := range Harnesses {
		y := 2 + i
		nameStyle := style.Harness(name)
		if i == p.harness {
			put(c, 2, y, "▸", style.Selected, 1)
			nameStyle = bold(nameStyle)
		}
		put(c, 4, y, name, nameStyle, 8)
		if p.hasLast && name == p.last.Harness {
			put(c, 12, y, lastMark, style.Dim, 6)
		}
		model := p.lastModel(name)
		if model == "" {
			model = "–"
		}
		put(c, 21, y, model, style.Dim, w-23)
	}
}

// drawModels draws the filter row, a rule, and the rows of the list that fit
// the box around the cursor.
func (p Picker) drawModels(c *canvas.Canvas, w, h int) {
	rows := p.visible()
	count := fmt.Sprintf("%d of %d", len(rows)-1, len(p.models))
	countX := w - 2 - utf8.RuneCountInString(count)
	put(c, countX, 1, count, style.Dim, len(count))
	drawInput(c, 1, 2, "filter", p.filter, countX-1)
	for x := 1; x < w-1; x++ {
		c.Set(x, 2, '─', style.Border)
	}
	nameWidth := 0
	for _, model := range rows {
		nameWidth = max(nameWidth, utf8.RuneCountInString(model))
	}
	nameWidth = min(nameWidth, w-14)
	lastHarness := p.hasLast && Harnesses[p.harness] == p.last.Harness
	list := h - 4
	first := max(0, min(p.cursor-list/2, len(rows)-list))
	for i := first; i < len(rows) && i < first+list; i++ {
		y := 3 + i - first
		nameStyle := ""
		if rows[i] == otherRow {
			nameStyle = style.Dim
		}
		if i == p.cursor {
			put(c, 2, y, "▸", style.Selected, 1)
			nameStyle = bold(nameStyle)
			if rows[i] != otherRow {
				nameStyle = style.Selected
			}
		}
		put(c, 4, y, rows[i], nameStyle, nameWidth)
		if lastHarness && rows[i] == p.last.Model {
			put(c, 6+nameWidth, y, lastMark, style.Dim, 6)
		}
	}
}

// drawInput draws label at x, then text and the ▏ cursor, ending before end.
// A text that does not fit keeps its end, after a … mark.
func drawInput(c *canvas.Canvas, y, x int, label, text string, end int) {
	put(c, x, y, label, style.Dim, end-x)
	x += utf8.RuneCountInString(label) + 1
	room := end - x - 1
	if room <= 0 {
		return
	}
	runes := []rune(text)
	if len(runes) > room {
		runes = append([]rune("…"), runes[len(runes)-room+1:]...)
	}
	put(c, x, y, string(runes)+"▏", "", room+1)
}

// put writes s at x, y, cut to limit cells with … at the end.
func put(c *canvas.Canvas, x, y int, s, style string, limit int) {
	c.Text(x, y, canvas.Clip(s, limit), style, limit)
}

// span is one run of title text with its SGR style.
type span struct{ text, style string }

// drawTitle writes the spans into the top edge, joined by a dim " · ".
func drawTitle(c *canvas.Canvas, spans []span) {
	x, end := 3, c.W-2
	c.Set(2, 0, ' ', "")
	for i, s := range spans {
		if i > 0 {
			put(c, x, 0, " · ", style.Dim, end-x)
			x += 3
		}
		put(c, x, 0, s.text, s.style, end-x)
		x += utf8.RuneCountInString(s.text)
	}
	if x < end {
		c.Set(x, 0, ' ', "")
	}
}

// bold adds bold to an SGR style.
func bold(s string) string {
	if s == "" {
		return "1"
	}
	return "1;" + s
}

func trimLast(text string) string {
	if text == "" {
		return text
	}
	_, size := utf8.DecodeLastRuneInString(text)
	return text[:len(text)-size]
}

func spinnerTick() tea.Cmd {
	return tea.Tick(spinnerDelay, func(time.Time) tea.Msg { return spinnerMsg{} })
}

// logFailures wraps discover. It logs each discovery error in full with the
// harness name, because the picker shows the error clipped.
func logFailures(discover func(string) ([]string, error), logger *slog.Logger) func(string) ([]string, error) {
	return func(harness string) ([]string, error) {
		models, err := discover(harness)
		if err != nil {
			logger.Error("model discovery failed", "harness", harness, "error", err.Error())
		}
		return models, err
	}
}

// NewPick returns the Pick function of Options. It runs the picker on in and
// out with discovery through run, and logs each discovery error to logger.
func NewPick(in io.Reader, out io.Writer, run Runner, logger *slog.Logger) func(Choice, bool) (Choice, bool, error) {
	return func(last Choice, hasLast bool) (Choice, bool, error) {
		discover := logFailures(func(harness string) ([]string, error) { return Discover(harness, run) }, logger)
		final, err := tea.NewProgram(NewPicker(last, hasLast, discover), tea.WithInput(in), tea.WithOutput(out)).Run()
		if err != nil {
			return Choice{}, false, err
		}
		choice, ok := final.(Picker).Result()
		return choice, ok, nil
	}
}
