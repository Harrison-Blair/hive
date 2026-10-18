package doctor

import (
	"context"
	"log/slog"
)

// The status of one check.
const (
	StatusOK   = "ok"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// Check is one line of the doctor result.
type Check struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Detail   string `json:"detail"`
}

// Run runs each check in order.
func Run(ctx context.Context, home string, logger *slog.Logger) Result {
	return Result{}
}
