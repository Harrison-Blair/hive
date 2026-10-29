package logs

import (
	"fmt"
	"strings"
)

// Filter selects events by their fields. An empty field matches each value.
type Filter struct {
	Level     string
	Component string
	TaskID    string
	AgentID   string
	RequestID string
}

// Match reports whether the event has each named field. An agent matches
// agent_id or target. A level matches the
// named level and each level above it. A malformed event has no fields, so it
// matches only an empty filter.
func (f Filter) Match(e Event) bool {
	if e.Fields == nil {
		return f.empty()
	}
	if f.Component != "" && fieldString(e.Fields, "component") != f.Component {
		return false
	}
	if f.TaskID != "" && fieldString(e.Fields, "task_id") != f.TaskID {
		return false
	}
	if f.AgentID != "" && fieldString(e.Fields, "agent_id") != f.AgentID && fieldString(e.Fields, "target") != f.AgentID {
		return false
	}
	if f.RequestID != "" && fieldString(e.Fields, "request_id") != f.RequestID {
		return false
	}
	if f.Level != "" {
		minimum, ok := levelRank(f.Level)
		if !ok {
			return false
		}
		actual, ok := levelRank(fieldString(e.Fields, "level"))
		if !ok || actual < minimum {
			return false
		}
	}
	return true
}

func (f Filter) empty() bool {
	return f.Level == "" && f.Component == "" && f.TaskID == "" && f.AgentID == "" && f.RequestID == ""
}

func fieldString(fields map[string]any, key string) string {
	value, ok := fields[key]
	if !ok || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

func levelRank(level string) (int, bool) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return 0, true
	case "INFO":
		return 1, true
	case "WARN", "WARNING":
		return 2, true
	case "ERROR":
		return 3, true
	default:
		return 0, false
	}
}
