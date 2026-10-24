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

// Agent kinds. A spawned agent has no kind.
const (
	AgentKindCoordinator = "coordinator"
	AgentKindSupervisor  = "supervisor"
)

// Profile labels of D1. The supervisor has no profile.
const (
	ProfileCoordinator = "coordinator"
	ProfilePlanner     = "planner"
	ProfileBuilder     = "builder"
	ProfileVerifier    = "verifier"
	ProfileGeneral     = "general"
)

// Agent is one registered coding-agent session in a herdr pane.
type Agent struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Kind    string `json:"kind"`
	Profile string `json:"profile"`
	// Worktree is the absolute worktree path of the task, or empty.
	Worktree string `json:"worktree"`
	// Owns is the project ID that the agent owns, or empty.
	Owns            string   `json:"owns"`
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
	// ReportedAt is the time of the last hive agent report of a general agent.
	ReportedAt string `json:"reported_at,omitempty"`
}

// MirrorRole returns the wave-A role mirror that D1 derives from the facts.
func (a Agent) MirrorRole() string {
	switch {
	case a.Kind == AgentKindCoordinator:
		return AgentKindCoordinator
	case a.Kind == AgentKindSupervisor:
		return AgentKindSupervisor
	case a.Owns != "":
		return "manager"
	case a.Task != "" && a.Profile == ProfileVerifier:
		return "verifier"
	case a.Task != "":
		return "worker"
	default:
		return "agent"
	}
}

// SyncRole writes the wave-A role mirror from the facts. It is the one
// function that sets role when hive writes an agent record.
func (a *Agent) SyncRole() {
	a.Role = a.MirrorRole()
}

// ApplyRoleFacts sets kind, profile, owns, and worktree from the wave-A role
// value and the worktree of the task record. It keeps role, because the
// version 2 migration keeps the mirror of a record that it changes. It
// returns false when the role is not one of the six D2 values, and changes
// nothing then.
func (a *Agent) ApplyRoleFacts(worktree string) bool {
	switch a.Role {
	case AgentKindCoordinator:
		a.Kind = AgentKindCoordinator
		a.Profile = ProfileCoordinator
		a.Owns = ""
		a.Worktree = ""
	case AgentKindSupervisor:
		a.Kind = AgentKindSupervisor
		a.Profile = ""
		a.Owns = ""
		a.Worktree = ""
	case "manager":
		a.Kind = ""
		a.Profile = ProfilePlanner
		a.Owns = a.Project
		a.Worktree = ""
	case "worker":
		a.Kind = ""
		a.Profile = ProfileBuilder
		a.Owns = ""
		a.Worktree = worktree
	case "verifier":
		a.Kind = ""
		a.Profile = ProfileVerifier
		a.Owns = ""
		a.Worktree = worktree
	case "agent":
		a.Kind = ""
		a.Profile = ProfileGeneral
		a.Owns = ""
		a.Worktree = ""
	default:
		return false
	}
	return true
}

// SetRoleFacts sets the facts of a new record from the wave-A role value and
// then writes the role mirror.
func (a *Agent) SetRoleFacts(worktree string) bool {
	if !a.ApplyRoleFacts(worktree) {
		return false
	}
	a.SyncRole()
	return true
}
