package memory

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	storelib "github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestNewMemoryWritesFileAndIndex(t *testing.T) {
	home := t.TempDir()
	input := memoryContent("base-is-dev", "The base branch of hive is dev.", "The base branch of hive is dev.")
	envelope := writeTestMemory(t, home, "base-is-dev", input)
	if envelope.Outcome != "ok" || envelope.Phase != "complete" || envelope.Error != nil {
		t.Fatalf("Write() = %#v; want success", envelope)
	}
	if len(envelope.Effects) != 2 {
		t.Fatalf("effects = %#v; want memory and index", envelope.Effects)
	}
	wantMemory := output.Effect{Resource: "memory", Path: "memory/base-is-dev.md", Action: "created"}
	if envelope.Effects[0] != wantMemory {
		t.Fatalf("memory effect = %#v; want %#v", envelope.Effects[0], wantMemory)
	}
	wantIndex := output.Effect{Resource: "index", Path: "memory/MEMORY.md", Action: "updated"}
	if envelope.Effects[1] != wantIndex {
		t.Fatalf("index effect = %#v; want %#v", envelope.Effects[1], wantIndex)
	}
	gotFile, err := os.ReadFile(filepath.Join(home, "memory", "base-is-dev.md"))
	if err != nil || string(gotFile) != input {
		t.Fatalf("memory file = %q, %v; want exact input", gotFile, err)
	}
	gotIndex, err := os.ReadFile(filepath.Join(home, "memory", "MEMORY.md"))
	if err != nil || string(gotIndex) != "- [base-is-dev](base-is-dev.md): The base branch of hive is dev.\n" {
		t.Fatalf("index = %q, %v", gotIndex, err)
	}
}

func TestReplacementReportsUpdated(t *testing.T) {
	home := t.TempDir()
	first := memoryContent("fact", "First fact.", "First body.")
	if envelope := writeTestMemory(t, home, "fact", first); envelope.Outcome != "ok" {
		t.Fatalf("first Write() = %#v", envelope)
	}
	second := memoryContent("fact", "Second fact.", "Second body.")
	envelope := writeTestMemory(t, home, "fact", second)
	if envelope.Outcome != "ok" {
		t.Fatalf("second Write() = %#v", envelope)
	}
	if len(envelope.Effects) != 2 || envelope.Effects[0].Action != "updated" {
		t.Fatalf("effects = %#v; want memory updated", envelope.Effects)
	}
	got, err := os.ReadFile(filepath.Join(home, "memory", "fact.md"))
	if err != nil || string(got) != second {
		t.Fatalf("memory file = %q, %v; want replacement", got, err)
	}
	gotIndex, err := os.ReadFile(filepath.Join(home, "memory", "MEMORY.md"))
	if err != nil || string(gotIndex) != "- [fact](fact.md): Second fact.\n" {
		t.Fatalf("index = %q, %v; want replacement description", gotIndex, err)
	}
}

func TestTwoMemoriesIndexNameOrder(t *testing.T) {
	home := t.TempDir()
	if envelope := writeTestMemory(t, home, "zulu", memoryContent("zulu", "Z.", "Z body.")); envelope.Outcome != "ok" {
		t.Fatalf("zulu Write() = %#v", envelope)
	}
	if envelope := writeTestMemory(t, home, "alpha", memoryContent("alpha", "A.", "A body.")); envelope.Outcome != "ok" {
		t.Fatalf("alpha Write() = %#v", envelope)
	}
	got, err := os.ReadFile(filepath.Join(home, "memory", "MEMORY.md"))
	want := "- [alpha](alpha.md): A.\n- [zulu](zulu.md): Z.\n"
	if err != nil || string(got) != want {
		t.Fatalf("index = %q, %v; want %q", got, err, want)
	}
}

func TestExistingInvalidFileStopsWrite(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(memoryDir, "bad.md")
	if err := writeFileFixture(t, badPath, []byte("---\nname: bad\n---\n")); err != nil {
		t.Fatal(err)
	}
	oldIndex := []byte("old index\n")
	if err := writeFileFixture(t, filepath.Join(memoryDir, "MEMORY.md"), oldIndex); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "new", memoryContent("new", "New fact.", "New body."))
	want := "memory file invalid: " + badPath + ": frontmatter: missing description"
	if envelope.Outcome != "failed" || envelope.Phase != "scan" || envelope.Error != want {
		t.Fatalf("Write() = %#v; want %q", envelope, want)
	}
	if _, err := os.Stat(filepath.Join(memoryDir, "new.md")); !os.IsNotExist(err) {
		t.Fatalf("new memory exists after invalid file: stat error = %v", err)
	}
	gotIndex, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md"))
	if err != nil || !bytes.Equal(gotIndex, oldIndex) {
		t.Fatalf("index = %q, %v; want old index", gotIndex, err)
	}
}

