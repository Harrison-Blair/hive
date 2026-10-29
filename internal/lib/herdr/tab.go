package herdr

import (
	"context"
	"encoding/json"
	"fmt"
)

// TabRename changes the label of one tab with tab.rename.
func TabRename(ctx context.Context, tabID, label string, opts ...Option) error {
	var result struct {
		Type *string         `json:"type"`
		Tab  json.RawMessage `json:"tab"`
	}
	params := struct {
		TabID string `json:"tab_id"`
		Label string `json:"label"`
	}{TabID: tabID, Label: label}
	if err := Call(ctx, "tab.rename", params, &result, opts...); err != nil {
		return err
	}
	if result.Type == nil || *result.Type != "tab_info" {
		return fmt.Errorf("bad reply: result type %q, want tab_info", stringValue(result.Type))
	}
	if isAbsentOrNull(result.Tab) {
		return fmt.Errorf("bad reply: tab_info tab is missing")
	}
	renamed, err := decodeTabInfo(result.Tab)
	if err != nil {
		return fmt.Errorf("bad reply: %s", err)
	}
	if renamed.TabID != tabID {
		return fmt.Errorf("bad reply: renamed tab ID %q, want %q", renamed.TabID, tabID)
	}
	return nil
}

func decodeTabInfo(raw json.RawMessage) (TabInfo, error) {
	if isAbsentOrNull(raw) {
		return TabInfo{}, fmt.Errorf("tab is missing")
	}
	var wire struct {
		TabID       *string `json:"tab_id"`
		WorkspaceID *string `json:"workspace_id"`
		Number      *uint   `json:"number"`
		Label       *string `json:"label"`
		Focused     *bool   `json:"focused"`
		PaneCount   *uint   `json:"pane_count"`
		AgentStatus *string `json:"agent_status"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return TabInfo{}, fmt.Errorf("tab: %s", err)
	}
	for _, field := range []struct {
		name  string
		value *string
	}{
		{"tab_id", wire.TabID},
		{"workspace_id", wire.WorkspaceID},
		{"agent_status", wire.AgentStatus},
	} {
		if field.value == nil || *field.value == "" {
			return TabInfo{}, fmt.Errorf("tab %s is missing or empty", field.name)
		}
	}
	if wire.Number == nil {
		return TabInfo{}, fmt.Errorf("tab number is missing")
	}
	if wire.Focused == nil {
		return TabInfo{}, fmt.Errorf("tab focused is missing")
	}
	if wire.PaneCount == nil {
		return TabInfo{}, fmt.Errorf("tab pane_count is missing")
	}
	tab := TabInfo{
		TabID:       *wire.TabID,
		WorkspaceID: *wire.WorkspaceID,
		Number:      *wire.Number,
		Focused:     *wire.Focused,
		PaneCount:   *wire.PaneCount,
		AgentStatus: *wire.AgentStatus,
	}
	if wire.Label != nil {
		tab.Label = *wire.Label
	}
	return tab, nil
}
