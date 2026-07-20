package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultsAndAliases(t *testing.T) {
	c := Default()
	if c.Root != "docs" || c.IndexFile != "README.md" || c.Markers.Prefix != "doc-ledger" || !c.ParentLink.FolderIndexes || c.ParentLink.IndexedFiles || !c.Demon.Run || !c.Index.Enabled || !c.Links.Enabled {
		t.Fatalf("unexpected defaults: %+v", c)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	text := `index_file = "!README.md"
[parent_link]
enabled = false
folder_indexes = true
[sections.files]
title = "Pages"
[editable]
extensions = [".md", ".mdx"]
`
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.IndexFile != "!README.md" || c.Files.IndexFile != "!README.md" || !c.ParentLink.FolderIndexes || c.ParentLink.IndexedFiles || c.Sections.FilesHeading != "Pages" || !reflect.DeepEqual(c.Files.EditableParentIndexExtensions, []string{".md", ".mdx"}) {
		t.Fatalf("aliases not preserved: %+v", c)
	}
}

func TestCodemapHeadingsLoadFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[codemap]\nheadings = [\"Implementation map\", \"Source map\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Codemap.Headings, []string{"Implementation map", "Source map"}) {
		t.Fatalf("codemap headings not loaded: %+v", loaded.Codemap)
	}
	if !strings.Contains(StarterText(), "[codemap]\nheadings =") {
		t.Fatal("starter config omitted codemap headings")
	}
	for _, section := range []string{"[index]\nenabled = true", "[links]\nenabled = true"} {
		if !strings.Contains(StarterText(), section) {
			t.Fatalf("starter config omitted %s", section)
		}
	}
}

func TestReverseIndexRootsLoadWithFoldersCompatibilityAlias(t *testing.T) {
	dir := t.TempDir()
	for name, section := range map[string]string{
		"roots.toml":   "[reverse_index]\nroots = [\"client\", \"services/game-server\"]\n",
		"folders.toml": "[reverse_index]\nfolders = [\"client\"]\n",
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(section), 0o644); err != nil {
			t.Fatal(err)
		}
		loaded, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(loaded.ReverseIndex.Roots) == 0 || loaded.ReverseIndex.Roots[0] != "client" {
			t.Fatalf("reverse-index roots not loaded from %s: %+v", name, loaded.ReverseIndex)
		}
	}
	if !strings.Contains(StarterText(), "[reverse_index]\nroots = []") {
		t.Fatal("starter config omitted reverse-index roots")
	}
}

func TestDemonRunDefaultsAndAtomicEditPreserveText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	original := "# keep this comment\nunknown = \"value\"\n\n[demon]\n# preserve me\nrun = true # trailing\n\n[watch]\ndebounce_seconds = 0.5\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDemonRun(path, false); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, want := range []string{"# keep this comment", "unknown = \"value\"", "# preserve me", "run = false # trailing", "debounce_seconds = 0.5"} {
		if !strings.Contains(text, want) {
			t.Fatalf("updated config missing %q: %s", want, text)
		}
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Demon.Run {
		t.Fatal("disabled demon was loaded as enabled")
	}
	if err := SetDemonRun(path, true); err != nil {
		t.Fatal(err)
	}
	loaded, err = Load(path)
	if err != nil || !loaded.Demon.Run {
		t.Fatalf("re-enable failed: %+v %v", loaded, err)
	}
}

func TestRepositoryFeatureSettingsLoadAndPreserveText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	original := "# keep this comment\ndocs_root = \"docs\"\n\n[index]\nenabled = true # index comment\n\n[links]\nenabled = true # links comment\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetIndexEnabled(path, false); err != nil {
		t.Fatal(err)
	}
	if err := SetLinksEnabled(path, false); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Index.Enabled || loaded.Links.Enabled {
		t.Fatalf("feature settings were not disabled: %+v", loaded)
	}
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# keep this comment", "enabled = false # index comment", "enabled = false # links comment"} {
		if !strings.Contains(string(text), want) {
			t.Fatalf("updated config missing %q: %s", want, text)
		}
	}
}

