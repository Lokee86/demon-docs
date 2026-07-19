package reverseindex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
)

type synchronizedBuffer struct {
	mu   sync.Mutex
	text strings.Builder
}

func (b *synchronizedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.Write(data)
}

func (b *synchronizedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.String()
}

func TestWatchReloadsNestedDocignoreAndAddsVisibleDirectories(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "services")
	apiRoot := filepath.Join(codeRoot, "api")
	generatedRoot := filepath.Join(apiRoot, "generated")
	mustWrite(t, filepath.Join(docsRoot, "feature.md"), "# Feature\n\n## Code map\n\n- `services/api/handler.go`\n- `services/api/generated/client.go`\n")
	mustWrite(t, filepath.Join(apiRoot, "handler.go"), "package api\n")
	mustWrite(t, filepath.Join(generatedRoot, "client.go"), "package generated\n")
	mustWrite(t, filepath.Join(apiRoot, ".docignore"), "generated/\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var output synchronizedBuffer
	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat(), 10*time.Millisecond, false, &output)
	}()

	waitFor(t, func() bool { return strings.Contains(output.String(), "watch watching") })
	if _, err := os.Stat(filepath.Join(generatedRoot, "README.md")); !os.IsNotExist(err) {
		t.Fatal("ignored generated directory was indexed before .docignore changed")
	}
	if err := os.WriteFile(filepath.Join(apiRoot, ".docignore"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		_, err := os.Stat(filepath.Join(generatedRoot, "README.md"))
		return err == nil
	})
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reverse-index watcher did not stop")
	}
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}
