package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
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

// checkHomeVersion compares the binary version with the tags at HEAD of the
// home (decision 19 of the releases spec).
func checkHomeVersion(ctx context.Context, home, binary, git string) Check {
	const name = "home version"
	if binary == "dev" {
		return newCheck(name, false, true, "development build: no release version to compare")
	}
	checkCtx, cancel := context.WithTimeout(ctx, checkBound)
	defer cancel()
	tags, err := gitLines(checkCtx, git, "-C", home, "tag", "--points-at", "HEAD")
	if err != nil {
		return newCheck(name, false, false, err.Error())
	}
	for _, tag := range tags {
		if tag == binary {
			return newCheck(name, false, true, binary)
		}
	}
	atHome := "home tags " + strings.Join(tags, ", ")
	if len(tags) == 0 {
		commit, err := gitLines(checkCtx, git, "-C", home, "rev-parse", "--short", "HEAD")
		if err != nil {
			return newCheck(name, false, false, err.Error())
		}
		atHome = "no tag at commit " + strings.Join(commit, "")
	}
	steps := fmt.Sprintf("update (README.md, Update): git -C HOME_DIR fetch --tags; git -C HOME_DIR checkout %[1]s; "+
		"go install github.com/Harrison-Blair/hive@%[1]s; restart the supervisor and the coordinator", binary)
	return newCheck(name, false, false, fmt.Sprintf("binary %s, %s; %s", binary, atHome, steps))
}

func gitLines(ctx context.Context, git string, args ...string) ([]string, error) {
	result, err := run.Command(ctx, "", git, args...)
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		detail := firstLine(result.Stderr)
		if detail == "" {
			detail = fmt.Sprintf("git exited with code %d", result.ExitCode)
		}
		return nil, errors.New(detail)
	}
	return strings.Fields(result.Stdout), nil
}

func checkSupervisor(home string) Check {
	store, err := storelib.Open(home)
	if err != nil {
		return newCheck("supervisor", false, false, err.Error())
	}
	var claim records.Claim
	err = store.ReadIn([]string{"claims"}, "supervisor", &claim)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return newCheck("supervisor", false, false, "supervisor down")
		}
		return newCheck("supervisor", false, false, err.Error())
	}
	alive, aliveErr := process.Alive(claim.PID, claim.PIDStart)
	if aliveErr != nil || !alive {
		return newCheck("supervisor", false, false, "supervisor down")
	}
	return newCheck("supervisor", false, true, fmt.Sprintf("supervisor live: pid %d", claim.PID))
}

func checkSupervisorBinary(home string) Check {
	registry, err := storelib.Open(home)
	if err != nil {
		return newCheck("supervisor binary", false, false, err.Error())
	}
	var claim records.Claim
	if err := registry.ReadIn([]string{"claims"}, "supervisor", &claim); err != nil {
		return newCheck("supervisor binary", false, false, "supervisor down")
	}
	alive, err := process.Alive(claim.PID, claim.PIDStart)
	if err != nil || !alive {
		return newCheck("supervisor binary", false, false, "supervisor down")
	}

	hivePath, err := exec.LookPath("hive")
	if err != nil {
		return newCheck("supervisor binary", false, false, "hive binary is not on PATH")
	}
	hiveInfo, err := os.Stat(hivePath)
	if err != nil {
		return newCheck("supervisor binary", false, false, fmt.Sprintf("stat hive binary %s: %s", hivePath, err))
	}

	procExe := filepath.Join("/proc", strconv.FormatInt(claim.PID, 10), "exe")
	executable, err := os.Readlink(procExe)
	if err != nil {
		return newCheck("supervisor binary", false, false, fmt.Sprintf("read supervisor executable: %s", err))
	}
	supervisorInfo, err := os.Stat(procExe)
	if err != nil {
		return newCheck("supervisor binary", false, false, fmt.Sprintf("stat supervisor executable: %s", err))
	}
	supervisorTime := supervisorInfo.ModTime().UTC().Format(time.RFC3339Nano)
	hiveTime := hiveInfo.ModTime().UTC().Format(time.RFC3339Nano)
	detail := fmt.Sprintf("supervisor %s modified %s; hive %s modified %s", executable, supervisorTime, hivePath, hiveTime)
	if strings.HasSuffix(executable, " (deleted)") {
		return newCheck("supervisor binary", false, false, detail+"; executable is deleted")
	}
	if supervisorInfo.ModTime().Before(hiveInfo.ModTime()) {
		return newCheck("supervisor binary", false, false, detail+"; supervisor binary is older")
	}
	return newCheck("supervisor binary", false, true, detail)
}

func checkStaleWorktrees(home string) Check {
	worktrees, err := listTaskWorktrees(home)
	if err != nil {
		return newCheck("stale worktrees", false, false, err.Error())
	}
	stale := make([]string, 0)
	for _, worktree := range worktrees {
		task := worktree.task
		if task.Status == "cancelled" || (task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Pushed) {
			stale = append(stale, worktree.name)
		}
	}
	if len(stale) == 0 {
		return newCheck("stale worktrees", false, true, "none")
	}
	return newCheck("stale worktrees", false, false, strings.Join(stale, ", "))
}

