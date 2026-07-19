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
		{[]string{"--help"}, []string{"ddocs reconciles folder indexes and repository-local links in Markdown documents with the filesystem.", "ddocs init --root docs", "ddocs demon --help", "ddocs config paths", "ddocs --version"}},
		{[]string{"init", "--help"}, []string{"Initialize a Demon Docs repository", "--root PATH", ".ddocs/config.toml", "must already exist", "[demon].run = true"}},
		{[]string{"status", "--help"}, []string{"Show the Demon Docs repository", "usage: ddocs status"}},
		{[]string{"fix", "-h"}, []string{"Reconcile selected indexes and links and write needed updates.", "-i, --indexes", "-l, --links", "--root PATH", "--config PATH", "--index-file NAME", "--draft-description-prefix TEXT", "--include PATTERN", "--exclude PATTERN", "--marker-prefix TEXT", "--parent-label TEXT", "--no-parent-link-folder-indexes", "wiki links such as [[guide]]", "local HTML href, src, and poster targets", "1. --config PATH", ".ddocs/config.toml", "./.demon-docs.toml", "./.doc-ledger.toml", "repository config is discovered by searching upward"}},
		{[]string{"check", "--help"}, []string{"Verify that selected indexes and links are already reconciled.", "-i, --indexes", "-l, --links", "--root PATH", "--no-parent-link-indexed-files", "undefined explicit or collapsed reference labels", "[Guide][guide]", "CLI flags override the selected config"}},
		{[]string{"watch", "-h"}, []string{"Watch runs in the foreground by default", "Each reconciliation diagnostic is printed as an individual message.", "--once", "--debounce-seconds FLOAT", "run one reconciliation pass and exit", "use ddocs demon for detached"}},
		{[]string{"demon", "--help"}, []string{"One fresh owner serves each local .ddocs repository", "run [--true|--false] [PATH]", "read-only ownership and feeder status", "ddocs demon __shell-hook bash", "linked Git worktree"}},
		{[]string{"demon", "run", "--help"}, []string{"register the current shell as a feeder", "--true", "clear a shutdown request", "--false", "remove all feeders", "linked worktree"}},
		{[]string{"demon", "--status", "--help"}, []string{"without creating runtime state", "running/stale/stopped", "active shell and agent counts", "watched docs root"}},
		{[]string{"demon", "--logs", "--help"}, []string{"oldest to newest", ".ddocs/runtime/logs", "five files", "1 MiB"}},
		{[]string{"demon", "__shell-hook", "--help"}, []string{"registers a shell feeder", "removes only that feeder", "eval \"$(ddocs demon __shell-hook bash)\"", "Invoke-Expression"}},
		{[]string{"config", "-h"}, []string{"paths", "show", "init", ".ddocs/config.toml", "Legacy local config lookup remains current-directory only."}},
		{[]string{"config", "paths", "-h"}, []string{"repository", ".ddocs/config.toml", ".demon-docs.toml", "demon-docs.toml", ".doc-ledger.toml", "doc-ledger.toml", "Global config candidates"}},
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
