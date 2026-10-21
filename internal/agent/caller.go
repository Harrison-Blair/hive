package agent

import (
	"errors"
	"fmt"

	identitylib "github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// CallerOptions holds the identity inputs of one hive command.
type CallerOptions struct {
	// Agent is the value of --agent.
	Agent string
	// Pane is the value of --pane when --pane names the caller.
	Pane string
	// HerdrPane is HERDR_PANE_ID. An empty value means the command runs
	// outside herdr.
	HerdrPane string
}

// Caller returns the caller of one hive command (decision 24). In a herdr
// pane, the agent of that pane is the caller: an --agent or a --pane that names
// another agent or pane fails. Outside herdr, --agent and --pane resolve the
// caller as before.
func Caller(locked identitylib.Reader, options CallerOptions) (records.Agent, error) {
	if options.HerdrPane == "" {
		return identitylib.ResolveHeld(locked, options.Agent, options.Pane)
	}
	if options.Pane != "" && options.Pane != options.HerdrPane {
		return records.Agent{}, fmt.Errorf("--pane %s is not the caller pane %s", options.Pane, options.HerdrPane)
	}
	caller, err := identitylib.ResolveHeld(locked, "", options.HerdrPane)
	if err != nil {
		if options.Agent != "" && errors.Is(err, identitylib.ErrCallerUnknown) {
			return records.Agent{}, fmt.Errorf("--agent %s is not the caller: pane %s has no live agent", options.Agent, options.HerdrPane)
		}
		return records.Agent{}, err
	}
	if options.Agent != "" && options.Agent != caller.Name && options.Agent != caller.Key {
		return records.Agent{}, fmt.Errorf("--agent %s is not the caller %s", options.Agent, caller.Name)
	}
	return caller, nil
}
