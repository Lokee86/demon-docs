package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestCommandHelpContract(t *testing.T) {
	tests := []struct {
		args []string
		want []string
	}{
		{[]string{"--help"}, []string{"ddocs reconciles local index files with a file tree.", "ddocs config paths", "ddocs --version"}},
		{[]string{"fix", "-h"}, []string{"Reconcile the docs tree and write any needed updates.", "--root PATH", "--config PATH", "--index-file NAME", "--draft-description-prefix TEXT", "--include PATTERN", "--exclude PATTERN", "--marker-prefix TEXT", "--parent-label TEXT", "--no-parent-link-folder-indexes", "1. --config PATH", "./.demon-docs.toml", "./.doc-ledger.toml", "there is no upward parent-directory search"}},
		{[]string{"check", "--help"}, []string{"Verify that the docs tree is already reconciled.", "--root PATH", "--no-parent-link-indexed-files", "CLI flags override the selected config"}},
		{[]string{"watch", "-h"}, []string{"Watch runs in the foreground by default", "--once", "--debounce-seconds FLOAT", "run one reconciliation pass and exit"}},
		{[]string{"config", "-h"}, []string{"paths", "show", "init", "Local config lookup is current-directory only.", "There is no upward parent-directory search."}},
		{[]string{"config", "paths", "-h"}, []string{"current-directory local config", ".demon-docs.toml", "demon-docs.toml", ".doc-ledger.toml", "doc-ledger.toml", "global user config path", "selected config path"}},
		{[]string{"config", "show", "--help"}, []string{"resolved selected config", "--config PATH", "--no-local-config", "--no-global-config"}},
		{[]string{"config", "init", "-h"}, []string{"global user config location", ".demon-docs.toml", "--local", "--global", "--force"}},
	}
	for _, test := range tests {
		t.Run(strings.Join(test.args, "_"), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(context.Background(), test.args, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%q", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("help wrote stderr: %q", stderr.String())
			}
			for _, text := range test.want {
				if !strings.Contains(stdout.String(), text) {
					t.Errorf("help missing %q:\n%s", text, stdout.String())
				}
			}
		})
	}
}

func TestOptionalStringOverridesDistinguishEmptyFromAbsent(t *testing.T) {
	c := config.Default()
	applyOverrides(&c, commonFlags{
		index:  optionalString{set: true},
		draft:  optionalString{set: true},
		prefix: optionalString{set: true},
		marker: optionalString{set: true},
		parent: optionalString{set: true},
	})
	if c.IndexFile != "" || c.Files.IndexFile != "" || c.Draft.Folder != "" || c.Draft.DescriptionPrefix != "" || c.Markers.Prefix != "" || c.ParentLink.Label != "" {
		t.Fatalf("empty overrides were ignored: %+v", c)
	}
}
