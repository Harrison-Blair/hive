package herdr

// Error is one error envelope of the herdr server.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns the code and the message of the error.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// The error codes that hive checks. reference/herdr/errors.md lists every
// code that the server can send.
const (
	CodePaneNotFound = "pane_not_found"
	CodeAgentBlocked = "agent_blocked"
)
