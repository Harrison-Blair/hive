// Package tui is the hive terminal interface that bare `hive` opens.
package tui

import (
	"io"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/logs"
)

// Model is the root Bubble Tea model.
type Model struct {
	events   chan logs.Event
	viewport viewport.Model
}

// New builds a log view for one home.
func New(home string, now func() time.Time) *Model {
	_ = home
	_ = now
	return &Model{events: make(chan logs.Event)}
}

// Stop releases the follower.
func (m *Model) Stop() {}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if press, ok := msg.(tea.KeyPressMsg); ok {
		switch press.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) View() tea.View {
	return tea.NewView("hive\n\npress q to quit\n")
}

func (m *Model) waitEvent() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.events
		return eventMsg{event: event, ok: ok}
	}
}

func (m *Model) waitFollow() tea.Cmd {
	return func() tea.Msg { return followErrMsg{} }
}

// Run starts the TUI on in and out and resolves the home itself.
func Run(in io.Reader, out io.Writer) error {
	_ = in
	_ = out
	return nil
}

// RunHome starts the TUI for one resolved home.
func RunHome(home string, in io.Reader, out io.Writer) error {
	_ = home
	_ = in
	_ = out
	return nil
}

type eventMsg struct {
	event logs.Event
	ok    bool
}

type followErrMsg struct {
	err error
}

type tickMsg struct{}
