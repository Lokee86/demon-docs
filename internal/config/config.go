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
type Review struct {
	UndoDepth      int
	UndoMaxAgeDays int
}
type Index struct{ Enabled bool }
type Links struct{ Enabled bool }
type ReverseIndex struct{ Roots []string }
type Codemap struct {
	Headings                []string
	RemoveUndiscoveredLinks bool
	RemoveLowScoreLinks     bool
}
type Template struct {
	ManagedSections                                                          []string
	IncludeOwnership, IncludeDoesNotBelong, IncludeRelatedDocs, IncludeNotes bool
}
type FrontmatterField struct {
	Type        string `toml:"type"`
	Required    bool   `toml:"required"`
	Immutable   bool   `toml:"immutable"`
	Generated   bool   `toml:"generated"`
	Default     any    `toml:"default"`
	DefaultFrom string `toml:"default_from"`
}
type FrontmatterRule struct {
	WhenField string `toml:"when_field"`
	Equals    any    `toml:"equals"`
	Require   string `toml:"require"`
}
type Frontmatter struct {
	Enabled        bool
	DefaultFormat  string
	AllowedFormats []string
	DefaultAuthor  string
	UnknownFields  string
	Fields         map[string]FrontmatterField
	Rules          []FrontmatterRule
}
type FormatPathRule struct {
	Pattern string `toml:"pattern"`
	Schema  string `toml:"schema"`
}
type Format struct {
	Enabled                bool
	SchemaDir              string
	DocumentSchemaDir      string
	DefaultSchema          string
	InvalidationSimilarity float64
	PathRules              []FormatPathRule
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
	Review          Review
	Index           Index
	Links           Links
	ReverseIndex    ReverseIndex
	Codemap         Codemap
	Frontmatter     Frontmatter
	Format          Format
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
		Review:       Review{UndoDepth: 100, UndoMaxAgeDays: 30},
		Index:        Index{Enabled: true},
		Links:        Links{Enabled: true},
		ReverseIndex: ReverseIndex{Roots: []string{}},
		Codemap: Codemap{Headings: []string{
			"Code map",
			"Codemap",
			"Code or source map",
			"Code and test map",
		}, RemoveUndiscoveredLinks: false, RemoveLowScoreLinks: false},
		Frontmatter: Frontmatter{
			DefaultFormat:  "yaml",
			AllowedFormats: []string{"yaml", "toml"},
			UnknownFields:  "remove",
		},
		Format: Format{
			Enabled:                false,
			SchemaDir:              ".ddocs/schemas",
			DocumentSchemaDir:      ".ddocs/document-schemas",
			DefaultSchema:          "general",
			InvalidationSimilarity: 0.5,
		},
	}
}

func StarterText() string {
	return "root = \"docs\"\n" + starterBody()
}

func RepositoryStarterText(docsRoot string) string {
	return "docs_root = " + strconv.Quote(docsRoot) + "\n" + starterBody()
}

