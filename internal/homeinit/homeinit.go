package homeinit

import (
	"errors"
	"fmt"
	"os"

	"github.com/Harrison-Blair/hive/internal/lib/home"
)

// Result describes the home configuration created or found by init.
type Result struct {
	Home        string
	ConfigPath  string
	ConfigSaved bool
}

// Initialize writes the home configuration for the outermost marked home.
func Initialize(cwd string) (Result, error) {
	markedHome, ok := home.FindMarkerAbove(cwd)
	if !ok {
		return Result{}, fmt.Errorf("not a hive home: no marker above %s", cwd)
	}
	configPath, err := home.ConfigPath()
	if err != nil {
		return Result{}, err
	}
	configuredHome, err := home.ReadConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Result{}, err
		}
		if err := home.WriteConfig(markedHome); err != nil {
			return Result{}, err
		}
		return Result{Home: markedHome, ConfigPath: configPath, ConfigSaved: true}, nil
	}
	resolvedHome, err := home.Validate(configuredHome)
	if err != nil {
		return Result{}, fmt.Errorf("home config %s: %w", configPath, err)
	}
	if resolvedHome != markedHome {
		return Result{}, fmt.Errorf("home conflict: config is %s, working directory is in %s", resolvedHome, markedHome)
	}
	return Result{Home: markedHome, ConfigPath: configPath}, nil
}
