package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/ddrepo"
	"github.com/Lokee86/demon-docs/internal/demon"
)

func TestIndexesAndLinksCanRunSeparately(t *testing.T) {
	t.Run("links only", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "page.md"), "[asset](asset.bin)\n")
		writeTestFile(t, filepath.Join(root, "asset.bin"), "asset")
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--root", root, "-l"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if _, err := os.Stat(filepath.Join(root, "INDEX.md")); !os.IsNotExist(err) {
			t.Fatalf("links-only run created an index: %v", err)
		}
		assertDDocsState(t, root)
	})

	t.Run("indexes only", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "page.md"), "# Page\n")
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--root", root, "--indexes", "--no-local-config", "--no-global-config"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if _, err := os.Stat(filepath.Join(root, "INDEX.md")); err != nil {
			t.Fatalf("indexes-only run did not create an index: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, ".ddocs")); !os.IsNotExist(err) {
			t.Fatalf("indexes-only run initialized link state: %v", err)
		}
	})
}

func TestFrontmatterOnlyCleanFixDoesNotRefreshLinkState(t *testing.T) {
	root := t.TempDir()
	configText := strings.Replace(frontmatterTestConfig(false, "yaml"), "[links]\nenabled = false", "[links]\nenabled = true", 1)
	writeTestFile(t, filepath.Join(root, ".ddocs", "config.toml"), configText)
	source := filepath.Join(root, "docs", "source.md")
	writeTestFile(t, source, "---\nauthor: Test Author\ncreated: 2026-07-20\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: general\nsummary: Existing\n---\n# Source\n\n[Target](target.md)\n")
	writeTestFile(t, filepath.Join(root, "docs", "target.md"), "# Target\n")
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"fix", "--root", root, "--links", "--no-local-config", "--no-global-config"}, &stdout, &stderr); code != 0 {
		t.Fatalf("baseline code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	before := snapshotLinkState(t, root)
	writeTestFile(t, source, "---\nauthor: Test Author\ncreated: 2026-07-20\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: general\nsummary: Existing\n---\n# Source\n\n[Target](other.md)\n")
	stdout.Reset()
	stderr.Reset()
	if code := Run(context.Background(), []string{"fix", "--root", root, "--frontmatter", "--no-local-config", "--no-global-config"}, &stdout, &stderr); code != 0 {
		t.Fatalf("clean frontmatter code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	after := snapshotLinkState(t, root)
	if len(before) != len(after) {
		t.Fatalf("clean frontmatter fix changed link record count: before=%d after=%d", len(before), len(after))
	}
	for name, data := range before {
		if string(after[name]) != string(data) {
			t.Fatalf("clean frontmatter fix changed link record %s", name)
		}
	}
}

func snapshotLinkState(t *testing.T, root string) map[string][]byte {
	t.Helper()
	repository, err := ddrepo.Open(filepath.Join(root, ".ddocs"))
	if err != nil {
		t.Fatal(err)
	}
	tx, err := repository.Begin()
	if err != nil {
		t.Fatal(err)
	}
	result := map[string][]byte{}
	for _, prefix := range []string{"meta/", "file/", "path/", "source/", "incoming/", "write/"} {
		names, err := tx.Names(prefix)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			data, err := tx.Read(name)
			if err != nil {
				t.Fatal(err)
			}
			result[name] = data
		}
	}
	return result
}

func TestIndexSelectorDoesNotSelectFrontmatterOrFormat(t *testing.T) {
	configured := config.Default()
	configured.Index.Enabled = true
	configured.Frontmatter.Enabled = true
	configured.Format.Enabled = true
	configured.Links.Enabled = true

	features := selectedFeatures("fix", commonFlags{indexesOnly: true}, configured)
	if !features.Indexes || features.Frontmatter || features.Format || features.Links || features.TrackLinks || features.Reverse {
		t.Fatalf("explicit index selector was not index-only: %+v", features)
	}

	features = selectedFeatures("fix", commonFlags{docsOnly: true}, configured)
	if !features.Indexes || !features.Frontmatter || !features.Format || features.Links || features.TrackLinks || features.Reverse {
		t.Fatalf("explicit docs selector did not select documentation policy systems: %+v", features)
	}
}

func TestDefaultFixSkipsFrontmatterAndFormatUnlessAllIsSelected(t *testing.T) {
	configured := config.Default()
	configured.Index.Enabled = true
	configured.Frontmatter.Enabled = true
	configured.Format.Enabled = true
	configured.Links.Enabled = true
	configured.ReverseIndex.Roots = []string{"services"}

	features := selectedFeatures("fix", commonFlags{}, configured)
	if !features.Indexes || features.Frontmatter || features.Format || !features.Links || !features.TrackLinks || !features.Reverse {
		t.Fatalf("bare fix did not select only its safe general systems: %+v", features)
	}

	features = selectedFeatures("fix", commonFlags{all: true}, configured)
	if !features.Indexes || !features.Frontmatter || !features.Format || !features.Links || !features.TrackLinks || !features.Reverse {
		t.Fatalf("all selector did not select every configured system: %+v", features)
	}
}

func TestDefaultCheckAndWatchStillSelectFrontmatterAndFormat(t *testing.T) {
	configured := config.Default()
	configured.Frontmatter.Enabled = true
	configured.Format.Enabled = true
	for _, command := range []string{"check", "watch"} {
		features := selectedFeatures(command, commonFlags{}, configured)
		if !features.Frontmatter || !features.Format {
			t.Fatalf("bare %s unexpectedly skipped frontmatter or format: %+v", command, features)
		}
	}
}

func TestReverseSpecificOptionsEnableReverseInDefaultSelection(t *testing.T) {
	features := selectedFeatures("fix", commonFlags{reverseRoots: stringsFlag{values: []string{"services"}}}, config.Default())
	if !features.Indexes || !features.Links || !features.Reverse {
		t.Fatalf("reverse root override did not enable reverse alongside default systems: %+v", features)
	}

	features = selectedFeatures("fix", commonFlags{reverseOnly: true}, config.Default())
	if features.Indexes || features.Links || features.TrackLinks || !features.Reverse {
		t.Fatalf("explicit reverse selector was not reverse-only: %+v", features)
	}

	disabled := config.Default()
	disabled.Index.Enabled = false
	disabled.Links.Enabled = false
	features = selectedFeatures("fix", commonFlags{}, disabled)
	if features.Indexes || features.Links || !features.TrackLinks {
		t.Fatalf("disabled features did not preserve internal link tracking: %+v", features)
	}
}

func TestLinksOnlyDoesNotRequireDocsRoot(t *testing.T) {
	repositoryRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repositoryRoot, ".ddocs", "config.toml"), "docs_root = \"missing-docs\"\n")
	writeTestFile(t, filepath.Join(repositoryRoot, "README.md"), "# Repository\n")
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	assertDDocsState(t, repositoryRoot)
}

