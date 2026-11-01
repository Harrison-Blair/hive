// Package create finds or makes the checkout of a branch.
package create

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/worktree"
)

// Options holds the values of one create.
type Options struct {
	// Dir selects the Git repository and the start commit of a new branch:
	// the working directory of the caller.
	Dir    string
	Branch string
	Base   string
	Logger *slog.Logger
}

// Run returns the checkout of the branch, and makes it below the home when no
// checkout has the branch. The result is the worktree.Checkout.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	envelope := run(ctx, home, options, logger)
	attrs := []any{"command", "worktree create", "branch", options.Branch, "outcome", envelope.Outcome, "phase", envelope.Phase}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("worktree create", attrs...)
	return envelope
}

func run(ctx context.Context, home string, options Options, logger *slog.Logger) output.Envelope {
	if options.Branch == "" {
		return output.Failed("validate", "BRANCH is required")
	}
	var checkout worktree.Checkout
	err := agent.Hold(home, func() error {
		var err error
		checkout, err = worktree.Ensure(ctx, home, options.Dir, options.Branch, options.Base, logger)
		return err
	})
	effects := worktree.Effects(checkout)
	switch {
	case err != nil && checkout.Created:
		// git made the checkout and then failed. The checkout is usable.
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "worktree", Effects: effects, Error: err.Error(), Result: checkout}
	case err != nil:
		envelope := output.Failed("worktree", err.Error())
		envelope.Effects = effects
		return envelope
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: checkout}
}

// Text renders the new branch and the checkout of one create. The branch
// line shows also after a failure, so the user knows that the branch stays.
func Text(envelope output.Envelope) string {
	var text strings.Builder
	for _, effect := range envelope.Effects {
		if effect.Resource == "branch" {
			fmt.Fprintf(&text, "%s branch %s\n", effect.Action, effect.Path)
		}
	}
	checkout, ok := envelope.Result.(worktree.Checkout)
	switch {
	case !ok:
	case checkout.Created:
		fmt.Fprintf(&text, "created worktree %s of branch %s\n", checkout.Path, checkout.Branch)
	default:
		fmt.Fprintf(&text, "branch %s has worktree %s\n", checkout.Branch, checkout.Path)
	}
	return text.String()
}