func checkUnpushedTaskBranches(ctx context.Context, home string) Check {
	worktrees, err := listTaskWorktrees(home)
	if err != nil {
		return newCheck("unpushed task branches", false, false, err.Error())
	}
	checkCtx, cancel := context.WithTimeout(ctx, checkBound)
	defer cancel()
	unpushed := make([]string, 0)
	for _, worktree := range worktrees {
		result, err := run.Command(checkCtx, worktree.path, "git", "rev-list", "--count", "@{upstream}..HEAD")
		if err != nil {
			unpushed = append(unpushed, worktree.name+": "+err.Error())
			continue
		}
		if result.ExitCode != 0 {
			detail := firstLine(result.Stderr)
			if detail == "" {
				detail = fmt.Sprintf("git rev-list exited with code %d", result.ExitCode)
			}
			unpushed = append(unpushed, worktree.name+": "+detail)
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(result.Stdout))
		if err != nil {
			unpushed = append(unpushed, worktree.name+": invalid ahead count")
			continue
		}
		if count > 0 {
			commitText := "commits"
			if count == 1 {
				commitText = "commit"
			}
			unpushed = append(unpushed, fmt.Sprintf("%s: %d %s ahead of upstream", worktree.name, count, commitText))
		}
	}
	if len(unpushed) == 0 {
		return newCheck("unpushed task branches", false, true, "none")
	}
	return newCheck("unpushed task branches", false, false, strings.Join(unpushed, ", "))
}

type taskWorktree struct {
	task records.Task
	path string
	name string
}

func listTaskWorktrees(home string) ([]taskWorktree, error) {
	registry, err := storelib.Open(home)
	if err != nil {
		return nil, err
	}
	raw, err := registry.List("tasks")
	if err != nil {
		return nil, err
	}
	worktrees := make([]taskWorktree, 0)
	for _, data := range raw {
		var task records.Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, fmt.Errorf("task record invalid: %s", err)
		}
		if task.Worktree == nil || task.Worktree.Path == "" {
			continue
		}
		path := filepath.Clean(task.Worktree.Path)
		relative, err := filepath.Rel(home, path)
		if err != nil {
			return nil, err
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		if len(parts) != 4 || parts[0] != "projects" || parts[2] != "worktrees" {
			continue
		}
		info, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			continue
		}
		worktrees = append(worktrees, taskWorktree{task: task, path: path, name: filepath.Join(parts[1], parts[3])})
	}
	return worktrees, nil
}

func checkStartingAgents(home string) Check {
	store, err := storelib.Open(home)
	if err != nil {
		return newCheck("starting agents", false, false, err.Error())
	}
	raw, err := store.ListIn([]string{"agents"}, nil)
	if err != nil {
		return newCheck("starting agents", false, false, err.Error())
	}
	count := 0
	for _, data := range raw {
		var agent records.Agent
		if err := json.Unmarshal(data, &agent); err != nil {
			return newCheck("starting agents", false, false, "record invalid: "+err.Error())
		}
		if agent.Status == records.AgentStarting {
			count++
		}
	}
	return newCheck("starting agents", false, count == 0, fmt.Sprintf("%d", count))
}

func checkUnresolvedAgents(home string) Check {
	store, err := storelib.Open(home)
	if err != nil {
		return newCheck("unresolved agents", false, false, err.Error())
	}
	raw, err := store.ListIn([]string{"agents"}, nil)
	if err != nil {
		return newCheck("unresolved agents", false, false, err.Error())
	}
	count := 0
	for _, data := range raw {
		var agent records.Agent
		if err := json.Unmarshal(data, &agent); err != nil {
			return newCheck("unresolved agents", false, false, "record invalid: "+err.Error())
		}
		if agent.Status == records.AgentUnresolved {
			count++
		}
	}
	return newCheck("unresolved agents", false, count == 0, fmt.Sprintf("%d", count))
}

// checkOwnerlessInboxes lists each unread inbox record in an inbox that has no
// live owner (decision 28).
func checkOwnerlessInboxes(home string) Check {
	const name = "ownerless inbox records"
	store, err := storelib.Open(home)
	if err != nil {
		return newCheck(name, false, false, err.Error())
	}
	raw, err := store.ListIn([]string{"agents"}, nil)
	if err != nil {
		return newCheck(name, false, false, err.Error())
	}
	owned := map[string]bool{}
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return newCheck(name, false, false, "record invalid: "+err.Error())
		}
		if record.Status == records.AgentLive {
			owned[agent.InboxKey(record)] = true
		}
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "state", "inbox"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return newCheck(name, false, false, err.Error())
	}
	unread := []string{}
	for _, entry := range entries {
		if !entry.IsDir() || owned[entry.Name()] {
			continue
		}
		found, err := inboxlib.All(home, entry.Name())
		if err != nil {
			return newCheck(name, false, false, err.Error())
		}
		for _, record := range found {
			if record.ReadAt == "" {
				unread = append(unread, fmt.Sprintf("%s/%d", entry.Name(), record.Seq))
			}
		}
	}
	detail := strconv.Itoa(len(unread))
	if len(unread) > 0 {
		detail += ": " + strings.Join(unread, ", ")
	}
	return newCheck(name, false, len(unread) == 0, detail)
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
