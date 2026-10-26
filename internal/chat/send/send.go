// Package send refuses the hidden CLI chat send.
package send

import (
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Refusal is the D14 text of a refused chat send.
const Refusal = "chat send is for the user only: type in the Chat screen of the hive TUI"

// Refuse logs one refused CLI chat send and returns its failure. It writes no
// chat record and sends no prompt.
func Refuse(home, paneID, workingDir string) output.Envelope {
	if logger, err := hivelog.Open(home, "chat"); err == nil {
		logger.Debug("chat send refused", "pane_id", paneID, "cwd", workingDir)
	}
	return chatlib.Failed("identity", Refusal, nil)
}
