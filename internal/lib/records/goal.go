package records

import "regexp"

var goalIDPattern = regexp.MustCompile(`^g[1-9][0-9]*-[a-z0-9]+(-[a-z0-9]+)*$`)

// Goal is one durable project goal.
type Goal struct {
	ID        string     `json:"id"`
	Project   string     `json:"project"`
	Title     string     `json:"title"`
	Grant     string     `json:"grant"`
	Status    string     `json:"status"`
	Waves     []GoalWave `json:"waves"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// GoalWave is one ordered delivery wave of a goal.
type GoalWave struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	DoneWhen  string   `json:"done_when"`
	Tasks     []string `json:"tasks"`
	NeedsUser string   `json:"needs_user,omitempty"`
	Status    string   `json:"status"`
}

// ValidGoalID reports whether id has the form g<n>-<slug>.
func ValidGoalID(id string) bool {
	return len(id) <= 200 && goalIDPattern.MatchString(id)
}
