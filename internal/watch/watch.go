package watch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/documentpolicy"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/reconcile"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/scan"
	"github.com/Lokee86/demon-docs/internal/validationcache"
	"github.com/fsnotify/fsnotify"
)

type eventWatcher interface {
	Add(string) error
	Remove(string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

var createWatcher = newEventWatcher
var useRecursiveTreeWatches = platformRecursiveTreeWatches
var repairObservedRename = links.RepairObservedRename

func Root(ctx context.Context, root string, c config.Config, debounce *float64, once bool, out io.Writer) error {
	return RootWithIgnoreRoot(ctx, root, root, c, debounce, once, out)
}

func RootWithIgnoreRoot(ctx context.Context, root, ignoreRoot string, c config.Config, debounce *float64, once bool, out io.Writer) error {
	return RootSelected(ctx, root, ignoreRoot, c, Features{Indexes: true, Frontmatter: c.Frontmatter.Enabled, Format: c.Format.Enabled}, debounce, once, out)
}

func RootSelected(ctx context.Context, docsRoot, repositoryRoot string, c config.Config, features Features, debounce *float64, once bool, out io.Writer) error {
	return RootSelectedWithRunLock(ctx, docsRoot, repositoryRoot, c, features, debounce, once, out, nil, nil)
}

func RootSelectedWithRunLock(ctx context.Context, docsRoot, repositoryRoot string, c config.Config, features Features, debounce *float64, once bool, out io.Writer, runLock sync.Locker, ready func() error) error {
	if features.Frontmatter {
		if err := frontmatter.ValidateConfig(c.Frontmatter); err != nil {
			return err
		}
	}
	if out == nil {
		out = io.Discard
	}
	if features.Links {
		features.TrackLinks = true
	}
	seconds := c.Watch.DebounceSeconds
	if debounce != nil {
		seconds = *debounce
	}
	watchRoot := docsRoot
	if features.TrackLinks {
		watchRoot = repositoryRoot
	}
	fmt.Fprintf(out, "%s ddocs watch watching %s pid=%d\n", timestamp(), watchRoot, os.Getpid())
	var watcher eventWatcher
	externalWatched := map[string]bool{}
	var externalDirectories []string
	run := func(changedPaths []string, fullValidation bool) error {
		if runLock != nil {
			runLock.Lock()
			defer runLock.Unlock()
		}
		changed := 0
		var diagnostics []string
		unresolved := 0
		frontmatterUnresolved := 0
		formatUnresolved := 0
		validationPathSet := map[string]bool{}
		addValidationPath := func(path string) {
			if repository.Contains(docsRoot, path) && strings.EqualFold(filepath.Ext(path), ".md") {
				validationPathSet[filepath.Clean(path)] = true
			}
		}
		for _, path := range changedPaths {
			addValidationPath(path)
		}
		validationPaths := func() []string {
			paths := make([]string, 0, len(validationPathSet))
			for path := range validationPathSet {
				paths = append(paths, path)
			}
			sort.Strings(paths)
			return paths
		}
		if features.TrackLinks {
			var plan links.Plan
			var err error
			if features.Links {
				plan, err = links.Reconcile(repositoryRoot)
			} else {
				plan, err = links.Track(repositoryRoot)
			}
			if err != nil {
				return err
			}
			if features.Links {
				count, err := links.ApplyAndSave(&plan)
				if err != nil {
					return err
				}
				changed += count
				for _, rewrite := range plan.Rewrites {
					addValidationPath(rewrite.Path)
				}
				for _, update := range plan.Updates {
					addValidationPath(update.Path)
				}
				diagnostics = append(diagnostics, plan.Messages...)
				unresolved = plan.Unresolved
			} else if err := links.Save(plan); err != nil {
				return err
			}
			externalDirectories = externalWatchDirectories(plan.Files)
			if watcher != nil {
				if err := addExternalWatches(watcher, externalDirectories, externalWatched); err != nil {
					return fmt.Errorf("watch external link targets: %w", err)
				}
			}
		}
		if features.Indexes {
			result, err := reconcile.TreeWithIgnoreRoot(docsRoot, repositoryRoot, c)
			if err != nil {
				return err
			}
			if err := reconcile.PrepareMissingWithin(result, docsRoot); err != nil {
				return err
			}
			for _, update := range result.Updates {
				addValidationPath(update.Path)
			}
		}
		if features.Frontmatter {
			paths := validationPaths()
			if fullValidation || len(paths) > 0 {
				var plan frontmatter.Plan
				var err error
				if fullValidation {
					plan, err = frontmatter.Build(repositoryRoot, docsRoot, c, true, time.Now())
				} else {
					plan, err = frontmatter.BuildScoped(repositoryRoot, docsRoot, c, true, time.Now(), paths)
					if errors.Is(err, validationcache.ErrScopedReuseUnavailable) {
						plan, err = frontmatter.Build(repositoryRoot, docsRoot, c, true, time.Now())
					}
				}
				if err != nil {
					return err
				}
				count, err := frontmatter.Apply(repositoryRoot, docsRoot, plan)
				if err != nil {
					return err
				}
				changed += count
				for _, update := range plan.Updates {
					addValidationPath(update.Path)
				}
				for _, diagnostic := range plan.Diagnostics {
					diagnostics = append(diagnostics, frontmatterDiagnostic(diagnostic))
					if !diagnostic.Warning && !diagnostic.Resolved {
						frontmatterUnresolved++
					}
				}
			}
		}
		if features.Format {
			paths := validationPaths()
			if fullValidation || len(paths) > 0 {
				var plan documentpolicy.Plan
				var err error
				if fullValidation {
					plan, err = documentpolicy.Build(repositoryRoot, docsRoot, c, true)
				} else {
					plan, err = documentpolicy.BuildScoped(repositoryRoot, docsRoot, c, true, paths)
					if errors.Is(err, validationcache.ErrScopedReuseUnavailable) {
						plan, err = documentpolicy.Build(repositoryRoot, docsRoot, c, true)
					}
				}
				if err != nil {
					return err
				}
				count, err := documentpolicy.Apply(plan, docsRoot)
				if err != nil {
					return err
				}
				changed += count
				for _, diagnostic := range plan.Diagnostics {
					diagnostics = append(diagnostics, formatDiagnostic(diagnostic))
					if !diagnostic.Warning && !diagnostic.Resolved {
						formatUnresolved++
					}
				}
			}
		}
		if features.Indexes {
			result, count, err := reconcile.ConvergeWithin(docsRoot, repositoryRoot, c)
			if err != nil {
				return err
			}
			changed += count
			diagnostics = append(diagnostics, result.Messages...)
		}
		if features.Indexes || features.Frontmatter || features.Format {
			refreshPlan, err := links.Track(repositoryRoot)
			if err != nil {
				return err
			}
			if features.TrackLinks || refreshPlan.Initialized {
				if err := links.Save(refreshPlan); err != nil {
					return err
				}
				externalDirectories = externalWatchDirectories(refreshPlan.Files)
				if watcher != nil {
					if err := addExternalWatches(watcher, externalDirectories, externalWatched); err != nil {
						return fmt.Errorf("watch external link targets: %w", err)
					}
				}
			}
		}
		fmt.Fprintf(out, "%s ddocs watch updated %d file(s)\n", timestamp(), changed)
		for _, message := range diagnostics {
			fmt.Fprintf(out, "%s ddocs watch: %s\n", timestamp(), message)
		}
		if unresolved > 0 {
			fmt.Fprintf(out, "%s ddocs watch unresolved links: %d\n", timestamp(), unresolved)
		}
		if frontmatterUnresolved > 0 {
			fmt.Fprintf(out, "%s ddocs watch unresolved frontmatter issue(s): %d\n", timestamp(), frontmatterUnresolved)
		}
		if formatUnresolved > 0 {
			fmt.Fprintf(out, "%s ddocs watch unresolved document-format issue(s): %d\n", timestamp(), formatUnresolved)
		}
		return nil
	}
	initialRetryDelay := time.Duration(seconds * float64(time.Second))
	if initialRetryDelay < 100*time.Millisecond {
		initialRetryDelay = 100 * time.Millisecond
	}
	if initialRetryDelay > time.Second {
		initialRetryDelay = time.Second
	}
	if err := runInitialUntilStable(ctx, func() error { return run(nil, true) }, initialRetryDelay, out); err != nil {
		return err
	}
	if once {
		return nil
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		return err
	}
	w, err := createWatcher()
	if err != nil {
		return fmt.Errorf("create watcher for %s: %w", watchRoot, err)
	}
	watcher = w
	defer w.Close()
	watchedDirs := map[string]bool{}
	formatWatched := map[string]bool{}
	if filepath.Clean(repositoryRoot) != filepath.Clean(watchRoot) {
		if err := w.Add(repositoryRoot); err != nil {
			return fmt.Errorf("watch repository root %s: %w", repositoryRoot, err)
		}
	}
	if err := addTree(w, watchRoot, watchRoot, c, policy, watchedDirs); err != nil {
		return err
	}
	if err := addExternalWatches(w, externalDirectories, externalWatched); err != nil {
		return fmt.Errorf("watch external link targets: %w", err)
	}
	if features.Format {
		if err := addFormatWatches(w, repositoryRoot, c, formatWatched); err != nil {
			return fmt.Errorf("watch document schemas: %w", err)
		}
	}
	scheduler := NewScopedScheduler(run, time.Duration(seconds*float64(time.Second)))
	// Close the startup handoff gap with one pass after watch registration.
	scheduler.MarkFullPass()
	if ready != nil {
		if err := ready(); err != nil {
			return fmt.Errorf("mark watcher ready: %w", err)
		}
	}
	pendingRenames := map[string]time.Time{}
	immediateRenameRepairs := 0
	bulkRenameObserved := false
	lastObservedRename := time.Time{}
	bulkRenameQuietPeriod := time.Duration(seconds * float64(time.Second))
	if bulkRenameQuietPeriod < 500*time.Millisecond {
		bulkRenameQuietPeriod = 500 * time.Millisecond
	}
	const immediateRenameRepairLimit = 1
	runObservedRename := func(oldPath, newPath string) (bool, int, error) {
		if runLock != nil {
			runLock.Lock()
			defer runLock.Unlock()
		}
		return repairObservedRename(repositoryRoot, oldPath, newPath)
	}
	interval := 100 * time.Millisecond
	if seconds > 0 {
		interval = time.Duration(seconds * float64(time.Second) / 2)
		if interval > 250*time.Millisecond {
			interval = 250 * time.Millisecond
		}
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-w.Errors():
			if !ok {
				return nil
			}
			if errors.Is(err, fsnotify.ErrEventOverflow) {
				fmt.Fprintf(out, "%s ddocs watch event buffer overflow; scheduling a complete reconciliation\n", timestamp())
				scheduler.MarkFullPass()
				continue
			}
			if err != nil {
				return fmt.Errorf("watch %s: %w", watchRoot, err)
			}
		case event, ok := <-w.Events():
			if !ok {
				return nil
			}
			if features.Links && repository.Contains(repositoryRoot, event.Name) {
				suppressed, err := links.ConsumePendingSuppression(repositoryRoot, event.Name)
				if err != nil {
					return fmt.Errorf("consume generated link rewrite event: %w", err)
				}
				if suppressed {
					continue
				}
			}
			isExternal := features.TrackLinks && externalEvent(event.Name, externalWatched)
			if policy.IsControlFile(event.Name) {
				updated, err := ignorepolicy.Load(repositoryRoot)
				if err != nil {
					return err
				}
				policy = updated
				if err := addTree(w, watchRoot, watchRoot, c, policy, watchedDirs); err != nil {
					return err
				}
				scheduler.MarkFullPass()
				continue
			}
			if features.Format && event.Op&fsnotify.Create != 0 && formatSchemaEvent(event.Name, repositoryRoot, c, false) {
				if err := addFormatWatches(w, repositoryRoot, c, formatWatched); err != nil {
					return fmt.Errorf("watch document schemas: %w", err)
				}
			}
			if event.Op&fsnotify.Create != 0 && repository.Contains(watchRoot, event.Name) {
				if st, err := os.Stat(event.Name); err == nil && st.IsDir() {
					ignored, err := policy.Ignored(event.Name, true)
					if err != nil {
						return err
					}
					if !ignored && !watchIgnored(event.Name, c) {
						if err := addTree(w, watchRoot, event.Name, c, policy, watchedDirs); err != nil {
							return err
						}
					}
				}
			}
			wasDirectory := watchedDirs[event.Name] || formatWatched[event.Name]
			if !wasDirectory && event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil {
					wasDirectory = info.IsDir()
				}
			}
			if features.Links && repository.Contains(repositoryRoot, event.Name) {
				now := time.Now()
				for path, observedAt := range pendingRenames {
					if now.Sub(observedAt) > 5*time.Second {
						delete(pendingRenames, path)
					}
				}
				if event.Op&fsnotify.Rename != 0 && !wasDirectory {
					pendingRenames[filepath.Clean(event.Name)] = now
				}
				if event.Op&fsnotify.Create != 0 {
					if info, statErr := os.Lstat(event.Name); statErr == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
						newDirectory := filepath.Clean(filepath.Dir(event.Name))
						oldPath := ""
						var newest time.Time
						for candidate, observedAt := range pendingRenames {
							if strings.EqualFold(filepath.Clean(filepath.Dir(candidate)), newDirectory) && (oldPath == "" || observedAt.After(newest)) {
								oldPath = candidate
								newest = observedAt
							}
						}
						if oldPath != "" {
							delete(pendingRenames, oldPath)
							lastObservedRename = time.Now()
							if immediateRenameRepairs < immediateRenameRepairLimit {
								immediateRenameRepairs++
								handled, changed, repairErr := runObservedRename(oldPath, event.Name)
								if repairErr != nil {
									return fmt.Errorf("repair observed rename %s -> %s: %w", oldPath, event.Name, repairErr)
								}
								if handled {
									fmt.Fprintf(out, "%s ddocs watch repaired rename immediately: %s -> %s; updated %d file(s)\n", timestamp(), oldPath, event.Name, changed)
								}
							} else {
								bulkRenameObserved = true
							}
						}
					}
				}
			}
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
				if wasDirectory {
					removeWatchTree(w, event.Name, watchedDirs)
					removeWatchTree(w, event.Name, externalWatched)
					removeWatchTree(w, event.Name, formatWatched)
				}
			}
			path, full, relevant := validationBatchForEvent(event, c, policy, docsRoot, repositoryRoot, features, wasDirectory, isExternal)
			if relevant {
				switch {
				case full:
					scheduler.MarkFullPass()
				case path != "":
					scheduler.MarkChangedPath(path)
				default:
					scheduler.MarkScopedPass()
				}
			}
		case <-ticker.C:
			if bulkRenameObserved && time.Since(lastObservedRename) < bulkRenameQuietPeriod {
				continue
			}
			ran, err := scheduler.RunIfPending()
			if err != nil {
				if links.IsTransientFilesystemRace(err) {
					fmt.Fprintf(out, "%s ddocs watch deferred stale reconciliation plan: %v\n", timestamp(), err)
					scheduler.MarkChanged()
					continue
				}
				return err
			}
			if ran {
				immediateRenameRepairs = 0
				bulkRenameObserved = false
				lastObservedRename = time.Time{}
			}
		}
	}
}

