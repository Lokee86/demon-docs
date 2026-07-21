package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/Lokee86/demon-docs/internal/config"
	md "github.com/Lokee86/demon-docs/internal/markdown"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/pathutil"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/scan"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func Tree(root string, c config.Config) (model.ReconcileResult, error) {
	return TreeWithIgnoreRoot(root, root, c)
}

func TreeWithIgnoreRoot(root, ignoreRoot string, c config.Config) (model.ReconcileResult, error) {
	tree, err := scan.TreeWithIgnoreRoot(root, ignoreRoot, c)
	if err != nil {
		return model.ReconcileResult{}, err
	}
	folders := orderedFolders(tree)
	texts := map[string]string{}
	entries := map[string][]*model.IndexEntry{}
	for _, f := range folders {
		if f.IndexPath == "" {
			continue
		}
		doc, err := textio.Read(f.IndexPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return model.ReconcileResult{}, fmt.Errorf("read index %s: %w", f.IndexPath, err)
		}
		texts[f.Path] = doc.Text
		entries[f.Path] = md.ParseEntries(f.IndexPath, doc.Text, c)
	}
	childTitles, err := rootChildTitles(tree.Root, folders, texts, c)
	if err != nil {
		return model.ReconcileResult{}, err
	}
	rootTitle := md.ManagedRootTitle(tree.Root, texts[tree.Root], childTitles)
	title := func(folder string) string {
		if filepath.Clean(folder) == filepath.Clean(tree.Root) {
			return rootTitle
		}
		return md.FolderTitle(folder, texts[folder])
	}
	for _, f := range folders {
		if f.IndexPath == "" {
			continue
		}
		if _, err := os.Stat(f.IndexPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return model.ReconcileResult{}, err
		}
		parent := ""
		if f.Path != tree.Root {
			parent = title(filepath.Dir(f.Path))
		}
		texts[f.Path] = md.MakeTemplate(f.Path, tree.Root, parent, c.IndexFile, c)
	}
	crossFiles := crossFileEntries(entries)
	crossFolders := crossFolderEntries(entries)
	fileCounts := unmatchedFileCounts(folders, entries, c)
	folderCounts := unmatchedFolderCounts(folders, entries, c)
	matched := map[*model.IndexEntry]bool{}
	result := model.ReconcileResult{}
	for _, f := range folders {
		if f.IndexPath != "" {
			current := texts[f.Path]
			desired := md.DesiredParent(f.IndexPath, tree.Root, title, c)
			newText, err := updateSections(f, md.UpdateParent(current, desired, c.ParentLink.Label), entries[f.Path], crossFiles, fileCounts, crossFolders, folderCounts, matched, c)
			if err != nil {
				return result, fmt.Errorf("reconcile index %s: %w", f.IndexPath, err)
			}
			_, statErr := os.Stat(f.IndexPath)
			if os.IsNotExist(statErr) || newText != current {
				var old *string
				if statErr == nil {
					x := current
					old = &x
				}
				result.Updates = append(result.Updates, model.FileUpdate{Path: f.IndexPath, OldText: old, NewText: newText})
			}
		}
		if f.IsStubs {
			continue
		}
		for _, p := range append(append([]string{}, f.DirectFiles...), f.StubFiles...) {
			if !config.IsParentEditable(p, c) {
				continue
			}
			doc, err := textio.Read(p)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return result, fmt.Errorf("read indexed file %s: %w", p, err)
			}
			desired := md.DesiredParent(p, tree.Root, title, c)
			next := md.UpdateParent(doc.Text, desired, c.ParentLink.Label)
			if next != doc.Text {
				x := doc.Text
				result.Updates = append(result.Updates, model.FileUpdate{Path: p, OldText: &x, NewText: next})
			}
		}
	}
	type staleEntry struct {
		indexPath, section, line string
	}
	var stale []staleEntry
	for folder, es := range entries {
		for _, e := range es {
			if !matched[e] {
				stale = append(stale, staleEntry{filepath.Join(folder, c.IndexFile), e.Section, e.OriginalLine})
			}
		}
	}
	sort.Slice(stale, func(i, j int) bool {
		left, right := stale[i], stale[j]
		if pathOrderKey(left.indexPath) != pathOrderKey(right.indexPath) {
			return pathOrderKey(left.indexPath) < pathOrderKey(right.indexPath)
		}
		if left.section != right.section {
			return left.section < right.section
		}
		return left.line < right.line
	})
	for _, entry := range stale {
		result.Messages = append(result.Messages, fmt.Sprintf("Removed stale %s entry from %s: %s", entry.section, entry.indexPath, entry.line))
	}
	return result, nil
}

