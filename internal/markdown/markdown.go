package markdown

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/Lokee86/doc-ledger/internal/config"
	"github.com/Lokee86/doc-ledger/internal/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var entryPattern = regexp.MustCompile(`^\s*-\s+\[([^\]]+)\]\(([^)]+)\)\s+-\s+(.*)$`)
var parentPatternCache = map[string]*regexp.Regexp{}
var sections = []string{"files", "stubs", "folders"}

type heading struct {
	Start, End, Level int
	Line, Title       string
}

type sourceRange struct{ Start, End int }

func fencedCodeRanges(source string) []sourceRange {
	data := []byte(source)
	doc := goldmark.DefaultParser().Parse(text.NewReader(data))
	var result []sourceRange
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		block, ok := n.(*ast.FencedCodeBlock)
		if !ok || block.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		lines := block.Lines()
		result = append(result, sourceRange{Start: lines.At(0).Start, End: lines.At(lines.Len() - 1).Stop})
		return ast.WalkContinue, nil
	})
	return result
}

func inRanges(offset int, ranges []sourceRange) bool {
	for _, r := range ranges {
		if offset >= r.Start && offset < r.End {
			return true
		}
	}
	return false
}

func structuralIndex(source, value string, from int, ranges []sourceRange) int {
	for from <= len(source) {
		relative := strings.Index(source[from:], value)
		if relative < 0 {
			return -1
		}
		position := from + relative
		if !inRanges(position, ranges) {
			return position
		}
		from = position + len(value)
	}
	return -1
}

func headings(source string) []heading {
	data := []byte(source)
	doc := goldmark.DefaultParser().Parse(text.NewReader(data))
	var result []heading
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		lines := h.Lines()
		if lines.Len() == 0 {
			return ast.WalkContinue, nil
		}
		seg := lines.At(0)
		start := seg.Start
		for start > 0 && data[start-1] != '\n' {
			start--
		}
		end := seg.Stop
		for end < len(data) && data[end] != '\n' {
			end++
		}
		line := string(data[start:end])
		title := strings.TrimSpace(string(h.Text(data)))
		result = append(result, heading{start, end, h.Level, line, title})
		return ast.WalkContinue, nil
	})
	return result
}

func FirstHeadingTitle(source string) string {
	hs := headings(source)
	if len(hs) == 0 {
		return ""
	}
	return hs[0].Title
}
func TitleFromName(path string) string {
	base := filepath.Base(path)
	base = strings.ReplaceAll(strings.ReplaceAll(base, "-", " "), "_", " ")
	parts := strings.Fields(base)
	for i, p := range parts {
		r := []rune(strings.ToLower(p))
		if len(r) > 0 {
			r[0] = unicode.ToUpper(r[0])
		}
		parts[i] = string(r)
	}
	return strings.Join(parts, " ")
}
func FolderTitle(path, source string) string {
	if source != "" {
		if h := FirstHeadingTitle(source); h != "" {
			return h
		}
	}
	return TitleFromName(path)
}
func ManagedRootTitle(path, source string, childTitles []string) string {
	uniq := map[string]bool{}
	one := ""
	for _, x := range childTitles {
		if x != "" {
			uniq[x] = true
			one = x
		}
	}
	if len(uniq) == 1 {
		return one
	}
	if source != "" {
		if h := FirstHeadingTitle(source); h != "" {
			return h
		}
	}
	return TitleFromName(path)
}
func DescriptionFromFile(path string, stub bool, c config.Config) string {
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	title := TitleFromName(stem)
	d := strings.ReplaceAll(c.Description.FileTemplate, "{title}", title)
	if stub {
		return c.Draft.DescriptionPrefix + d
	}
	return d
}
func DescriptionFromFolder(path string, c config.Config) string {
	return strings.ReplaceAll(c.Description.FolderTemplate, "{title}", TitleFromName(path))
}
func MarkerStart(prefix, section string) string {
	return fmt.Sprintf("<!-- %s:%s:start -->", prefix, section)
}
func MarkerEnd(prefix, section string) string {
	return fmt.Sprintf("<!-- %s:%s:end -->", prefix, section)
}
func titleMap(c config.Config) map[string]string {
	return map[string]string{"files": "## " + c.Sections.FilesHeading, "stubs": "## " + c.Sections.StubsHeading, "folders": "## " + c.Sections.FoldersHeading}
}
func startMap(c config.Config) map[string]string {
	return map[string]string{"files": MarkerStart(c.Markers.Prefix, "files"), "stubs": MarkerStart(c.Markers.Prefix, "stubs"), "folders": MarkerStart(c.Markers.Prefix, "folders")}
}
func endMap(c config.Config) map[string]string {
	return map[string]string{"files": MarkerEnd(c.Markers.Prefix, "files"), "stubs": MarkerEnd(c.Markers.Prefix, "stubs"), "folders": MarkerEnd(c.Markers.Prefix, "folders")}
}
func aliases(c config.Config) map[string][]string {
	f := []string{}
	for _, x := range c.Sections.LegacyFilesHeadings {
		f = append(f, "## "+x)
	}
	d := []string{}
	for _, x := range c.Sections.LegacyFoldersHeadings {
		d = append(d, "## "+x)
	}
	return map[string][]string{"files": f, "folders": d}
}
func sectionFromHeading(line string, c config.Config) string {
	for k, v := range titleMap(c) {
		if line == v {
			return k
		}
	}
	for k, values := range aliases(c) {
		for _, v := range values {
			if line == v {
				return k
			}
		}
	}
	return ""
}

