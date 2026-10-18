package inbox

import (
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Send appends one inbox record.
func Send(home string, options inboxlib.AppendOptions) output.Envelope {
	return inboxlib.Append(home, options)
}
