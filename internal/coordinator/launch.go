package coordinator

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// Name is the agent name of the coordinator record.
const Name = "coordinator"

// Options holds the inputs of one bare hive run.
type Options struct {
	// Home is the resolved hive home.
	Home string
	// Dir is the working directory of the hive process. The harness runs in
	// it.
	Dir string
	// Pane is the herdr pane ID of this process (HERDR_PANE_ID).
	Pane string
	// Pick shows the picker with the last choice. ok is false when the user
	// quits.
	Pick func(last Choice, hasLast bool) (choice Choice, ok bool, err error)
	// Exec replaces this process with the harness, in dir and with the
	// environment env. ExecHarness is the production value.
	Exec func(dir, binary string, args, env []string) error
	// Guide is the hive guide. The harness gets it in its system prompt.
	Guide  string
	Now    func() time.Time
	Logger *slog.Logger
}

// Main runs bare hive with home, the picker on in and out, the pane of
// HERDR_PANE_ID, and the real exec. The harness runs in dir and gets
// guideText in its system prompt.
func Main(home, dir, guideText string, in io.Reader, out io.Writer) error {
	requestID, err := hivelog.NewRequestID()
	if err != nil {
		return err
	}
	defer hivelog.SetRequestID(requestID)()
	logger, err := hivelog.Open(home, "cli")
	if err != nil {
		return err
	}
	return Run(Options{
		Home:   home,
		Dir:    dir,
		Pane:   os.Getenv("HERDR_PANE_ID"),
		Pick:   NewPick(in, out, ExecRunner, logger.With("command", "hive", "agent_id", Name)),
		Exec:   ExecHarness,
		Guide:  guideText,
		Now:    time.Now,
		Logger: logger,
	})
}

// Run shows the picker, saves the choice and the coordinator record, labels
// the tab of the pane coordinator, and then replaces this process with the
// harness. A quit returns nil and writes
// nothing.
func Run(opts Options) error {
	logger := opts.Logger.With("command", "hive", "agent_id", Name)
	last, hasLast, err := LoadChoice(opts.Home)
	if err != nil {
		logger.Warn("last choice unreadable", "error", err.Error())
		hasLast = false
	}
	choice, ok, err := opts.Pick(last, hasLast)
	if err != nil {
		logger.Error("picker failed", "error", err.Error())
		return fmt.Errorf("picker: %w", err)
	}
	if !ok {
		logger.Info("picker quit")
		return nil
	}
	logger.Info("picker choice", "harness", choice.Harness, "model", choice.Model)
	profile, err := harness.Lookup(choice.Harness)
	if err != nil {
		return err
	}
	if err := SaveChoice(opts.Home, choice); err != nil {
		return err
	}
	record := agent.Record{
		Name:      Name,
		Pane:      opts.Pane,
		Harness:   choice.Harness,
		Model:     choice.Model,
		Cwd:       opts.Dir,
		StartedAt: opts.Now(),
	}
	if err := agent.Save(opts.Home, record); err != nil {
		logger.Error("coordinator record save failed", "error", err.Error())
		return err
	}
	logger.Info("coordinator record saved", "pane", record.Pane, "harness", record.Harness, "model", record.Model)
	if opts.Pane != "" {
		labelTab(opts.Pane, logger)
	}
	binary := profile.ProcessNames()[0]
	inbox, err := profile.InboxArgs(opts.Home)
	if err != nil {
		logger.Error("inbox failed", "harness", choice.Harness, "error", err.Error())
		return fmt.Errorf("message path of %s: %w", choice.Harness, err)
	}
	guideArgs := profile.GuideArgs(opts.Guide)
	args := profile.Args(choice.Model, append(inbox, guideArgs...))
	logger.Info("exec harness", "binary", binary, "args", guide.Redact(args, guideArgs), "cwd", opts.Dir)
	if err := opts.Exec(opts.Dir, binary, args, homeEnv(os.Environ(), opts.Home)); err != nil {
		logger.Error("exec harness failed", "binary", binary, "error", err.Error())
		return fmt.Errorf("start %s: %w", binary, err)
	}
	return nil
}

// labelTab renames the tab of pane to the coordinator name. A failure only
// logs a warning, because the harness can run without the label.
func labelTab(pane string, logger *slog.Logger) {
	ctx := context.Background()
	info, err := herdr.PaneGet(ctx, pane, herdr.WithLogger(logger))
	if err != nil {
		logger.Warn("coordinator tab label failed", "pane", pane, "error", err.Error())
		return
	}
	if err := herdr.TabRename(ctx, info.TabID, Name, herdr.WithLogger(logger)); err != nil {
		logger.Warn("coordinator tab label failed", "pane", pane, "tab", info.TabID, "error", err.Error())
		return
	}
	logger.Info("coordinator tab labeled", "pane", pane, "tab", info.TabID)
}

// homeEnv returns environ with HIVE_HOME set to home, so the hive commands of
// the harness use the same home when it runs outside the home.
func homeEnv(environ []string, home string) []string {
	env := make([]string, 0, len(environ)+1)
	for _, pair := range environ {
		if !strings.HasPrefix(pair, "HIVE_HOME=") {
			env = append(env, pair)
		}
	}
	return append(env, "HIVE_HOME="+home)
}

// ExecHarness changes to dir and replaces this process with binary and env.
// It returns only on an error.
func ExecHarness(dir, binary string, args, env []string) error {
	path, err := exec.LookPath(binary)
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	return syscall.Exec(path, append([]string{binary}, args...), env)
}
