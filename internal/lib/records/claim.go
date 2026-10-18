package records

// Claim is the record that gives one agent a singleton role.
type Claim struct {
	Key        string `json:"key"`
	Agent      string `json:"agent"`
	Generation uint64 `json:"generation"`
	PID        int64  `json:"pid"`
	PIDStart   uint64 `json:"pid_start"`
	ClaimedAt  string `json:"claimed_at"`
}
