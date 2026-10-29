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
)

// Harnesses are the rows of step 1, in order.
var Harnesses = []string{"claude", "codex", "pi"}

const (
	otherRow     = "Other…"
	lastMark     = "✓ last"
	maxErrorLen  = 48
	defaultRows  = 15
	spinnerDelay = 100 * time.Millisecond
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
	rows    int

	choice Choice
	done   bool
}

// NewPicker returns a picker on step 1 with the cursor on the last harness.
func NewPicker(last Choice, hasLast bool, discover func(string) ([]string, error)) Picker {
	p := Picker{last: last, hasLast: hasLast, discover: discover, rows: defaultRows}
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
		p.rows = max(3, msg.Height-6)
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
	if p.hasLast && p.last.Harness == msg.harness {
		if i := slices.Index(p.models, p.last.Model); i >= 0 {
			p.cursor = i
		}
	}
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

// View draws the box of the current step.
func (p Picker) View() tea.View {
	harness := Harnesses[p.harness]
	title := "hive · " + harness + " · model"
	var lines []string
	switch p.step {
	case stepHarness:
		title = "hive · harness"
		for i, name := range Harnesses {
			lines = append(lines, row(i == p.harness, name, p.hasLast && name == p.last.Harness, harnessWidth()))
		}
		lines = append(lines, "", "↑↓ move · enter next · esc quit")
	case stepLoading:
		lines = append(lines, spinnerFrames[p.frame]+" listing models…", "", "esc back")
	case stepModels:
		lines = p.modelLines(harness)
	case stepFailure:
		lines = append(lines, "! model list unavailable: "+clip(p.failure, maxErrorLen), "type a model name: "+p.typed+"▏", "", "enter start · esc back")
	case stepOther:
		lines = append(lines, "model: "+p.typed+"▏", "", "enter start · esc back")
	}
	return tea.NewView(box(title, lines))
}

func (p Picker) modelLines(harness string) []string {
	rows := p.visible()
	width := 0
	for _, model := range rows {
		width = max(width, utf8.RuneCountInString(model))
	}
	count := fmt.Sprintf("%d of %d", len(rows)-1, len(p.models))
	filter := p.filter + "▏"
	gap := max(2, width+4+len(lastMark)-utf8.RuneCountInString(filter)-len(count))
	lines := []string{filter + strings.Repeat(" ", gap) + count}
	first := max(0, min(p.cursor-p.rows/2, len(rows)-p.rows))
	for i := first; i < len(rows) && i < first+p.rows; i++ {
		isLast := p.hasLast && harness == p.last.Harness && rows[i] == p.last.Model
		lines = append(lines, row(i == p.cursor, rows[i], isLast, width))
	}
	return append(lines, "", "↑↓ move · type filter · enter start · esc back")
}

func harnessWidth() int {
	width := 0
	for _, name := range Harnesses {
		width = max(width, len(name))
	}
	return width + 10
}

// row draws one list row: ▸ marks the cursor and ✓ last marks the last choice.
func row(cursor bool, name string, last bool, width int) string {
	text := "  " + name
	if cursor {
		text = "▸ " + name
	}
	if last {
		text += strings.Repeat(" ", width-utf8.RuneCountInString(name)+1) + lastMark
	}
	return text
}

// box draws lines in a single-line border with the title in the top edge.
func box(title string, lines []string) string {
	width := utf8.RuneCountInString(title) + 2
	for _, line := range lines {
		width = max(width, utf8.RuneCountInString(line))
	}
	var b strings.Builder
	b.WriteString("┌ " + title + " " + strings.Repeat("─", width-utf8.RuneCountInString(title)) + "┐\n")
	for _, line := range lines {
		b.WriteString("│ " + line + strings.Repeat(" ", width-utf8.RuneCountInString(line)) + " │\n")
	}
	b.WriteString("└" + strings.Repeat("─", width+2) + "┘\n")
	return b.String()
}

func clip(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	return string([]rune(text)[:limit-1]) + "…"
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
