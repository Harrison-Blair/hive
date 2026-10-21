package records

// Claim status values. A claim with no status is live.
const (
	ClaimStarting = "starting"
	ClaimLive     = "live"
)

// Claim is the record that gives one agent a singleton role.
type Claim struct {
	Key        string `json:"key"`
	Agent      string `json:"agent"`
	Generation uint64 `json:"generation"`
	PID        int64  `json:"pid"`
	PIDStart   uint64 `json:"pid_start"`
	ClaimedAt  string `json:"claimed_at"`
	// Status is starting while a launcher starts the coordinator harness.
	// Then PID and PIDStart name the launcher and Agent is empty.
	Status string `json:"status,omitempty"`
	// PaneID is the target pane of a starting claim.
	PaneID string `json:"pane_id,omitempty"`
}
