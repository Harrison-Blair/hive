package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readSchemaFile(t *testing.T, home string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, "schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

func writeSchemaFile(t *testing.T, home, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, "schema.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestOpenWritesSchemaVersionFiveInAFreshHome(t *testing.T) {
	home := t.TempDir()
	openStore(t, home)
	if got := readSchemaFile(t, home); len(got) != 1 || got["version"] != float64(5) {
		t.Fatalf("schema.json = %v; want {\"version\":5}", got)
	}
}

func TestOpenAcceptsSchemaVersionFive(t *testing.T) {
	home := t.TempDir()
	writeSchemaFile(t, home, `{"version":5}`)
	openStore(t, home)
	if got := readSchemaFile(t, home); got["version"] != float64(5) {
		t.Fatalf("schema.json = %v; want version 5", got)
	}
}

func TestOpenRefusesOtherSchemaVersions(t *testing.T) {
	for _, tc := range []struct {
		body string
		want string
	}{
		{`{"version":4}`, "home schema 4 is not supported; this hive needs 5"},
		{`{"version":6}`, "home schema 6 is not supported; this hive needs 5"},
	} {
		t.Run(tc.body, func(t *testing.T) {
			home := t.TempDir()
			writeSchemaFile(t, home, tc.body)
			_, err := Open(home)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("Open() error = %v; want %q", err, tc.want)
			}
			if got := readSchemaFile(t, home); got["version"] == float64(5) {
				t.Fatalf("schema.json = %v; want it unchanged", got)
			}
		})
	}
}
