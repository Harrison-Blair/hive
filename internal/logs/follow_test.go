package logs

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

type testClock struct {
	nanos atomic.Int64
}

func (c *testClock) set(when time.Time) {
	c.nanos.Store(when.UnixNano())
}

func (c *testClock) now() time.Time {
	return time.Unix(0, c.nanos.Load()).In(time.Local)
}

func nextEvent(t *testing.T, events <-chan Event, timeout time.Duration) Event {
	t.Helper()
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed before an event arrived")
		}
		return event
	case <-time.After(timeout):
		t.Fatal("no event arrived before the timeout")
		return Event{}
	}
}

func cancelFollow(t *testing.T, cancel context.CancelFunc, errors <-chan error) {
	t.Helper()
	cancel()
	select {
	case err, ok := <-errors:
		if ok && err != nil {
			t.Errorf("cancelled follower returned %v; want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled follower did not stop")
	}
}

func TestFollowNewLineWithinOneSecond(t *testing.T) {
	clock := &testClock{}
	today := localNoon(time.Now())
	clock.set(today)
	reader, home := newTestReader(t, clock.now())
	reader.now = clock.now
	writeDaily(t, home, today, eventLine(t, today, map[string]any{"msg": "old"}))

	_, offset, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	events, errors := reader.Follow(ctx, offset)

	line := eventLine(t, today, map[string]any{"msg": "new"}) + "\n"
	writeErr := make(chan error, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		file, err := os.OpenFile(dailyPath(home, today), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			writeErr <- err
			return
		}
		_, err = file.WriteString(line)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		writeErr <- err
	}()
	if got := eventMessage(nextEvent(t, events, time.Second)); got != "new" {
		t.Fatalf("message = %q; want new", got)
	}
	if err := <-writeErr; err != nil {
		t.Fatal(err)
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowLineAppendedBeforeStartPrintsOnce(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	writeDaily(t, home, today, eventLine(t, today, map[string]any{"msg": "old"}))
	_, offset, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	appendDaily(t, home, today, eventLine(t, today, map[string]any{"msg": "between"})+"\n")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	events, errors := reader.Follow(ctx, offset)

	if got := eventMessage(nextEvent(t, events, time.Second)); got != "between" {
		t.Fatalf("message = %q; want between", got)
	}
	select {
	case event := <-events:
		t.Fatalf("duplicate event %q", eventMessage(event))
	case <-time.After(600 * time.Millisecond):
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowKeepsPartialLineUntilNewline(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	line := eventLine(t, today, map[string]any{"msg": "split"})
	firstHalf := line[:len(line)/2]
	secondHalf := line[len(line)/2:]
	if err := os.WriteFile(dailyPath(home, today), []byte(firstHalf), 0o644); err != nil {
		t.Fatal(err)
	}
	_, offset, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	events, errors := reader.Follow(ctx, offset)

	select {
	case event := <-events:
		t.Fatalf("partial line printed early: %q", eventMessage(event))
	case <-time.After(600 * time.Millisecond):
	}
	appendDaily(t, home, today, secondHalf)
	select {
	case event := <-events:
		t.Fatalf("line without newline printed early: %q", eventMessage(event))
	case <-time.After(600 * time.Millisecond):
	}
	appendDaily(t, home, today, "\n")
	if got := eventMessage(nextEvent(t, events, time.Second)); got != "split" {
		t.Fatalf("message = %q; want split", got)
	}
	select {
	case event := <-events:
		t.Fatalf("duplicate event %q", eventMessage(event))
	case <-time.After(600 * time.Millisecond):
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowMovesToNextDailyFile(t *testing.T) {
	clock := &testClock{}
	dayOne := localNoon(time.Now())
	dayTwo := dayOne.AddDate(0, 0, 1)
	clock.set(dayOne)
	reader, home := newTestReader(t, clock.now())
	reader.now = clock.now
	writeDaily(t, home, dayOne, eventLine(t, dayOne, map[string]any{"msg": "old"}))
	_, offset, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	events, errors := reader.Follow(ctx, offset)

	appendDaily(t, home, dayOne, eventLine(t, dayOne, map[string]any{"msg": "rest"})+"\n")
	clock.set(dayTwo)
	writeDaily(t, home, dayTwo, eventLine(t, dayTwo, map[string]any{"msg": "new day"}))

	if got := eventMessage(nextEvent(t, events, 2*time.Second)); got != "rest" {
		t.Fatalf("first message = %q; want rest", got)
	}
	if got := eventMessage(nextEvent(t, events, 2*time.Second)); got != "new day" {
		t.Fatalf("second message = %q; want new day", got)
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowAppliesCutoff(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	cutoff := today.Add(time.Hour)
	writeDaily(t, home, today, eventLine(t, cutoff.Add(-time.Second), map[string]any{"msg": "before"}))

	events, offset, err := reader.Read(context.Background(), cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("initial events = %d; want 0", len(events))
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	followed, errors := reader.Follow(ctx, offset)

	appendDaily(t, home, today, eventLine(t, cutoff.Add(-time.Second), map[string]any{"msg": "appended before"})+"\n")
	appendDaily(t, home, today, eventLine(t, cutoff, map[string]any{"msg": "at"})+"\n")
	appendDaily(t, home, today, eventLine(t, cutoff.Add(time.Second), map[string]any{"msg": "after"})+"\n")

	if got := eventMessage(nextEvent(t, followed, 2*time.Second)); got != "at" {
		t.Fatalf("first message = %q; want at", got)
	}
	if got := eventMessage(nextEvent(t, followed, 2*time.Second)); got != "after" {
		t.Fatalf("second message = %q; want after", got)
	}
	select {
	case event := <-followed:
		t.Fatalf("unexpected event %q", eventMessage(event))
	case <-time.After(600 * time.Millisecond):
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowAppliesCutoffAfterToday(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	cutoff := localNoon(today.AddDate(0, 0, 1))
	writeDaily(t, home, today, eventLine(t, cutoff.Add(-2*time.Second), map[string]any{"msg": "before"}))

	events, offset, err := reader.Read(context.Background(), cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("initial events = %d; want 0", len(events))
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	followed, errors := reader.Follow(ctx, offset)

	appendDaily(t, home, today, eventLine(t, cutoff.Add(-time.Second), map[string]any{"msg": "appended before"})+"\n")
	appendDaily(t, home, today, eventLine(t, cutoff, map[string]any{"msg": "at"})+"\n")
	appendDaily(t, home, today, eventLine(t, cutoff.Add(time.Second), map[string]any{"msg": "after"})+"\n")

	if got := eventMessage(nextEvent(t, followed, 2*time.Second)); got != "at" {
		t.Fatalf("first message = %q; want at", got)
	}
	if got := eventMessage(nextEvent(t, followed, 2*time.Second)); got != "after" {
		t.Fatalf("second message = %q; want after", got)
	}
	select {
	case event := <-followed:
		t.Fatalf("unexpected event %q", eventMessage(event))
	case <-time.After(600 * time.Millisecond):
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowUsesAbsoluteLineNumbers(t *testing.T) {
	today := localNoon(time.Now())
	reader, home := newTestReader(t, today)
	writeDaily(t, home, today,
		eventLine(t, today, map[string]any{"msg": "one"}),
		eventLine(t, today, map[string]any{"msg": "two"}),
	)
	_, offset, err := reader.Read(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	events, errors := reader.Follow(ctx, offset)

	appendDaily(t, home, today, "not json\n")
	event := nextEvent(t, events, time.Second)
	if !event.Malformed() {
		t.Fatalf("event = %#v; want malformed", event)
	}
	if event.Line() != 3 {
		t.Fatalf("line = %d; want 3", event.Line())
	}
	cancelFollow(t, cancel, errors)
}

func TestFollowCancelledContextReturnsNilError(t *testing.T) {
	today := localNoon(time.Now())
	reader, _ := newTestReader(t, today)
	ctx, cancel := context.WithCancel(context.Background())
	events, errors := reader.Follow(ctx, Offset{})
	cancel()
	select {
	case err, ok := <-errors:
		if ok && err != nil {
			t.Fatalf("error = %v; want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled follower did not stop")
	}
	for range events {
	}
}
