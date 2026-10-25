// Package prune deletes each stopped agent record that no open task and no
// live agent refers to, after it routes the unread mail of that agent.
package prune

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// PrunedAgent describes one pruned agent record.
type PrunedAgent struct {
	Agent        string `json:"agent"`
	Status       string `json:"status"`
	InboxRecords int    `json:"inbox_records"`
	MovedRecords int    `json:"moved_records"`
	MovedTo      string `json:"moved_to"`
}

// Result is the set of records that one prune deleted.
type Result struct {
	Pruned []PrunedAgent `json:"pruned"`
}

// Run prunes each eligible agent record of one home.
func Run(ctx context.Context, home string) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	writer, err := hivelog.NewWriter(home, "prune", time.Now)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)

	registry, err := store.Open(home)
	if err != nil {
		return Result{}, err
	}
	result := Result{Pruned: []PrunedAgent{}}
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		return runLocked(home, locked, logger, &result)
	})
	return result, err
}

func runLocked(home string, locked *store.LockedStore, logger *slog.Logger, result *Result) error {
	agents, err := listAgents(locked)
	if err != nil {
		return err
	}
	tasks, err := listTasks(locked)
	if err != nil {
		return err
	}
	referenced := references(agents, tasks)
	candidates := make([]records.Agent, 0)
	for _, record := range agents {
		if !prunable(record.Status) || referenced[record.Key] {
			continue
		}
		candidates = append(candidates, record)
	}
	if len(candidates) == 0 {
		return nil
	}
	byKey := make(map[string]records.Agent, len(agents))
	for _, record := range agents {
		byKey[record.Key] = record
	}
	candidateSet := make(map[string]bool, len(candidates))
	for _, record := range candidates {
		candidateSet[record.Key] = true
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		left := parentDepth(candidates[i].Key, candidateSet, byKey)
		right := parentDepth(candidates[j].Key, candidateSet, byKey)
		if left != right {
			return left > right
		}
		return candidates[i].Key < candidates[j].Key
	})

	removed := map[string]bool{}
	for _, record := range candidates {
		pruned, err := pruneOne(home, locked, logger, agents, removed, record)
		if err != nil {
			return err
		}
		removed[record.Key] = true
		result.Pruned = append(result.Pruned, pruned)
	}
	return nil
}

func pruneOne(home string, locked *store.LockedStore, logger *slog.Logger, agents []records.Agent, removed map[string]bool, record records.Agent) (PrunedAgent, error) {
	key := inbox.InboxKey(record)
	all, err := inbox.AllHeld(locked, key)
	if err != nil {
		return PrunedAgent{}, err
	}
	destination := destinationAddress(agents, removed, record)
	moved := 0
	if destination != "" {
		remaining := remainingAgents(agents, removed)
		for _, item := range all {
			if item.ReadAt != "" {
				continue
			}
			if err := copyUnread(locked, remaining, destination, item); err != nil {
				return PrunedAgent{}, err
			}
			moved++
		}
	}
	if err := locked.RemoveIn([]string{"agents"}, record.Key); err != nil {
		return PrunedAgent{}, err
	}
	if !hasInboxKey(remainingAgents(agents, removed), record.Key, key) {
		if err := removeInboxFolder(home, key); err != nil {
			return PrunedAgent{}, err
		}
	}
	movedTo := ""
	if moved > 0 {
		movedTo = destination
	}
	logger.Debug("agent pruned",
		"agent_id", record.Key,
		"status", record.Status,
		"inbox_records", len(all),
		"moved_records", moved,
		"moved_to", movedTo,
	)
	return PrunedAgent{
		Agent:        record.Key,
		Status:       record.Status,
		InboxRecords: len(all),
		MovedRecords: moved,
		MovedTo:      movedTo,
	}, nil
}

// destinationAddress is the D9 destination of the unread mail of one record.
func destinationAddress(agents []records.Agent, removed map[string]bool, record records.Agent) string {
	if record.Project != "" {
		return inbox.ProjectAddressPrefix + record.Project
	}
	if record.Parent != "" {
		for _, candidate := range agents {
			if candidate.Key != record.Parent || removed[candidate.Key] {
				continue
			}
			return inbox.InboxKey(candidate)
		}
	}
	return records.AgentKindCoordinator
}

