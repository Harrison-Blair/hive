package task

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// CreateText renders the effects of task creation and the new ID.
func CreateText(envelope output.Envelope) string {
	if record, ok := envelope.Result.(records.Task); ok && envelope.Error == nil {
		return fmt.Sprintf("id: %s\nbrief: %s\n", record.ID, briefRelPath(record.ID))
	}
	return effectsText(envelope.Effects)
}

// ListText renders a task list.
func ListText(envelope output.Envelope) string {
	tasks, ok := envelope.Result.([]records.Task)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(tasks) == 0 {
		return "no tasks\n"
	}
	lines := make([]string, 0, len(tasks))
	for _, task := range tasks {
		line := fmt.Sprintf("%s\t%s\t%s\t%s",
			task.ID, task.Status, oneLine(task.ActingAgent), oneLine(task.Title))
		if task.NeedsUser != "" {
			line += " [needs user]"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}

// ShowText renders one task record and its brief path.
func ShowText(envelope output.Envelope) string {
	result, ok := envelope.Result.(ShowResult)
	if !ok || envelope.Error != nil {
		return ""
	}
	task := result.Task
	lines := []string{
		"id: " + task.ID,
		"project: " + oneLine(task.Project),
		"title: " + oneLine(task.Title),
		"mail: " + projectAddress(task.Project),
		"acting_agent: " + oneLine(task.ActingAgent),
		"acting_agent_status: " + agentStatus(result.Agent),
		"status: " + task.Status,
		"depends_on: " + strings.Join(task.DependsOn, ","),
		"verification: " + verificationText(task.Verification),
		"verification_history: " + verificationHistoryText(task.VerificationHistory),
		"worktree: " + worktreeText(task.Worktree),
		"delivery: " + deliveryText(task.Delivery),
		"cancel_reason: " + oneLine(task.CancelReason),
		"created_at: " + task.CreatedAt,
		"updated_at: " + task.UpdatedAt,
	}
	if task.NeedsUser != "" {
		lines = append(lines, "needs user: "+oneLine(task.NeedsUser))
	}
	if paths := TaskFeaturePaths(task); paths != nil {
		lines = append(lines,
			"feature: "+task.Feature,
			"brief: "+paths.Brief,
			"evidence: "+paths.Evidence,
		)
		for _, warning := range result.Warnings {
			lines = append(lines, "warning: "+warning)
		}
	} else {
		lines = append(lines, "brief: "+briefRelPath(task.ID))
	}
	return strings.Join(lines, "\n") + "\n"
}

// FailureReport renders the known effects and the inspection commands of one
// finish failure. It never replays the finish group.
func FailureReport(envelope output.Envelope, taskID, agent string) string {
	project := ""
	if result, ok := envelope.Result.(records.Task); ok {
		taskID = result.ID
		project = result.Project
		if agent == "" {
			if result.Verification != nil && result.Verification.Agent != "" {
				agent = result.Verification.Agent
			} else {
				agent = result.ActingAgent
			}
		}
	}
	lines := []string{
		"outcome: " + envelope.Outcome,
		"phase: " + envelope.Phase,
	}
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("effect: %s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if envelope.Error != nil {
		lines = append(lines, "error: "+oneLine(fmt.Sprint(envelope.Error)))
	}
	if taskID != "" {
		lines = append(lines, "inspect: hive task show "+taskID)
	}
	if agent != "" {
		lines = append(lines, "inspect: hive agent show "+oneLine(agent))
	}
	if project != "" {
		lines = append(lines, "inspect: hive inbox --inbox "+projectAddress(project)+" --all")
	} else {
		lines = append(lines, "inspect: hive inbox")
	}
	return strings.Join(lines, "\n") + "\n"
}

// ShowBriefText returns the brief text of one show result.
func ShowBriefText(envelope output.Envelope) string {
	result, ok := envelope.Result.(ShowResult)
	if !ok || envelope.Error != nil {
		return ""
	}
	return result.Brief
}

// TransitionText renders one transition result.
func TransitionText(envelope output.Envelope) string {
	record, ok := envelope.Result.(records.Task)
	if !ok || envelope.Error != nil {
		return ""
	}
	return fmt.Sprintf("id: %s\nstatus: %s\n", record.ID, record.Status)
}

// effectsText renders each completed effect.
func effectsText(effects []output.Effect) string {
	lines := make([]string, 0, len(effects))
	for _, effect := range effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// verificationText renders the last verification of a task.
func verificationText(verification *records.Verification) string {
	if verification == nil {
		return "-"
	}
	text := fmt.Sprintf("%s by %s at %s", verification.Result, oneLine(verification.Agent), verification.At)
	if verification.Reason != "" {
		text += ": " + oneLine(verification.Reason)
	}
	return text
}

// verificationHistoryText renders the verification history list.
func verificationHistoryText(history []records.VerificationHistory) string {
	if history == nil {
		return "[]"
	}
	return jsonText(history)
}

// worktreeText renders the worktree record.
func worktreeText(worktree *records.Worktree) string {
	if worktree == nil {
		return "-"
	}
	return jsonText(worktree)
}

// deliveryText renders the delivery record.
func deliveryText(delivery *records.Delivery) string {
	if delivery == nil {
		return "-"
	}
	return jsonText(delivery)
}

// agentStatus renders the status of the acting agent.
func agentStatus(agent *records.Agent) string {
	if agent == nil {
		return "no record"
	}
	return agent.Status
}

// jsonText renders one value as compact JSON.
func jsonText(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "-"
	}
	return string(data)
}