func addTree(w eventWatcher, root, start string, c config.Config, policy ignorepolicy.Policy, watched map[string]bool) error {
	if useRecursiveTreeWatches {
		start = filepath.Clean(start)
		for watchedRoot := range watched {
			if repository.Contains(watchedRoot, start) {
				return nil
			}
		}
		if err := w.Add(start); err != nil {
			return fmt.Errorf("watch directory tree %s: %w", start, err)
		}
		watched[start] = true
		return nil
	}
	return filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root {
			ignored, err := policy.Ignored(path, true)
			if err != nil {
				return err
			}
			if ignored || watchIgnored(path, c) {
				return filepath.SkipDir
			}
		}
		if watched[path] {
			return nil
		}
		if err := w.Add(path); err != nil {
			return fmt.Errorf("watch directory %s: %w", path, err)
		}
		watched[path] = true
		return nil
	})
}

func Relevant(path string, c config.Config, root string) bool {
	return RelevantWithIgnoreRoot(path, c, root, root)
}

func RelevantWithIgnoreRoot(path string, c config.Config, root, ignoreRoot string) bool {
	policy, err := ignorepolicy.Load(ignoreRoot)
	return err == nil && relevantWithPolicy(path, c, policy, root)
}

func relevantWithPolicy(path string, c config.Config, policy ignorepolicy.Policy, root string) bool {
	if policy.IsControlFile(path) {
		return true
	}
	if !repository.Contains(root, path) {
		return false
	}
	ignored, err := policy.Ignored(path, false)
	if err != nil || ignored || watchIgnored(path, c) {
		return false
	}
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		ignored, err := policy.Ignored(path, true)
		return err == nil && !ignored
	}
	ok, err := scan.IsIndexable(root, path, c)
	return err == nil && ok
}

