package log

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// testBodyBound is the literal body budget of the contract. The expected
// values of these tests never read the production constant.
const testBodyBound = 4096

// bodyMode selects where a body attribute enters the handler. A test uses one
// real Writer and one real daily file for each case.
type bodyMode struct {
	name      string
	inherited []slog.Attr
	groups    []string
	attr      slog.Attr
	key       string
}

func ordinaryBodyMode(input string) bodyMode {
	return bodyMode{name: "ordinary Body", attr: Body(input), key: "body"}
}

func inheritedBodyMode(input string) bodyMode {
	return bodyMode{name: "inherited WithAttrs", inherited: []slog.Attr{Body(input)}, key: "body"}
}

func groupedBodyMode(input string) bodyMode {
	return bodyMode{name: "grouped Body", attr: Body(input), groups: []string{"payload"}, key: "payload.body"}
}

func nestedGroupedBodyMode(input string) bodyMode {
	return bodyMode{name: "nested grouped Body", attr: Body(input), groups: []string{"outer", "inner"}, key: "outer.inner.body"}
}

// emitBodyEvent writes one real debug event with NewWriter, slog and Body.
// It then reads the real daily JSONL file and decodes it with encoding/json.
func emitBodyEvent(t *testing.T, mode bodyMode) (map[string]any, []byte) {
	t.Helper()
	home := t.TempDir()
	clock := func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local) }
	writer, err := NewWriter(home, "cli", clock)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(writer)
	for _, group := range mode.groups {
		logger = logger.WithGroup(group)
	}
	if len(mode.inherited) != 0 {
		inherited := make([]any, len(mode.inherited))
		for index, attr := range mode.inherited {
			inherited[index] = attr
		}
		logger = logger.With(inherited...)
		logger.Debug("body event", "command", "body-bound-test", "case", t.Name())
	} else {
		logger.Debug("body event", "command", "body-bound-test", "case", t.Name(), mode.attr)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(home, ".hive", "logs", "2026-01-02.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(raw, []byte{'\n'}) != 1 || !bytes.HasSuffix(raw, []byte{'\n'}) {
		t.Fatalf("the daily file must hold one JSON line: %q", raw)
	}
	var event map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&event); err != nil {
		t.Fatalf("the daily file is not one JSON event: %v", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("the daily file holds more than one JSON value: %v", err)
	}
	return event, raw
}

