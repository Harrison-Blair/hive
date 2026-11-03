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
	"github.com/charmbracelet/x/ansi"
)

// Harnesses are the rows of step 1, in order.
var Harnesses = []string{"claude", "codex", "pi"}

const (
	otherRow     = "Other…"
	lastMark     = "✓ last"
	spinnerDelay = 100 * time.Millisecond
	boxWidth     = 62
	boxHeight    = 15
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
	harness := Harnesses[p.harness]
	title := []span{{"hive", style.Accent}, {harness, style.Harness(harness)}, {"model", ""}}
	keys := "[model] enter start · esc back"
	switch p.step {
	case stepHarness:
		title = []span{{"hive", style.Accent}, {"harness", ""}}
		keys = "[harness] ↑↓ move · enter next · esc quit"
		p.drawHarnesses(c, w, h)
	case stepLoading:
		keys = "[model] esc back"
		y := min(2, h-2)
		put(c, 2, y, spinnerFrames[p.frame], style.Working, 1)
		put(c, 4, y, "listing models…", style.Dim, w-6)
	case stepModels:
		keys = "[model] ↑↓ move · type filter · enter start · esc back"
		p.drawModels(c, w, h)
	case stepFailure:
		// In a short box the input row moves up over the cause.
		y := min(5, h-2)
		if y > 2 {
			put(c, 2, 2, "! model list unavailable", style.Error, w-4)
		}
		if y > 3 {
			put(c, 4, 3, p.failure, style.Dim, w-6)
		}
		drawInput(c, y, 2, "model", p.typed, w-2)
	case stepOther:
		drawInput(c, min(2, h-2), 2, "model", p.typed, w-2)
	}
	// The content stays inside the box: the edges overwrite it, and the
	// key row is cleared before the key line.
	c.Cells[h] = make([]canvas.Cell, w)
	if h > 0 {
		c.Box(0, 0, w, h, []rune("╭╮╰╯─│"), style.Border)
		drawTitle(c, title)
	}
	put(c, 1, h, keys, style.Dim, w-1)
	lines := c.Lines()
	if p.width > 0 && p.height > 0 {
		lines = fit(lines, p.width, p.height)
	}
	return tea.NewView(strings.Join(lines, "\n"))
}

// fit is the last guard of View: it cuts each line to width cells and drops
// the lines below height.
func fit(lines []string, width, height int) []string {
	lines = lines[:min(len(lines), height)]
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return lines
}

// size gives the box size: boxWidth × boxHeight, cut to fit the pane with one
// row left for the shell prompt and one for the key line. The height is 0,
// for the key line alone, when the pane has no room for a box.
func (p Picker) size() (int, int) {
	w, h := boxWidth, boxHeight
	if p.width > 0 {
		w = min(w, p.width)
	}
	if p.height > 0 {
		h = min(h, p.height-2)
	}
	if w < 2 || h < 2 {
		h = 0
	}
	return w, h
}

// drawHarnesses draws one row per harness with the last model of each. In a
// short box the rows start on the first inner row and scroll to the cursor.
func (p Picker) drawHarnesses(c *canvas.Canvas, w, h int) {
	top := 2
	if h-1-top < 1 {
		top = 1
	}
	rows := h - 1 - top
	first := max(0, min(p.harness-rows/2, len(Harnesses)-rows))
	for i := first; i < len(Harnesses) && i < first+rows; i++ {
		name := Harnesses[i]
		y := top + i - first
		nameStyle := style.Harness(name)
		if i == p.harness {
			nameStyle = bold(nameStyle)
		}
		model := p.lastModel(name)
		if model == "" {
			model = "–"
		}
		drawRow(c, y, w, listRow{
			cursor: i == p.harness, name: name, nameStyle: nameStyle, field: 8,
			last: p.hasLast && name == p.last.Harness, second: model,
		})
	}
}

