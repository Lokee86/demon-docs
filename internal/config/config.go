package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Lokee86/demon-docs/internal/repository"
)

const DefaultIndexFile = "README.md"

type Marker struct{ Prefix string }
type ParentLink struct {
	Label                       string
	FolderIndexes, IndexedFiles bool
}
type Sections struct {
	FilesHeading, StubsHeading, FoldersHeading string
	LegacyFilesHeadings, LegacyFoldersHeadings []string
}
type Draft struct{ Folder, DescriptionPrefix string }
type Files struct {
	IndexFile                                                       string
	IncludePatterns, ExcludePatterns, EditableParentIndexExtensions []string
}
type Description struct{ FileTemplate, FolderTemplate string }
type Watch struct {
	DebounceSeconds              float64
	IgnoredDirs, IgnoredSuffixes []string
}
type Demon struct{ Run bool }
type ReverseIndex struct{ Roots []string }
type Codemap struct{ Headings []string }
type Template struct {
	ManagedSections                                                          []string
	IncludeOwnership, IncludeDoesNotBelong, IncludeRelatedDocs, IncludeNotes bool
}
type Config struct {
	Root, IndexFile string
	Markers         Marker
	ParentLink      ParentLink
	Sections        Sections
	Draft           Draft
	Files           Files
	Description     Description
	Watch           Watch
	Template        Template
	Demon           Demon
	ReverseIndex    ReverseIndex
	Codemap         Codemap
}

func Default() Config {
	return Config{
		Root: "docs", IndexFile: DefaultIndexFile,
		Markers: Marker{"doc-ledger"}, ParentLink: ParentLink{"Parent index", true, false},
		Sections: Sections{"Direct Files", "Stub Files", "Direct Folders", []string{"Top-Level Files"}, []string{"Top-Level Folders"}},
		Draft:    Draft{"stubs", "Stub: "}, Files: Files{DefaultIndexFile, []string{"**/*.md"}, []string{}, []string{".md"}},
		Description:  Description{"{title} documentation.", "{title} documentation."},
		Watch:        Watch{0.75, []string{".cache", "__pycache__"}, []string{"~", ".swp", ".tmp", ".bak"}},
		Template:     Template{[]string{"files", "stubs", "folders"}, true, true, true, true},
		Demon:        Demon{Run: true},
		ReverseIndex: ReverseIndex{Roots: []string{}},
		Codemap: Codemap{Headings: []string{
			"Code map",
			"Codemap",
			"Code or source map",
			"Code and test map",
		}},
	}
}

func StarterText() string {
	return "root = \"docs\"\n" + starterBody()
}

func RepositoryStarterText(docsRoot string) string {
	return "docs_root = " + strconv.Quote(docsRoot) + "\n" + starterBody()
}

func starterBody() string {
	return "index_file = \"README.md\"\n\n[reverse_index]\nroots = []\n\n[codemap]\nheadings = [\"Code map\", \"Codemap\", \"Code or source map\", \"Code and test map\"]\n\n[demon]\nrun = true\n\n[parent_link]\nfolder_indexes = true\nindexed_files = false\n\n[drafts]\nfolder = \"stubs\"\ndescription_prefix = \"Stub: \"\n\n[watch]\ndebounce_seconds = 0.75\nignored_dirs = [\".cache\", \"__pycache__\"]\nignored_suffixes = [\"~\", \".swp\", \".tmp\", \".bak\"]\n"
}

type rawConfig struct {
	Root      *string `toml:"root"`
	DocsRoot  *string `toml:"docs_root"`
	IndexFile *string `toml:"index_file"`
	Markers   *struct {
		Prefix *string `toml:"prefix"`
	} `toml:"markers"`
	ParentLink *struct {
		Label         *string `toml:"label"`
		Enabled       *bool   `toml:"enabled"`
		FolderIndexes *bool   `toml:"folder_indexes"`
		IndexedFiles  *bool   `toml:"indexed_files"`
	} `toml:"parent_link"`
	Sections map[string]map[string]any `toml:"sections"`
	Drafts   *struct {
		Folder            *string `toml:"folder"`
		DescriptionPrefix *string `toml:"description_prefix"`
	} `toml:"drafts"`
	Files *struct {
		IncludePatterns *[]string `toml:"include_patterns"`
		ExcludePatterns *[]string `toml:"exclude_patterns"`
	} `toml:"files"`
	Editable *struct {
		ParentIndexExtensions *[]string `toml:"parent_index_extensions"`
		Extensions            *[]string `toml:"extensions"`
	} `toml:"editable"`
	Descriptions *struct {
		FileTemplate   *string `toml:"file_template"`
		FolderTemplate *string `toml:"folder_template"`
	} `toml:"descriptions"`
	Watch *struct {
		DebounceSeconds *float64  `toml:"debounce_seconds"`
		IgnoredDirs     *[]string `toml:"ignored_dirs"`
		IgnoredSuffixes *[]string `toml:"ignored_suffixes"`
	} `toml:"watch"`
	Demon *struct {
		Run *bool `toml:"run"`
	} `toml:"demon"`
	ReverseIndex *struct {
		Roots   *[]string `toml:"roots"`
		Folders *[]string `toml:"folders"`
	} `toml:"reverse_index"`
	Codemap *struct {
		Headings *[]string `toml:"headings"`
	} `toml:"codemap"`
	Aliases *struct {
		Files   *[]string `toml:"files"`
		Folders *[]string `toml:"folders"`
	} `toml:"aliases"`
	Template *struct {
		ManagedSections      *[]string `toml:"managed_sections"`
		IncludeOwnership     *bool     `toml:"include_ownership"`
		IncludeDoesNotBelong *bool     `toml:"include_does_not_belong"`
		IncludeRelatedDocs   *bool     `toml:"include_related_docs"`
		IncludeNotes         *bool     `toml:"include_notes"`
	} `toml:"template"`
}