func TestExistingOversizedFileStopsWrite(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(memoryDir, "large.md")
	if err := writeFileFixture(t, badPath, bytes.Repeat([]byte("a"), 1048577)); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "new", memoryContent("new", "New fact.", "New body."))
	if envelope.Outcome != "failed" || !strings.HasPrefix(envelope.Error.(string), "memory file invalid: "+badPath+": ") {
		t.Fatalf("Write() = %#v; want memory file invalid", envelope)
	}
	if _, err := os.Stat(filepath.Join(memoryDir, "new.md")); !os.IsNotExist(err) {
		t.Fatalf("new memory exists after oversized file: stat error = %v", err)
	}
}

func TestMemoryNameValidation(t *testing.T) {
	names := []string{
		"",
		"Base",
		"base_is",
		" base",
		"base ",
		"base/dev",
		"..",
		"%2e%2e%2f",
		"quote'",
		"#",
		"?",
		"%",
		"&",
		";",
		"+",
		"lé",
		"bad\nname",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			envelope := writeTestMemory(t, home, name, memoryContent("fact", "A fact.", "Body."))
			if envelope.Outcome != "failed" {
				t.Fatalf("Write(%q) = %#v; want failure", name, envelope)
			}
			if strings.Contains(envelope.Error.(string), "\n") {
				t.Fatalf("Write(%q) error = %q; want one physical line", name, envelope.Error)
			}
			if _, err := os.Stat(filepath.Join(home, "memory")); !os.IsNotExist(err) {
				t.Fatalf("Write(%q) made a memory directory: stat error = %v", name, err)
			}
		})
	}
}

func TestVeryLongMemoryNameFailsWithoutPanic(t *testing.T) {
	home := t.TempDir()
	name := strings.Repeat("a", 10000)
	envelope := writeTestMemory(t, home, name, memoryContent(name, "A fact.", "Body."))
	if envelope.Outcome != "failed" {
		t.Fatalf("Write(long name) = %#v; want failure", envelope)
	}
	if strings.Contains(envelope.Error.(string), "\n") {
		t.Fatalf("Write(long name) error has more than one line")
	}
}

func TestStaticIndexDirectoryRefusedBeforeWrite(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	indexPath := filepath.Join(memoryDir, "MEMORY.md")
	if err := os.MkdirAll(indexPath, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "new", memoryContent("new", "New fact.", "New body."))
	if envelope.Outcome != "failed" {
		t.Fatalf("Write() = %#v; want refusal", envelope)
	}
	if info, err := os.Lstat(indexPath); err != nil || !info.IsDir() {
		t.Fatalf("MEMORY.md = %#v, %v; want the directory kept", info, err)
	}
	if _, err := os.Stat(filepath.Join(memoryDir, "new.md")); !os.IsNotExist(err) {
		t.Fatalf("new memory exists after index refusal: stat error = %v", err)
	}
}

func TestStaticIndexLinkRefusedBeforeWrite(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "index.md")
	old := []byte("old\n")
	if err := writeFileFixture(t, target, old); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(memoryDir, "MEMORY.md")
	if err := os.Symlink(target, indexPath); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "new", memoryContent("new", "New fact.", "New body."))
	if envelope.Outcome != "failed" {
		t.Fatalf("Write() = %#v; want refusal", envelope)
	}
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, old) {
		t.Fatalf("index target = %q, %v; want old bytes", got, err)
	}
	if info, err := os.Lstat(indexPath); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("MEMORY.md = %#v, %v; want the symbolic link kept", info, err)
	}
	if _, err := os.Stat(filepath.Join(memoryDir, "new.md")); !os.IsNotExist(err) {
		t.Fatalf("new memory exists after index refusal: stat error = %v", err)
	}
}

func TestStaticMemoryDirectoryLinkRefused(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, "memory")); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "new", memoryContent("new", "New fact.", "New body."))
	if envelope.Outcome != "failed" {
		t.Fatalf("Write() = %#v; want refusal", envelope)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("outside directory entries = %#v, %v; want no file", entries, err)
	}
}

