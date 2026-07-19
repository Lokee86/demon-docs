package demon

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	MaxLogBytes = 1 << 20
	LogFiles    = 5
)

type RotatingLog struct {
	mu   sync.Mutex
	path string
	file *os.File
}

func OpenLog(paths Paths) (*RotatingLog, error) {
	if err := os.MkdirAll(paths.Logs, 0o755); err != nil {
		return nil, err
	}
	l := &RotatingLog{path: paths.Log}
	if err := l.open(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *RotatingLog) open() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

func (l *RotatingLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		if err := l.open(); err != nil {
			return 0, err
		}
	}
	written := 0
	for len(p) > 0 {
		info, err := l.file.Stat()
		if err != nil {
			return written, err
		}
		remaining := MaxLogBytes - info.Size()
		if remaining <= 0 {
			if err := l.rotate(); err != nil {
				return written, err
			}
			continue
		}
		n := len(p)
		if int64(n) > remaining {
			n = int(remaining)
		}
		count, err := l.file.Write(p[:n])
		written += count
		p = p[count:]
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (l *RotatingLog) rotate() error {
	if err := l.file.Close(); err != nil {
		return err
	}
	for i := LogFiles - 1; i >= 1; i-- {
		from := l.path
		if i > 1 {
			from = l.path + "." + itoa(i-1)
		}
		to := l.path + "." + itoa(i)
		if _, err := os.Stat(from); err == nil {
			_ = os.Remove(to)
			if err := os.Rename(from, to); err != nil {
				return err
			}
		}
	}
	return l.open()
}

func (l *RotatingLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

func (l *RotatingLog) Reader() (io.Reader, error) { return os.Open(filepath.Clean(l.path)) }

func itoa(v int) string {
	if v == 1 {
		return "1"
	}
	if v == 2 {
		return "2"
	}
	if v == 3 {
		return "3"
	}
	return "4"
}