func TestFeatureSettingsAddMissingSections(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("docs_root = \"docs\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetIndexEnabled(path, false); err != nil {
		t.Fatal(err)
	}
	if err := SetLinksEnabled(path, false); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[index]\nenabled = false", "[links]\nenabled = false"} {
		if !strings.Contains(string(text), want) {
			t.Fatalf("missing feature section %q: %s", want, text)
		}
	}
}

func TestDemonRunAddsMissingSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("root = \"docs\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDemonRun(path, false); err != nil {
		t.Fatal(err)
	}
	text, _ := os.ReadFile(path)
	if !strings.Contains(string(text), "[demon]\nrun = false\n") {
		t.Fatalf("missing demon section: %q", text)
	}
}
func TestSelectionIsCurrentDirectoryOnly(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, ".doc-ledger.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if LocalPath(child) != "" {
		t.Fatal("local lookup searched parent")
	}
	if Discover(child) == "" {
		t.Fatal("legacy discovery should still search parent")
	}
}

func TestDiscoverWithinDoesNotCrossRepositoryBoundary(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "docs", "guide")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".demon-docs.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverWithin(child, filepath.Join(root, "docs")); got != "" {
		t.Fatalf("discovery crossed boundary: %s", got)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", ".demon-docs.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverWithin(child, filepath.Join(root, "docs")); got != filepath.Join(root, "docs", ".demon-docs.toml") {
		t.Fatalf("bounded discovery missed config: %s", got)
	}
}
func TestStarterConfigLoads(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"legacy.toml": StarterText(),
		"repo.toml":   RepositoryStarterText("manual"),
	} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		loaded, err := Load(p)
		if err != nil {
			t.Fatal(err)
		}
		if name == "repo.toml" && loaded.Root != "manual" {
			t.Fatalf("docs_root not loaded: %+v", loaded)
		}
	}
}

func TestFrontmatterAbsentRemainsDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("root = \"docs\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Frontmatter.Enabled {
		t.Fatalf("frontmatter unexpectedly enabled: %+v", loaded.Frontmatter)
	}
	if loaded.Frontmatter.DefaultFormat != "yaml" || !reflect.DeepEqual(loaded.Frontmatter.AllowedFormats, []string{"yaml", "toml"}) || loaded.Frontmatter.UnknownFields != "remove" {
		t.Fatalf("unexpected frontmatter defaults: %+v", loaded.Frontmatter)
	}
	if loaded.Frontmatter.Fields != nil || loaded.Frontmatter.Rules != nil {
		t.Fatalf("legacy config unexpectedly gained a schema: %+v", loaded.Frontmatter)
	}
}

func TestFormatConfigurationLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	text := `[format]
enabled = true
schema_dir = ".ddocs/policies"
document_schema_dir = ".ddocs/document-policies"
default_schema = "service"
invalidation_similarity = 0.75

[[format.path_rules]]
pattern = "docs/planning/**"
schema = "planning"
`
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Format.Enabled || loaded.Format.SchemaDir != ".ddocs/policies" || loaded.Format.DocumentSchemaDir != ".ddocs/document-policies" || loaded.Format.DefaultSchema != "service" || loaded.Format.InvalidationSimilarity != 0.75 {
		t.Fatalf("format settings not loaded: %+v", loaded.Format)
	}
	if len(loaded.Format.PathRules) != 1 || loaded.Format.PathRules[0] != (FormatPathRule{Pattern: "docs/planning/**", Schema: "planning"}) {
		t.Fatalf("format path rules not loaded: %+v", loaded.Format.PathRules)
	}
}

func TestFormatAbsentRemainsDisabledButKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("root = \"docs\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Format.Enabled || loaded.Format.SchemaDir != ".ddocs/schemas" || loaded.Format.DocumentSchemaDir != ".ddocs/document-schemas" || loaded.Format.DefaultSchema != "general" || loaded.Format.InvalidationSimilarity != 0.5 {
		t.Fatalf("unexpected legacy format defaults: %+v", loaded.Format)
	}
}

func TestFrontmatterConfigurationLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	text := `[frontmatter]
enabled = true
default_format = "toml"
allowed_formats = ["toml"]
default_author = "Demon Docs"
unknown_fields = "warn"

[frontmatter.fields.document_id]
type = "uuid"
required = true
immutable = true
generated = true

[frontmatter.fields.author]
type = "string"
required = true
default_from = "frontmatter.default_author"

[frontmatter.fields.policy_exempt]
type = "boolean"
default = false

[[frontmatter.rules]]
when_field = "policy_exempt"
equals = true
require = "policy_exempt_reason"
`
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Frontmatter.Enabled || loaded.Frontmatter.DefaultFormat != "toml" || !reflect.DeepEqual(loaded.Frontmatter.AllowedFormats, []string{"toml"}) || loaded.Frontmatter.DefaultAuthor != "Demon Docs" || loaded.Frontmatter.UnknownFields != "warn" {
		t.Fatalf("frontmatter settings not loaded: %+v", loaded.Frontmatter)
	}
	id := loaded.Frontmatter.Fields["document_id"]
	if id.Type != "uuid" || !id.Required || !id.Immutable || !id.Generated {
		t.Fatalf("document_id field not loaded: %+v", id)
	}
	author := loaded.Frontmatter.Fields["author"]
	if author.DefaultFrom != "frontmatter.default_author" || !author.Required {
		t.Fatalf("author field not loaded: %+v", author)
	}
	exempt := loaded.Frontmatter.Fields["policy_exempt"]
	if exempt.Type != "boolean" || exempt.Default != false {
		t.Fatalf("policy_exempt field not loaded: %+v", exempt)
	}
	if len(loaded.Frontmatter.Rules) != 1 || loaded.Frontmatter.Rules[0].WhenField != "policy_exempt" || loaded.Frontmatter.Rules[0].Equals != true || loaded.Frontmatter.Rules[0].Require != "policy_exempt_reason" {
		t.Fatalf("frontmatter rule not loaded: %+v", loaded.Frontmatter.Rules)
	}
}

func TestStarterConfigIncludesDefaultFrontmatterSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "starter.toml")
	if err := os.WriteFile(path, []byte(StarterText()), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Frontmatter.Enabled || loaded.Frontmatter.DefaultFormat != "yaml" || !reflect.DeepEqual(loaded.Frontmatter.AllowedFormats, []string{"yaml", "toml"}) || loaded.Frontmatter.UnknownFields != "remove" {
		t.Fatalf("starter frontmatter settings incorrect: %+v", loaded.Frontmatter)
	}
	expected := map[string]FrontmatterField{
		"document_id":          {Type: "uuid", Required: true, Immutable: true, Generated: true},
		"author":               {Type: "string", Required: true, DefaultFrom: "frontmatter.default_author"},
		"document_type":        {Type: "string", Required: true, Default: "general"},
		"created":              {Type: "date", Required: true, Immutable: true, Generated: true},
		"summary":              {Type: "string", Required: true},
		"policy_exempt":        {Type: "boolean", Default: false},
		"policy_exempt_reason": {Type: "string"},
	}
	if !reflect.DeepEqual(loaded.Frontmatter.Fields, expected) {
		t.Fatalf("starter frontmatter fields incorrect: %#v", loaded.Frontmatter.Fields)
	}
	if len(loaded.Frontmatter.Rules) != 1 || loaded.Frontmatter.Rules[0] != (FrontmatterRule{WhenField: "policy_exempt", Equals: true, Require: "policy_exempt_reason"}) {
		t.Fatalf("starter frontmatter rules incorrect: %#v", loaded.Frontmatter.Rules)
	}
}
