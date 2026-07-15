package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionAndUnknownCommandExitCodes(t *testing.T) {
	var out, err bytes.Buffer
	if code := Run(context.Background(), []string{"--version"}, &out, &err); code != 0 || out.String() != "doc-ledger 0.1.1\n" {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), err.String())
	}
	out.Reset()
	err.Reset()
	if code := Run(context.Background(), []string{"nope"}, &out, &err); code != 2 || !strings.Contains(err.String(), "invalid choice") {
		t.Fatalf("code=%d err=%q", code, err.String())
	}
}

func TestHelpUsesStdoutAndSuccess(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"fix", "--help"}, {"config", "paths", "--help"}, {"config", "show", "--help"}, {"config", "init", "--help"}} {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), args, &out, &errOut); code != 0 || out.Len() == 0 || errOut.Len() != 0 {
			t.Fatalf("args=%v code=%d out=%q err=%q", args, code, out.String(), errOut.String())
		}
	}
}

func TestMissingCommandAndUnexpectedArgumentsFail(t *testing.T) {
	for _, args := range [][]string{nil, {"fix", "extra"}, {"config", "paths", "extra"}, {"config", "show", "extra"}, {"config", "init", "--local", "extra"}, {"config", "init", "--local", "--global"}} {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), args, &out, &errOut); code != 2 || errOut.Len() == 0 {
			t.Fatalf("args=%v code=%d out=%q err=%q", args, code, out.String(), errOut.String())
		}
	}
}

func TestFixCheckAndOverrides(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "page.md"), []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	args := []string{"fix", "--root", root, "--index-file", "!README.md", "--parent-link-indexed-files"}
	if code := Run(context.Background(), args, &out, &errOut); code != 0 {
		t.Fatalf("code=%d err=%s", code, errOut.String())
	}
	if _, err := os.Stat(filepath.Join(root, "!README.md")); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile(filepath.Join(root, "page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "./!README.md") {
		t.Fatal(string(page))
	}
	out.Reset()
	errOut.Reset()
	if code := Run(context.Background(), []string{"check", "--root", root, "--index-file", "!README.md", "--parent-link-indexed-files"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "check passed") {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
}
func TestCheckReportsDriftWithoutWriting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "page.md")
	if err := os.WriteFile(path, []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"check", "--root", root}, &out, &errOut); code != 1 {
		t.Fatalf("code=%d out=%s err=%s", code, out.String(), errOut.String())
	}
	if _, err := os.Stat(filepath.Join(root, "README.md")); !os.IsNotExist(err) {
		t.Fatal("check wrote index")
	}
}
func TestConfigInitAndShow(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"config", "init", "--local"}, &out, &errOut); code != 0 {
		t.Fatalf("%d %s", code, errOut.String())
	}
	out.Reset()
	if code := Run(context.Background(), []string{"config", "show"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "index_file = 'README.md'") {
		t.Fatalf("code=%d out=%s err=%s", code, out.String(), errOut.String())
	}
}
