package store

import (
	"context"
	"encoding/json"
	"errors"
)

var ErrLockTimeout = errors.New("store lock timeout")
var ErrLockReleased = errors.New("store lock released")

type StoredRecord struct {
	Collection []string
	Name       string
	Data       json.RawMessage
}

type LockedStore struct{}

func (s *Store) WithLock(ctx context.Context, fn func(*LockedStore) error) error {
	return nil
}

func (s *Store) WalkRecords(ctx context.Context, fn func(StoredRecord) error) error {
	return nil
}

func (s *Store) ReadIn(collection []string, id string, target any) error {
	return nil
}

func (s *Store) WriteIn(collection []string, id string, value any) error {
	return nil
}

func (s *Store) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	return nil
}

func (s *Store) ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error) {
	return nil, nil
}

func (l *LockedStore) Read(kind, id string, target any) error {
	return nil
}

func (l *LockedStore) Write(kind, id string, value any) error {
	return nil
}

func (l *LockedStore) Update(kind, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	return nil
}

func (l *LockedStore) List(kind string) ([]json.RawMessage, error) {
	return nil, nil
}

func (l *LockedStore) Next(name string) (uint64, error) {
	return 0, nil
}

func (l *LockedStore) ReadIn(collection []string, id string, target any) error {
	return nil
}

func (l *LockedStore) WriteIn(collection []string, id string, value any) error {
	return nil
}

func (l *LockedStore) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	return nil
}

func (l *LockedStore) ListIn(collection []string, include func(filename string) bool) ([]json.RawMessage, error) {
	return nil, nil
}
