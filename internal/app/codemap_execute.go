package app

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/Lokee86/demon-docs/internal/codemaprun"
	"github.com/Lokee86/demon-docs/internal/documentpolicy"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func codemapExecutionHelp(w io.Writer, command string) {
	rootRule := "optional; may name one Markdown file or a directory; defaults to the configured docs root"
	if command != "fix" {
		rootRule = "required; may name one Markdown file or a directory"
	}
	dryRun := ""
	if command == "fix" {
		dryRun = "  --dry-run          report planned updates without writing\n"
	}
	rootUsage := "--root PATH"
	if command == "fix" {
		rootUsage = "[--root PATH]"
	}
	fmt.Fprintf(w, "usage: ddocs codemaps %s [-h] %s [--config PATH] [--heading TEXT]", command, rootUsage)
	if command == "fix" {
		fmt.Fprint(w, " [--dry-run]")
	}
	fmt.Fprintf(w, "\n\n%s codemap sections using deterministic repository evidence. Codemap operations run only through this explicit foreground command; the watcher and daemon never execute them.\n\noptions:\n  -h, --help         show this help message and exit\n  --root PATH        %s\n  --config PATH      explicit ddocs config file\n  --no-local-config  skip current-directory local config\n  --no-global-config skip the global user config\n  --heading TEXT     replace configured codemap headings; repeat as needed\n%s", codemapExecutionVerb(command), rootRule, dryRun)
}

func codemapExecutionVerb(command string) string {
	switch command {
	case "fix":
		return "Adopt, generate, and update"
	case "check":
		return "Check"
	default:
		return "Inspect"
	}
}

func runCodemapExecution(ctx context.Context, command string, args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		codemapExecutionHelp(out, command)
		return 0
	}
	fs := flag.NewFlagSet("ddocs codemaps "+command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var flags commonFlags
	var root optionalString
	var headings stringsFlag
	dryRun := false
	fs.Var(&root, "root", "Markdown file or directory")
	fs.Var(&flags.config, "config", "explicit ddocs config file")
	fs.BoolVar(&flags.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&flags.noGlobal, "no-global-config", false, "skip the global user config")
	fs.Var(&headings, "heading", "accepted codemap heading")
	if command == "fix" {
		fs.BoolVar(&dryRun, "dry-run", false, "report planned updates without writing")
	}
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs codemaps %s: error: %v\n", command, err)
		return 2
	}
	if fs.NArg() != 0 {
		writeUnrecognized(errOut, fs.Args())
		return 2
	}
	if command != "fix" && !root.set {
		fmt.Fprintf(errOut, "usage: ddocs codemaps %s [-h] --root PATH\n", command)
		fmt.Fprintf(errOut, "ddocs codemaps %s: error: the following arguments are required: --root\n", command)
		return 2
	}

	resolved, configPath, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	scope, err := resolveScope(optionalString{}, resolved.Root, configPath)
	if err != nil {
		return fail(errOut, err)
	}
	if !repository.DocsRootExists(scope) {
		fmt.Fprintf(errOut, "ddocs error: docs root does not exist: %s\n", scope.DocsRoot)
		return 2
	}
	targetRoot := scope.DocsRoot
	if root.set {
		targetRoot, err = resolveCodemapRoot(scope, root.value)
		if err != nil {
			return fail(errOut, err)
		}
	}
	files, err := codemapTargetFiles(scope.RepositoryRoot, targetRoot)
	if err != nil {
		return fail(errOut, err)
	}
	acceptedHeadings := append([]string(nil), resolved.Codemap.Headings...)
	if len(headings.values) > 0 {
		acceptedHeadings = append([]string(nil), headings.values...)
	}
	plan, err := codemaprun.Build(ctx, codemaprun.Options{
		RepositoryRoot:          scope.RepositoryRoot,
		DocsRoot:                scope.DocsRoot,
		TargetFiles:             files,
		Headings:                acceptedHeadings,
		MarkerPrefix:            resolved.Markers.Prefix,
		RemoveUndiscoveredLinks: resolved.Codemap.RemoveUndiscoveredLinks,
		RemoveLowScoreLinks:     resolved.Codemap.RemoveLowScoreLinks,
		Schema: documentpolicy.CodemapSchemaProvider{
			RepositoryRoot: scope.RepositoryRoot,
			Config:         resolved,
			Headings:       acceptedHeadings,
		},
	})
	if err != nil {
		return fail(errOut, err)
	}

	switch command {
	case "inspect":
		writeCodemapInspection(out, plan)
		return 0
	case "check":
		if plan.ChangedCount() == 0 {
			fmt.Fprintln(out, "ddocs codemaps check passed")
			return 0
		}
		fmt.Fprintln(out, "ddocs codemaps check failed")
		for _, document := range plan.Documents {
			if document.Changed {
				fmt.Fprintln(out, document.Path)
			}
		}
		return 1
	case "fix":
		if dryRun {
			fmt.Fprintf(out, "ddocs codemaps fix would update %d file(s)\n", plan.ChangedCount())
			writeCodemapSummary(out, plan)
			return 0
		}
		if err := codemaprun.Apply(plan); err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "ddocs codemaps fix updated %d file(s)\n", plan.ChangedCount())
		writeCodemapSummary(out, plan)
		return 0
	default:
		return fail(errOut, fmt.Errorf("unsupported codemap command %s", command))
	}
}
