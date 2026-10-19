package records

import "regexp"

var taskIDPattern = regexp.MustCompile(`^[0-9]+-[a-z0-9]+(-[a-z0-9]+)*$`)

// maxTaskIDLength is 20 counter digits, one hyphen, and a 40-character slug.
const maxTaskIDLength = 61

// Task is one durable task record.
type Task struct {
	ID                  string                `json:"id"`
	Project             string                `json:"project"`
	Title               string                `json:"title"`
	ControllingRole     string                `json:"controlling_role"`
	ActingAgent         string                `json:"acting_agent"`
	Status              string                `json:"status"`
	DependsOn           []string              `json:"depends_on"`
	Verification        *Verification         `json:"verification"`
	VerificationHistory []VerificationHistory `json:"verification_history"`
	Worktree            *Worktree             `json:"worktree"`
	Delivery            *Delivery             `json:"delivery"`
	CreatedAt           string                `json:"created_at"`
	UpdatedAt           string                `json:"updated_at"`
	// CancelReason holds the reason of a cancel command. It is empty for a
	// task that no cancel command moved.
	CancelReason string `json:"cancel_reason,omitempty"`
}

// VerificationHistory keeps one verification that is no longer current.
type VerificationHistory struct {
	Verification *Verification `json:"verification"`
	Reason       string        `json:"reason"`
	Caller       string        `json:"caller"`
	At           string        `json:"at"`
}

// Verification is the last independent check of one task.
type Verification struct {
	Agent  string `json:"agent"`
	Result string `json:"result"`
	Reason string `json:"reason"`
	At     string `json:"at"`
}

// Worktree holds the worktree fields of a task. T09 fills them.
type Worktree struct {
	WorkspaceID string `json:"workspace_id"`
	TabID       string `json:"tab_id"`
	RootPane    string `json:"root_pane"`
	Path        string `json:"path"`
	Branch      string `json:"branch"`
	CreatedAt   string `json:"created_at"`
	RemovedAt   string `json:"removed_at"`
}

// Delivery holds the delivery fields of a task. T12 fills them.
type Delivery struct {
	Mode     string `json:"mode"`
	Target   string `json:"target"`
	Grant    string `json:"grant"`
	Commit   string `json:"commit"`
	Pushed   bool   `json:"pushed"`
	PRURL    string `json:"pr_url,omitempty"`
	Evidence string `json:"evidence,omitempty"`
	At       string `json:"at"`
}

// ValidTaskID reports whether id has the task form <n>-<slug>.
func ValidTaskID(id string) bool {
	return len(id) <= maxTaskIDLength && taskIDPattern.MatchString(id)
}
