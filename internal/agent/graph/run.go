package graph

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Run shows the graph on in and out until the user quits. Each refresh reads
// the entries with the agent list operation.
func Run(home string, logger *slog.Logger, in io.Reader, out io.Writer) error {
	fetch := func(ctx context.Context) ([]list.Entry, error) {
		envelope := list.Run(ctx, home, list.Options{Logger: logger})
		if envelope.Outcome != output.OutcomeOK {
			return nil, fmt.Errorf("%s: %v", envelope.Phase, envelope.Error)
		}
		entries, ok := envelope.Result.([]list.Entry)
		if !ok {
			return nil, errors.New("agent list returned no entries")
		}
		return entries, nil
	}
	model := New(Options{Fetch: fetch, Now: time.Now})
	_, err := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out)).Run()
	return err
}
