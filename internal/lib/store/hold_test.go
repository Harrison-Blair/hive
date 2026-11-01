package store

import (
	"errors"
	"testing"
	"time"
)

func TestHoldRunsOneHolderAtATimeAndAllowsWrites(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	inside := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- s.Hold("layout", time.Second, func() error {
			close(inside)
			<-release
			return s.WriteIn([]string{"things"}, "one", map[string]string{"a": "b"})
		})
	}()
	<-inside
	if err := s.Hold("layout", 50*time.Millisecond, func() error { return nil }); !errors.Is(err, errLockTimeout) {
		t.Fatalf("second Hold while held = %v; want a lock timeout", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first Hold = %v; want nil, with a record write inside", err)
	}
	ran := false
	if err := s.Hold("layout", time.Second, func() error { ran = true; return nil }); err != nil || !ran {
		t.Fatalf("Hold after release = %v, ran %v; want it to run", err, ran)
	}
	if err := s.Hold("Bad/Name", time.Second, func() error { return nil }); err == nil {
		t.Fatal("Hold with a bad name = nil; want an error")
	}
}
