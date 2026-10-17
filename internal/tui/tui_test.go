package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestQuitKeysQuit(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{
		{Code: 'q', Text: "q"},
		{Code: 'c', Mod: tea.ModCtrl},
	} {
		_, cmd := Model{}.Update(key)
		if cmd == nil {
			t.Fatalf("%q: got no command, want quit", key.String())
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%q: got %T, want tea.QuitMsg", key.String(), cmd())
		}
	}
}

func TestOtherKeysDoNotQuit(t *testing.T) {
	_, cmd := Model{}.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if cmd != nil {
		t.Fatalf("got command %T, want none", cmd())
	}
}

func TestViewNamesHive(t *testing.T) {
	if got := (Model{}).View().Content; !strings.Contains(got, "hive") {
		t.Fatalf("view %q does not name hive", got)
	}
}

func TestRunExitsOnQ(t *testing.T) {
	done := make(chan error, 1)
	go func() { done <- Run(strings.NewReader("q"), &bytes.Buffer{}) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not exit after q")
	}
}