func Load(path string) (Config, error) {
	c := Default()
	var raw rawConfig
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return c, fmt.Errorf("config %s: %w", path, err)
	}
	if raw.Root != nil {
		c.Root = *raw.Root
	}
	if raw.DocsRoot != nil {
		c.Root = *raw.DocsRoot
	}
	if raw.IndexFile != nil {
		c.IndexFile = *raw.IndexFile
		c.Files.IndexFile = *raw.IndexFile
	}
	if raw.Markers != nil && raw.Markers.Prefix != nil {
		c.Markers.Prefix = *raw.Markers.Prefix
	}
	if p := raw.ParentLink; p != nil {
		if p.Label != nil {
			c.ParentLink.Label = *p.Label
		}
		if p.Enabled != nil {
			if p.FolderIndexes == nil {
				c.ParentLink.FolderIndexes = *p.Enabled
			}
			if p.IndexedFiles == nil {
				c.ParentLink.IndexedFiles = *p.Enabled
			}
		}
		if p.FolderIndexes != nil {
			c.ParentLink.FolderIndexes = *p.FolderIndexes
		}
		if p.IndexedFiles != nil {
			c.ParentLink.IndexedFiles = *p.IndexedFiles
		}
	}
	for section, field := range map[string]*string{"files": &c.Sections.FilesHeading, "stubs": &c.Sections.StubsHeading, "folders": &c.Sections.FoldersHeading} {
		if values := raw.Sections[section]; values != nil {
			for _, key := range []string{"heading", "title", "name"} {
				if value, ok := values[key]; ok {
					*field = fmt.Sprint(value)
					break
				}
			}
		}
	}
	if d := raw.Drafts; d != nil {
		if d.Folder != nil {
			c.Draft.Folder = *d.Folder
		}
		if d.DescriptionPrefix != nil {
			c.Draft.DescriptionPrefix = *d.DescriptionPrefix
		}
	}
	if f := raw.Files; f != nil {
		if f.IncludePatterns != nil {
			c.Files.IncludePatterns = *f.IncludePatterns
		}
		if f.ExcludePatterns != nil {
			c.Files.ExcludePatterns = *f.ExcludePatterns
		}
	}
	if e := raw.Editable; e != nil {
		if e.ParentIndexExtensions != nil {
			c.Files.EditableParentIndexExtensions = *e.ParentIndexExtensions
		} else if e.Extensions != nil {
			c.Files.EditableParentIndexExtensions = *e.Extensions
		}
	}
	if d := raw.Descriptions; d != nil {
		if d.FileTemplate != nil {
			c.Description.FileTemplate = *d.FileTemplate
		}
		if d.FolderTemplate != nil {
			c.Description.FolderTemplate = *d.FolderTemplate
		}
	}
	if w := raw.Watch; w != nil {
		if w.DebounceSeconds != nil {
			c.Watch.DebounceSeconds = *w.DebounceSeconds
		}
		if w.IgnoredDirs != nil {
			c.Watch.IgnoredDirs = *w.IgnoredDirs
		}
		if w.IgnoredSuffixes != nil {
			c.Watch.IgnoredSuffixes = *w.IgnoredSuffixes
		}
	}
	if d := raw.Demon; d != nil && d.Run != nil {
		c.Demon.Run = *d.Run
	}
	if r := raw.ReverseIndex; r != nil {
		if r.Roots != nil {
			c.ReverseIndex.Roots = *r.Roots
		} else if r.Folders != nil {
			c.ReverseIndex.Roots = *r.Folders
		}
	}
	if m := raw.Codemap; m != nil && m.Headings != nil {
		c.Codemap.Headings = *m.Headings
	}
	if a := raw.Aliases; a != nil {
		if a.Files != nil {
			c.Sections.LegacyFilesHeadings = *a.Files
		}
		if a.Folders != nil {
			c.Sections.LegacyFoldersHeadings = *a.Folders
		}
	}
	if t := raw.Template; t != nil {
		if t.ManagedSections != nil {
			c.Template.ManagedSections = *t.ManagedSections
		}
		if t.IncludeOwnership != nil {
			c.Template.IncludeOwnership = *t.IncludeOwnership
		}
		if t.IncludeDoesNotBelong != nil {
			c.Template.IncludeDoesNotBelong = *t.IncludeDoesNotBelong
		}
		if t.IncludeRelatedDocs != nil {
			c.Template.IncludeRelatedDocs = *t.IncludeRelatedDocs
		}
		if t.IncludeNotes != nil {
			c.Template.IncludeNotes = *t.IncludeNotes
		}
	}
	return c, nil
}