// drawModels draws the filter row, a rule, and the rows of the list that fit
// the box around the cursor.
func (p Picker) drawModels(c *canvas.Canvas, w, h int) {
	rows := p.visible()
	count := fmt.Sprintf("%d of %d", len(rows)-1, len(p.models))
	// In a short box the list keeps one row: the rule goes first, then the
	// filter row.
	top := 3
	for top > 1 && h-1-top < 1 {
		top--
	}
	if top >= 2 {
		drawFilter(c, 1, w, p.filter, count)
	}
	for x := 1; top >= 3 && x < w-1; x++ {
		c.Set(x, 2, '─', style.Border)
	}
	nameWidth := 0
	for _, model := range rows {
		nameWidth = max(nameWidth, ansi.StringWidth(model))
	}
	lastModel := p.lastModel(Harnesses[p.harness])
	list := h - 1 - top
	first := max(0, min(p.cursor-list/2, len(rows)-list))
	for i := first; i < len(rows) && i < first+list; i++ {
		y := top + i - first
		nameStyle := ""
		if rows[i] == otherRow {
			nameStyle = style.Dim
		}
		if i == p.cursor {
			nameStyle = bold(nameStyle)
			if rows[i] != otherRow {
				nameStyle = style.Selected
			}
		}
		drawRow(c, y, w, listRow{
			cursor: i == p.cursor, name: rows[i], nameStyle: nameStyle, field: nameWidth, gap: 2,
			last: lastModel != "" && rows[i] == lastModel,
		})
	}
}

// listRow is one row of the harness or the model list.
type listRow struct {
	cursor          bool
	name, nameStyle string
	// field is the width of the name column, and gap the cells between it
	// and the ✓ last column, so the marks of each row line up.
	field, gap int
	last       bool
	// second is the dim column after the ✓ last column, or "".
	second string
}

// drawRow draws one list row in a box of width w. The ▸ cursor mark never
// drops. When the row is too narrow, it drops the second column first, then
// ✓ last, and then cuts the name with …, keeping at least its first cell.
func drawRow(c *canvas.Canvas, y, w int, r listRow) {
	if r.cursor {
		put(c, 2, y, "▸", style.Selected, 1)
	}
	room := w - 2 - 4 // the cells from the name to the right margin
	// The name column narrows to keep the ✓ last column while the name fits.
	field := min(r.field, room-r.gap-ansi.StringWidth(lastMark))
	if ansi.StringWidth(r.name) > field {
		drawName(c, y, r.name, r.nameStyle, room)
		return
	}
	put(c, 4, y, r.name, r.nameStyle, field)
	lastX := 4 + field + r.gap
	if r.last {
		put(c, lastX, y, lastMark, style.Dim, ansi.StringWidth(lastMark))
	}
	secondX := lastX + ansi.StringWidth(lastMark) + 3
	if limit := w - 2 - secondX; r.second != "" && limit >= 2 {
		put(c, secondX, y, r.second, style.Dim, limit)
	}
}

// drawName cuts name to limit cells with …, and keeps its first cell when
// there is room for less than the first cell and the … mark.
func drawName(c *canvas.Canvas, y int, name, nameStyle string, limit int) {
	if limit < 2 {
		c.Text(4, y, name, nameStyle, 1)
		return
	}
	put(c, 4, y, name, nameStyle, limit)
}

// drawInput draws label at x, then text and the ▏ cursor, ending before end.
func drawInput(c *canvas.Canvas, y, x int, label, text string, end int) {
	put(c, x, y, label, style.Dim, end-x)
	x += ansi.StringWidth(label) + 1
	put(c, x, y, field(text, end-x), "", end-x)
}

// drawFilter draws the filter row in a box of width w: the filter label, the
// typed text with the ▏ cursor, and the count at the right. When the row is
// too narrow, it drops the count first, then the label, and then cuts the
// start of the text. The ▏ cursor always stays.
func drawFilter(c *canvas.Canvas, y, w int, text, count string) {
	const label = "filter"
	end := w - 2
	room := end - 2
	input := ansi.StringWidth(text) + 1
	labelWidth := ansi.StringWidth(label) + 1
	countWidth := ansi.StringWidth(count)
	if labelWidth+input > room {
		put(c, 2, y, field(text, room), "", room)
		return
	}
	put(c, 2, y, label, style.Dim, labelWidth)
	put(c, 2+labelWidth, y, text+"▏", "", input)
	if labelWidth+input+1+countWidth <= room {
		put(c, end-countWidth, y, count, style.Dim, countWidth)
	}
}

// field gives text and the ▏ cursor in room cells. A text that does not fit
// keeps its end, after a … mark.
func field(text string, room int) string {
	room-- // the ▏ cursor
	if room < 0 {
		return ""
	}
	if width := ansi.StringWidth(text); width > room {
		if room == 0 {
			return "▏"
		}
		// A wide character on the cut stays whole, so cut one more cell
		// until the end fits after the … mark.
		tail := text
		for cut := width - room + 1; ansi.StringWidth(tail) > room-1; cut++ {
			tail = ansi.TruncateLeft(text, cut, "")
		}
		text = "…" + tail
	}
	return text + "▏"
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
		x += ansi.StringWidth(s.text)
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