// standardJSONDecoded returns the string value that encoding/json decodes for
// the input. It is the normal replacement behavior of the encoder, not a copy
// of the production cap.
func standardJSONDecoded(t *testing.T, input string) string {
	t.Helper()
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var decoded string
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

// expectedBodyPrefix makes the expected value from the standard decoded
// sequence: the largest prefix of complete characters within the literal
// budget, and a flag that says if the sequence was shortened.
func expectedBodyPrefix(t *testing.T, input string) (string, bool) {
	t.Helper()
	decoded := standardJSONDecoded(t, input)
	if len(decoded) <= testBodyBound {
		return decoded, false
	}
	cut := 0
	for index := range decoded {
		if index > testBodyBound {
			break
		}
		cut = index
	}
	return decoded[:cut], true
}

func bodySHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// captureBodyObservation keeps the real JSONL line and the observation when
// T01_BODY_BOUND_CAPTURE names a directory. The red run uses it to retain the
// exact failing output.
func captureBodyObservation(t *testing.T, mode bodyMode, input string, event map[string]any, raw []byte, body, want string, cut bool) {
	t.Helper()
	dir := os.Getenv("T01_BODY_BOUND_CAPTURE")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	truncated, present := event["truncated"]
	observation := map[string]any{
		"test":                   t.Name(),
		"mode":                   mode.name,
		"key":                    mode.key,
		"input_bytes":            len(input),
		"input_valid_utf8":       utf8.ValidString(input),
		"input_sha256":           bodySHA256([]byte(input)),
		"standard_decoded_bytes": len(standardJSONDecoded(t, input)),
		"decoded_bytes":          len(body),
		"decoded_valid_utf8":     utf8.ValidString(body),
		"decoded_sha256":         bodySHA256([]byte(body)),
		"expected_bytes":         len(want),
		"expected_sha256":        bodySHA256([]byte(want)),
		"expected_truncated":     cut,
		"truncated_present":      present,
		"truncated":              truncated,
		"body_equal_expected":    body == want,
		"jsonl_bytes":            len(raw),
		"jsonl_sha256":           bodySHA256(raw),
	}
	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	name := strings.NewReplacer("/", "-", " ", "_").Replace(t.Name())
	if err := os.WriteFile(filepath.Join(dir, name+".json"), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".jsonl"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// checkBodyCase checks one input. The expected value comes from the standard
// JSON decode and the literal budget, never from the production helper.
func checkBodyCase(t *testing.T, input string, mode bodyMode) (string, bool) {
	t.Helper()
	event, raw := emitBodyEvent(t, mode)
	want, wantCut := expectedBodyPrefix(t, input)
	value, ok := event[mode.key]
	if !ok {
		t.Fatalf("%s: the event has no %q field: %v", mode.name, mode.key, event)
	}
	body, ok := value.(string)
	if !ok {
		t.Fatalf("%s: the decoded %q value is not a string: %T", mode.name, mode.key, value)
	}
	captureBodyObservation(t, mode, input, event, raw, body, want, wantCut)

	if len(body) > testBodyBound {
		t.Errorf("%s: decoded body has %d bytes; the literal bound is %d", mode.name, len(body), testBodyBound)
	}
	if !utf8.ValidString(body) {
		t.Errorf("%s: decoded body is not valid UTF-8", mode.name)
	}
	if body != want {
		t.Errorf("%s: decoded body differs from the largest complete prefix within %d bytes of the standard JSON decoded input", mode.name, testBodyBound)
	}
	truncated, present := event["truncated"]
	if wantCut {
		if !present || truncated != true {
			t.Errorf("%s: truncated=%v present=%t; want true", mode.name, truncated, present)
		}
	} else if present {
		t.Errorf("%s: the unchanged decoded sequence has truncated=%v", mode.name, truncated)
	}
	if eventTime, ok := event["time"].(string); !ok || !strings.HasPrefix(eventTime, "2026-01-02T03:04:05") {
		t.Errorf("%s: event time = %v; want the injected local time", mode.name, event["time"])
	}
	if event["level"] != "DEBUG" || event["msg"] != "body event" || event["component"] != "cli" {
		t.Errorf("%s: event fields = %#v", mode.name, event)
	}
	prefix := ""
	if mode.key != "body" {
		prefix = strings.TrimSuffix(mode.key, ".body")
	}
	commandKey, caseKey := "command", "case"
	if prefix != "" {
		commandKey = prefix + "." + commandKey
		caseKey = prefix + "." + caseKey
	}
	if event[commandKey] != "body-bound-test" || event[caseKey] != t.Name() {
		t.Errorf("%s: body event lost metadata: %#v", mode.name, event)
	}
	return body, wantCut
}

func TestBodyBoundASCII(t *testing.T) {
	t.Run("below", func(t *testing.T) {
		input := strings.Repeat("a", 4095)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if cut || body != input {
			t.Errorf("a 4095-byte ASCII body changed: cut=%t", cut)
		}
	})
	t.Run("exact", func(t *testing.T) {
		input := strings.Repeat("a", 4096)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if cut || body != input {
			t.Errorf("a 4096-byte ASCII body changed: cut=%t", cut)
		}
	})
	t.Run("above", func(t *testing.T) {
		input := strings.Repeat("a", 5000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != strings.Repeat("a", 4096) {
			t.Errorf("a 5000-byte ASCII body did not keep 4096 a bytes: cut=%t bytes=%d", cut, len(body))
		}
	})
}

func TestBodyBoundRepeatedEuro(t *testing.T) {
	t.Run("unaligned", func(t *testing.T) {
		input := strings.Repeat("€", 3000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != strings.Repeat("€", 1365) {
			t.Errorf("an unaligned euro body kept %d bytes; want 1365 euro characters (4095 bytes)", len(body))
		}
	})
	t.Run("exact", func(t *testing.T) {
		input := strings.Repeat("€", 1365) + "x"
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if cut || body != input {
			t.Errorf("an exactly 4096-byte euro body changed: cut=%t bytes=%d", cut, len(body))
		}
	})
	t.Run("aligned", func(t *testing.T) {
		input := "a" + strings.Repeat("€", 3000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != "a"+strings.Repeat("€", 1365) {
			t.Errorf("an aligned euro body kept %d bytes; want 4096 bytes of complete characters", len(body))
		}
	})
}

func TestBodyBoundFourByteRunes(t *testing.T) {
	t.Run("unaligned", func(t *testing.T) {
		input := "a" + strings.Repeat("😀", 1025)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != "a"+strings.Repeat("😀", 1023) {
			t.Errorf("an unaligned four-byte body kept %d bytes; want 4093 bytes of complete characters", len(body))
		}
	})
	t.Run("aligned", func(t *testing.T) {
		input := strings.Repeat("😀", 1024)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if cut || body != input {
			t.Errorf("an aligned 4096-byte four-byte body changed: cut=%t", cut)
		}
	})
}

func TestBodyBoundInvalidBytes(t *testing.T) {
	t.Run("short heavy", func(t *testing.T) {
		input := strings.Repeat("\xff", 2000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != strings.Repeat("\uFFFD", 1365) {
			t.Errorf("2000 invalid bytes decoded to %d bytes; want 1365 U+FFFD characters (4095 bytes)", len(body))
		}
	})
	t.Run("below", func(t *testing.T) {
		input := strings.Repeat("\xff", 1000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		want := strings.Repeat("\uFFFD", 1000)
		if cut || body != want {
			t.Errorf("1000 invalid bytes decoded to %d bytes; want the unchanged 3000-byte replacement sequence", len(body))
		}
	})
	t.Run("exact", func(t *testing.T) {
		input := strings.Repeat("€", 1364) + "x" + "\xff"
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		want := strings.Repeat("€", 1364) + "x" + "\uFFFD"
		if cut || body != want {
			t.Errorf("an exactly 4096-byte decoded sequence changed: cut=%t bytes=%d", cut, len(body))
		}
	})
	t.Run("boundary", func(t *testing.T) {
		input := strings.Repeat("€", 1365) + "\xff"
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != strings.Repeat("€", 1365) {
			t.Errorf("an invalid byte at the bound kept %d bytes; want 4095 bytes of complete characters", len(body))
		}
	})
}

func TestBodyBoundJSONEscapes(t *testing.T) {
	t.Run("below", func(t *testing.T) {
		input := strings.Repeat("\"\\\n\t<> &"+string(rune(1)), 100)
		if len(input) >= 4096 {
			t.Fatalf("the escape control is %d bytes; it must stay below the bound", len(input))
		}
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if cut || body != input {
			t.Errorf("an escaped ASCII body below the bound changed: cut=%t bytes=%d", cut, len(body))
		}
	})
	t.Run("above", func(t *testing.T) {
		input := strings.Repeat("\"", 5000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != strings.Repeat("\"", 4096) {
			t.Errorf("an escaped ASCII body above the bound kept %d bytes; want 4096 quote bytes", len(body))
		}
	})
}

func TestBodyBoundAttributePaths(t *testing.T) {
	t.Run("ordinary", func(t *testing.T) {
		input := strings.Repeat("€", 3000)
		body, cut := checkBodyCase(t, input, ordinaryBodyMode(input))
		if !cut || body != strings.Repeat("€", 1365) {
			t.Errorf("the ordinary body path kept %d bytes", len(body))
		}
	})
	t.Run("inherited", func(t *testing.T) {
		input := strings.Repeat("€", 3000)
		body, cut := checkBodyCase(t, input, inheritedBodyMode(input))
		if !cut || body != strings.Repeat("€", 1365) {
			t.Errorf("the inherited WithAttrs body path kept %d bytes", len(body))
		}
	})
	t.Run("grouped", func(t *testing.T) {
		input := strings.Repeat("€", 3000)
		body, cut := checkBodyCase(t, input, groupedBodyMode(input))
		if !cut || body != strings.Repeat("€", 1365) {
			t.Errorf("the grouped body path kept %d bytes", len(body))
		}
	})
	t.Run("nested grouped invalid", func(t *testing.T) {
		input := strings.Repeat("\xff", 2000)
		body, cut := checkBodyCase(t, input, nestedGroupedBodyMode(input))
		if !cut || body != strings.Repeat("\uFFFD", 1365) {
			t.Errorf("the nested grouped body path kept %d bytes", len(body))
		}
	})
	t.Run("inherited group", func(t *testing.T) {
		input := strings.Repeat("€", 3000)
		mode := inheritedBodyMode(input)
		mode.groups = []string{"payload"}
		mode.key = "payload.body"
		body, cut := checkBodyCase(t, input, mode)
		if !cut || body != strings.Repeat("€", 1365) {
			t.Errorf("the inherited and grouped body path kept %d bytes", len(body))
		}
	})
}
