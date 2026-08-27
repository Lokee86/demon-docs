package codemaparcana

import "sync"

type lockedBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (buffer *lockedBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	buffer.data = append(buffer.data, data...)
	buffer.mu.Unlock()
	return len(data), nil
}

func (buffer *lockedBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return string(append([]byte(nil), buffer.data...))
}
