package store

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	writeLimitBelow    = 1048575
	writeLimitAt       = 1048576
	writeLimitAbove    = 1048577
	writeLimitFarAbove = 2097152

	writeLimitFrame        = "{\n  \"padding\": \"\"\n}\n"
	writeLimitEscapePrefix = "{\n  \"escape\": \"\\u0001\",\n  \"padding\": \""
	writeLimitEscapeSuffix = "\",\n  \"unicode\": \"\u00e9\"\n}\n"
)

func writeLimitRecordData(t *testing.T, size int) []byte {
	t.Helper()
	padding := size - len(writeLimitFrame)
	if padding < 0 {
		t.Fatalf("encoded size %d is smaller than the frame %d", size, len(writeLimitFrame))
	}
	data := []byte(writeLimitFrame[:len(writeLimitFrame)-len("\"\n}\n")] + strings.Repeat("x", padding) + "\"\n}\n")
	if len(data) != size {
		t.Fatalf("fixture encoded size = %d; want %d", len(data), size)
	}
	return data
}

func writeLimitRecordValue(t *testing.T, size int) map[string]string {
	t.Helper()
	padding := size - len(writeLimitFrame)
	if padding < 0 {
		t.Fatalf("encoded size %d is smaller than the frame %d", size, len(writeLimitFrame))
	}
	return map[string]string{"padding": strings.Repeat("x", padding)}
}

func writeLimitPath(store *Store, kind, id string) string {
	return filepath.Join(store.home, kind, id+".json")
}

func TestWriteLimitAcceptsEncodedBoundary(t *testing.T) {
	for _, test := range []struct {
		name string
		size int
	}{
		{"below", writeLimitBelow},
		{"exact", writeLimitAt},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := openStore(t, t.TempDir())
			value := writeLimitRecordValue(t, test.size)
			if err := store.WriteIn([]string{"tasks"}, "generic", value); err != nil {
				t.Fatalf("WriteIn() rejected %d encoded bytes: %v", test.size, err)
			}
			pathIn := writeLimitPath(store, "tasks", "generic")
			dataIn, err := os.ReadFile(pathIn)
			if err != nil {
				t.Fatal(err)
			}
			if len(dataIn) != test.size {
				t.Fatalf("WriteIn() encoded %d bytes; want %d", len(dataIn), test.size)
			}
			var generic map[string]string
			if err := store.ReadIn([]string{"tasks"}, "generic", &generic); err != nil {
				t.Fatalf("ReadIn() rejected the boundary record: %v", err)
			}
			if len(generic["padding"]) != test.size-len(writeLimitFrame) {
				t.Fatalf("ReadIn() padding = %d bytes; want %d", len(generic["padding"]), test.size-len(writeLimitFrame))
			}
		})
	}
}

func TestWriteLimitRejectsEncodedCreate(t *testing.T) {
	for _, test := range []struct {
		name string
		size int
	}{
		{"one byte above", writeLimitAbove},
		{"two MiB", writeLimitFarAbove},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := openStore(t, t.TempDir())
			value := writeLimitRecordValue(t, test.size)
			data := writeLimitRecordData(t, test.size)
			update := func(json.RawMessage) (json.RawMessage, error) {
				return json.RawMessage(data), nil
			}
			for _, test := range []struct {
				name  string
				id    string
				path  string
				write func() error
			}{
				{
					name: "generic WriteIn",
					id:   "generic-write",
					write: func() error {
						return store.WriteIn([]string{"tasks"}, "generic-write", value)
					},
				},
				{
					name: "generic UpdateIn",
					id:   "generic-update",
					write: func() error {
						return store.UpdateIn([]string{"tasks"}, "generic-update", update)
					},
				},
			} {
				path := writeLimitPath(store, "tasks", test.id)
				err := test.write()
				want := "record too large: " + path
				if err == nil || err.Error() != want {
					t.Errorf("%s error = %v; want %q", test.name, err, want)
					continue
				}
				if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
					t.Errorf("%s made a record at %s: %v", test.name, path, statErr)
				}
			}
		})
	}
}

