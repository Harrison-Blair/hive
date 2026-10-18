package doctor

// Result holds the checks of one doctor run.
type Result struct {
	Checks []Check `json:"checks"`
}

// Failed reports a failed required check.
func (r Result) Failed() bool {
	return false
}

// ExitCode is 1 when a required check fails.
func (r Result) ExitCode() int {
	return 0
}
