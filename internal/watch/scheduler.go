package watch

import (
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Scheduler struct {
	mu        sync.Mutex
	run       func() error
	scopedRun func([]string, bool) error
	debounce  time.Duration
	pending   int
	paths     map[string]bool
	fullPass  bool
	running   bool
	last      time.Time
	now       func() time.Time
}

func NewScheduler(run func() error, debounce time.Duration) *Scheduler {
	return &Scheduler{run: run, debounce: debounce, now: time.Now, paths: map[string]bool{}}
}

// NewScopedScheduler preserves the scheduler's debounce and coalescing
// semantics while passing a deterministic snapshot of changed paths to run.
func NewScopedScheduler(run func([]string, bool) error, debounce time.Duration) *Scheduler {
	return &Scheduler{scopedRun: run, debounce: debounce, now: time.Now, paths: map[string]bool{}}
}

func (s *Scheduler) MarkChanged() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending++
	s.fullPass = true
	s.paths = map[string]bool{}
	s.last = s.now()
}

// MarkChangedPath queues one path for a scoped run. A later full mark
// conservatively overrides the path set for that pending batch.
func (s *Scheduler) MarkChangedPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending++
	if !s.fullPass {
		s.paths[filepath.Clean(path)] = true
	}
	s.last = s.now()
}

// MarkScopedPass queues selected subsystem work without claiming that any
// validation-owned Markdown source changed.
func (s *Scheduler) MarkScopedPass() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending++
	s.last = s.now()
}

func (s *Scheduler) MarkFullPass() { s.MarkChanged() }

func (s *Scheduler) RunIfPending() (bool, error) {
	s.mu.Lock()
	if s.running || s.pending == 0 || (s.debounce > 0 && s.now().Sub(s.last) < s.debounce) {
		s.mu.Unlock()
		return false, nil
	}
	s.running = true
	s.pending = 0
	paths := make([]string, 0, len(s.paths))
	for path := range s.paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	fullPass := s.fullPass
	s.paths = map[string]bool{}
	s.fullPass = false
	s.mu.Unlock()
	var err error
	if s.scopedRun != nil {
		err = s.scopedRun(paths, fullPass)
	} else {
		err = s.run()
	}
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	return true, err
}