func TestWriteLimitPreservesRecordOnRejectedReplacement(t *testing.T) {
	data := writeLimitRecordData(t, writeLimitAbove)
	update := func(json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(data), nil
	}
	value := writeLimitRecordValue(t, writeLimitAbove)
	for _, test := range []struct {
		name  string
		id    string
		path  string
		write func(*Store) error
	}{
		{
			name: "generic WriteIn",
			id:   "generic-write",
			write: func(store *Store) error {
				return store.WriteIn([]string{"tasks"}, "generic-write", value)
			},
		},
		{
			name: "generic UpdateIn",
			id:   "generic-update",
			write: func(store *Store) error {
				return store.UpdateIn([]string{"tasks"}, "generic-update", update)
			},
		},
		{
			name: "locked WriteIn",
			id:   "locked-write-in",
			write: func(store *Store) error {
				return store.WithLock(context.Background(), func(locked *LockedStore) error {
					return locked.WriteIn([]string{"tasks"}, "locked-write-in", value)
				})
			},
		},
		{
			name: "locked UpdateIn",
			id:   "locked-update-in",
			write: func(store *Store) error {
				return store.WithLock(context.Background(), func(locked *LockedStore) error {
					return locked.UpdateIn([]string{"tasks"}, "locked-update-in", update)
				})
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := openStore(t, t.TempDir())
			if err := store.WriteIn([]string{"tasks"}, test.id, map[string]string{"value": "old"}); err != nil {
				t.Fatal(err)
			}
			path := writeLimitPath(store, "tasks", test.id)
			old, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			err = test.write(store)
			want := "record too large: " + path
			if err == nil || err.Error() != want {
				t.Fatalf("%s error = %v; want %q", test.name, err, want)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(old, after) {
				t.Fatalf("%s changed the old record: got %d bytes, want %d", test.name, len(after), len(old))
			}
			var record map[string]string
			if err := store.ReadIn([]string{"tasks"}, test.id, &record); err != nil {
				t.Fatalf("%s left the old record unreadable: %v", test.name, err)
			}
			if record["value"] != "old" {
				t.Fatalf("%s old value = %q; want old", test.name, record["value"])
			}
		})
	}
}

func TestWriteLimitKeepsEarlierGroupedWrite(t *testing.T) {
	store := openStore(t, t.TempDir())
	value := writeLimitRecordValue(t, writeLimitAbove)
	var writeErr error
	err := store.WithLock(context.Background(), func(locked *LockedStore) error {
		if err := locked.WriteIn([]string{"tasks"}, "first", map[string]string{"value": "small"}); err != nil {
			return err
		}
		writeErr = locked.WriteIn([]string{"tasks"}, "second", value)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	second := writeLimitPath(store, "tasks", "second")
	want := "record too large: " + second
	if writeErr == nil || writeErr.Error() != want {
		t.Fatalf("later write error = %v; want %q", writeErr, want)
	}
	if _, statErr := os.Lstat(second); !os.IsNotExist(statErr) {
		t.Fatalf("later write made a record at %s: %v", second, statErr)
	}
	var first map[string]string
	if err := store.ReadIn([]string{"tasks"}, "first", &first); err != nil {
		t.Fatalf("earlier write is unreadable: %v", err)
	}
	if first["value"] != "small" {
		t.Fatalf("earlier value = %q; want small", first["value"])
	}
}

func TestWriteLimitCountsEncodedEscapesAndUTF8(t *testing.T) {
	store := openStore(t, t.TempDir())
	padding := writeLimitAt - len(writeLimitEscapePrefix) - len(writeLimitEscapeSuffix)
	if padding < 0 {
		t.Fatalf("encoded size %d is smaller than the frame %d", writeLimitAt, len(writeLimitEscapePrefix)+len(writeLimitEscapeSuffix))
	}
	value := map[string]string{
		"escape":  "\x01",
		"padding": strings.Repeat("x", padding),
		"unicode": "\u00e9",
	}
	if err := store.WriteIn([]string{"tasks"}, "escaped", value); err != nil {
		t.Fatalf("WriteIn() rejected %d encoded bytes: %v", writeLimitAt, err)
	}
	path := writeLimitPath(store, "tasks", "escaped")
	old, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := writeLimitEscapePrefix + strings.Repeat("x", padding) + writeLimitEscapeSuffix
	if string(old) != want {
		t.Fatalf("encoded bytes = %d, want %d; escape and UTF-8 framing differs", len(old), len(want))
	}
	if len(old) != writeLimitAt {
		t.Fatalf("encoded bytes = %d; want %d", len(old), writeLimitAt)
	}
	var record map[string]string
	if err := store.ReadIn([]string{"tasks"}, "escaped", &record); err != nil {
		t.Fatalf("ReadIn() rejected the escaped boundary record: %v", err)
	}
	if record["escape"] != "\x01" {
		t.Fatalf("ReadIn() escape = %q; want the control byte", record["escape"])
	}
	if record["unicode"] != "\u00e9" {
		t.Fatalf("ReadIn() unicode = %q; want the UTF-8 rune", record["unicode"])
	}

	value["padding"] = strings.Repeat("x", padding+1)
	err = store.WriteIn([]string{"tasks"}, "escaped", value)
	wantErr := "record too large: " + path
	if err == nil || err.Error() != wantErr {
		t.Fatalf("WriteIn() error = %v; want %q", err, wantErr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(old, after) {
		t.Fatalf("rejected write changed the old record: got %d bytes, want %d", len(after), len(old))
	}
}