// SetDemonRun changes only the [demon].run setting and atomically replaces the
// config file. The line-oriented edit intentionally keeps comments, unknown
// keys, and formatting outside this setting intact.
func SetDemonRun(path string, enabled bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)
	updated := setDemonRunText(text, enabled)
	if updated == text {
		return nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config.toml-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(updated); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func setDemonRunText(text string, enabled bool) string {
	value := strconv.FormatBool(enabled)
	lines := strings.SplitAfter(text, "\n")
	section := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\r\n"))
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if trimmed == "[demon]" {
				section = i
				continue
			}
			if section >= 0 {
				lines = append(lines[:i], append([]string{"run = " + value + "\n"}, lines[i:]...)...)
				return strings.Join(lines, "")
			}
		}
		if section >= 0 {
			content := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if strings.HasPrefix(strings.TrimSpace(content), "run") {
				prefix := content[:len(content)-len(strings.TrimLeft(content, " \t"))]
				rest := strings.TrimSpace(strings.TrimSpace(content)[len("run"):])
				if strings.HasPrefix(rest, "=") {
					comment := ""
					if at := strings.Index(rest, "#"); at >= 0 {
						comment = " " + strings.TrimSpace(rest[at:])
					}
					ending := "\n"
					if strings.HasSuffix(line, "\r\n") {
						ending = "\r\n"
					}
					lines[i] = prefix + "run = " + value + comment + ending
					return strings.Join(lines, "")
				}
			}
		}
	}
	if section >= 0 {
		ending := ""
		if len(text) > 0 && !strings.HasSuffix(text, "\n") {
			ending = "\n"
		}
		lines = append(lines, ending+"run = "+value+"\n")
		return strings.Join(lines, "")
	}
	separator := ""
	if text != "" && !strings.HasSuffix(text, "\n") {
		separator = "\n"
	}
	return text + separator + "\n[demon]\nrun = " + value + "\n"
}

func LocalPath(cwd string) string {
	for _, name := range []string{".demon-docs.toml", "demon-docs.toml", ".doc-ledger.toml", "doc-ledger.toml"} {
		p := filepath.Join(cwd, name)
		if exists(p) {
			return p
		}
	}
	return ""
}
func Discover(start string) string {
	p, _ := filepath.Abs(start)
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		p = filepath.Dir(p)
	}
	for {
		if found := LocalPath(p); found != "" {
			return found
		}
		parent := filepath.Dir(p)
		if parent == p {
			return ""
		}
		p = parent
	}
}

// DiscoverWithin searches legacy local configs upward without crossing an
// initialized Demon Docs repository boundary.
func DiscoverWithin(start, boundary string) string {
	p, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	if info, statErr := os.Stat(p); statErr == nil && !info.IsDir() {
		p = filepath.Dir(p)
	}
	boundary, err = filepath.Abs(boundary)
	if err != nil {
		return ""
	}
	for {
		if found := LocalPath(p); found != "" {
			return found
		}
		if p == boundary {
			return ""
		}
		parent := filepath.Dir(p)
		if parent == p || !repository.Contains(boundary, parent) && parent != boundary {
			return ""
		}
		p = parent
	}
}
func GlobalPath(env func(string) string, home string) string {
	return filepath.Join(configHome(env, home), "demon-docs", "config.toml")
}
func LegacyGlobalPath(env func(string) string, home string) string {
	return filepath.Join(configHome(env, home), "doc-ledger", "config.toml")
}
func configHome(env func(string) string, home string) string {
	if xdg := env("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	return filepath.Join(home, ".config")
}
func Select(cwd, explicit string, noLocal, noGlobal bool, env func(string) string, home string) string {
	if explicit != "" {
		return explicit
	}
	if !noLocal {
		if location, ok := repository.Discover(cwd); ok {
			return location.ConfigPath
		}
		if root, ok := repository.FindMarker(cwd); ok {
			if p := DiscoverWithin(cwd, root); p != "" {
				return p
			}
		}
		if p := LocalPath(cwd); p != "" {
			return p
		}
	}
	if !noGlobal {
		for _, p := range []string{GlobalPath(env, home), LegacyGlobalPath(env, home)} {
			if exists(p) {
				return p
			}
		}
	}
	return ""
}
func IsParentEditable(path string, c Config) bool {
	ext := filepath.Ext(path)
	for _, x := range c.Files.EditableParentIndexExtensions {
		if ext == x {
			return true
		}
	}
	return false
}
func exists(path string) bool { _, err := os.Stat(path); return err == nil }
