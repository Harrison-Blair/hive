package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/run"
	storelib "github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	checkBound  = 5 * time.Second
	maxDetail   = 1000
	writeAccess = 2
)

// The status of one check.
const (
	StatusOK   = "ok"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// Check is one line of the doctor result.
type Check struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Detail   string `json:"detail"`
}

// Result holds the checks of one doctor run.
type Result struct {
	Checks []Check `json:"checks"`
}

// Failed reports a failed required check.
func (r Result) Failed() bool {
	for _, check := range r.Checks {
		if check.Required && check.Status == StatusFail {
			return true
		}
	}
	return false
}

// ExitCode is 1 when a required check fails.
func (r Result) ExitCode() int {
	if r.Failed() {
		return 1
	}
	return 0
}

func newCheck(name string, required bool, pass bool, detail string) Check {
	status := StatusOK
	if !pass {
		status = StatusWarn
		if required {
			status = StatusFail
		}
	}
	return Check{Name: name, Status: status, Required: required, Detail: plainDetail(detail)}
}

func checkHerdrSocket(ctx context.Context, logger *slog.Logger) Check {
	path := os.Getenv("HERDR_SOCKET_PATH")
	if path == "" {
		return newCheck("herdr socket", true, false, "HERDR_SOCKET_PATH is not set")
	}
	checkCtx, cancel := context.WithTimeout(ctx, checkBound)
	defer cancel()
	protocol, err := herdr.Ping(checkCtx, herdr.WithLogger(logger))
	if err != nil {
		return newCheck("herdr socket", true, false, err.Error())
	}
	return newCheck("herdr socket", true, true, fmt.Sprintf("%s protocol %d", path, protocol))
}

func checkHerdrPane() Check {
	pane := os.Getenv("HERDR_PANE_ID")
	if os.Getenv("HERDR_ENV") != "1" || pane == "" {
		return newCheck("herdr pane", false, false, "outside a herdr pane")
	}
	return newCheck("herdr pane", false, true, pane)
}

func checkHome(home string) Check {
	marker := filepath.Join(home, ".hive", "README.md")
	info, err := os.Lstat(marker)
	if err != nil || !info.Mode().IsRegular() {
		return newCheck("home", true, false, fmt.Sprintf("marker is missing: %s", marker))
	}
	hiveDir := filepath.Join(home, ".hive")
	if err := syscall.Access(hiveDir, writeAccess); err != nil {
		return newCheck("home", true, false, fmt.Sprintf("home is not writable: %s", hiveDir))
	}
	return newCheck("home", true, true, home)
}

func checkStateStore(home string) Check {
	store, err := storelib.Open(home)
	if err != nil {
		return newCheck("state store", true, false, err.Error())
	}
	records := 0
	err = store.WalkRecords(context.Background(), func(record storelib.StoredRecord) error {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(record.Data, &object); err != nil {
			return fmt.Errorf("record invalid: %s", recordPath(home, record))
		}
		records++
		return nil
	})
	if err != nil {
		return newCheck("state store", true, false, err.Error())
	}
	return newCheck("state store", true, true, fmt.Sprintf("%d records", records))
}

func recordPath(home string, record storelib.StoredRecord) string {
	parts := append([]string{home, ".hive", "state"}, record.Collection...)
	return filepath.Join(append(parts, record.Name)...)
}

func checkGit(ctx context.Context) Check {
	checkCtx, cancel := context.WithTimeout(ctx, checkBound)
	defer cancel()
	result, err := run.Command(checkCtx, "", "git", "--version")
	if err != nil {
		return newCheck("git", true, false, err.Error())
	}
	if result.ExitCode != 0 {
		return newCheck("git", true, false, fmt.Sprintf("git --version exited with code %d", result.ExitCode))
	}
	return newCheck("git", true, true, firstLine(result.Stdout))
}

func checkSupervisor(home string) Check {
	store, err := storelib.Open(home)
	if err != nil {
		return newCheck("supervisor", false, false, err.Error())
	}
	var claim json.RawMessage
	err = store.ReadIn([]string{"claims"}, "supervisor", &claim)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return newCheck("supervisor", false, false, "no supervisor claim")
		}
		return newCheck("supervisor", false, false, err.Error())
	}
	return newCheck("supervisor", false, true, "supervisor claim present")
}

func checkHarness(ctx context.Context, name string) Check {
	checkCtx, cancel := context.WithTimeout(ctx, checkBound)
	defer cancel()
	result, err := run.Command(checkCtx, "", "sh", "-c", `command -v "$1"`, "sh", name)
	if err != nil || result.ExitCode != 0 {
		return newCheck("harness "+name, false, false, "not on PATH")
	}
	return newCheck("harness "+name, false, true, firstLine(result.Stdout))
}

func checkGH(ctx context.Context) Check {
	checkCtx, cancel := context.WithTimeout(ctx, checkBound)
	defer cancel()
	version, err := run.Command(checkCtx, "", "gh", "--version")
	if err != nil {
		return newCheck("gh", false, false, "not on PATH")
	}
	if version.ExitCode != 0 {
		return newCheck("gh", false, false, fmt.Sprintf("gh --version exited with code %d", version.ExitCode))
	}
	status, err := run.Command(checkCtx, "", "gh", "auth", "status")
	if err != nil || status.ExitCode != 0 {
		return newCheck("gh", false, false, "not authenticated")
	}
	return newCheck("gh", false, true, firstLine(version.Stdout))
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

func plainDetail(text string) string {
	cleaned := strings.Map(func(value rune) rune {
		if value < 0x20 || value == 0x7f {
			return '?'
		}
		return value
	}, text)
	if len(cleaned) > maxDetail {
		return cleaned[:maxDetail] + "..."
	}
	return cleaned
}
