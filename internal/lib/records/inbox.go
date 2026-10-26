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
	InboxReport     = "report"
)

// Inbox is one durable record in one inbox.
type Inbox struct {
	Seq   uint64 `json:"seq"`
	Inbox string `json:"inbox"`
	Kind  string `json:"kind"`
	Task  string `json:"task"`
	Agent string `json:"agent"`
	Text  string `json:"text"`
	// Path is the absolute path of the file of a report record.
	Path             string `json:"path,omitempty"`
	CreatedAt        string `json:"created_at"`
	ReadAt           string `json:"read_at"`
	NudgedAt         string `json:"nudged_at"`
	NudgedGeneration uint64 `json:"nudged_generation"`
}