func TestStaticLeafLinkRefused(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "fact.md")
	old := []byte("---\nname: fact\ndescription: Outside fact.\n---\nOutside body.\n")
	if err := writeFileFixture(t, target, old); err != nil {
		t.Fatal(err)
	}
	leaf := filepath.Join(memoryDir, "fact.md")
	if err := os.Symlink(target, leaf); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "fact", memoryContent("fact", "A fact.", "Body."))
	if envelope.Outcome != "failed" {
		t.Fatalf("Write() = %#v; want refusal", envelope)
	}
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, old) {
		t.Fatalf("leaf target = %q, %v; want old bytes", got, err)
	}
	if info, err := os.Lstat(leaf); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("leaf = %#v, %v; want the symbolic link kept", info, err)
	}
}

func TestPartialIndexForNewMemory(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldIndex := []byte("old index\n")
	if err := writeFileFixture(t, filepath.Join(memoryDir, "MEMORY.md"), oldIndex); err != nil {
		t.Fatal(err)
	}
	input := memoryContent("new", "New fact.", "New body.")
	envelope := Write(home, WriteOptions{
		Scope:      "home",
		Name:       "new",
		Input:      strings.NewReader(input),
		Deadline:   time.Second,
		IndexWrite: func(string, []byte) error { return errors.New("injected index failure") },
	})
	if envelope.Outcome != "partial" || envelope.Phase != "index" {
		t.Fatalf("Write() = %#v; want partial index", envelope)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0].Action != "created" || envelope.Effects[0].Resource != "memory" {
		t.Fatalf("effects = %#v; want one created memory effect", envelope.Effects)
	}
	if envelope.Error != "injected index failure" {
		t.Fatalf("error = %v; want injected failure", envelope.Error)
	}
	if got, err := os.ReadFile(filepath.Join(memoryDir, "new.md")); err != nil || string(got) != input {
		t.Fatalf("memory file = %q, %v; want retained fact", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md")); err != nil || !bytes.Equal(got, oldIndex) {
		t.Fatalf("index = %q, %v; want old index", got, err)
	}

	green := writeTestMemory(t, home, "new", memoryContent("new", "Regenerated fact.", "New body."))
	if green.Outcome != "ok" {
		t.Fatalf("later Write() = %#v; want success", green)
	}
	got, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md"))
	if err != nil || string(got) != "- [new](new.md): Regenerated fact.\n" {
		t.Fatalf("regenerated index = %q, %v", got, err)
	}
}

func TestPartialIndexForReplacement(t *testing.T) {
	home := t.TempDir()
	first := memoryContent("fact", "First fact.", "First body.")
	if envelope := writeTestMemory(t, home, "fact", first); envelope.Outcome != "ok" {
		t.Fatalf("first Write() = %#v", envelope)
	}
	oldIndex, err := os.ReadFile(filepath.Join(home, "memory", "MEMORY.md"))
	if err != nil {
		t.Fatal(err)
	}
	second := memoryContent("fact", "Second fact.", "Second body.")
	envelope := Write(home, WriteOptions{
		Scope:      "home",
		Name:       "fact",
		Input:      strings.NewReader(second),
		Deadline:   time.Second,
		IndexWrite: func(string, []byte) error { return errors.New("injected index failure") },
	})
	if envelope.Outcome != "partial" || envelope.Phase != "index" {
		t.Fatalf("Write() = %#v; want partial index", envelope)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0].Action != "updated" {
		t.Fatalf("effects = %#v; want one updated memory effect", envelope.Effects)
	}
	if got, err := os.ReadFile(filepath.Join(home, "memory", "fact.md")); err != nil || string(got) != second {
		t.Fatalf("memory file = %q, %v; want replacement", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(home, "memory", "MEMORY.md")); err != nil || !bytes.Equal(got, oldIndex) {
		t.Fatalf("index = %q, %v; want old index", got, err)
	}
}

func TestMemoryOversizedInputWritesNothing(t *testing.T) {
	home := t.TempDir()
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Name:     "fact",
		Input:    strings.NewReader(strings.Repeat("x", 1048577)),
		Deadline: time.Second,
	})
	if envelope.Outcome != "failed" || envelope.Phase != "input" || envelope.Error != "input too large" {
		t.Fatalf("Write() = %#v; want input too large", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, "memory")); !os.IsNotExist(err) {
		t.Fatalf("memory directory exists after oversized input: stat error = %v", err)
	}
}

