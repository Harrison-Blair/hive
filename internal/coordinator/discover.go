package coordinator

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs one external command and returns its standard output.
type Runner func(name string, args ...string) ([]byte, error)

// ExecRunner runs the command with os/exec. It drops the standard error, where
// pi writes its pattern warnings.
func ExecRunner(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// Discover returns the model names of one harness:
//   - pi: each row of pi --list-models as provider/model.
//   - codex: the openai-codex rows of pi --list-models, with no provider.
//   - claude: the opencode/claude-* rows of opencode models, with no prefix.
//
// A failed command or an empty list is an error. The picker then takes a typed
// model.
func Discover(harness string, run Runner) ([]string, error) {
	var command []string
	var parse func([]byte) []string
	switch harness {
	case "pi":
		command, parse = []string{"pi", "--list-models"}, parsePi
	case "codex":
		command, parse = []string{"pi", "--list-models"}, parseCodex
	case "claude":
		command, parse = []string{"opencode", "models"}, parseOpencodeClaude
	default:
		return nil, fmt.Errorf("no model discovery for harness %q", harness)
	}
	out, err := run(command[0], command[1:]...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", strings.Join(command, " "), err)
	}
	models := parse(out)
	if len(models) == 0 {
		return nil, fmt.Errorf("%s: no models for %s", strings.Join(command, " "), harness)
	}
	return models, nil
}

// piRows returns the provider and model columns of each table row. It skips
// the header row and each line with fewer than two columns.
func piRows(out []byte) [][2]string {
	var rows [][2]string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || (fields[0] == "provider" && fields[1] == "model") {
			continue
		}
		rows = append(rows, [2]string{fields[0], fields[1]})
	}
	return rows
}

func parsePi(out []byte) []string {
	var models []string
	for _, row := range piRows(out) {
		models = append(models, row[0]+"/"+row[1])
	}
	return models
}

func parseCodex(out []byte) []string {
	var models []string
	for _, row := range piRows(out) {
		if row[0] == "openai-codex" {
			models = append(models, row[1])
		}
	}
	return models
}

func parseOpencodeClaude(out []byte) []string {
	var models []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "opencode/claude-") {
			models = append(models, strings.TrimPrefix(line, "opencode/"))
		}
	}
	return models
}
