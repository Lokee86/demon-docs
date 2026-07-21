//go:build windows

package watch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sys/windows"
)

const platformRecursiveTreeWatches = true

const windowsNotifyMask = windows.FILE_NOTIFY_CHANGE_FILE_NAME |
	windows.FILE_NOTIFY_CHANGE_DIR_NAME |
	windows.FILE_NOTIFY_CHANGE_ATTRIBUTES |
	windows.FILE_NOTIFY_CHANGE_SIZE |
	windows.FILE_NOTIFY_CHANGE_LAST_WRITE |
	windows.FILE_NOTIFY_CHANGE_CREATION

type windowsRecursiveWatcher struct {
	events chan fsnotify.Event
	errors chan error
	done   chan struct{}

	mu      sync.Mutex
	watches map[string]*windowsRecursiveWatch
	closed  bool
	wg      sync.WaitGroup
}

type windowsRecursiveWatch struct {
	path   string
	handle windows.Handle
	done   chan struct{}
	once   sync.Once
}

func newEventWatcher() (eventWatcher, error) {
	return &windowsRecursiveWatcher{
		events:  make(chan fsnotify.Event, 4096),
		errors:  make(chan error, 16),
		done:    make(chan struct{}),
		watches: map[string]*windowsRecursiveWatch{},
	}, nil
}

func (w *windowsRecursiveWatcher) Events() <-chan fsnotify.Event { return w.events }
func (w *windowsRecursiveWatcher) Errors() <-chan error          { return w.errors }

func (w *windowsRecursiveWatcher) Add(path string) error {
	clean, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	clean = filepath.Clean(clean)
	info, err := os.Stat(clean)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("recursive Windows watch requires a directory: %s", clean)
	}

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return fsnotify.ErrClosed
	}
	if _, exists := w.watches[clean]; exists {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	name, err := windows.UTF16PtrFromString(clean)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(
		name,
		windows.FILE_LIST_DIRECTORY,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return err
	}
	watch := &windowsRecursiveWatch{path: clean, handle: handle, done: make(chan struct{})}

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		_ = windows.CloseHandle(handle)
		return fsnotify.ErrClosed
	}
	if _, exists := w.watches[clean]; exists {
		w.mu.Unlock()
		_ = windows.CloseHandle(handle)
		return nil
	}
	w.watches[clean] = watch
	w.wg.Add(1)
	w.mu.Unlock()

	go w.readLoop(watch)
	return nil
}

func (w *windowsRecursiveWatcher) Remove(path string) error {
	clean, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	clean = filepath.Clean(clean)

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	watch, exists := w.watches[clean]
	if exists {
		delete(w.watches, clean)
	}
	w.mu.Unlock()
	if !exists {
		return fmt.Errorf("%w: %s", fsnotify.ErrNonExistentWatch, clean)
	}
	watch.stop()
	<-watch.done
	return nil
}

func (w *windowsRecursiveWatcher) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	close(w.done)
	watches := make([]*windowsRecursiveWatch, 0, len(w.watches))
	for _, watch := range w.watches {
		watches = append(watches, watch)
	}
	w.watches = map[string]*windowsRecursiveWatch{}
	w.mu.Unlock()

	for _, watch := range watches {
		watch.stop()
	}
	w.wg.Wait()
	close(w.events)
	close(w.errors)
	return nil
}

func (watch *windowsRecursiveWatch) stop() {
	watch.once.Do(func() {
		_ = windows.CancelIoEx(watch.handle, nil)
		_ = windows.CloseHandle(watch.handle)
	})
}

func (w *windowsRecursiveWatcher) readLoop(watch *windowsRecursiveWatch) {
	defer w.wg.Done()
	defer close(watch.done)
	defer func() {
		w.mu.Lock()
		if current := w.watches[watch.path]; current == watch {
			delete(w.watches, watch.path)
		}
		w.mu.Unlock()
	}()

	buffer := make([]byte, 64*1024)
	for {
		var bytesReturned uint32
		err := windows.ReadDirectoryChanges(
			watch.handle,
			&buffer[0],
			uint32(len(buffer)),
			true,
			windowsNotifyMask,
			&bytesReturned,
			nil,
			0,
		)
		if err != nil {
			if w.stopped() || errors.Is(err, windows.ERROR_OPERATION_ABORTED) || errors.Is(err, windows.ERROR_INVALID_HANDLE) {
				return
			}
			w.sendError(os.NewSyscallError("ReadDirectoryChangesW", err))
			return
		}
		if bytesReturned == 0 {
			w.sendError(fsnotify.ErrEventOverflow)
			continue
		}
		w.emitBuffer(watch.path, buffer[:bytesReturned])
	}
}

func (w *windowsRecursiveWatcher) emitBuffer(root string, buffer []byte) {
	var offset uint32
	for offset < uint32(len(buffer)) {
		raw := (*windows.FileNotifyInformation)(unsafe.Pointer(&buffer[offset]))
		nameLength := int(raw.FileNameLength / 2)
		name := windows.UTF16ToString(unsafe.Slice(&raw.FileName, nameLength))
		fullPath := filepath.Join(root, name)
		var op fsnotify.Op
		switch raw.Action {
		case windows.FILE_ACTION_ADDED:
			op = fsnotify.Create
		case windows.FILE_ACTION_REMOVED:
			op = fsnotify.Remove
		case windows.FILE_ACTION_MODIFIED:
			op = fsnotify.Write
		case windows.FILE_ACTION_RENAMED_OLD_NAME:
			op = fsnotify.Rename
		case windows.FILE_ACTION_RENAMED_NEW_NAME:
			op = fsnotify.Create
		}
		if op != 0 {
			w.sendEvent(fsnotify.Event{Name: fullPath, Op: op})
		}
		if raw.NextEntryOffset == 0 {
			return
		}
		offset += raw.NextEntryOffset
	}
}

func (w *windowsRecursiveWatcher) stopped() bool {
	select {
	case <-w.done:
		return true
	default:
		return false
	}
}

func (w *windowsRecursiveWatcher) sendEvent(event fsnotify.Event) {
	select {
	case w.events <- event:
	case <-w.done:
	}
}

func (w *windowsRecursiveWatcher) sendError(err error) {
	select {
	case w.errors <- err:
	case <-w.done:
	}
}
