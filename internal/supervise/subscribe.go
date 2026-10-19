package supervise

import (
	"context"
	"sort"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// plainSubscriptions are the six plain entries of one subscribe request.
var plainSubscriptions = []string{
	"pane.agent_detected",
	"pane.exited",
	"pane.closed",
	"tab.closed",
	"workspace.closed",
	"pane.moved",
}

// subscriptionParams returns the selector list of the watched panes.
func subscriptionParams(panes []string) herdr.SubscribeParams {
	subscriptions := make([]herdr.Subscription, 0, len(panes)+len(plainSubscriptions))
	for _, pane := range panes {
		subscriptions = append(subscriptions, herdr.Subscription{Type: "pane.agent_status_changed", PaneID: pane})
	}
	for _, kind := range plainSubscriptions {
		subscriptions = append(subscriptions, herdr.Subscription{Type: kind})
	}
	return herdr.SubscribeParams{Subscriptions: subscriptions}
}

// watchedPanes returns the sorted unique panes of each live record.
func (s *Supervisor) watchedPanes() []string {
	seen := map[string]bool{}
	for _, record := range s.agents {
		if (record.Status == records.AgentLive || record.Status == records.AgentStarting) && record.PaneID != "" {
			seen[record.PaneID] = true
		}
	}
	panes := make([]string, 0, len(seen))
	for pane := range seen {
		panes = append(panes, pane)
	}
	sort.Strings(panes)
	return panes
}

// resubscribe opens a new subscription when the watched set changed, closes
// the old one, and applies the current status of each watched pane.
func (s *Supervisor) resubscribe(ctx context.Context) error {
	panes := s.watchedPanes()
	if s.stream != nil && sameSet(s.watched, panes) {
		return nil
	}
	stream, err := s.ops.Subscribe(ctx, subscriptionParams(panes))
	if err != nil {
		return err
	}
	old := s.stream
	s.stream = stream
	s.watched = setOf(panes)
	if old != nil {
		_ = old.Close()
	}
	s.logger.Debug("supervisor resubscribe", "command", "supervise", "panes", len(panes))
	return s.applyStatuses(ctx)
}

// applyStatuses reads the current status of each watched pane after a rebuild.
func (s *Supervisor) applyStatuses(ctx context.Context) error {
	for _, pane := range s.watchedPanes() {
		record, ok := s.liveByPane(pane)
		if !ok {
			continue
		}
		info, err := s.ops.GetAgent(ctx, pane)
		if err != nil {
			continue
		}
		if record.TerminalID != "" && info.TerminalID != record.TerminalID {
			continue
		}
		if info.AgentStatus == "" || info.AgentStatus == record.AgentStatus {
			continue
		}
		if err := s.applyStatus(ctx, herdr.Event{Event: "rebuild"}, record, info.AgentStatus); err != nil {
			return err
		}
	}
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	return s.settleQuiet(ctx)
}

func sameSet(current map[string]bool, panes []string) bool {
	if len(current) != len(panes) {
		return false
	}
	for _, pane := range panes {
		if !current[pane] {
			return false
		}
	}
	return true
}

func setOf(panes []string) map[string]bool {
	set := make(map[string]bool, len(panes))
	for _, pane := range panes {
		set[pane] = true
	}
	return set
}