func TestWatchOnceHonorsLinksOnly(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "page.md"), "[asset](asset.bin)\n")
	writeTestFile(t, filepath.Join(root, "asset.bin"), "asset")
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"watch", "--root", root, "-l", "--once"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "INDEX.md")); !os.IsNotExist(err) {
		t.Fatalf("links-only watch created an index: %v", err)
	}
	assertDDocsState(t, root)
}

func TestFeatureToggleCommandsPersistRepositoryConfig(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".ddocs", "config.toml")
	writeTestFile(t, configPath, "# preserve\ndocs_root = \"docs\"\n\n[index]\nenabled = true\n\n[links]\nenabled = true\n")
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, root, func(string) {
		for _, command := range [][]string{{"index", "disable"}, {"links", "--false"}} {
			var stdout, stderr bytes.Buffer
			if code := Run(context.Background(), command, &stdout, &stderr); code != 0 {
				t.Fatalf("command=%v code=%d stdout=%q stderr=%q", command, code, stdout.String(), stderr.String())
			}
		}
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"index", "status"}, &stdout, &stderr); code != 0 || stdout.String() != "index: disabled\n" {
			t.Fatalf("status code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Index.Enabled || loaded.Links.Enabled {
		t.Fatalf("feature toggles were not persisted: %+v", loaded)
	}
	text, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "# preserve") {
		t.Fatalf("toggle rewrite discarded config text: %s", text)
	}
}

func TestFeatureToggleRequestsRunningDemonReload(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".ddocs", "config.toml")
	writeTestFile(t, configPath, config.RepositoryStarterText("docs"))
	runtime := demon.New(root)
	owner, claimed, err := runtime.Claim(os.Getpid())
	if err != nil || !claimed {
		t.Fatalf("claim owner: claimed=%t err=%v", claimed, err)
	}
	defer runtime.Release(owner)
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"index", "disable"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	if !runtime.ShutdownRequested() {
		t.Fatal("feature toggle did not request a running demon reload")
	}
}