func TestMemoryTimeoutWritesNothing(t *testing.T) {
	home := t.TempDir()
	reader, writer := io.Pipe()
	defer func() { _ = reader.Close() }()
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Name:     "fact",
		Input:    reader,
		Deadline: 100 * time.Millisecond,
	})
	_ = writer.Close()
	if envelope.Outcome != "failed" || envelope.Phase != "input" || envelope.Error != "input timeout" {
		t.Fatalf("Write() = %#v; want input timeout", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, "memory")); !os.IsNotExist(err) {
		t.Fatalf("memory directory exists after timeout: stat error = %v", err)
	}
}

func TestWriteLogsEffectsWithoutContent(t *testing.T) {
	home := t.TempDir()
	var log bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	input := memoryContent("fact", "Secret description.", "Secret body.")
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Name:     "fact",
		Input:    strings.NewReader(input),
		Deadline: time.Second,
		Logger:   logger,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("Write() = %#v", envelope)
	}
	text := log.String()
	if !strings.Contains(text, `"resource":"memory"`) || !strings.Contains(text, `"action":"created"`) {
		t.Fatalf("log = %q; want completed memory effect", text)
	}
	if strings.Contains(text, "Secret") {
		t.Fatalf("log = %q; must not contain memory content", text)
	}
}

func TestInvalidSeparatorStdinStopsWrite(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"missing name", "---\nname:fact\ndescription: D.\n---\nBody.\n", "frontmatter: missing name"},
		{"missing description", "---\nname: fact\ndescription:D.\n---\nBody.\n", "frontmatter: missing description"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			memoryDir := filepath.Join(home, "memory")
			if err := os.MkdirAll(memoryDir, 0o755); err != nil {
				t.Fatal(err)
			}
			oldIndex := []byte("old index\n")
			if err := writeFileFixture(t, filepath.Join(memoryDir, "MEMORY.md"), oldIndex); err != nil {
				t.Fatal(err)
			}
			envelope := writeTestMemory(t, home, "fact", tc.input)
			if envelope.Outcome != "failed" || envelope.Phase != "frontmatter" || envelope.Error != tc.want {
				t.Fatalf("Write() = %#v; want %q", envelope, tc.want)
			}
			if _, err := os.Stat(filepath.Join(memoryDir, "fact.md")); !os.IsNotExist(err) {
				t.Fatalf("fact exists after invalid separator: stat error = %v", err)
			}
			gotIndex, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md"))
			if err != nil || !bytes.Equal(gotIndex, oldIndex) {
				t.Fatalf("index = %q, %v; want old index", gotIndex, err)
			}
		})
	}
}

func TestExistingInvalidSeparatorStopsWrite(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(memoryDir, "bad.md")
	if err := writeFileFixture(t, badPath, []byte("---\nname:bad\ndescription:D.\n---\nBad body.\n")); err != nil {
		t.Fatal(err)
	}
	oldIndex := []byte("old index\n")
	if err := writeFileFixture(t, filepath.Join(memoryDir, "MEMORY.md"), oldIndex); err != nil {
		t.Fatal(err)
	}
	envelope := writeTestMemory(t, home, "new", memoryContent("new", "New fact.", "New body."))
	want := "memory file invalid: " + badPath + ": frontmatter: missing name"
	if envelope.Outcome != "failed" || envelope.Phase != "scan" || envelope.Error != want {
		t.Fatalf("Write() = %#v; want %q", envelope, want)
	}
	if _, err := os.Stat(filepath.Join(memoryDir, "new.md")); !os.IsNotExist(err) {
		t.Fatalf("new memory exists after invalid existing file: stat error = %v", err)
	}
	gotIndex, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md"))
	if err != nil || !bytes.Equal(gotIndex, oldIndex) {
		t.Fatalf("index = %q, %v; want old index", gotIndex, err)
	}
}

