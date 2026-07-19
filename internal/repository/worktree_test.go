package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapLinkedWorktreeCopiesOnlyConfigAndFreshObjects(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	linked := filepath.Join(base, "linked")
	if err := os.MkdirAll(filepath.Join(primary, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Initialize(primary, "docs_root = \"docs\"\n\n[demon]\nrun = true\n"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(primary, ".git", "worktrees", "linked"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linked, ".git"), []byte("gitdir: "+filepath.Join(primary, ".git", "worktrees", "linked")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(primary, ".git", "worktrees", "linked", "commondir"), []byte("../..\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(linked, "docs", "guide")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	location, detected, err := BootstrapLinkedWorktree(nested)
	if err != nil || !detected {
		t.Fatalf("location=%+v detected=%t err=%v", location, detected, err)
	}
	if location.Root != linked || !fileExists(location.ConfigPath) {
		t.Fatalf("unexpected location: %+v", location)
	}
	primaryConfig, _ := os.ReadFile(filepath.Join(primary, ".ddocs", "config.toml"))
	linkedConfig, _ := os.ReadFile(location.ConfigPath)
	if string(primaryConfig) != string(linkedConfig) {
		t.Fatal("linked config differs from primary")
	}
	if fileExists(filepath.Join(linked, ".ddocs", "runtime", "owner.json")) {
		t.Fatal("runtime state was copied")
	}
	if _, err := os.Stat(filepath.Join(linked, ".ddocs", "objects")); err != nil {
		t.Fatalf("linked object repository missing: %v", err)
	}
}

func TestDetectLinkedWorktreeIsReadOnlyBeforeBootstrap(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	linked := filepath.Join(base, "linked")
	if err := os.MkdirAll(filepath.Join(primary, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Initialize(primary, "docs_root = \"docs\"\n\n[demon]\nrun = true\n"); err != nil {
		t.Fatal(err)
	}
	gitWorktree := filepath.Join(primary, ".git", "worktrees", "linked")
	if err := os.MkdirAll(gitWorktree, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(linked, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linked, ".git"), []byte("gitdir: "+gitWorktree+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitWorktree, "commondir"), []byte("../..\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	location, detected, err := DetectLinkedWorktree(filepath.Join(linked, "nested"))
	if err != nil || !detected {
		t.Fatalf("location=%+v detected=%t err=%v", location, detected, err)
	}
	if location.Root != linked || location.ConfigPath != filepath.Join(primary, DirectoryName, ConfigName) {
		t.Fatalf("unexpected read-only location: %+v", location)
	}
	if _, err := os.Stat(filepath.Join(linked, DirectoryName)); !os.IsNotExist(err) {
		t.Fatalf("read-only detection created linked marker: %v", err)
	}
}
