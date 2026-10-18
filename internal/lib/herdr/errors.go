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

// The error codes that hive handles.
const (
	CodeInvalidRequest     = "invalid_request"
	CodeUnknownMethod      = "unknown_method"
	CodeWorkspaceNotFound  = "workspace_not_found"
	CodeTabNotFound        = "tab_not_found"
	CodePaneNotFound       = "pane_not_found"
	CodeAgentNotFound      = "agent_not_found"
	CodeAgentBlocked       = "agent_blocked"
	CodeAgentNotReady      = "agent_not_ready"
	CodeAgentPromptStalled = "agent_prompt_stalled"
	CodeTimeout            = "timeout"
	CodeServerNotRunning   = "server_not_running"
)
