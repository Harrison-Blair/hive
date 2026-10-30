package store

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestRemoveInDeletesOneRecordAndKeepsOthers(t *testing.T) {
	home := t.TempDir()
	registry := openStore(t, home)
	for _, id := range []string{"coordinator", "supervisor"} {
		if err := registry.WriteIn([]string{"claims"}, id, map[string]string{"key": id}); err != nil {
			t.Fatal(err)
		}
	}
	err := registry.WithLock(context.Background(), func(locked *LockedStore) error {
		return locked.RemoveIn([]string{"claims"}, "coordinator")
	})
	if err != nil {
		t.Fatalf("RemoveIn() = %v", err)
	}
	var claim map[string]string
	if err := registry.ReadIn([]string{"claims"}, "coordinator", &claim); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read removed claim = %v, %#v; want not exist", err, claim)
	}
	if err := registry.ReadIn([]string{"claims"}, "supervisor", &claim); err != nil || claim["key"] != "supervisor" {
		t.Fatalf("read kept claim = %v, %#v; want supervisor", err, claim)
	}
}

func TestRemoveInMissingRecordIsNotExist(t *testing.T) {
	home := t.TempDir()
	registry := openStore(t, home)
	err := registry.WithLock(context.Background(), func(locked *LockedStore) error {
		return locked.RemoveIn([]string{"claims"}, "coordinator")
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("RemoveIn() = %v; want not exist", err)
	}
}

func TestRemoveInAfterCallbackFails(t *testing.T) {
	home := t.TempDir()
	registry := openStore(t, home)
	var held *LockedStore
	if err := registry.WithLock(context.Background(), func(locked *LockedStore) error {
		held = locked
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := held.RemoveIn([]string{"claims"}, "coordinator"); !errors.Is(err, ErrLockReleased) {
		t.Fatalf("RemoveIn() after callback = %v; want ErrLockReleased", err)
	}
}