func TestWriteLogsSafeFrontmatterCategory(t *testing.T) {
	home := t.TempDir()
	var log bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sentinel := "T07FIX01SYNTHETICNAME"
	bodySentinel := "T07FIX01SYNTHETICBODY"
	input := memoryContent(sentinel, "Description.", bodySentinel)
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Name:     "fact",
		Input:    strings.NewReader(input),
		Deadline: time.Second,
		Logger:   logger,
	})
	want := "frontmatter: name " + sentinel + " differs from fact"
	if envelope.Outcome != "failed" || envelope.Phase != "frontmatter" || envelope.Error != want {
		t.Fatalf("Write() = %#v; want user error %q", envelope, want)
	}
	text := log.String()
	if strings.Contains(text, sentinel) || strings.Contains(text, bodySentinel) {
		t.Fatalf("log = %q; must not contain parsed values or content", text)
	}
	if !strings.Contains(text, `"msg":"error"`) || !strings.Contains(text, `"phase":"frontmatter"`) || !strings.Contains(text, `"error":"frontmatter: name differs"`) {
		t.Fatalf("log = %q; want safe frontmatter error event", text)
	}
}

func TestExistingInvalidFileLogsNoParsedValues(t *testing.T) {
	home := t.TempDir()
	memoryDir := filepath.Join(home, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := "T07FIX01EXISTINGSYNTHETICNAME"
	bodySentinel := "T07FIX01EXISTINGSYNTHETICBODY"
	badPath := filepath.Join(memoryDir, "bad.md")
	if err := writeFileFixture(t, badPath, []byte("---\nname: "+sentinel+"\ndescription: D.\n---\n"+bodySentinel+"\n")); err != nil {
		t.Fatal(err)
	}
	var log bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Name:     "new",
		Input:    strings.NewReader(memoryContent("new", "New fact.", "New body.")),
		Deadline: time.Second,
		Logger:   logger,
	})
	if envelope.Outcome != "failed" || envelope.Phase != "scan" || !strings.HasPrefix(envelope.Error.(string), "memory file invalid: "+badPath+": ") {
		t.Fatalf("Write() = %#v; want memory file invalid", envelope)
	}
	text := log.String()
	if strings.Contains(text, sentinel) || strings.Contains(text, bodySentinel) {
		t.Fatalf("log = %q; must not contain parsed values or content", text)
	}
	if !strings.Contains(text, `"msg":"error"`) || !strings.Contains(text, `"phase":"scan"`) || !strings.Contains(text, "memory file invalid: ") {
		t.Fatalf("log = %q; want safe memory file error event", text)
	}
}

func TestWriteLogsProjectID(t *testing.T) {
	home := t.TempDir()
	registerLogProject(t, home, "demo")
	var log bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	envelope := Write(home, WriteOptions{
		Scope:    "project:demo",
		Name:     "fact",
		Input:    strings.NewReader(memoryContent("fact", "A fact.", "Body.")),
		Deadline: time.Second,
		Logger:   logger,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("Write() = %#v; want success", envelope)
	}
	failure := Write(home, WriteOptions{
		Scope:    "project:demo",
		Name:     "other",
		Input:    strings.NewReader("---\nname: other\n---\n"),
		Deadline: time.Second,
		Logger:   logger,
	})
	if failure.Outcome != "failed" || failure.Phase != "frontmatter" {
		t.Fatalf("Write() = %#v; want frontmatter failure", failure)
	}
	partial := Write(home, WriteOptions{
		Scope:      "project:demo",
		Name:       "third",
		Input:      strings.NewReader(memoryContent("third", "Third fact.", "Body.")),
		Deadline:   time.Second,
		Logger:     logger,
		IndexWrite: func(string, []byte) error { return errors.New("injected index failure") },
	})
	if partial.Outcome != "partial" || partial.Phase != "index" {
		t.Fatalf("Write() = %#v; want partial index", partial)
	}
	projectEvents := 0
	for _, event := range logEvents(t, log.String()) {
		if event["scope"] != "project:demo" {
			continue
		}
		projectEvents++
		if event["project_id"] != "demo" {
			t.Fatalf("event = %#v; want project_id demo", event)
		}
	}
	if projectEvents < 4 {
		t.Fatalf("project events = %d; want memory, index, frontmatter, and index error events", projectEvents)
	}
}

func logEvents(t *testing.T, text string) []map[string]string {
	t.Helper()
	events := []map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		if line == "" {
			continue
		}
		var event map[string]string
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line = %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func registerLogProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := storelib.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("projects", id, records.Project{ID: id, Path: home, Base: "dev"}); err != nil {
		t.Fatal(err)
	}
}

func writeTestMemory(t *testing.T, home, name, input string) output.Envelope {
	t.Helper()
	return Write(home, WriteOptions{
		Scope:    "home",
		Name:     name,
		Input:    strings.NewReader(input),
		Deadline: time.Second,
	})
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

func memoryContent(name, description, body string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n" + body + "\n"
}