func watchIgnored(path string, c config.Config) bool {
	clean := filepath.Clean(path)
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		for _, name := range c.Watch.IgnoredDirs {
			if part == name {
				return true
			}
		}
	}
	name := filepath.Base(clean)
	if strings.HasPrefix(name, ".#") {
		return true
	}
	for _, suffix := range c.Watch.IgnoredSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func frontmatterDiagnostic(diagnostic frontmatter.Diagnostic) string {
	location := diagnostic.Path
	if diagnostic.Field != "" {
		location += ":" + diagnostic.Field
	}
	status := "issue"
	if diagnostic.Warning {
		status = "warning"
	} else if diagnostic.Resolved {
		status = "repaired"
	}
	return fmt.Sprintf("Frontmatter %s at %s: %s", status, location, diagnostic.Message)
}

func formatDiagnostic(diagnostic documentpolicy.Diagnostic) string {
	location := diagnostic.Path
	if diagnostic.Section != "" {
		location += ":" + diagnostic.Section
	}
	status := "issue"
	if diagnostic.Warning {
		status = "warning"
	} else if diagnostic.Resolved {
		status = "repaired"
	}
	return fmt.Sprintf("Document format %s at %s: %s", status, location, diagnostic.Message)
}

func runInitialUntilStable(ctx context.Context, run func() error, retryDelay time.Duration, out io.Writer) error {
	for {
		err := run()
		if err == nil {
			return nil
		}
		if !links.IsTransientFilesystemRace(err) {
			return err
		}
		fmt.Fprintf(out, "%s ddocs watch deferred stale initial reconciliation plan: %v\n", timestamp(), err)
		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func timestamp() string { return time.Now().Format("2006-01-02T15:04:05") }