func Apply(result model.ReconcileResult) (int, error) {
	return apply(result)
}

func ApplyWithin(result model.ReconcileResult, root string) (int, error) {
	for _, update := range result.Updates {
		if !repository.Contains(root, update.Path) {
			return 0, fmt.Errorf("refusing to write outside docs root: %s", update.Path)
		}
	}
	return applyWithin(result)
}

// PrepareMissingWithin creates planned index content for indexes that do not yet
// exist. It never recreates a missing parent directory: a directory move can
// invalidate a prepared reconciliation plan while the daemon is still running.
func PrepareMissingWithin(result model.ReconcileResult, root string) error {
	for _, update := range result.Updates {
		if update.OldText != nil {
			continue
		}
		if !repository.Contains(root, update.Path) {
			return fmt.Errorf("refusing to prepare index outside docs root: %s", update.Path)
		}
		parentExists, err := existingParent(update.Path)
		if err != nil {
			return err
		}
		if !parentExists {
			continue
		}
		if _, err := applyUpdate(update, true); err != nil {
			return fmt.Errorf("prepare index %s: %w", update.Path, err)
		}
	}
	return nil
}

const maxIndexConvergencePasses = 8

// ConvergeWithin rebuilds and applies index plans until the tree is stable.
// Newly prepared indexes can require a follow-up pass once all generated titles
// and parent relationships are visible to the scanner.
func ConvergeWithin(root, ignoreRoot string, c config.Config) (model.ReconcileResult, int, error) {
	changed := 0
	messages := []string{}
	for pass := 0; pass < maxIndexConvergencePasses; pass++ {
		result, err := TreeWithIgnoreRoot(root, ignoreRoot, c)
		if err != nil {
			return result, changed, err
		}
		messages = append(messages, result.Messages...)
		if len(result.Updates) == 0 {
			result.Messages = messages
			return result, changed, nil
		}
		count, err := ApplyWithin(result, root)
		if err != nil {
			return result, changed, err
		}
		changed += count
	}
	result, err := TreeWithIgnoreRoot(root, ignoreRoot, c)
	if err != nil {
		return result, changed, err
	}
	messages = append(messages, result.Messages...)
	result.Messages = messages
	if len(result.Updates) > 0 {
		return result, changed, fmt.Errorf("index reconciliation did not converge after %d passes", maxIndexConvergencePasses)
	}
	return result, changed, nil
}

func apply(result model.ReconcileResult) (int, error) {
	changed := 0
	for _, u := range result.Updates {
		if u.OldText != nil && *u.OldText == u.NewText {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(u.Path), 0o755); err != nil {
			return changed, fmt.Errorf("create parent for %s: %w", u.Path, err)
		}
		written, err := applyUpdate(u, false)
		if err != nil {
			return changed, err
		}
		if written {
			changed++
		}
	}
	return changed, nil
}

func applyWithin(result model.ReconcileResult) (int, error) {
	changed := 0
	for _, update := range result.Updates {
		if update.OldText != nil && *update.OldText == update.NewText {
			continue
		}
		parentExists, err := existingParent(update.Path)
		if err != nil {
			return changed, err
		}
		if !parentExists {
			continue
		}
		written, err := applyUpdate(update, true)
		if err != nil {
			return changed, err
		}
		if written {
			changed++
		}
	}
	return changed, nil
}

