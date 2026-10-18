package process

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func statLine(pid int, comm string, start string) string {
	fields := []string{
		"R", "1", "1", "1", "0", "-1", "4194304", "0", "0", "0", "0", "0", "0", "0", "0", "20", "0", "1", "0",
		start,
	}
	return strings.Join([]string{"1", "(" + comm + ")", strings.Join(fields, " ")}, " ") + "\n"
}

func TestStartAndAliveRealProcess(t *testing.T) {
	command := exec.Command("sleep", "30")
	if err := command.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	defer func() {
		_ = command.Process.Kill()
		_, _ = command.Process.Wait()
	}()
	pid := int64(command.Process.Pid)
	start, err := Start(pid)
	if err != nil {
		t.Fatalf("Start(%d) error = %v", pid, err)
	}
	if start == 0 {
		t.Fatalf("Start(%d) = 0; want a nonzero start time", pid)
	}
	alive, err := Alive(pid, start)
	if err != nil {
		t.Fatalf("Alive(%d, %d) error = %v", pid, start, err)
	}
	if !alive {
		t.Fatalf("Alive(%d, %d) = false; want true", pid, start)
	}
	alive, err = Alive(pid, start+1)
	if err != nil {
		t.Fatalf("Alive(%d, %d) error = %v", pid, start+1, err)
	}
	if alive {
		t.Fatalf("Alive(%d, %d) = true; want false for a different start", pid, start+1)
	}
}

func TestStartCurrentProcess(t *testing.T) {
	start, err := Start(int64(os.Getpid()))
	if err != nil {
		t.Fatalf("Start(self) error = %v", err)
	}
	if start == 0 {
		t.Fatalf("Start(self) = 0; want a nonzero start time")
	}
}

func TestStartRejectsInvalidPID(t *testing.T) {
	for _, pid := range []int64{0, -1, -999} {
		if _, err := Start(pid); err == nil {
			t.Fatalf("Start(%d) error = nil; want an error", pid)
		}
	}
	if _, err := Start(1 << 40); err == nil {
		t.Fatalf("Start(1<<40) error = nil; want an error")
	}
}

func TestParseStatParenthesesAndSpaces(t *testing.T) {
	start, err := parseStat([]byte(statLine(42, "a (b) c", "12345")))
	if err != nil {
		t.Fatalf("parseStat error = %v", err)
	}
	if start != 12345 {
		t.Fatalf("parseStat = %d; want 12345", start)
	}
	start, err = parseStat([]byte(statLine(42, "name with spaces", "99")))
	if err != nil {
		t.Fatalf("parseStat spaces error = %v", err)
	}
	if start != 99 {
		t.Fatalf("parseStat spaces = %d; want 99", start)
	}
}

func TestParseStatRejectsMalformedInput(t *testing.T) {
	cases := [][]byte{
		{},
		[]byte("1 (x) R\n"),
		[]byte("1 (x\n"),
		[]byte("garbage"),
		[]byte("1 (x) R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 notanumber\n"),
	}
	for _, data := range cases {
		if _, err := parseStat(data); err == nil {
			t.Fatalf("parseStat(%q) error = nil; want an error", data)
		}
	}
	large := make([]byte, maxStatBytes+1)
	for i := range large {
		large[i] = 'x'
	}
	if _, err := readStart(strings.NewReader(string(large))); err == nil {
		t.Fatalf("readStart(large) error = nil; want an error")
	}
}

func TestReadStartPartialReader(t *testing.T) {
	line := statLine(42, "comm", "777")
	reader := &chunkReader{data: []byte(line), size: 3}
	start, err := readStart(reader)
	if err != nil {
		t.Fatalf("readStart(chunked) error = %v", err)
	}
	if start != 777 {
		t.Fatalf("readStart(chunked) = %d; want 777", start)
	}
	start, err = readStart(strings.NewReader(line))
	if err != nil {
		t.Fatalf("readStart whole error = %v", err)
	}
	if start != 777 {
		t.Fatalf("readStart whole = %d; want 777", start)
	}
}

func TestAliveMissingProcess(t *testing.T) {
	alive, err := Alive(1<<40, 1)
	if err == nil {
		t.Fatalf("Alive(missing) error = nil; want an error")
	}
	if alive {
		t.Fatalf("Alive(missing) = true; want false")
	}
}

type chunkReader struct {
	data []byte
	size int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := r.size
	if n > len(r.data) {
		n = len(r.data)
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, nil
	}
	return n, nil
}

func TestReadStartSlowReader(t *testing.T) {
	line := statLine(42, "comm", "555")
	start, err := readStart(&slowReader{reader: strings.NewReader(line), delay: time.Millisecond})
	if err != nil {
		t.Fatalf("readStart(slow) error = %v", err)
	}
	if start != 555 {
		t.Fatalf("readStart(slow) = %d; want 555", start)
	}
}

type slowReader struct {
	reader *strings.Reader
	delay  time.Duration
}

func (r *slowReader) Read(p []byte) (int, error) {
	time.Sleep(r.delay)
	return r.reader.Read(p)
}
