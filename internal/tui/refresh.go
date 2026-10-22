package tui

import (
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
)

type refreshSnapshot struct {
	header  string
	screens [7]string
}

type refreshCheckMsg struct {
	request dataRequest
	header  string
	screen  string
}

func (m *Model) startRefreshCheck() tea.Cmd {
	if m.refreshChecking || m.demo != nil {
		return nil
	}
	m.refreshChecking = true
	home := m.home
	request := m.currentScreen().DataRequest()
	day := startOfDay(m.logDate)
	tasks := m.tasks.data
	return func() tea.Msg {
		return refreshCheckMsg{
			request: request,
			header:  pathSignature(headerRefreshPaths(home)),
			screen:  pathSignature(screenRefreshPaths(home, request, day, tasks)),
		}
	}
}

func snapshotRefreshState(home string, day time.Time, tasks tasksData) refreshSnapshot {
	snapshot := refreshSnapshot{header: pathSignature(headerRefreshPaths(home))}
	for request := chatDataRequest; request <= frictionDataRequest; request++ {
		snapshot.screens[request] = pathSignature(screenRefreshPaths(home, request, day, tasks))
	}
	return snapshot
}

func headerRefreshPaths(home string) []string {
	state := filepath.Join(home, ".hive", "state")
	paths := []string{
		filepath.Join(state, "agents"),
		filepath.Join(state, "claims"),
		filepath.Join(state, "projects"),
		filepath.Join(state, "tasks"),
	}
	if hivePath, err := exec.LookPath("hive"); err == nil {
		paths = append(paths, hivePath)
	}
	return paths
}

func screenRefreshPaths(home string, request dataRequest, day time.Time, tasks tasksData) []string {
	state := filepath.Join(home, ".hive", "state")
	switch request {
	case chatDataRequest:
		return []string{
			filepath.Join(state, "agents"),
			filepath.Join(state, "chat"),
			filepath.Join(state, "inbox"),
		}
	case logsDataRequest:
		return []string{filepath.Join(home, ".hive", "logs", day.Format("2006-01-02")+".jsonl")}
	case agentsDataRequest:
		return []string{
			filepath.Join(state, "agents"),
			filepath.Join(state, "claims"),
			filepath.Join(state, "goals"),
			filepath.Join(state, "inbox"),
			filepath.Join(state, "projects"),
		}
	case goalsQueueDataRequest:
		return []string{filepath.Join(state, "goals"), filepath.Join(state, "tasks")}
	case healthDataRequest:
		return []string{filepath.Join(home, ".hive", "README.md"), state}
	case tasksDataRequest:
		paths := []string{
			filepath.Join(state, "goals"),
			filepath.Join(state, "projects"),
			filepath.Join(state, "tasks"),
		}
		for _, item := range tasks.items {
			if item.projectPath == "" || item.record.Feature == "" {
				continue
			}
			paths = append(paths, filepath.Join(item.projectPath, ".hive", "spec", item.record.Feature, "tasks", item.record.ID, "evidence"))
		}
		sort.Strings(paths)
		return paths
	case frictionDataRequest:
		return []string{filepath.Join(home, "reference", "dogfood", "friction")}
	default:
		return nil
	}
}

func pathSignature(paths []string) string {
	hash := fnv.New64a()
	for _, root := range paths {
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			_, _ = hash.Write([]byte(path))
			if walkErr != nil {
				_, _ = hash.Write([]byte(walkErr.Error()))
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				_, _ = hash.Write([]byte(err.Error()))
				return nil
			}
			_, _ = hash.Write([]byte(info.Mode().String()))
			var buffer [32]byte
			encoded := strconv.AppendInt(buffer[:0], info.Size(), 10)
			encoded = strconv.AppendInt(encoded, info.ModTime().UnixNano(), 10)
			_, _ = hash.Write(encoded)
			return nil
		})
	}
	return strconv.FormatUint(hash.Sum64(), 16)
}
