package records

// Inbox record kinds.
const (
	InboxCompleted  = "completed"
	InboxVerified   = "verified"
	InboxBlocked    = "blocked"
	InboxExited     = "exited"
	InboxQuiet      = "quiet"
	InboxUnresolved = "unresolved"
	InboxMessage    = "message"
)

// Inbox is one durable record in the inbox of a role.
type Inbox struct {
	Seq       uint64 `json:"seq"`
	Role      string `json:"role"`
	Kind      string `json:"kind"`
	Task      string `json:"task"`
	Agent     string `json:"agent"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	ReadAt    string `json:"read_at"`
	NudgedAt  string `json:"nudged_at"`
}
