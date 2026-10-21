package records

// Agent status values.
const (
	AgentStarting   = "starting"
	AgentLive       = "live"
	AgentStale      = "stale"
	AgentStopped    = "stopped"
	AgentExited     = "exited"
	AgentUnresolved = "unresolved"
)

// Agent is one registered coding-agent session in a herdr pane.
type Agent struct {
	Key             string   `json:"key"`
	Name            string   `json:"name"`
	Role            string   `json:"role"`
	Project         string   `json:"project"`
	Task            string   `json:"task"`
	Harness         string   `json:"harness"`
	Model           string   `json:"model"`
	PaneID          string   `json:"pane_id"`
	PaneHistory     []string `json:"pane_history,omitempty"`
	WorkspaceID     string   `json:"workspace_id"`
	TabID           string   `json:"tab_id"`
	TerminalID      string   `json:"terminal_id"`
	Cwd             string   `json:"cwd"`
	Status          string   `json:"status"`
	Parent          string   `json:"parent,omitempty"`
	Replaces        string   `json:"replaces,omitempty"`
	StatusChangedAt string   `json:"status_changed_at,omitempty"`
	// AgentStatus is the last observed harness status, separate from Status.
	AgentStatus   string `json:"agent_status,omitempty"`
	Finished      bool   `json:"finished"`
	QuietAfterMin int    `json:"quiet_after_min"`
	Generation    uint64 `json:"generation"`
	StartedAt     string `json:"started_at"`
	UpdatedAt     string `json:"updated_at"`
	LastError     string `json:"last_error,omitempty"`
	Spawned       bool   `json:"spawned"`
	StartedByHive bool   `json:"started_by_hive,omitempty"`
	// TabCreatedByHive marks a general agent whose tab hive created for it.
	TabCreatedByHive bool `json:"tab_created_by_hive,omitempty"`
	// FirstPrompt is the kept first prompt of a general agent that spawn did
	// not send. The supervisor sends it.
	FirstPrompt string `json:"first_prompt,omitempty"`
	// FirstPromptAttemptAt and FirstPromptSentAt are steps 2 and 4 of the
	// supervisor send of a kept first prompt.
	FirstPromptAttemptAt string `json:"first_prompt_attempt_at,omitempty"`
	FirstPromptSentAt    string `json:"first_prompt_sent_at,omitempty"`
	ProcessPID           int64  `json:"process_pid,omitempty"`
	ProcessName          string `json:"process_name,omitempty"`
	ProcessStart         uint64 `json:"process_start,omitempty"`
	AgentSessionRef      string `json:"agent_session_ref,omitempty"`
}
