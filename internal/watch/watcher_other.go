//go:build !windows

package watch

import "github.com/fsnotify/fsnotify"

const platformRecursiveTreeWatches = false

type fsnotifyWatcher struct{ *fsnotify.Watcher }

func (w fsnotifyWatcher) Events() <-chan fsnotify.Event { return w.Watcher.Events }
func (w fsnotifyWatcher) Errors() <-chan error          { return w.Watcher.Errors }

func newEventWatcher() (eventWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return fsnotifyWatcher{w}, nil
}
