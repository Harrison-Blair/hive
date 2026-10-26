package memory

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadMemoryFileBoundary(t *testing.T) {
	dir := t.TempDir()
	for _, size := range []int{1048575, 1048576} {
		path := filepath.Join(dir, "ok.md")
		if err := writeFileFixture(t, path, bytes.Repeat([]byte("a"), size)); err != nil {
			t.Fatal(err)
		}
		data, err := readMemoryFile(path)
		if err != nil {
			t.Fatalf("readMemoryFile(%d bytes) error = %v; want success", size, err)
		}
		if len(data) != size {
			t.Fatalf("readMemoryFile(%d bytes) length = %d", size, len(data))
		}
	}
	path := filepath.Join(dir, "large.md")
	if err := writeFileFixture(t, path, bytes.Repeat([]byte("a"), 1048577)); err != nil {
		t.Fatal(err)
	}
	if _, err := readMemoryFile(path); err == nil || err.Error() != "file too large" {
		t.Fatalf("readMemoryFile(1048577 bytes) error = %v; want file too large", err)
	}
}

func writeFileFixture(t *testing.T, path string, data []byte) error {
	t.Helper()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
