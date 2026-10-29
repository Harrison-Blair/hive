package coordinator

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// Name is the agent name of the coordinator record.
const Name = "coordinator"

// Options holds the inputs of one bare hive run.
type Options struct {
	// Home is the resolved hive home. The harness runs in it.
	Home string
	// Pane is the herdr pane ID of this process (HERDR_PANE_ID).
	Pane string
	// Pick shows the picker with the last choice. ok is false when the user
	// quits.
	Pick func(last Choice, hasLast bool) (choice Choice, ok bool, err error)
	// Exec replaces this process with the harness. ExecHarness is the
	// production value.
	Exec   func(dir, binary string, args []string) error
	Now    func() time.Time
	Logger *slog.Logger
}

// Main runs bare hive in home with the picker on in and out, the pane of
// HERDR_PANE_ID, and the real exec.
func Main(home string, in io.Reader, out io.Writer) error {
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
		Pane:   os.Getenv("HERDR_PANE_ID"),
		Pick:   NewPick(in, out, ExecRunner),
		Exec:   ExecHarness,
		Now:    time.Now,
		Logger: logger,
	})
}

// Run shows the picker, saves the choice and the coordinator record, and then
// replaces this process with the harness. A quit returns nil and writes
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
		Cwd:       opts.Home,
		StartedAt: opts.Now(),
	}
	if err := agent.Save(opts.Home, record); err != nil {
		logger.Error("coordinator record save failed", "error", err.Error())
		return err
	}
	logger.Info("coordinator record saved", "pane", record.Pane, "harness", record.Harness, "model", record.Model)
	binary := profile.ProcessNames()[0]
	args := profile.Args(choice.Model, nil)
	logger.Info("exec harness", "binary", binary, "args", args, "cwd", opts.Home)
	if err := opts.Exec(opts.Home, binary, args); err != nil {
		logger.Error("exec harness failed", "binary", binary, "error", err.Error())
		return fmt.Errorf("start %s: %w", binary, err)
	}
	return nil
}

// ExecHarness changes to dir and replaces this process with binary. It
// returns only on an error.
func ExecHarness(dir, binary string, args []string) error {
	path, err := exec.LookPath(binary)
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	return syscall.Exec(path, append([]string{binary}, args...), os.Environ())
}
