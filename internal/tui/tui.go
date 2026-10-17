// Package tui is the hive terminal interface that bare `hive` opens.
package tui

import (
	"io"

	tea "charm.land/bubbletea/v2"
)

// Model is the root Bubble Tea model.
type Model struct{}

func (Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (Model) View() tea.View {
	return tea.NewView("hive\n\npress q to quit\n")
}

// Run starts the TUI on in and out and returns when it quits.
func Run(in io.Reader, out io.Writer) error {
	_, err := tea.NewProgram(Model{}, tea.WithInput(in), tea.WithOutput(out)).Run()
	return err
}
