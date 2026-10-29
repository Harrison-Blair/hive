package homeinit

import (
	"errors"
	"fmt"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/home"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Result is the result of a successful init: the home in the config.
type Result struct {
	Home string `json:"home"`
}

// FindHome returns the outermost marked home above cwd.
func FindHome(cwd string) (string, error) {
	markedHome, ok := home.FindMarkerAbove(cwd)
	if !ok {
		return "", fmt.Errorf("not a hive home: no marker above %s", cwd)
	}
	return markedHome, nil
}

// Initialize writes the home configuration for markedHome when there is no
// config. A config that names another home fails.
func Initialize(markedHome string) output.Envelope {
	configPath, err := home.ConfigPath()
	if err != nil {
		return output.Failed("command", err.Error())
	}
	configuredHome, err := home.ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return output.Failed("command", err.Error())
		}
		if err := home.WriteConfig(markedHome); err != nil {
			return output.Failed("command", err.Error())
		}
		effects := []output.Effect{{Resource: "config", Path: configPath, Action: "created"}}
		return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Result: Result{Home: markedHome}}
	}
	resolvedHome, err := home.Validate(configuredHome)
	if err != nil {
		return output.Failed("command", fmt.Sprintf("home config %s: %v", configPath, err))
	}
	if resolvedHome != markedHome {
		return output.Failed("command", fmt.Sprintf("home conflict: config is %s, working directory is in %s", resolvedHome, markedHome))
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: []output.Effect{}, Result: Result{Home: markedHome}}
}

// Text renders the envelope of Initialize for a person. A failure has no text.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok {
		return ""
	}
	if len(envelope.Effects) != 0 {
		return fmt.Sprintf("Initialized hive home at %s\n", result.Home)
	}
	return fmt.Sprintf("Hive home already initialized at %s\n", result.Home)
}
