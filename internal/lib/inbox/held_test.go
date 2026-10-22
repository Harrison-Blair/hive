package inbox

import (
	"context"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type outputEnvelope struct {
	envelope output.Envelope
}

func TestHeldAppendAndConsumeStayOnOneHandle(t *testing.T) {
	home := t.TempDir()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var appended outputEnvelope
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		appended.envelope = AppendHeld(locked, AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "held"})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if appended.envelope.Outcome != "ok" {
		t.Fatalf("AppendHeld() = %#v", appended.envelope)
	}
	var consumed outputEnvelope
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		consumed.envelope = ConsumeHeld(locked, "coordinator")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if consumed.envelope.Outcome != "ok" {
		t.Fatalf("ConsumeHeld() = %#v", consumed.envelope)
	}
	var all []records.Inbox
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var inner error
		all, inner = AllHeld(locked, "coordinator")
		return inner
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ReadAt == "" || all[0].Text != "held" {
		t.Fatalf("AllHeld() = %#v; want one marked record", all)
	}
}

func TestValidateAppendUsesTheLiteralByteBound(t *testing.T) {
	cases := []struct {
		name string
		text string
		ok   bool
	}{
		{"4096", strings.Repeat("a", 4096), true},
		{"4097", strings.Repeat("a", 4097), false},
		{"4096-utf8", strings.Repeat("\u00e9", 2048), true},
		{"4097-utf8", "a" + strings.Repeat("\u00e9", 2048), false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAppend(AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Text: test.text})
			if (err == nil) != test.ok {
				t.Fatalf("ValidateAppend(%d bytes) = %v; want ok=%t", len(test.text), err, test.ok)
			}
		})
	}
}
