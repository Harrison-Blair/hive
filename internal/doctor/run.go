package doctor

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/Harrison-Blair/hive/internal/lib/run"
	storelib "github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/lib/version"
)

// Run runs each check in order.
func Run(ctx context.Context, home string, logger *slog.Logger) Result {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = run.WithLogger(ctx, logger)
	checks := []Check{
		checkHerdrSocket(ctx, logger),
		checkHerdrPane(),
		checkHome(home),
		checkStateStore(home),
		checkSchemaVersion(home),
		checkStartingAgents(home),
		checkUnresolvedAgents(home),
		checkOwnerlessInboxes(home),
		checkGit(ctx),
		checkHomeVersion(ctx, home, version.Info().Version, "git"),
		checkSupervisor(home),
		checkSupervisorBinary(home),
		checkStaleWorktrees(home),
		checkUnpushedTaskBranches(ctx, home),
		checkHarness(ctx, "claude"),
		checkHarness(ctx, "codex"),
		checkHarness(ctx, "opencode"),
		checkHarness(ctx, "pi"),
		checkHarness(ctx, "cursor-agent"),
		checkGH(ctx),
	}
	return Result{Checks: checks}
}

// checkSchemaVersion shows the version of the state store after the migration
// that the store open runs.
func checkSchemaVersion(home string) Check {
	version, err := storelib.SchemaVersion(home)
	if err != nil {
		return newCheck("schema version", false, false, err.Error())
	}
	return newCheck("schema version", false, true, strconv.Itoa(version))
}