func copyUnread(locked *store.LockedStore, agents []records.Agent, address string, item records.Inbox) error {
	key := inbox.ResolveAddress(agents, address)
	seq, err := locked.Next("inbox:" + key)
	if err != nil {
		return err
	}
	moved := item
	moved.Seq = seq
	moved.Role = key
	moved.ReadAt = ""
	moved.NudgedAt = ""
	moved.NudgedGeneration = 0
	return locked.WriteIn([]string{"inbox", key}, strconv.FormatUint(seq, 10), moved)
}

func removeInboxFolder(home, key string) error {
	path := filepath.Join(home, ".hive", "state", "inbox", key)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("store path invalid: %s", path)
	}
	return os.RemoveAll(path)
}

func prunable(status string) bool {
	return status == records.AgentStopped || status == records.AgentExited || status == records.AgentStale
}

func references(agents []records.Agent, tasks []records.Task) map[string]bool {
	referenced := map[string]bool{}
	byKey := make(map[string]records.Agent, len(agents))
	for _, record := range agents {
		byKey[record.Key] = record
	}
	for _, record := range agents {
		if record.Status != records.AgentLive {
			continue
		}
		markParentChain(referenced, agents, byKey, record.Parent)
		if record.Replaces != "" {
			referenced[record.Replaces] = true
		}
	}
	for _, task := range tasks {
		if !openTask(task) {
			continue
		}
		for _, record := range agents {
			if task.ActingAgent != "" && (record.Name == task.ActingAgent || record.Key == task.ActingAgent) {
				referenced[record.Key] = true
			}
			if record.Task == task.ID {
				referenced[record.Key] = true
			}
		}
	}
	return referenced
}

// markParentChain marks the parent of one live agent and each record on the
// replaces chain from that parent to the live successor.
func markParentChain(referenced map[string]bool, agents []records.Agent, byKey map[string]records.Agent, key string) {
	seen := map[string]bool{}
	for key != "" && !seen[key] {
		seen[key] = true
		referenced[key] = true
		parent, ok := byKey[key]
		if ok && parent.Status == records.AgentLive {
			return
		}
		next, ok := successor(agents, key)
		if !ok {
			return
		}
		key = next.Key
	}
}

func successor(agents []records.Agent, key string) (records.Agent, bool) {
	var newest records.Agent
	found := false
	for _, record := range agents {
		if record.Replaces != key {
			continue
		}
		if !found || record.StartedAt > newest.StartedAt ||
			(record.StartedAt == newest.StartedAt && record.Key > newest.Key) {
			newest = record
			found = true
		}
	}
	return newest, found
}

func openTask(task records.Task) bool {
	if task.Status == "cancelled" {
		return false
	}
	if task.Status == "verified" && task.Delivery != nil && task.Delivery.Pushed {
		return false
	}
	return true
}

func parentDepth(key string, candidates map[string]bool, byKey map[string]records.Agent) int {
	depth := 0
	seen := map[string]bool{key: true}
	current := key
	for {
		record, ok := byKey[current]
		if !ok || record.Parent == "" || !candidates[record.Parent] || seen[record.Parent] {
			return depth
		}
		seen[record.Parent] = true
		depth++
		current = record.Parent
	}
}

func remainingAgents(agents []records.Agent, removed map[string]bool) []records.Agent {
	remaining := make([]records.Agent, 0, len(agents))
	for _, record := range agents {
		if removed[record.Key] {
			continue
		}
		remaining = append(remaining, record)
	}
	return remaining
}

func hasInboxKey(agents []records.Agent, skip, key string) bool {
	for _, record := range agents {
		if record.Key != skip && inbox.InboxKey(record) == key {
			return true
		}
	}
	return false
}

func listAgents(locked *store.LockedStore) ([]records.Agent, error) {
	raw, err := locked.ListIn([]string{"agents"}, nil)
	if err != nil {
		return nil, err
	}
	found := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}
		found = append(found, record)
	}
	return found, nil
}

func listTasks(locked *store.LockedStore) ([]records.Task, error) {
	raw, err := locked.ListIn([]string{"tasks"}, nil)
	if err != nil {
		return nil, err
	}
	found := make([]records.Task, 0, len(raw))
	for _, data := range raw {
		var task records.Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}
		found = append(found, task)
	}
	return found, nil
}
