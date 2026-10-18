package inbox

import (
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Read returns all records or consumes the unread records of one role.
func Read(home, role string, all bool) output.Envelope {
	if !all {
		return inboxlib.Consume(home, role)
	}
	found, err := inboxlib.All(home, role)
	if err != nil {
		return output.Envelope{Outcome: "failed", Phase: "read", Effects: []output.Effect{}, Error: err.Error(), Result: nil}
	}
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{},
		Error:   nil,
		Result:  inboxlib.Result{Records: found},
	}
}
