// Package show reads one task record, its brief, and its acting agent.
package show

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Options holds the values of one task read.
type Options struct {
	Brief  bool
	Logger *slog.Logger
}

// Result holds one task record and its brief text.
type Result struct {
	records.Task
	Brief    string         `json:"brief"`
	Agent    *records.Agent `json:"agent,omitempty"`
	Warnings []string       `json:"warnings,omitempty"`
}

// Run returns one task record and its brief without taking the lock.
func Run(home, id string, options Options) output.Envelope {
	log := tasklib.OpenLog(home, options.Logger)
	defer log.Close()
	fail := func(phase, message string) output.Envelope {
		log.Failure("task show", "", id, "", phase, message)
		return tasklib.FailedEnvelope(phase, message)
	}
	if !records.ValidTaskID(id) {
		return fail("validate", "task not found: "+tasklib.QuoteIfControl(id))
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var record records.Task
	var agent *records.Agent
	var readErr error
	if err := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if err := locked.ReadIn([]string{tasklib.TasksCollection}, id, &record); err != nil {
			readErr = err
			return nil
		}
		if record.ActingAgent == "" {
			return nil
		}
		found, ok, err := identity.FindHeld(locked, record.ActingAgent)
		if err != nil {
			readErr = err
			return nil
		}
		if ok {
			agent = &found
		}
		return nil
	}); err != nil {
		return fail("store", err.Error())
	}
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return fail("read", "task not found: "+tasklib.QuoteIfControl(id))
		}
		return fail("read", readErr.Error())
	}
	if err := tasklib.CheckBriefDir(home); err != nil {
		log.Failure("task show", record.Project, id, "", "brief", err.Error())
		return tasklib.FailedEnvelope("brief", err.Error())
	}
	brief, err := tasklib.ReadBrief(tasklib.BriefPath(home, id), nil, tasklib.BriefDeadline)
	if err != nil {
		log.Failure("task show", record.Project, id, "", "brief", err.Error())
		return tasklib.FailedEnvelope("brief", err.Error())
	}
	warnings := featureFileWarnings(home, record)
	return tasklib.OKEnvelope(Result{Task: record, Brief: string(brief), Agent: agent, Warnings: warnings})
}

func featureFileWarnings(home string, record records.Task) []string {
	paths := tasklib.TaskFeaturePaths(record)
	if paths == nil {
		return nil
	}
	warnings := make([]string, 0, 2)
	for _, path := range []string{paths.README, paths.Brief} {
		info, err := os.Lstat(filepath.Join(home, filepath.FromSlash(path)))
		if err != nil || !info.Mode().IsRegular() {
			warnings = append(warnings, "linked file missing: "+path)
		}
	}
	return warnings
}

// Text renders one task record and its brief path.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok || envelope.Error != nil {
		return ""
	}
	task := result.Task
	lines := []string{
		"id: " + task.ID,
		"project: " + tasklib.OneLine(task.Project),
		"title: " + tasklib.OneLine(task.Title),
		"mail: " + tasklib.ProjectAddress(task.Project),
		"acting_agent: " + tasklib.OneLine(task.ActingAgent),
		"acting_agent_status: " + agentStatus(result.Agent),
		"status: " + task.Status,
		"depends_on: " + strings.Join(task.DependsOn, ","),
		"verification: " + verificationText(task.Verification),
		"verification_history: " + verificationHistoryText(task.VerificationHistory),
		"worktree: " + worktreeText(task.Worktree),
		"delivery: " + deliveryText(task.Delivery),
		"cancel_reason: " + tasklib.OneLine(task.CancelReason),
		"created_at: " + task.CreatedAt,
		"updated_at: " + task.UpdatedAt,
	}
	if task.NeedsUser != "" {
		lines = append(lines, "needs user: "+tasklib.OneLine(task.NeedsUser))
	}
	if paths := tasklib.TaskFeaturePaths(task); paths != nil {
		lines = append(lines,
			"feature: "+task.Feature,
			"brief: "+paths.Brief,
			"evidence: "+paths.Evidence,
		)
		for _, warning := range result.Warnings {
			lines = append(lines, "warning: "+warning)
		}
	} else {
		lines = append(lines, "brief: "+tasklib.BriefRelPath(task.ID))
	}
	return strings.Join(lines, "\n") + "\n"
}

// BriefText returns the brief text of one show result.
func BriefText(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok || envelope.Error != nil {
		return ""
	}
	return result.Brief
}

// verificationText renders the last verification of a task.
func verificationText(verification *records.Verification) string {
	if verification == nil {
		return "-"
	}
	text := fmt.Sprintf("%s by %s at %s", verification.Result, tasklib.OneLine(verification.Agent), verification.At)
	if verification.Reason != "" {
		text += ": " + tasklib.OneLine(verification.Reason)
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
