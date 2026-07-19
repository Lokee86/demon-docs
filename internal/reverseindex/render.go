package reverseindex

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	md "github.com/Lokee86/demon-docs/internal/markdown"
	"github.com/Lokee86/demon-docs/internal/pathutil"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func renderBlock(repositoryRoot, indexPath, folder string, files []string, f facts, c config.Config) string {
	lines := []string{markerStart(c), "", "## Code Files"}
	folderRelative, _ := filepath.Rel(repositoryRoot, folder)
	if docs := sortedReferences(f.folderDocs[filepath.ToSlash(filepath.Clean(folderRelative))]); len(docs) > 0 {
		lines = append(lines, "", "Folder documentation:")
		for _, document := range docs {
			lines = append(lines, "- "+documentLink(indexPath, repositoryRoot, document, f.titles))
		}
	}
	for _, file := range files {
		relative, _ := filepath.Rel(repositoryRoot, file)
		relative = filepath.ToSlash(filepath.Clean(relative))
		target, _ := pathutil.Relative(file, filepath.Dir(indexPath))
		lines = append(lines, "", fmt.Sprintf("- [%s](%s)", filepath.Base(file), target))
		for _, document := range sortedReferences(f.fileDocs[relative]) {
			lines = append(lines, "  - "+documentLink(indexPath, repositoryRoot, document, f.titles))
		}
	}
	lines = append(lines, "", markerEnd(c))
	return strings.Join(lines, "\n")
}

func documentLink(indexPath, repositoryRoot, document string, titles map[string]string) string {
	full := filepath.Join(repositoryRoot, filepath.FromSlash(document))
	target, _ := pathutil.Relative(full, filepath.Dir(indexPath))
	title := titles[document]
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(document), filepath.Ext(document))
	}
	return fmt.Sprintf("[%s](%s)", title, target)
}

func replaceManaged(source, block, folder string, c config.Config) (string, error) {
	start, end := markerStart(c), markerEnd(c)
	startAt, endAt := strings.Index(source, start), strings.Index(source, end)
	if startAt >= 0 || endAt >= 0 {
		if startAt < 0 || endAt < startAt {
			return "", fmt.Errorf("incomplete reverse-index markers")
		}
		endAt += len(end)
		return source[:startAt] + block + source[endAt:], nil
	}
	if source == "" {
		return fmt.Sprintf("# %s\n\nThis index maps code files to their documentation.\n\n%s", md.TitleFromName(folder), block), nil
	}
	return strings.TrimRight(source, "\r\n") + "\n\n" + block, nil
}

func documentTitle(repositoryRoot, relative string) string {
	doc, err := textio.Read(filepath.Join(repositoryRoot, filepath.FromSlash(relative)))
	if err == nil {
		if title := md.FirstHeadingTitle(doc.Text); title != "" {
			return title
		}
	}
	return strings.TrimSuffix(filepath.Base(relative), filepath.Ext(relative))
}

func addReference(target map[string]map[string]struct{}, path, document string) {
	if target[path] == nil {
		target[path] = map[string]struct{}{}
	}
	target[path][document] = struct{}{}
}

func sortedReferences(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func markerStart(c config.Config) string {
	return fmt.Sprintf("<!-- %s:%s:start -->", c.Markers.Prefix, section)
}
func markerEnd(c config.Config) string {
	return fmt.Sprintf("<!-- %s:%s:end -->", c.Markers.Prefix, section)
}
