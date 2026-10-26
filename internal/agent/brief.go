package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// firstPrompt returns the first prompt of one spawned agent (D3). It gives the
// facts of the record first, then the spawner text.
func firstPrompt(record records.Agent, task *records.Task, spawner string) string {
	return firstPromptWithInstructions(record, task, spawner, "", false)
}

// firstPromptWithInstructions builds the first prompt. When the harness takes
// the profile and the guide in the first prompt, that text goes before the
// facts.
func firstPromptWithInstructions(record records.Agent, task *records.Task, spawner, instructions string, instructionsInPrompt bool) string {
	facts := []string{"You are " + record.Name + "."}
	if record.Parent != "" {
		facts = append(facts, "Your parent is "+record.Parent+".")
	}
	if record.Project != "" {
		facts = append(facts, "Your project is "+record.Project+".")
	}
	if record.Task != "" {
		facts = append(facts, "Your task is "+record.Task+".")
	}
	if record.Worktree != "" {
		facts = append(facts, "Your worktree is "+record.Worktree+".")
	}
	if task != nil {
		brief, evidence := taskPromptPaths(*task)
		if brief != "" {
			facts = append(facts, "Your brief is "+brief+".")
		}
		if evidence != "" {
			facts = append(facts, "Your evidence is "+evidence+".")
		}
	}
	body := strings.Join(facts, " ")
	if instructionsInPrompt && instructions != "" {
		body = instructions + "\n\n" + body
	}
	text := marker.Prefix + body
	if strings.TrimSpace(spawner) != "" {
		text += "\n\n" + spawner
	}
	return text
}

// instructionsText returns the text that an agent reads before its first turn:
// the profile, then the guide, then the profile path and the guide command
// (D4). It returns the empty string when the home has no profile file.
func instructionsText(home, profile string) (string, string) {
	if profile == "" {
		return "", ""
	}
	data, err := os.ReadFile(filepath.Join(home, "profiles", profile+".md"))
	if errors.Is(err, os.ErrNotExist) {
		return "", ""
	}
	if err != nil {
		return "", "profile: " + err.Error()
	}
	parts := []string{strings.TrimSpace(string(data)), strings.TrimSpace(guide.Text())}
	parts = append(parts, fmt.Sprintf("Read `profiles/%s.md` and run `hive guide` after a context reset.", profile))
	return strings.Join(parts, "\n\n"), ""
}

// instructionFile is one private file that holds the instructions of one
// start attempt. The harness reads it while it starts (D4).
type instructionFile struct {
	home string
	path string
}

// prepareInstructions writes the instructions to a new private file in the
// .hive/tmp/ folder of home and returns the start arguments that name it. It
// returns no file when the harness takes the instructions in the first
// prompt. The file is closed when the function returns.
func prepareInstructions(home string, profile harness.Profile, instructions string) ([]string, *instructionFile, error) {
	flag := profile.SystemPromptArg()
	if flag == "" || instructions == "" {
		return nil, nil, nil
	}
	absolute, err := filepath.Abs(home)
	if err != nil {
		return nil, nil, fmt.Errorf("instruction file: %w", err)
	}
	directory := filepath.Join(absolute, ".hive", "tmp")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, nil, fmt.Errorf("instruction file: %w", err)
	}
	file, err := os.CreateTemp(directory, "instructions-*.md")
	if err != nil {
		return nil, nil, fmt.Errorf("instruction file: %w", err)
	}
	owned := &instructionFile{home: absolute, path: file.Name()}
	_, writeErr := file.WriteString(instructions)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = owned.remove()
		return nil, nil, fmt.Errorf("instruction file: %w", err)
	}
	return []string{flag, owned.path}, owned, nil
}

// remove deletes the file of this attempt only.
func (f *instructionFile) remove() error {
	if err := os.Remove(f.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// keptEffect names the kept file relative to the home.
func (f *instructionFile) keptEffect() output.Effect {
	path := f.path
	if relative, err := filepath.Rel(f.home, f.path); err == nil {
		path = filepath.ToSlash(relative)
	}
	return output.Effect{Resource: "instructions", Path: path, Action: "kept"}
}

// taskPromptPaths gives the brief and evidence paths of one task. The evidence
// path is empty for a task with no feature, because its brief names it.
func taskPromptPaths(task records.Task) (string, string) {
	if task.Feature == "" {
		return ".hive/state/briefs/" + task.ID + ".md", ""
	}
	root := ".hive/spec/" + task.Feature + "/tasks/" + task.ID
	return root + "/brief.md", root + "/evidence/"
}

// shellQuote wraps one value in single quotes for the POSIX shell. Each single
// quote in the value is replaced by the four characters '\”.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
