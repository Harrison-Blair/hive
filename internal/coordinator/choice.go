package coordinator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Choice is the harness and the model of one coordinator start.
type Choice struct {
	Harness string `json:"harness"`
	Model   string `json:"model"`
}

func choicePath(home string) string {
	return filepath.Join(home, "coordinator-choice.json")
}

// LoadChoice reads the last choice. ok is false when no choice is saved.
func LoadChoice(home string) (choice Choice, ok bool, err error) {
	data, err := os.ReadFile(choicePath(home))
	if errors.Is(err, os.ErrNotExist) {
		return Choice{}, false, nil
	}
	if err != nil {
		return Choice{}, false, err
	}
	if err := json.Unmarshal(data, &choice); err != nil {
		return Choice{}, false, fmt.Errorf("read %s: %w", choicePath(home), err)
	}
	return choice, true, nil
}

// SaveChoice writes the last choice atomically.
func SaveChoice(home string, choice Choice) error {
	data, err := json.Marshal(choice)
	if err != nil {
		return err
	}
	target := choicePath(home)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	temp := target + ".tmp"
	if err := os.WriteFile(temp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(temp, target)
}
