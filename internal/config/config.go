package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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
}

func Default() Config {
	return Config{
		Root: "docs", IndexFile: DefaultIndexFile,
		Markers: Marker{"doc-ledger"}, ParentLink: ParentLink{"Parent index", true, false},
		Sections: Sections{"Direct Files", "Stub Files", "Direct Folders", []string{"Top-Level Files"}, []string{"Top-Level Folders"}},
		Draft:    Draft{"stubs", "Stub: "}, Files: Files{DefaultIndexFile, []string{"**/*.md"}, []string{}, []string{".md"}},
		Description: Description{"{title} documentation.", "{title} documentation."},
		Watch:       Watch{0.75, []string{".git", ".cache", "__pycache__"}, []string{"~", ".swp", ".tmp", ".bak"}},
		Template:    Template{[]string{"files", "stubs", "folders"}, true, true, true, true},
	}
}

func StarterText() string {
	return "root = \"docs\"\nindex_file = \"README.md\"\n\n[parent_link]\nfolder_indexes = true\nindexed_files = false\n\n[drafts]\nfolder = \"stubs\"\ndescription_prefix = \"Stub: \"\n\n[watch]\ndebounce_seconds = 0.75\nignored_dirs = [\".git\", \".cache\", \"__pycache__\"]\nignored_suffixes = [\"~\", \".swp\", \".tmp\", \".bak\"]\n"
}

type rawConfig struct {
	Root      *string `toml:"root"`
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

func LocalPath(cwd string) string {
	for _, name := range []string{".doc-ledger.toml", "doc-ledger.toml"} {
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
func GlobalPath(env func(string) string, home string) string {
	if xdg := env("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "doc-ledger", "config.toml")
	}
	return filepath.Join(home, ".config", "doc-ledger", "config.toml")
}
func Select(cwd, explicit string, noLocal, noGlobal bool, env func(string) string, home string) string {
	if explicit != "" {
		return explicit
	}
	if !noLocal {
		if p := LocalPath(cwd); p != "" {
			return p
		}
	}
	if !noGlobal {
		p := GlobalPath(env, home)
		if exists(p) {
			return p
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