func TestDisabledSelectedIndexWatchDoesNotFallBackToOtherSystems(t *testing.T) {
	root := t.TempDir()
	docsRoot := filepath.Join(root, "docs")
	configPath := filepath.Join(root, ".ddocs", "config.toml")
	writeTestFile(t, configPath, starterWithoutFrontmatter("docs"))
	if err := config.SetIndexEnabled(configPath, false); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docsRoot, "page.md"), "# Page\n")
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"watch", "--docs", "--once"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	if _, err := os.Stat(filepath.Join(docsRoot, "INDEX.md")); !os.IsNotExist(err) {
		t.Fatalf("disabled index watch created an index: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".ddocs", "refs", "ddocs", "state")); !os.IsNotExist(err) {
		t.Fatalf("disabled index-only watch unexpectedly tracked links: %v", err)
	}
}

func TestDisabledIndexLeavesIndexOrdinaryAndLinkManaged(t *testing.T) {
	root := t.TempDir()
	docsRoot := filepath.Join(root, "docs")
	configPath := filepath.Join(root, ".ddocs", "config.toml")
	writeTestFile(t, configPath, starterWithoutFrontmatter("docs"))
	if err := config.SetIndexEnabled(configPath, false); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docsRoot, "INDEX.md"), "# Existing index\n\n[Page](page.md)\n")
	writeTestFile(t, filepath.Join(docsRoot, "page.md"), "# Page\n")
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix"}, &stdout, &stderr); code != 0 {
			t.Fatalf("baseline code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if err := os.Rename(filepath.Join(docsRoot, "page.md"), filepath.Join(docsRoot, "moved.md")); err != nil {
			t.Fatal(err)
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix"}, &stdout, &stderr); code != 0 {
			t.Fatalf("repair code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	text, err := os.ReadFile(filepath.Join(docsRoot, "INDEX.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "[Page](moved.md)") {
		t.Fatalf("existing index was not treated as a normal link source: %s", text)
	}
	if strings.Contains(string(text), "doc-ledger") {
		t.Fatalf("disabled indexing still inserted managed index content: %s", text)
	}
}

func TestDisabledLinksKeepStateWithoutRewriting(t *testing.T) {
	root := t.TempDir()
	docsRoot := filepath.Join(root, "docs")
	configPath := filepath.Join(root, ".ddocs", "config.toml")
	writeTestFile(t, configPath, starterWithoutFrontmatter("docs"))
	if err := config.SetIndexEnabled(configPath, false); err != nil {
		t.Fatal(err)
	}
	if err := config.SetLinksEnabled(configPath, false); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(docsRoot, "INDEX.md")
	writeTestFile(t, readme, "[Page](page.md)\n")
	writeTestFile(t, filepath.Join(docsRoot, "page.md"), "# Page\n")
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix"}, &stdout, &stderr); code != 0 {
			t.Fatalf("baseline code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if err := os.Rename(filepath.Join(docsRoot, "page.md"), filepath.Join(docsRoot, "moved.md")); err != nil {
			t.Fatal(err)
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix"}, &stdout, &stderr); code != 0 {
			t.Fatalf("tracking code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		unchanged, err := os.ReadFile(readme)
		if err != nil {
			t.Fatal(err)
		}
		if string(unchanged) != "[Page](page.md)\n" {
			t.Fatalf("disabled links rewrote the document: %s", unchanged)
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"links", "enable"}, &stdout, &stderr); code != 0 {
			t.Fatalf("enable code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix"}, &stdout, &stderr); code != 0 {
			t.Fatalf("repair code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	updated, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) != "[Page](moved.md)\n" {
		t.Fatalf("re-enabled links did not use persistent tracking state: %s", updated)
	}
}

func starterWithoutFrontmatter(root string) string {
	text := strings.Replace(config.RepositoryStarterText(root), "[frontmatter]\nenabled = true", "[frontmatter]\nenabled = false", 1)
	return strings.Replace(text, "[format]\nenabled = true", "[format]\nenabled = false", 1)
}

func assertDDocsState(t *testing.T, root string) {
	t.Helper()
	for _, path := range []string{
		filepath.Join(root, ".ddocs", "objects"),
		filepath.Join(root, ".ddocs", "refs", "ddocs", "state"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("ddocs repository state is missing at %s: %v", path, err)
		}
	}
}