func existingParent(path string) (bool, error) {
	parent := filepath.Dir(path)
	info, err := os.Stat(parent)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect parent for %s: %w", path, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("parent is not a directory for %s", path)
	}
	return true, nil
}

func applyUpdate(update model.FileUpdate, rejectStale bool) (bool, error) {
	data := textio.EncodeNew(update.NewText)
	if update.OldText == nil {
		if rejectStale {
			file, err := os.OpenFile(update.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if os.IsExist(err) {
				return false, nil
			}
			if err != nil {
				return false, fmt.Errorf("create %s: %w", update.Path, err)
			}
			if _, err := file.Write(data); err != nil {
				_ = file.Close()
				return false, fmt.Errorf("write %s: %w", update.Path, err)
			}
			if err := file.Close(); err != nil {
				return false, fmt.Errorf("close %s: %w", update.Path, err)
			}
			return true, nil
		}
	} else {
		doc, err := textio.Read(update.Path)
		if os.IsNotExist(err) && rejectStale {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("read before write %s: %w", update.Path, err)
		}
		if rejectStale && doc.Text != *update.OldText {
			return false, nil
		}
		data = doc.Encode(update.NewText)
	}
	if err := os.WriteFile(update.Path, data, 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", update.Path, err)
	}
	return true, nil
}

func orderedFolders(tree model.DocsTree) []*model.FolderInfo {
	out := make([]*model.FolderInfo, 0, len(tree.Folders))
	for _, f := range tree.Folders {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		di := len(strings.Split(filepath.Clean(out[i].Path), string(filepath.Separator)))
		dj := len(strings.Split(filepath.Clean(out[j].Path), string(filepath.Separator)))
		if di != dj {
			return di < dj
		}
		left, right := out[i].Path, out[j].Path
		if runtime.GOOS == "windows" {
			left, right = strings.ToLower(left), strings.ToLower(right)
		}
		return left < right
	})
	return out
}

func pathOrderKey(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}

func rootChildTitles(root string, folders []*model.FolderInfo, texts map[string]string, c config.Config) ([]string, error) {
	var result []string
	rootIndex := filepath.Join(root, c.IndexFile)
	for _, f := range folders {
		if f.IndexPath != "" {
			if source, ok := texts[f.Path]; ok {
				if t := parentTitleForRoot(f.IndexPath, source, rootIndex, c.ParentLink.Label); t != "" {
					result = append(result, t)
				}
			}
		}
		for _, p := range append(append([]string{}, f.DirectFiles...), f.StubFiles...) {
			if !config.IsParentEditable(p, c) {
				continue
			}
			doc, err := textio.Read(p)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if t := parentTitleForRoot(p, doc.Text, rootIndex, c.ParentLink.Label); t != "" {
				result = append(result, t)
			}
		}
	}
	return result, nil
}
func parentTitleForRoot(path, source, rootIndex, label string) string {
	for _, line := range strings.Split(source, "\n") {
		title, target := md.ParentLineParts(line, label)
		if title == "" {
			continue
		}
		absolute, _ := filepath.Abs(filepath.Join(filepath.Dir(path), filepath.FromSlash(target)))
		rootAbs, _ := filepath.Abs(rootIndex)
		if filepath.Clean(absolute) == filepath.Clean(rootAbs) {
			return title
		}
	}
	return ""
}

