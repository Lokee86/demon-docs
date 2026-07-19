package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func moveHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs mv [-h] [--root PATH] [--dry-run] SOURCE DESTINATION\n\nMove one file or directory and rewrite affected local links without requiring a Demon Docs repository.\n\noptions:\n  -h, --help   show this help message and exit\n  --root PATH  repository boundary to scan; defaults to the detected initialized repository root or the current directory\n  --dry-run    print the planned move and rewrites without changing files\n\nSOURCE and DESTINATION resolve from the current directory and must remain inside the repository boundary. If DESTINATION is an existing directory, SOURCE is moved beneath it. Destination parents must already exist.\n\nThe command preserves recognized Markdown, image, reference-definition, wiki, and local HTML link syntax. It changes only affected destination paths and does not create .ddocs/ state.")
}

func runMove(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		moveHelp(out)
		return 0
	}
	fs := flag.NewFlagSet("ddocs mv", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var root optionalString
	var dryRun bool
	fs.Var(&root, "root", "repository boundary to scan")
	fs.BoolVar(&dryRun, "dry-run", false, "print the planned move without writing")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs mv: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 2 {
		fmt.Fprintln(errOut, "usage: ddocs mv [-h] [--root PATH] [--dry-run] SOURCE DESTINATION")
		fmt.Fprintln(errOut, "ddocs mv: error: SOURCE and DESTINATION are required")
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	repositoryRoot := cwd
	if location, ok := repository.Discover(cwd); ok {
		repositoryRoot = location.Root
	}
	if root.set {
		repositoryRoot = root.value
		if !filepath.IsAbs(repositoryRoot) {
			repositoryRoot = filepath.Join(cwd, repositoryRoot)
		}
	}
	repositoryRoot, err = filepath.Abs(repositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	source := fs.Arg(0)
	destination := fs.Arg(1)
	if !filepath.IsAbs(source) {
		source = filepath.Join(cwd, source)
	}
	if !filepath.IsAbs(destination) {
		destination = filepath.Join(cwd, destination)
	}
	plan, err := links.PlanMove(repositoryRoot, source, destination)
	if err != nil {
		return fail(errOut, err)
	}
	verb := "move"
	updateVerb := "update"
	rewriteVerb := "rewrite"
	if !dryRun {
		if err := links.ApplyMove(plan); err != nil {
			return fail(errOut, err)
		}
		verb = "moved"
		updateVerb = "updated"
		rewriteVerb = "rewrote"
	}
	fmt.Fprintf(out, "%s: %s -> %s\n", verb, displayMovePath(plan.RepositoryRoot, plan.Source), displayMovePath(plan.RepositoryRoot, plan.Destination))
	fmt.Fprintf(out, "%s %d Markdown file(s)\n", updateVerb, len(plan.Updates))
	fmt.Fprintf(out, "%s %d link(s)\n", rewriteVerb, plan.RewrittenLinks)
	for _, update := range plan.Updates {
		fmt.Fprintf(out, "  %s (%d link(s))\n", displayMovePath(plan.RepositoryRoot, update.Path), update.Links)
	}
	return 0
}

func displayMovePath(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || len(relative) > 3 && relative[:3] == ".."+string(filepath.Separator) {
		return filepath.Clean(path)
	}
	return filepath.ToSlash(relative)
}
