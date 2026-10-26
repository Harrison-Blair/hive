package task

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	// briefLimit is the largest brief that ReadBrief accepts.
	briefLimit = 1 << 20
	// BriefDeadline is the default total deadline of one brief read.
	BriefDeadline = 30 * time.Second
	// maxSlugLength is the largest slug.
	maxSlugLength = 40
	// briefDirName is the state directory of the briefs.
	briefDirName = "briefs"
	// TasksCollection is the task collection below the state root.
	TasksCollection = "tasks"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidSlug reports whether a value follows the slug rule of task and feature
// names.
func ValidSlug(value string) bool {
	return len(value) <= maxSlugLength && slugPattern.MatchString(value)
}

// ReadBrief reads one brief with a byte cap and a total deadline. A path that
// is not a regular file fails before the open.
func ReadBrief(path string, in io.Reader, deadline time.Duration) ([]byte, error) {
	if deadline <= 0 {
		deadline = BriefDeadline
	}
	if path == "-" {
		if in == nil {
			return nil, errors.New("read brief: no input")
		}
		return readBounded(in, deadline)
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("brief not found: %s", QuoteIfControl(path))
	}
	if err != nil {
		return nil, fmt.Errorf("read brief: %s", OneLine(err.Error()))
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("brief is not a regular file: %s", QuoteIfControl(path))
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read brief: %s", OneLine(err.Error()))
	}
	defer func() { _ = file.Close() }()
	return readBounded(file, deadline)
}

// readBounded reads at most briefLimit bytes before deadline.
func readBounded(reader io.Reader, deadline time.Duration) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(reader, briefLimit+1))
		done <- result{data: data, err: err}
	}()
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	select {
	case res := <-done:
		if res.err != nil {
			return nil, fmt.Errorf("read brief: %s", OneLine(res.err.Error()))
		}
		if len(res.data) > briefLimit {
			return nil, fmt.Errorf("brief too large: %d bytes", len(res.data))
		}
		return res.data, nil
	case <-timer.C:
		return nil, fmt.Errorf("read brief: deadline exceeded after %s", deadline)
	}
}

// WriteBrief writes one brief atomically. It is the only writer of a brief path.
func WriteBrief(home, id string, data []byte) error {
	if err := CheckBriefDir(home); err != nil {
		return err
	}
	path := BriefPath(home, id)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	if err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		return fmt.Errorf("brief path invalid: %s", QuoteIfControl(path))
	}
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	temp := file.Name()
	defer func() { _ = os.Remove(temp) }()
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("write brief: %s", OneLine(err.Error()))
	}
	return nil
}

// CheckBriefDir rejects a static symbolic link or a wrong type in the
// computed brief ancestors. A missing directory is valid because WriteBrief
// creates it.
func CheckBriefDir(home string) error {
	current := home
	for _, name := range []string{".hive", "state", briefDirName} {
		current = filepath.Join(current, name)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("brief path invalid: %s", QuoteIfControl(current))
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("brief path invalid: %s", QuoteIfControl(current))
		}
	}
	return nil
}

// BriefPath returns the absolute path of one brief.
func BriefPath(home, id string) string {
	return filepath.Join(home, ".hive", "state", briefDirName, id+".md")
}

// BriefRelPath returns the home-relative path of one brief.
func BriefRelPath(id string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", briefDirName, id+".md"))
}

// TaskRelPath returns the home-relative path of one task record.
func TaskRelPath(id string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", TasksCollection, id+".json"))
}

// WriteWorktreeBrief writes one file below a worktree or the home.
func WriteWorktreeBrief(worktreePath, relativePath string, data []byte) error {
	destination := filepath.Join(worktreePath, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create directory: %s", err)
	}
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return fmt.Errorf("write file: %s", err)
	}
	return nil
}
