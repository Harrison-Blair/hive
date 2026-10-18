package doctor

import (
	"context"
	"log/slog"
)

// Run runs each check in order.
func Run(ctx context.Context, home string, logger *slog.Logger) Result {
	if ctx == nil {
		ctx = context.Background()
	}
	checks := []Check{
		checkHerdrSocket(ctx, logger),
		checkHerdrPane(),
		checkHome(home),
		checkStateStore(home),
		checkGit(ctx),
		checkSupervisor(home),
		checkHarness(ctx, "claude"),
		checkHarness(ctx, "codex"),
		checkHarness(ctx, "opencode"),
		checkHarness(ctx, "pi"),
		checkHarness(ctx, "cursor-agent"),
		checkGH(ctx),
	}
	return Result{Checks: checks}
}
