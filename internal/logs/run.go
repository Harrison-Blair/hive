package logs

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Options are the flag values of hive logs.
type Options struct {
	Since     string
	Level     string
	Component string
	Agent     string
	Request   string
	Limit     int
	Follow    bool
	JSON      bool
}

// OptionError reports an invalid option value.
type OptionError struct {
	Message string
}

func (e *OptionError) Error() string { return e.Message }

// Run prints the matching events of the home and, with Follow, each new match
// until ctx ends. A read or follow error writes one hive: line to errOut and
// returns the error. A bad option returns an OptionError and writes nothing.
func Run(ctx context.Context, home string, opts Options, out, errOut io.Writer) error {
	if opts.Limit < 0 {
		return &OptionError{Message: "--limit must be zero or more"}
	}
	cutoff, err := parseSince(opts.Since, time.Now())
	if err != nil {
		return err
	}
	printer := &printer{
		out:    out,
		errOut: errOut,
		json:   opts.JSON,
		filter: Filter{
			Level:     opts.Level,
			Component: opts.Component,
			AgentID:   opts.Agent,
			RequestID: opts.Request,
		},
	}
	reader := NewReader(home)
	events, offset, err := reader.Read(ctx, cutoff)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		_, _ = fmt.Fprintf(errOut, "hive: %v\n", err)
		return err
	}
	if err := printer.first(events, opts.Limit); err != nil {
		return err
	}
	if !opts.Follow {
		return nil
	}

	ctx, stop := context.WithCancel(ctx)
	defer stop()
	eventCh, errCh := reader.Follow(ctx, offset)
	for event := range eventCh {
		if err := printer.one(event); err != nil {
			stop()
			return err
		}
	}
	if err := <-errCh; err != nil {
		if ctx.Err() != nil {
			return nil
		}
		_, _ = fmt.Fprintf(errOut, "hive: %v\n", err)
		return err
	}
	return nil
}

type printer struct {
	out    io.Writer
	errOut io.Writer
	json   bool
	filter Filter
}

// first prints the matches of the initial read. The limit applies to this
// print only. In JSON mode each malformed line writes a diagnostic and the
// limit counts only valid events.
func (p *printer) first(events []Event, limit int) error {
	if p.json {
		valid := make([]Event, 0, len(events))
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

	matches := make([]Event, 0, len(events))
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

func (p *printer) one(event Event) error {
	if !p.filter.Match(event) {
		return nil
	}
	return p.match(event)
}

func (p *printer) match(event Event) error {
	if event.Malformed() {
		if p.json {
			return p.malformed(event)
		}
		_, err := fmt.Fprintln(p.out, MalformedText(event))
		return err
	}
	if p.json {
		return p.valid(event)
	}
	_, err := fmt.Fprintln(p.out, TextLine(event))
	return err
}

func (p *printer) valid(event Event) error {
	_, err := fmt.Fprintln(p.out, string(event.Raw))
	return err
}

func (p *printer) malformed(event Event) error {
	_, err := fmt.Fprintf(p.errOut, "hive: malformed line %s:%d\n", event.Path(), event.Line())
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
		return time.Time{}, &OptionError{Message: fmt.Sprintf("invalid --since value %q: use a duration or an RFC 3339 time", value)}
	}
	return parsed, nil
}