func EnsureManaged(source string, c config.Config) string {
	titles, starts, ends := titleMap(c), startMap(c), endMap(c)
	hs := headings(source)
	ranges := fencedCodeRanges(source)
	legacy := false
	for _, h := range hs {
		if sectionFromHeading(h.Line, c) != "" {
			legacy = true
			break
		}
	}
	anyMarkers := false
	for _, s := range sections {
		if structuralIndex(source, starts[s], 0, ranges) >= 0 || structuralIndex(source, ends[s], 0, ranges) >= 0 {
			anyMarkers = true
		}
	}
	if legacy && !anyMarkers {
		source = wrapLegacy(source, c)
		hs = headings(source)
		ranges = fencedCodeRanges(source)
	} else if legacy {
		source = normalizeLegacy(source, c)
		hs = headings(source)
		ranges = fencedCodeRanges(source)
	}
	missing := []string{}
	for _, s := range sections {
		present := structuralIndex(source, starts[s], 0, ranges) >= 0 && structuralIndex(source, ends[s], 0, ranges) >= 0
		if !present {
			for _, h := range hs {
				if h.Line == titles[s] {
					present = true
					break
				}
			}
		}
		if !present {
			missing = append(missing, s)
		}
	}
	if len(missing) == 0 {
		return source
	}
	parts := []string{}
	for _, s := range missing {
		parts = append(parts, titles[s]+"\n\n"+starts[s]+"\n\n"+ends[s])
	}
	block := strings.Join(parts, "\n\n")
	anchor := -1
	for _, h := range headings(source) {
		if h.Line == "## Related Docs" || h.Line == "## Notes" {
			anchor = h.Start
			break
		}
	}
	if anchor < 0 {
		if source != "" {
			return source + "\n\n" + block
		}
		return block
	}
	before, after := source[:anchor], source[anchor:]
	if before != "" && !strings.HasSuffix(before, "\n\n") {
		before = strings.TrimRight(before, "\n") + "\n\n"
	}
	return before + block + "\n\n" + strings.TrimLeft(after, " \t\r\n")
}

