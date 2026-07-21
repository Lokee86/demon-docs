package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/reverseindex"
	"github.com/Lokee86/demon-docs/internal/watch"
)

type reverseOptions struct {
	roots  []string
	format codemap.Format
}

type synchronizedWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (w *synchronizedWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(data)
}

func resolveReverseOptions(flags commonFlags, c config.Config, scope repository.Scope) (reverseOptions, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return reverseOptions{}, err
	}
	roots, err := reverseindex.ResolveRoots(
		scope.RepositoryRoot,
		scope.DocsRoot,
		cwd,
		flags.reverseRoots.values,
		c.ReverseIndex.Roots,
	)
	if err != nil {
		return reverseOptions{}, err
	}
	headings := c.Codemap.Headings
	if len(flags.codemapHeadings.values) > 0 {
		headings = flags.codemapHeadings.values
	}
	if len(headings) == 0 {
		return reverseOptions{}, fmt.Errorf("no codemap headings configured; set [codemap].headings or pass --codemap-heading")
	}
	format := codemap.DefaultFormat()
	format.SectionHeadings = append([]string(nil), headings...)
	return reverseOptions{roots: roots, format: format}, nil
}

func runSelectedWatch(
	ctx context.Context,
	scope repository.Scope,
	c config.Config,
	features watch.Features,
	reverse reverseOptions,
	debounce *float64,
	once bool,
	out io.Writer,
	ready func() error,
) error {
	if !features.Reverse {
		return watch.RootSelectedWithRunLock(ctx, scope.DocsRoot, scope.RepositoryRoot, c, features, debounce, once, out, nil, ready)
	}
	seconds := c.Watch.DebounceSeconds
	if debounce != nil {
		seconds = *debounce
	}
	if !features.Indexes && !features.Frontmatter && !features.Links && !features.TrackLinks {
		return reverseindex.WatchWithRunLock(
			ctx,
			scope.RepositoryRoot,
			scope.DocsRoot,
			reverse.roots,
			c,
			reverse.format,
			time.Duration(seconds*float64(time.Second)),
			once,
			out,
			nil,
			ready,
		)
	}
	baseFeatures := features
	baseFeatures.Reverse = false
	if once {
		if err := watch.RootSelected(ctx, scope.DocsRoot, scope.RepositoryRoot, c, baseFeatures, debounce, true, out); err != nil {
			return err
		}
		return reverseindex.Watch(
			ctx,
			scope.RepositoryRoot,
			scope.DocsRoot,
			reverse.roots,
			c,
			reverse.format,
			time.Duration(seconds*float64(time.Second)),
			true,
			out,
		)
	}

	watchContext, cancel := context.WithCancel(ctx)
	defer cancel()
	errors := make(chan error, 2)
	safeOut := &synchronizedWriter{writer: out}
	runLock := &sync.Mutex{}
	readyLock := &sync.Mutex{}
	readyCount := 0
	markReady := func() error {
		readyLock.Lock()
		defer readyLock.Unlock()
		readyCount++
		if readyCount == 2 && ready != nil {
			return ready()
		}
		return nil
	}
	go func() {
		errors <- watch.RootSelectedWithRunLock(watchContext, scope.DocsRoot, scope.RepositoryRoot, c, baseFeatures, debounce, false, safeOut, runLock, markReady)
	}()
	go func() {
		errors <- reverseindex.WatchWithRunLock(
			watchContext,
			scope.RepositoryRoot,
			scope.DocsRoot,
			reverse.roots,
			c,
			reverse.format,
			time.Duration(seconds*float64(time.Second)),
			false,
			safeOut,
			runLock,
			markReady,
		)
	}()
	first := <-errors
	cancel()
	second := <-errors
	if first != nil {
		return first
	}
	return second
}

func writeReverseIndexDiagnostics(out io.Writer, diagnostics []string) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(out, "diagnostic: %s\n", diagnostic)
	}
}
