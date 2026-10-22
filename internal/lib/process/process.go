// Package process reads the bounded identity of one operating-system process.
//
// process.go owns the /proc/<pid>/stat read and the alive check.
package process

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

const maxStatBytes = 4096

// Start returns the start time of one process.
func Start(pid int64) (uint64, error) {
	if pid <= 0 {
		return 0, fmt.Errorf("invalid process id: %d", pid)
	}
	path := fmt.Sprintf("/proc/%d/stat", pid)
	info, err := os.Lstat(path)
	if err != nil {
		return 0, fmt.Errorf("read process start: %s", err)
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("read process start: %s is not a regular file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("read process start: %s", err)
	}
	defer func() { _ = file.Close() }()
	return readStart(file)
}

// Alive reports whether the process has the recorded start time. A different
// start time means that the platform reused the PID.
func Alive(pid int64, start uint64) (bool, error) {
	if start == 0 {
		return false, errors.New("invalid process start: 0")
	}
	current, err := Start(pid)
	if err != nil {
		return false, err
	}
	return current == start, nil
}

func readStart(reader io.Reader) (uint64, error) {
	if reader == nil {
		return 0, errors.New("read process start: nil reader")
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxStatBytes+1))
	if err != nil {
		return 0, fmt.Errorf("read process start: %s", err)
	}
	if len(data) > maxStatBytes {
		return 0, fmt.Errorf("read process start: file is larger than %d bytes", maxStatBytes)
	}
	if len(data) == 0 {
		return 0, errors.New("read process start: empty file")
	}
	return parseStat(data)
}

func parseStat(data []byte) (uint64, error) {
	end := bytes.LastIndexByte(data, ')')
	if end < 0 || end == len(data)-1 {
		return 0, errors.New("read process start: malformed stat line")
	}
	fields := bytes.Fields(bytes.TrimSpace(data[end+1:]))
	if len(fields) < 20 {
		return 0, errors.New("read process start: short stat line")
	}
	start, err := strconv.ParseUint(string(fields[19]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("read process start: invalid start time: %s", err)
	}
	if start == 0 {
		return 0, errors.New("read process start: zero start time")
	}
	return start, nil
}
