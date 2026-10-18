package supervise

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

const inotifyEventSize = 16

// watcher watches the inbox and agent directories with inotify.
type watcher struct {
	fd     int
	mu     sync.Mutex
	paths  map[int32]string
	events chan struct{}
	errs   chan error
	done   chan struct{}
	once   sync.Once
}

const watchMask = syscall.IN_CREATE | syscall.IN_MOVED_TO | syscall.IN_MODIFY |
	syscall.IN_CLOSE_WRITE | syscall.IN_DELETE | syscall.IN_MOVED_FROM

func newWatcher(roots ...string) (*watcher, error) {
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC)
	if err != nil {
		return nil, err
	}
	w := &watcher{
		fd:     fd,
		paths:  map[int32]string{},
		events: make(chan struct{}, 1),
		errs:   make(chan error, 1),
		done:   make(chan struct{}),
	}
	for _, root := range roots {
		if err := w.addTree(root); err != nil {
			_ = syscall.Close(fd)
			return nil, err
		}
	}
	go w.read()
	return w, nil
}

func (w *watcher) add(path string) error {
	wd, err := syscall.InotifyAddWatch(w.fd, path, watchMask)
	if err != nil {
		return err
	}
	w.mu.Lock()
	w.paths[int32(wd)] = path
	w.mu.Unlock()
	return nil
}

func (w *watcher) addTree(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	if err := w.add(root); err != nil {
		return err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := w.addTree(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (w *watcher) read() {
	buf := make([]byte, 4096*8)
	for {
		n, err := syscall.Read(w.fd, buf)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			select {
			case w.errs <- err:
			case <-w.done:
			}
			return
		}
		if n < inotifyEventSize {
			continue
		}
		w.parse(buf[:n])
		w.signal()
	}
}

func (w *watcher) parse(buf []byte) {
	offset := 0
	for offset+inotifyEventSize <= len(buf) {
		wd := int32(binary.LittleEndian.Uint32(buf[offset:]))
		mask := binary.LittleEndian.Uint32(buf[offset+4:])
		length := int(binary.LittleEndian.Uint32(buf[offset+12:]))
		if offset+inotifyEventSize+length > len(buf) {
			return
		}
		name := ""
		if length > 0 {
			name = strings.TrimRight(string(buf[offset+inotifyEventSize:offset+inotifyEventSize+length]), "\x00")
		}
		if mask&(syscall.IN_CREATE|syscall.IN_MOVED_TO) != 0 && name != "" {
			w.mu.Lock()
			parent := w.paths[wd]
			w.mu.Unlock()
			if parent != "" {
				_ = w.addTree(filepath.Join(parent, name))
			}
		}
		offset += inotifyEventSize + length
	}
}

func (w *watcher) signal() {
	select {
	case w.events <- struct{}{}:
	default:
	}
}

// Events returns the channel that signals a state change.
func (w *watcher) Events() <-chan struct{} {
	return w.events
}

// Close closes the inotify descriptor.
func (w *watcher) Close() {
	w.once.Do(func() {
		close(w.done)
		_ = syscall.Close(w.fd)
	})
}