func updateSections(f *model.FolderInfo, source string, existing []*model.IndexEntry, crossFiles map[string][]*model.IndexEntry, fileCounts map[string]int, crossFolders map[string][]*model.IndexEntry, folderCounts map[string]int, matched map[*model.IndexEntry]bool, c config.Config) (string, error) {
	ensured := md.EnsureManaged(source, c)
	byTarget := entriesByTarget(f.IndexPath, existing)
	files := []string{}
	for _, p := range f.DirectFiles {
		files = append(files, renderFile(f.IndexPath, p, false, byTarget, existing, crossFiles, fileCounts, matched, c))
	}
	stubs := []string{}
	for _, p := range f.StubFiles {
		stubs = append(stubs, renderFile(f.IndexPath, p, true, byTarget, existing, crossFiles, fileCounts, matched, c))
	}
	folders := []string{}
	for _, p := range f.Subfolders {
		folders = append(folders, renderFolder(f.IndexPath, p, byTarget, crossFolders, folderCounts, matched, c))
	}
	var err error
	ensured, err = md.ReplaceManaged(ensured, "files", files, c)
	if err != nil {
		return "", err
	}
	ensured, err = md.ReplaceManaged(ensured, "stubs", stubs, c)
	if err != nil {
		return "", err
	}
	return md.ReplaceManaged(ensured, "folders", folders, c)
}
func targetKey(section, path string) string {
	abs, _ := filepath.Abs(path)
	return section + "\x00" + filepath.Clean(abs)
}
func entriesByTarget(index string, entries []*model.IndexEntry) map[string]*model.IndexEntry {
	r := map[string]*model.IndexEntry{}
	for _, e := range entries {
		r[targetKey(e.Section, filepath.Join(filepath.Dir(index), filepath.FromSlash(e.LinkTarget)))] = e
	}
	return r
}
func canonical(index, target string) string {
	r, _ := pathutil.Relative(target, filepath.Dir(index))
	return r
}
func renderFile(index, path string, stub bool, by map[string]*model.IndexEntry, existing []*model.IndexEntry, cross map[string][]*model.IndexEntry, counts map[string]int, matched map[*model.IndexEntry]bool, c config.Config) string {
	section := "files"
	if stub {
		section = "stubs"
	}
	if e := by[targetKey(section, path)]; e != nil {
		matched[e] = true
		return md.RenderFile(e.LinkText, canonical(index, path), e.Description)
	}
	if !stub {
		if e := oldStub(index, path, existing, c.Draft.Folder); e != nil {
			matched[e] = true
			return md.RenderFile(e.LinkText, canonical(index, path), promote(e.Description, c))
		}
	} else if e := oldDirect(index, path, existing); e != nil {
		matched[e] = true
		return md.RenderFile(e.LinkText, canonical(index, path), ensureStub(e.Description, c))
	}
	if counts[filepath.Base(path)] == 1 && len(cross[filepath.Base(path)]) == 1 {
		e := cross[filepath.Base(path)][0]
		matched[e] = true
		d := promote(e.Description, c)
		if stub {
			d = ensureStub(e.Description, c)
		}
		return md.RenderFile(filepath.Base(path), canonical(index, path), d)
	}
	return md.RenderFile(filepath.Base(path), canonical(index, path), md.DescriptionFromFile(path, stub, c))
}
func renderFolder(index, folder string, by map[string]*model.IndexEntry, cross map[string][]*model.IndexEntry, counts map[string]int, matched map[*model.IndexEntry]bool, c config.Config) string {
	target := filepath.Join(folder, c.IndexFile)
	if e := by[targetKey("folders", target)]; e != nil {
		matched[e] = true
		return md.RenderFolder(e.LinkText, canonical(index, target), e.Description)
	}
	if counts[filepath.Base(folder)] == 1 && len(cross[filepath.Base(folder)]) == 1 {
		e := cross[filepath.Base(folder)][0]
		matched[e] = true
		return md.RenderFolder(e.LinkText, canonical(index, target), e.Description)
	}
	return md.RenderFolder(filepath.Base(folder), canonical(index, target), md.DescriptionFromFolder(folder, c))
}
func oldStub(index, path string, entries []*model.IndexEntry, draft string) *model.IndexEntry {
	expected := targetKey("stubs", filepath.Join(filepath.Dir(index), draft, filepath.Base(path)))
	for _, e := range entries {
		if targetKey(e.Section, filepath.Join(filepath.Dir(index), filepath.FromSlash(e.LinkTarget))) == expected {
			return e
		}
	}
	return nil
}
func oldDirect(index, path string, entries []*model.IndexEntry) *model.IndexEntry {
	expected := targetKey("files", filepath.Join(filepath.Dir(index), filepath.Base(path)))
	for _, e := range entries {
		if targetKey(e.Section, filepath.Join(filepath.Dir(index), filepath.FromSlash(e.LinkTarget))) == expected {
			return e
		}
	}
	return nil
}
func promote(d string, c config.Config) string {
	if !strings.HasPrefix(d, c.Draft.DescriptionPrefix) {
		return d
	}
	d = strings.TrimPrefix(d, c.Draft.DescriptionPrefix)
	r := []rune(d)
	if len(r) > 0 && unicode.IsLower(r[0]) {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}
func ensureStub(d string, c config.Config) string {
	if strings.HasPrefix(d, c.Draft.DescriptionPrefix) {
		return d
	}
	return c.Draft.DescriptionPrefix + d
}
func crossFileEntries(all map[string][]*model.IndexEntry) map[string][]*model.IndexEntry {
	r := map[string][]*model.IndexEntry{}
	for _, es := range all {
		for _, e := range es {
			if e.Section != "files" && e.Section != "stubs" {
				continue
			}
			p := filepath.Join(filepath.Dir(e.IndexPath), filepath.FromSlash(e.LinkTarget))
			if _, err := os.Stat(p); err == nil {
				continue
			}
			r[filepath.Base(p)] = append(r[filepath.Base(p)], e)
		}
	}
	return r
}
func crossFolderEntries(all map[string][]*model.IndexEntry) map[string][]*model.IndexEntry {
	r := map[string][]*model.IndexEntry{}
	for _, es := range all {
		for _, e := range es {
			if e.Section != "folders" {
				continue
			}
			p := filepath.Join(filepath.Dir(e.IndexPath), filepath.FromSlash(e.LinkTarget))
			if _, err := os.Stat(p); err == nil {
				continue
			}
			name := filepath.Base(filepath.Dir(p))
			r[name] = append(r[name], e)
		}
	}
	return r
}
func unmatchedFileCounts(folders []*model.FolderInfo, all map[string][]*model.IndexEntry, c config.Config) map[string]int {
	r := map[string]int{}
	for _, f := range folders {
		if f.IndexPath == "" {
			continue
		}
		es := all[f.Path]
		by := entriesByTarget(f.IndexPath, es)
		for _, pair := range []struct {
			paths []string
			stub  bool
		}{{f.DirectFiles, false}, {f.StubFiles, true}} {
			for _, p := range pair.paths {
				section := "files"
				if pair.stub {
					section = "stubs"
				}
				stable := by[targetKey(section, p)] != nil
				if !stable {
					if pair.stub {
						stable = oldDirect(f.IndexPath, p, es) != nil
					} else {
						stable = oldStub(f.IndexPath, p, es, c.Draft.Folder) != nil
					}
				}
				if !stable {
					r[filepath.Base(p)]++
				}
			}
		}
	}
	return r
}
func unmatchedFolderCounts(folders []*model.FolderInfo, all map[string][]*model.IndexEntry, c config.Config) map[string]int {
	r := map[string]int{}
	for _, f := range folders {
		if f.IndexPath == "" || f.IsStubs {
			continue
		}
		es := all[f.Path]
		by := entriesByTarget(f.IndexPath, es)
		for _, child := range f.Subfolders {
			if by[targetKey("folders", filepath.Join(child, c.IndexFile))] == nil {
				r[filepath.Base(child)]++
			}
		}
	}
	return r
}
