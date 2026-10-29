// Package logs wires the logs command.
package logs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/home"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	loglib "github.com/Harrison-Blair/hive/internal/logs"
	"github.com/spf13/cobra"
)

type options struct {
	since     string
	level     string
	component string
	agent     string
	request   string
	limit     int
	follow    bool
}

// New creates the logs command.
func New() *cobra.Command {
	opts := &options{}
	command := &cobra.Command{
		Use:   "logs",
		Short: "read the daily log files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, opts)
		},
	}
	flags := command.Flags()
	flags.StringVar(&opts.since, "since", "", "duration or RFC 3339 cutoff")
	flags.StringVar(&opts.level, "level", "", "minimum level")
	flags.StringVar(&opts.component, "component", "", "component")
	flags.StringVar(&opts.agent, "agent", "", "agent name: matches agent_id or target")
	flags.StringVar(&opts.request, "request", "", "request ID")
	flags.IntVar(&opts.limit, "limit", 200, "maximum matches to print")
	flags.BoolVar(&opts.follow, "follow", false, "print new events as they arrive")
	return command
}

func run(cmd *cobra.Command, opts *options) (runErr error) {
	root := cmd.Root()
	jsonOutput, err := root.PersistentFlags().GetBool("json")
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(cmd, jsonOutput, err, 1)
	}
	resolvedHome, err := home.Resolve(os.Getenv("HIVE_HOME"), cwd)
	if err != nil {
		return fail(cmd, jsonOutput, err, 1)
	}
	requestID, err := hivelog.NewRequestID()
	if err != nil {
		return fail(cmd, jsonOutput, err, 1)
	}
	restoreRequestID := hivelog.SetRequestID(requestID)
	defer restoreRequestID()
	logger, err := common.Logger(resolvedHome, "")
	if err != nil {
		return fail(cmd, jsonOutput, err, 1)
	}
	started := time.Now()
	logger.Debug("command start", "command", "logs")
	// failure is the error text of a failed run for the command end event.
	var failure string
	defer func() {
		outcome := output.OutcomeOK
		if runErr != nil {
			var status *common.StatusError
			if !errors.As(runErr, &status) || status.Status != 0 {
				outcome = output.OutcomeFailed
			}
		}
		attrs := []any{"command", "logs", "duration_ms", time.Since(started).Milliseconds(), "outcome", outcome}
		if outcome != output.OutcomeOK && failure != "" {
			attrs = append(attrs, "error", failure)
		}
		logger.Debug("command end", attrs...)
	}()
	if opts.limit < 0 {
		failure = "--limit must be zero or more"
		return common.Usage(cmd, failure)
	}
	cutoff, err := parseSince(opts.since, time.Now())
	if err != nil {
		failure = err.Error()
		return common.Usage(cmd, failure)
	}
	filter := loglib.Filter{
		Level:     opts.level,
		Component: opts.component,
		AgentID:   opts.agent,
		RequestID: opts.request,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer stop()

	reader := loglib.NewReader(resolvedHome)
	events, offset, err := reader.Read(ctx, cutoff)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		failure = err.Error()
		return fail(cmd, jsonOutput, err, 1)
	}
	printer := &printer{
		out:    cmd.OutOrStdout(),
		errOut: cmd.ErrOrStderr(),
		json:   jsonOutput,
		filter: filter,
	}
	if err := printer.first(events, opts.limit); err != nil {
		failure = err.Error()
		return &common.StatusError{Status: 1}
	}
	if !opts.follow {
		return nil
	}

	eventCh, errCh := reader.Follow(ctx, offset)
	for event := range eventCh {
		if err := printer.one(event); err != nil {
			failure = err.Error()
			stop()
			return &common.StatusError{Status: 1}
		}
	}
	if err := <-errCh; err != nil {
		if ctx.Err() != nil {
			return nil
		}
		failure = err.Error()
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hive logs: %v\n", err)
		return &common.StatusError{Status: 1}
	}
	return nil
}

type printer struct {
	out    io.Writer
	errOut io.Writer
	json   bool
	filter loglib.Filter
}

// first prints the matches of the initial read. The limit applies to this
// print only. In JSON mode each malformed line writes a diagnostic and the
// limit counts only valid events.
func (p *printer) first(events []loglib.Event, limit int) error {
	if p.json {
		valid := make([]loglib.Event, 0, len(events))
		for _, event := range events {
			if !p.filter.Match(event) {
				continue
			}
			if event.Malformed() {
				if err := p.malformed(event); err != nil {
					return err
				}
				continue
			}
			valid = append(valid, event)
		}
		if limit < len(valid) {
			valid = valid[len(valid)-limit:]
		}
		for _, event := range valid {
			if err := p.valid(event); err != nil {
				return err
			}
		}
		return nil
	}

	matches := make([]loglib.Event, 0, len(events))
	for _, event := range events {
		if p.filter.Match(event) {
			matches = append(matches, event)
		}
	}
	if limit < len(matches) {
		matches = matches[len(matches)-limit:]
	}
	for _, event := range matches {
		if err := p.match(event); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) one(event loglib.Event) error {
	if !p.filter.Match(event) {
		return nil
	}
	return p.match(event)
}

func (p *printer) match(event loglib.Event) error {
	if event.Malformed() {
		if p.json {
			return p.malformed(event)
		}
		_, err := fmt.Fprintln(p.out, loglib.MalformedText(event))
		return err
	}
	if p.json {
		return p.valid(event)
	}
	_, err := fmt.Fprintln(p.out, loglib.TextLine(event))
	return err
}

func (p *printer) valid(event loglib.Event) error {
	_, err := fmt.Fprintln(p.out, string(event.Raw))
	return err
}

func (p *printer) malformed(event loglib.Event) error {
	_, err := fmt.Fprintf(p.errOut, "hive logs: malformed line %s:%d\n", event.Path(), event.Line())
	return err
}

func parseSince(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return now.Add(-duration), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --since value %q: use a duration or an RFC 3339 time", value)
	}
	return parsed, nil
}

func fail(cmd *cobra.Command, jsonOutput bool, err error, status int) error {
	prefix := "hive: "
	if jsonOutput {
		prefix = "hive logs: "
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s%v\n", prefix, err)
	return &common.StatusError{Status: status}
}