func wrapLegacy(source string, c config.Config) string {
	hs := headings(source)
	if len(hs) == 0 {
		return source
	}
	starts, ends, titles := startMap(c), endMap(c), titleMap(c)
	out := strings.Builder{}
	cursor := 0
	for i, h := range hs {
		s := sectionFromHeading(h.Line, c)
		if s == "" {
			continue
		}
		out.WriteString(source[cursor:h.Start])
		out.WriteString(titles[s])
		bodyStart := h.End
		if bodyStart < len(source) && source[bodyStart] == '\n' {
			bodyStart++
		}
		bodyEnd := len(source)
		if i+1 < len(hs) {
			bodyEnd = hs[i+1].Start
		}
		body := source[bodyStart:bodyEnd]
		body = strings.TrimSuffix(body, "\n")
		out.WriteString("\n" + starts[s])
		if body != "" {
			out.WriteString("\n\n" + body)
		}
		out.WriteString("\n" + ends[s])
		if bodyEnd < len(source) {
			out.WriteString("\n")
		}
		cursor = bodyEnd
	}
	out.WriteString(source[cursor:])
	return out.String()
}
func normalizeLegacy(source string, c config.Config) string {
	result := source
	hs := headings(source)
	for i := len(hs) - 1; i >= 0; i-- {
		h := hs[i]
		s := sectionFromHeading(h.Line, c)
		if s == "" {
			continue
		}
		canonical := titleMap(c)[s]
		if h.Line != canonical {
			result = result[:h.Start] + canonical + result[h.End:]
		}
	}
	return result
}

func ReplaceManaged(source, section string, lines []string, c config.Config) (string, error) {
	valid := false
	for _, s := range sections {
		if s == section {
			valid = true
		}
	}
	if !valid {
		return "", fmt.Errorf("unknown managed section: %s", section)
	}
	source = EnsureManaged(source, c)
	startMarker, endMarker := startMap(c)[section], endMap(c)[section]
	ranges := fencedCodeRanges(source)
	start := structuralIndex(source, startMarker, 0, ranges)
	end := structuralIndex(source, endMarker, start+len(startMarker), ranges)
	spanEnd := 0
	if end >= 0 {
		spanEnd = end + len(endMarker)
	} else {
		spanEnd = findSectionEnd(source, start, section, c)
	}
	block := []string{startMarker}
	if len(lines) > 0 {
		block = append(block, "")
		block = append(block, lines...)
	}
	block = append(block, endMarker)
	return source[:start] + strings.Join(block, "\n") + source[spanEnd:], nil
}
func findSectionEnd(source string, start int, section string, c config.Config) int {
	for _, h := range headings(source) {
		if h.Start > start {
			return h.Start
		}
	}
	starts := startMap(c)
	ranges := fencedCodeRanges(source)
	for _, s := range sections {
		if s != section {
			if p := structuralIndex(source, starts[s], start+len(starts[section]), ranges); p >= 0 {
				return p
			}
		}
	}
	return len(source)
}

func ParseEntries(indexPath, source string, c config.Config) []*model.IndexEntry {
	starts, ends := startMap(c), endMap(c)
	ranges := fencedCodeRanges(source)
	headingSections := map[int]string{}
	for _, h := range headings(source) {
		headingSections[h.Start] = sectionFromHeading(h.Line, c)
	}
	result := []*model.IndexEntry{}
	current := ""
	offset := 0
	for _, line := range strings.Split(source, "\n") {
		if inRanges(offset, ranges) {
			offset += len(line) + 1
			continue
		}
		if s := markerSection(line, starts); s != "" {
			current = s
			offset += len(line) + 1
			continue
		}
		if containsValue(ends, line) {
			current = ""
			offset += len(line) + 1
			continue
		}
		if s, ok := headingSections[offset]; ok {
			current = s
			offset += len(line) + 1
			continue
		}
		if current != "" {
			if m := entryPattern.FindStringSubmatch(line); m != nil {
				result = append(result, &model.IndexEntry{
					IndexPath: indexPath, Section: current, LinkText: m[1],
					LinkTarget: m[2], Description: m[3], OriginalLine: line,
				})
			}
		}
		offset += len(line) + 1
	}
	return result
}
func markerSection(line string, m map[string]string) string {
	for k, v := range m {
		if line == v {
			return k
		}
	}
	return ""
}
func containsValue(m map[string]string, v string) bool {
	for _, x := range m {
		if x == v {
			return true
		}
	}
	return false
}
func RenderFile(text, target, description string) string {
	return fmt.Sprintf("- [%s](%s) - %s", text, target, description)
}
func RenderFolder(text, target, description string) string {
	return RenderFile(text, target, description)
}