func starterBody() string {
	return "index_file = \"README.md\"\n\n[index]\nenabled = true\n\n[format]\nenabled = true\nschema_dir = \".ddocs/schemas\"\ndocument_schema_dir = \".ddocs/document-schemas\"\ndefault_schema = \"general\"\ninvalidation_similarity = 0.5\n\n[[format.path_rules]]\npattern = \"**/README.md\"\nschema = \"index\"\n\n[[format.path_rules]]\npattern = \"**/!INDEX.md\"\nschema = \"index\"\n\n[[format.path_rules]]\npattern = \"**/planning/**\"\nschema = \"planning\"\n\n[[format.path_rules]]\npattern = \"**/services/**\"\nschema = \"service\"\n\n[links]\nenabled = true\n\n[reverse_index]\nroots = []\n\n[codemap]\nheadings = [\"Code map\", \"Codemap\", \"Code or source map\", \"Code and test map\"]\nremove_undiscovered_links = false\nremove_low_score_links = false\n\n[demon]\nrun = true\n\n[review]\nundo_depth = 100\nundo_max_age_days = 30\n\n[parent_link]\nfolder_indexes = true\nindexed_files = false\n\n[drafts]\nfolder = \"stubs\"\ndescription_prefix = \"Stub: \"\n\n[watch]\ndebounce_seconds = 0.75\nignored_dirs = [\".cache\", \"__pycache__\"]\nignored_suffixes = [\"~\", \".swp\", \".tmp\", \".bak\"]\n\n[frontmatter]\nenabled = true\ndefault_format = \"yaml\"\nallowed_formats = [\"yaml\", \"toml\"]\ndefault_author = \"\"\nunknown_fields = \"remove\"\n\n[frontmatter.fields.document_id]\ntype = \"uuid\"\nrequired = true\nimmutable = true\ngenerated = true\n\n[frontmatter.fields.author]\ntype = \"string\"\nrequired = true\ndefault_from = \"frontmatter.default_author\"\n\n[frontmatter.fields.document_type]\ntype = \"string\"\nrequired = true\ndefault = \"general\"\n\n[frontmatter.fields.created]\ntype = \"date\"\nrequired = true\nimmutable = true\ngenerated = true\n\n[frontmatter.fields.summary]\ntype = \"string\"\nrequired = true\n\n[frontmatter.fields.policy_exempt]\ntype = \"boolean\"\ndefault = false\n\n[frontmatter.fields.policy_exempt_reason]\ntype = \"string\"\n\n[[frontmatter.rules]]\nwhen_field = \"policy_exempt\"\nequals = true\nrequire = \"policy_exempt_reason\"\n"
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
	Review *struct {
		UndoDepth      *int `toml:"undo_depth"`
		UndoMaxAgeDays *int `toml:"undo_max_age_days"`
	} `toml:"review"`
	Index *struct {
		Enabled *bool `toml:"enabled"`
	} `toml:"index"`
	Links *struct {
		Enabled *bool `toml:"enabled"`
	} `toml:"links"`
	ReverseIndex *struct {
		Roots   *[]string `toml:"roots"`
		Folders *[]string `toml:"folders"`
	} `toml:"reverse_index"`
	Codemap *struct {
		Headings                *[]string `toml:"headings"`
		RemoveUndiscoveredLinks *bool     `toml:"remove_undiscovered_links"`
		RemoveLowScoreLinks     *bool     `toml:"remove_low_score_links"`
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
	Frontmatter *struct {
		Enabled        *bool                       `toml:"enabled"`
		DefaultFormat  *string                     `toml:"default_format"`
		AllowedFormats *[]string                   `toml:"allowed_formats"`
		DefaultAuthor  *string                     `toml:"default_author"`
		UnknownFields  *string                     `toml:"unknown_fields"`
		Fields         map[string]FrontmatterField `toml:"fields"`
		Rules          []FrontmatterRule           `toml:"rules"`
	} `toml:"frontmatter"`
	Format *struct {
		Enabled                *bool            `toml:"enabled"`
		SchemaDir              *string          `toml:"schema_dir"`
		DocumentSchemaDir      *string          `toml:"document_schema_dir"`
		DefaultSchema          *string          `toml:"default_schema"`
		InvalidationSimilarity *float64         `toml:"invalidation_similarity"`
		PathRules              []FormatPathRule `toml:"path_rules"`
	} `toml:"format"`
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
	if review := raw.Review; review != nil {
		if review.UndoDepth != nil {
			c.Review.UndoDepth = *review.UndoDepth
		}
		if review.UndoMaxAgeDays != nil {
			c.Review.UndoMaxAgeDays = *review.UndoMaxAgeDays
		}
	}
	if index := raw.Index; index != nil && index.Enabled != nil {
		c.Index.Enabled = *index.Enabled
	}
	if links := raw.Links; links != nil && links.Enabled != nil {
		c.Links.Enabled = *links.Enabled
	}
	if r := raw.ReverseIndex; r != nil {
		if r.Roots != nil {
			c.ReverseIndex.Roots = *r.Roots
		} else if r.Folders != nil {
			c.ReverseIndex.Roots = *r.Folders
		}
	}
	if m := raw.Codemap; m != nil {
		if m.Headings != nil {
			c.Codemap.Headings = *m.Headings
		}
		if m.RemoveUndiscoveredLinks != nil {
			c.Codemap.RemoveUndiscoveredLinks = *m.RemoveUndiscoveredLinks
		}
		if m.RemoveLowScoreLinks != nil {
			c.Codemap.RemoveLowScoreLinks = *m.RemoveLowScoreLinks
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
	if f := raw.Format; f != nil {
		if f.Enabled != nil {
			c.Format.Enabled = *f.Enabled
		}
		if f.SchemaDir != nil {
			c.Format.SchemaDir = *f.SchemaDir
		}
		if f.DocumentSchemaDir != nil {
			c.Format.DocumentSchemaDir = *f.DocumentSchemaDir
		}
		if f.DefaultSchema != nil {
			c.Format.DefaultSchema = *f.DefaultSchema
		}
		if f.InvalidationSimilarity != nil {
			c.Format.InvalidationSimilarity = *f.InvalidationSimilarity
		}
		if f.PathRules != nil {
			c.Format.PathRules = f.PathRules
		}
	}
	if f := raw.Frontmatter; f != nil {
		c.Frontmatter.Enabled = true
		if f.Enabled != nil {
			c.Frontmatter.Enabled = *f.Enabled
		}
		if f.DefaultFormat != nil {
			c.Frontmatter.DefaultFormat = *f.DefaultFormat
		}
		if f.AllowedFormats != nil {
			c.Frontmatter.AllowedFormats = *f.AllowedFormats
		}
		if f.DefaultAuthor != nil {
			c.Frontmatter.DefaultAuthor = *f.DefaultAuthor
		}
		if f.UnknownFields != nil {
			c.Frontmatter.UnknownFields = *f.UnknownFields
		}
		if f.Fields != nil {
			c.Frontmatter.Fields = f.Fields
		}
		if f.Rules != nil {
			c.Frontmatter.Rules = f.Rules
		}
	}
	return c, nil
}

// SetDemonRun changes only the [demon].run setting and atomically replaces the
// config file. The line-oriented edit intentionally keeps comments, unknown
// keys, and formatting outside this setting intact.
func SetDemonRun(path string, enabled bool) error {
	return setBoolean(path, "demon", "run", enabled)
}

func SetIndexEnabled(path string, enabled bool) error {
	return setBoolean(path, "index", "enabled", enabled)
}

func SetLinksEnabled(path string, enabled bool) error {
	return setBoolean(path, "links", "enabled", enabled)
}

func setBoolean(path, section, key string, enabled bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)
	updated := setBooleanText(text, section, key, enabled)
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
	return setBooleanText(text, "demon", "run", enabled)
}

func setBooleanText(text, sectionName, key string, enabled bool) string {
	value := strconv.FormatBool(enabled)
	lines := strings.SplitAfter(text, "\n")
	section := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\r\n"))
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if trimmed == "["+sectionName+"]" {
				section = i
				continue
			}
			if section >= 0 {
				lines = append(lines[:i], append([]string{key + " = " + value + "\n"}, lines[i:]...)...)
				return strings.Join(lines, "")
			}
		}
		if section >= 0 {
			content := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if strings.HasPrefix(strings.TrimSpace(content), key) {
				prefix := content[:len(content)-len(strings.TrimLeft(content, " \t"))]
				rest := strings.TrimSpace(strings.TrimSpace(content)[len(key):])
				if strings.HasPrefix(rest, "=") {
					comment := ""
					if at := strings.Index(rest, "#"); at >= 0 {
						comment = " " + strings.TrimSpace(rest[at:])
					}
					ending := "\n"
					if strings.HasSuffix(line, "\r\n") {
						ending = "\r\n"
					}
					lines[i] = prefix + key + " = " + value + comment + ending
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
		lines = append(lines, ending+key+" = "+value+"\n")
		return strings.Join(lines, "")
	}
	separator := ""
	if text != "" && !strings.HasSuffix(text, "\n") {
		separator = "\n"
	}
	return text + separator + "\n[" + sectionName + "]\n" + key + " = " + value + "\n"
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
