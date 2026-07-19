package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/reverseindex"
)

func reverseIndexHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs reverse-index [-h] {check,fix,watch} ...\n\nGenerate code-folder indexes and documentation backlinks from authored codemaps.\n\npositional arguments:\n  {check,fix,watch}\n    check               report reverse indexes that need reconciliation\n    fix                 write reconciled reverse indexes\n    watch               reconcile immediately and watch for changes\n\noptions:\n  -h, --help            show this help message and exit")
}

func reverseIndexCommandHelp(w io.Writer, command string) {
	watchUsage := ""
	watchOptions := ""
	if command == "watch" {
		watchUsage = " [--once] [--debounce-seconds FLOAT]"
		watchOptions = "  --once                run one reconciliation pass and exit\n  --debounce-seconds FLOAT\n                        override the watcher debounce interval\n"
	}
	fmt.Fprintf(w, "usage: ddocs reverse-index %s [-h] [--root PATH] [--config PATH]\n                                  [--no-local-config] [--no-global-config]\n                                  [--index-file NAME] [--heading TEXT]\n                                  [--target-base BASE] [--target-root PATH]%s\n                                  [PATH ...]\n\nRecursively reconcile configured reverse-index roots, or the positional directory paths supplied to this command. Relative positional paths are resolved from the current working directory.\n\noptions:\n  -h, --help          show this help message and exit\n  --root PATH         override the configured docs root\n  --config PATH       explicit ddocs config file\n  --no-local-config   skip current-directory local config\n  --no-global-config  skip the global user config\n  --index-file NAME   override the code-folder index filename\n  --heading TEXT      accepted codemap heading; repeat to replace defaults\n  --target-base BASE  resolve targets from repository or document\n  --target-root PATH  repository-relative component root; repeat as needed\n%s\nPaths:\n  PATH                directory to reverse index recursively; positional paths replace [reverse_index].roots\n", command, watchUsage, watchOptions)
}

func runReverseIndex(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, "usage: ddocs reverse-index [-h] {check,fix,watch} ...")
		fmt.Fprintln(errOut, "ddocs reverse-index: error: the following arguments are required: reverse_index_command")
		return 2
	}
	if args[0] == "-h" || args[0] == "--help" {
		reverseIndexHelp(out)
		return 0
	}
	command := args[0]
	if command != "check" && command != "fix" && command != "watch" {
		fmt.Fprintln(errOut, "usage: ddocs reverse-index [-h] {check,fix,watch} ...")
		fmt.Fprintf(errOut, "ddocs reverse-index: error: argument reverse_index_command: invalid choice: '%s' (choose from check, fix, watch)\n", command)
		return 2
	}
	if helpRequested(args[1:]) {
		reverseIndexCommandHelp(out, command)
		return 0
	}

	fs := flag.NewFlagSet("ddocs reverse-index "+command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var flags commonFlags
	var headings stringsFlag
	var targetRoots stringsFlag
	targetBase := string(codemap.TargetBaseRepository)
	once := false
	debounce := -1.0
	fs.Var(&flags.root, "root", "override the configured docs root")
	fs.Var(&flags.config, "config", "explicit ddocs config file")
	fs.BoolVar(&flags.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&flags.noGlobal, "no-global-config", false, "skip the global user config")
	fs.Var(&flags.index, "index-file", "override the code-folder index filename")
	fs.Var(&headings, "heading", "accepted codemap heading")
	fs.StringVar(&targetBase, "target-base", targetBase, "repository or document")
	fs.Var(&targetRoots, "target-root", "repository-relative component root")
	if command == "watch" {
		fs.BoolVar(&once, "once", false, "run one reconciliation pass and exit")
		fs.Float64Var(&debounce, "debounce-seconds", -1, "override watcher debounce")
	}
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(errOut, "ddocs reverse-index %s: error: %v\n", command, err)
		return 2
	}
	if targetBase != string(codemap.TargetBaseRepository) && targetBase != string(codemap.TargetBaseDocument) {
		fmt.Fprintf(errOut, "ddocs reverse-index %s: error: invalid --target-base %q; expected repository or document\n", command, targetBase)
		return 2
	}
	resolved, configPath, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	applyOverrides(&resolved, flags)
	scope, err := resolveScope(flags.root, resolved.Root, configPath)
	if err != nil {
		return fail(errOut, err)
	}
	if !repository.DocsRootExists(scope) {
		fmt.Fprintf(errOut, "ddocs error: docs root does not exist: %s\n", scope.DocsRoot)
		return 2
	}
	format := codemap.DefaultFormat()
	if len(headings.values) > 0 {
		format.SectionHeadings = headings.values
	}
	format.TargetBase = codemap.TargetBase(targetBase)
	format.TargetRoots = targetRoots.values
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	roots, err := reverseindex.ResolveRoots(scope.RepositoryRoot, scope.DocsRoot, cwd, fs.Args(), resolved.ReverseIndex.Roots)
	if err != nil {
		return fail(errOut, err)
	}

	if command == "watch" {
		seconds := resolved.Watch.DebounceSeconds
		if debounce >= 0 {
			seconds = debounce
		}
		if err := reverseindex.Watch(ctx, scope.RepositoryRoot, scope.DocsRoot, roots, resolved, format, time.Duration(seconds*float64(time.Second)), once, out); err != nil {
			return fail(errOut, err)
		}
		return 0
	}

	plan, err := reverseindex.Build(scope.RepositoryRoot, scope.DocsRoot, roots, resolved, format)
	if err != nil {
		return fail(errOut, err)
	}
	if command == "fix" {
		changed, err := reverseindex.Apply(scope.RepositoryRoot, plan)
		if err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "ddocs reverse-index fix updated %d file(s) across %d code folder(s)\n", changed, plan.IndexCount)
		writeReverseIndexDiagnostics(out, plan.Diagnostics)
		return 0
	}
	if plan.Failed() {
		fmt.Fprintln(out, "ddocs reverse-index check failed")
		for _, update := range plan.Updates {
			fmt.Fprintln(out, update.Path)
		}
		writeReverseIndexDiagnostics(out, plan.Diagnostics)
		return 1
	}
	fmt.Fprintf(out, "ddocs reverse-index check passed across %d code folder(s)\n", plan.IndexCount)
	writeReverseIndexDiagnostics(out, plan.Diagnostics)
	return 0
}

func writeReverseIndexDiagnostics(out io.Writer, diagnostics []string) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(out, "diagnostic: %s\n", diagnostic)
	}
}