func MakeTemplate(folder, root, parentTitle, indexFile string, c config.Config) string {
	title := TitleFromName(folder)
	starts, ends, titles := startMap(c), endMap(c), titleMap(c)
	lines := []string{"# " + title, "", "This index summarizes the " + strings.ToLower(title) + " docs."}
	if filepath.Clean(folder) != filepath.Clean(root) && parentTitle != "" && c.ParentLink.FolderIndexes {
		lines = append(lines, "", fmt.Sprintf("%s: [%s](../%s)", c.ParentLink.Label, parentTitle, indexFile))
	}
	if c.Template.IncludeOwnership {
		lines = append(lines, "", "## Ownership", "", "Describe who maintains these docs.")
	}
	if c.Template.IncludeDoesNotBelong {
		lines = append(lines, "", "## Does Not Belong", "", "List content that belongs somewhere else.")
	}
	for _, s := range sections {
		lines = append(lines, "", titles[s], starts[s], ends[s])
	}
	if c.Template.IncludeRelatedDocs {
		lines = append(lines, "", "## Related Docs", "", "Add hand-picked links that help readers continue.")
	}
	if c.Template.IncludeNotes {
		lines = append(lines, "", "## Notes", "", "Add brief context that does not fit above.")
	}
	return strings.Join(lines, "\n")
}

func DesiredParent(path, root string, title func(string) string, c config.Config) string {
	if filepath.Clean(path) == filepath.Join(filepath.Clean(root), c.IndexFile) {
		return ""
	}
	if filepath.Base(path) == c.IndexFile {
		if !c.ParentLink.FolderIndexes {
			return ""
		}
		return fmt.Sprintf("%s: [%s](../%s)", c.ParentLink.Label, title(filepath.Dir(filepath.Dir(path))), c.IndexFile)
	}
	if !config.IsParentEditable(path, c) || !c.ParentLink.IndexedFiles {
		return ""
	}
	if filepath.Base(filepath.Dir(path)) == c.Draft.Folder {
		return fmt.Sprintf("%s: [%s](../%s)", c.ParentLink.Label, title(filepath.Dir(filepath.Dir(path))), c.IndexFile)
	}
	return fmt.Sprintf("%s: [%s](./%s)", c.ParentLink.Label, title(filepath.Dir(path)), c.IndexFile)
}
func UpdateParent(source, desired, label string) string {
	re := parentPatternCache[label]
	if re == nil {
		re = regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(label) + `:\s+.*$`)
		parentPatternCache[label] = re
	}
	loc := re.FindStringIndex(source)
	if loc != nil {
		if desired != "" {
			return source[:loc[0]] + desired + source[loc[1]:]
		}
		start, end := loc[0], loc[1]
		if start >= 2 && source[start-2:start] == "\n\n" && end+1 < len(source) && source[end:end+2] == "\n\n" {
			end += 2
		}
		return source[:start] + source[end:]
	}
	if desired == "" {
		return source
	}
	hs := headings(source)
	if len(hs) == 0 {
		if source != "" {
			return desired + "\n\n" + source
		}
		return desired
	}
	h := hs[0]
	suffix := source[h.End:]
	suffix = strings.TrimPrefix(suffix, "\n")
	suffix = strings.TrimPrefix(suffix, "\n")
	result := source[:h.End] + "\n\n" + desired
	if suffix != "" {
		return result + "\n\n" + suffix
	}
	if strings.HasSuffix(source, "\n") {
		return result + "\n"
	}
	return result
}
func ParentLineParts(line, label string) (string, string) {
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(label) + `:\s+\[([^\]]+)\]\(([^)]+)\)\s*$`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}
