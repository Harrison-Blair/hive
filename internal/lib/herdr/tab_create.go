package herdr

import (
	"context"
	"encoding/json"
	"fmt"
)

// TabCreateParams is one tab.create request.
type TabCreateParams struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
	Cwd         string `json:"cwd,omitempty"`
	Label       string `json:"label,omitempty"`
	Focus       bool   `json:"focus"`
}

// TabCreated is the new tab and its root pane.
type TabCreated struct {
	Tab      TabInfo  `json:"tab"`
	RootPane PaneInfo `json:"root_pane"`
}

// TabCreate makes one tab and its root pane in a workspace.
func TabCreate(ctx context.Context, params TabCreateParams, opts ...Option) (TabCreated, error) {
	var result struct {
		Type     *string         `json:"type"`
		Tab      json.RawMessage `json:"tab"`
		RootPane json.RawMessage `json:"root_pane"`
	}
	if err := Call(ctx, "tab.create", params, &result, opts...); err != nil {
		return TabCreated{}, err
	}
	if result.Type == nil || *result.Type != "tab_created" {
		return TabCreated{}, fmt.Errorf("bad reply: result type %q, want tab_created", stringValue(result.Type))
	}
	if isAbsentOrNull(result.Tab) {
		return TabCreated{}, fmt.Errorf("bad reply: tab_created tab is missing")
	}
	var tab TabInfo
	if err := json.Unmarshal(result.Tab, &tab); err != nil {
		return TabCreated{}, fmt.Errorf("bad reply: tab: %s", err)
	}
	pane, err := decodePaneInfo(result.RootPane)
	if err != nil {
		return TabCreated{}, fmt.Errorf("bad reply: %s", err)
	}
	return TabCreated{Tab: tab, RootPane: pane}, nil
}
