package watch

import (
	"sync"
	"time"
)

type Scheduler struct {
	mu       sync.Mutex
	run      func() error
	debounce time.Duration
	pending  int
	running  bool
	last     time.Time
	now      func() time.Time
}

func NewScheduler(run func() error, debounce time.Duration) *Scheduler {
	return &Scheduler{run: run, debounce: debounce, now: time.Now}
}

func (s *Scheduler) MarkChanged() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending++
	s.last = s.now()
}

func (s *Scheduler) RunIfPending() (bool, error) {
	s.mu.Lock()
	if s.running || s.pending == 0 || (s.debounce > 0 && s.now().Sub(s.last) < s.debounce) {
		s.mu.Unlock()
		return false, nil
	}
	s.running = true
	s.pending = 0
	s.mu.Unlock()
	err := s.run()
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	return true, err
}
