package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/reconcile"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/reverseindex"
	"github.com/Lokee86/demon-docs/internal/watch"
)

const Version = "0.2.0"

type stringsFlag struct{ values []string }

func (s *stringsFlag) String() string     { return strings.Join(s.values, ",") }
func (s *stringsFlag) Set(v string) error { s.values = append(s.values, v); return nil }

type optionalString struct {
	set   bool
	value string
}

func (s *optionalString) String() string { return s.value }
func (s *optionalString) Set(value string) error {
	s.set = true
	s.value = value
	return nil
}

type optionalBool struct{ set, value bool }

func (b *optionalBool) String() string { return strconv.FormatBool(b.value) }
func (b *optionalBool) Set(v string) error {
	x, err := strconv.ParseBool(v)
	if err == nil {
		b.set = true
		b.value = x
	}
	return err
}
func (b *optionalBool) IsBoolFlag() bool { return true }

type commonFlags struct {
	root, config, index, draft, prefix, marker, parent optionalString
	noLocal, noGlobal                                  bool
	docsOnly, linksOnly, reverseOnly                   bool
	includes, excludes                                 stringsFlag
	reverseRoots, codemapHeadings                      stringsFlag
	folderLinks, fileLinks                             optionalBool
}

func addCommon(fs *flag.FlagSet, c *commonFlags) {
	fs.Var(&c.root, "root", "docs root directory to reconcile")
	fs.Var(&c.config, "config", "explicit ddocs config file")
	fs.BoolVar(&c.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&c.noGlobal, "no-global-config", false, "skip the global user config")
	fs.BoolVar(&c.docsOnly, "d", false, "reconcile documentation indexes")
	fs.BoolVar(&c.docsOnly, "docs", false, "reconcile documentation indexes")
	fs.BoolVar(&c.docsOnly, "i", false, "compatibility alias for --docs")
	fs.BoolVar(&c.docsOnly, "indexes", false, "compatibility alias for --docs")
	fs.BoolVar(&c.linksOnly, "l", false, "reconcile links")
	fs.BoolVar(&c.linksOnly, "links", false, "reconcile links")
	fs.BoolVar(&c.reverseOnly, "r", false, "reconcile reverse indexes")
	fs.BoolVar(&c.reverseOnly, "reverse", false, "reconcile reverse indexes")
	fs.Var(&c.reverseRoots, "reverse-root", "override configured reverse-index roots; repeat as needed")
	fs.Var(&c.codemapHeadings, "codemap-heading", "override configured codemap headings; repeat as needed")
	fs.Var(&c.index, "index-file", "override the folder index filename")
	fs.Var(&c.draft, "draft-folder", "override the draft folder name")
	fs.Var(&c.prefix, "draft-description-prefix", "override the draft file description prefix")
	fs.Var(&c.includes, "include", "add an include pattern for indexed files")
	fs.Var(&c.excludes, "exclude", "add an exclude pattern for indexed files")
	fs.Var(&c.marker, "marker-prefix", "override the managed marker prefix")
	fs.Var(&c.parent, "parent-label", "override the parent link label")
	fs.Var(&c.folderLinks, "parent-link-folder-indexes", "enable parent links in folder indexes")
	fs.Var(boolNeg{&c.folderLinks}, "no-parent-link-folder-indexes", "disable parent links in folder indexes")
	fs.Var(&c.fileLinks, "parent-link-indexed-files", "enable parent links in indexed files")
	fs.Var(boolNeg{&c.fileLinks}, "no-parent-link-indexed-files", "disable parent links in indexed files")
}

type boolNeg struct{ b *optionalBool }

func (n boolNeg) String() string {
	if n.b == nil {
		return "false"
	}
	return strconv.FormatBool(!n.b.value)
}
func (n boolNeg) Set(value string) error {
	if value != "true" {
		return fmt.Errorf("ignored explicit argument %q", value)
	}
	n.b.set = true
	n.b.value = false
	return nil
}
func (n boolNeg) IsBoolFlag() bool { return true }

func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, topUsageLine)
		fmt.Fprintln(errOut, "ddocs: error: the following arguments are required: command")
		return 2
	}
	if args[0] == "--help" || args[0] == "-h" {
		topHelp(out)
		return 0
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Fprintf(out, "ddocs %s\n", Version)
		return 0
	}
	switch args[0] {
	case "init":
		return runInit(args[1:], out, errOut)
	case "status":
		return runStatus(args[1:], out, errOut)
	case "mv":
		return runMove(args[1:], out, errOut)
	case "fix", "check", "watch":
		return runTree(ctx, args[0], args[1:], out, errOut)
	case "codemap":
		return runCodemap(ctx, args[1:], out, errOut)
	case "suggestions":
		return runSuggestions(ctx, args[1:], out, errOut)
	case "changes":
		return runChanges(args[1:], out, errOut)
	case "config":
		return runConfig(args[1:], out, errOut)
	case "demon":
		return runDemon(ctx, args[1:], out, errOut)
	default:
		fmt.Fprintln(errOut, topUsageLine)
		choices := "init, status, mv, fix, check, watch, codemap, suggestions, changes, config, demon"
		if runtime.GOOS == "windows" {
			choices = "'init', 'status', 'mv', 'fix', 'check', 'watch', 'codemap', 'suggestions', 'changes', 'config', 'demon'"
		}
		fmt.Fprintf(errOut, "ddocs: error: argument command: invalid choice: '%s' (choose from %s)\n", args[0], choices)
		return 2
	}
}
func topHelp(w io.Writer) {
	fmt.Fprintf(w, "%s\n\nddocs reconciles documentation indexes, reverse indexes, and repository-local links in Markdown documents with the filesystem.\n\npositional arguments:\n  {init,status,mv,fix,check,watch,codemap,suggestions,changes,config,demon}\n    init                initialize a Demon Docs repository\n    status              show the detected repository and docs root\n    mv                  move a file or directory and rewrite affected links\n    fix                 reconcile selected systems and write updates\n    check               verify selected systems without writing\n    watch               reconcile selected systems and watch for changes\n    codemap             extract and benchmark authored code-map relationships\n    suggestions         inspect and decide unresolved repair suggestions\n    changes             inspect, undo, and block applied repairs\n    config              inspect config path selection and resolved config\n    demon               manage the repository-local self-managing watcher\n\nreconciliation selectors:\n  -d, --docs            documentation-folder indexes\n  -l, --links           repository-local Markdown links\n  -r, --reverse         code-folder reverse indexes\n  -i, --indexes         compatibility alias for --docs\n\nUse selectors with check, fix, or watch. Run `ddocs check --help` for selector defaults, reverse-root overrides, and codemap-heading configuration.\n\noptions:\n  -h, --help            show this help message and exit\n  -v, --version         show program's version number and exit\n\nExamples:\n  ddocs init --root docs\n  ddocs status\n  ddocs mv --dry-run docs/old.md docs/new.md\n  ddocs mv docs/old.md docs/new.md\n  ddocs fix\n  ddocs check -r\n  ddocs fix --reverse --reverse-root services/game-server\n  ddocs watch -d -r\n  ddocs demon --help\n  ddocs demon --status\n  ddocs demon run\n  ddocs codemap export\n  ddocs suggestions docs/guide.md\n  ddocs changes docs/guide.md\n  ddocs config paths\n  ddocs --version\n", topUsageLine)
}

func initHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs init [-h] --root PATH\n\nInitialize a Demon Docs repository in the current directory.\n\noptions:\n  -h, --help   show this help message and exit\n  --root PATH  docs root, relative to the repository root\n\nThe command creates .ddocs/config.toml. The current directory becomes the repository root, and the docs root must already exist inside it. New repository configs enable the self-managing watcher by default with [demon].run = true.")
}

func runInit(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		initHelp(out)
		return 0
	}
	fs := flag.NewFlagSet("ddocs init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var root optionalString
	fs.Var(&root, "root", "docs root, relative to the repository root")
	if err := fs.Parse(args); err != nil {
		writeInitParseError(errOut, err)
		return 2
	}
	if fs.NArg() != 0 {
		writeUnrecognized(errOut, fs.Args())
		return 2
	}
	if !root.set {
		fmt.Fprintln(errOut, "usage: ddocs init [-h] --root PATH")
		fmt.Fprintln(errOut, "ddocs init: error: the following arguments are required: --root")
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return fail(errOut, err)
	}
	if existing, ok := repository.FindMarker(cwd); ok {
		fmt.Fprintf(errOut, "ddocs error: demon-docs repository already initialized at %s\n", existing)
		return 2
	}
	relative, absolute, err := repository.ResolveDocsRoot(cwd, root.value)
	if err != nil {
		return fail(errOut, err)
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(errOut, "ddocs error: docs root does not exist: %s\n", absolute)
		return 2
	}
	configPath, err := repository.Initialize(cwd, config.RepositoryStarterText(relative))
	if err != nil {
		return fail(errOut, err)
	}
	fmt.Fprintf(out, "initialized demon-docs repository at %s\nconfig: %s\ndocs root: %s\n", cwd, configPath, relative)
	return 0
}

func statusHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs status [-h]\n\nShow the Demon Docs repository detected from the current directory.\n\noptions:\n  -h, --help  show this help message and exit")
}

func runStatus(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		statusHelp(out)
		return 0
	}
	if len(args) != 0 {
		writeUnrecognized(errOut, args)
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	location, ok := repository.Discover(cwd)
	if !ok {
		fmt.Fprintln(errOut, "ddocs error: no Demon Docs repository found")
		return 2
	}
	resolved, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	scope, err := repository.ResolveScope(repository.ScopeOptions{
		WorkingDirectory: cwd,
		ConfigPath:       location.ConfigPath,
		ConfiguredRoot:   resolved.Root,
	})
	if err != nil {
		return fail(errOut, err)
	}
	fmt.Fprintf(out, "repository root = %s\ndocs root = %s\nconfig = %s\ndocignore = %s\ndocs root exists = %t\ndocignore exists = %t\n", scope.RepositoryRoot, scope.DocsRoot, scope.ConfigPath, scope.IgnorePath, repository.DocsRootExists(scope), pathExists(scope.IgnorePath))
	return 0
}

func writeInitParseError(w io.Writer, err error) {
	message := err.Error()
	if match := unknownFlagPattern.FindStringSubmatch(message); match != nil {
		writeUnrecognized(w, []string{"--" + match[1]})
		return
	}
	fmt.Fprintln(w, "usage: ddocs init [-h] --root PATH")
	if match := missingValuePattern.FindStringSubmatch(message); match != nil {
		fmt.Fprintf(w, "ddocs init: error: argument --%s: expected one argument\n", match[1])
		return
	}
	fmt.Fprintf(w, "ddocs init: error: %s\n", message)
}

func treeUsage(command string) string {
	usage := fmt.Sprintf("usage: ddocs %s [-h] [-d] [-l] [-r] [--root PATH] [--config PATH]\n", command) +
		"                     [--no-local-config] [--no-global-config] [--index-file NAME]\n" +
		"                     [--reverse-root PATH] [--codemap-heading TEXT]\n" +
		"                     [--draft-folder NAME] [--draft-description-prefix TEXT]\n" +
		"                     [--include PATTERN] [--exclude PATTERN]\n" +
		"                     [--marker-prefix TEXT] [--parent-label TEXT]\n" +
		"                     [--parent-link-folder-indexes | --no-parent-link-folder-indexes]\n" +
		"                     [--parent-link-indexed-files | --no-parent-link-indexed-files]"
	if command == "watch" {
		usage += "\n                     [--once] [--debounce-seconds FLOAT]"
	}
	return usage
}

func treeHelp(w io.Writer, command string) {
	usage := treeUsage(command)
	description := map[string]string{
		"fix":   "Reconcile selected documentation indexes, links, and reverse indexes and write needed updates.",
		"check": "Verify that selected documentation indexes, links, and reverse indexes are already reconciled. Report orphan managed Markdown documents when links are selected.",
		"watch": "Watch runs in the foreground by default, runs selected reconciliation immediately, and then watches for relevant filesystem changes. Each reconciliation diagnostic is printed as an individual message.",
	}[command]
	watchOptions := ""
	watchNotes := ""
	if command == "watch" {
		watchOptions = "  --once                run one reconciliation pass and exit\n  --debounce-seconds FLOAT\n                        override the watcher debounce interval\n"
		watchNotes = "\nWatcher lifecycle:\n  - ddocs watch remains attached to the current terminal\n  - use ddocs demon for detached, repository-local self-management\n"
	}
	fmt.Fprintf(w, "%s\n\n%s\n\noptions:\n  -h, --help            show this help message and exit\n  -d, --docs            reconcile documentation-folder indexes\n  -l, --links           reconcile repository-local Markdown links\n  -r, --reverse         reconcile code-folder reverse indexes\n  -i, --indexes         compatibility alias for --docs\n  --root PATH           docs root directory to reconcile\n  --reverse-root PATH   replace [reverse_index].roots for this run; repeat as needed\n  --codemap-heading TEXT\n                        replace [codemap].headings for this run; repeat as needed\n  --config PATH         explicit ddocs config file\n  --no-local-config     skip current-directory local config\n  --no-global-config    skip the global user config\n  --index-file NAME     override the folder index filename\n  --draft-folder NAME   override the draft folder name\n  --draft-description-prefix TEXT\n                        override the draft file description prefix\n  --include PATTERN     add an include pattern for indexed files\n  --exclude PATTERN     add an exclude pattern for indexed files\n  --marker-prefix TEXT  override the managed marker prefix\n  --parent-label TEXT   override the parent link label\n  --parent-link-folder-indexes, --no-parent-link-folder-indexes\n                        enable parent links in folder indexes\n  --parent-link-indexed-files, --no-parent-link-indexed-files\n                        enable parent links in indexed files\n%s\nSelector rules:\n  - when any selector is supplied, only selected systems run\n  - without selectors, docs and links run\n  - without selectors, reverse also runs when reverse roots are configured or supplied\n\nLink reconciliation:\n  - validates and repairs Markdown links, images, and reference definitions\n  - supports wiki links such as [[guide]], [[docs/guide|Guide]], and ![[image.png]]\n  - validates local HTML href, src, and poster targets\n  - reports undefined explicit or collapsed reference labels such as [Guide][guide] and [guide][]\n  - leaves shortcut references such as [guide] untreated unless a definition exists\n\nReverse-index rules:\n  - -r/--reverse requires [reverse_index].roots or at least one --reverse-root\n  - relative --reverse-root paths resolve from the current working directory\n  - absolute --reverse-root paths must remain inside the repository\n  - --codemap-heading matching is case-insensitive and replaces configured headings\n  - reverse reconciliation errors when no matching codemap section exists\n  - a matching codemap section with no code targets is reported separately\n%s\nExamples:\n  ddocs %s -d\n  ddocs %s -l\n  ddocs %s -r --reverse-root services/game-server\n  ddocs %s -d -r\n  ddocs %s -r --codemap-heading \"Implementation map\"\n\nConfig selection order:\n  1. --config PATH\n  2. nearest .ddocs/config.toml found upward\n  3. ./.demon-docs.toml\n  4. ./demon-docs.toml\n  5. ./.doc-ledger.toml\n  6. ./doc-ledger.toml\n  7. global user config (demon-docs, then doc-ledger compatibility)\n  8. built-in defaults\n\nConfig rules:\n  - repository config is discovered by searching upward\n  - legacy local config is current-directory only\n  - local and global configs are not merged\n  - CLI flags override the selected config\n", usage, description, watchOptions, watchNotes, command, command, command, command, command)
}

var (
	unknownFlagPattern  = regexp.MustCompile(`flag provided but not defined: -([^ ]+)`)
	missingValuePattern = regexp.MustCompile(`flag needs an argument: -([^ ]+)`)
	invalidValuePattern = regexp.MustCompile(`invalid value "([^"]*)" for flag -([^:]+):`)
	invalidBoolPattern  = regexp.MustCompile(`invalid boolean value "([^"]*)" for -([^:]+):`)
)

func writeTreeParseError(w io.Writer, command string, err error) {
	message := err.Error()
	if match := unknownFlagPattern.FindStringSubmatch(message); match != nil {
		fmt.Fprintln(w, topUsageLine)
		fmt.Fprintf(w, "ddocs: error: unrecognized arguments: --%s\n", match[1])
		return
	}
	fmt.Fprintln(w, treeUsage(command))
	if match := missingValuePattern.FindStringSubmatch(message); match != nil {
		fmt.Fprintf(w, "ddocs %s: error: argument --%s: expected one argument\n", command, match[1])
		return
	}
	match := invalidValuePattern.FindStringSubmatch(message)
	if match == nil {
		match = invalidBoolPattern.FindStringSubmatch(message)
	}
	if match != nil {
		value, name := match[1], match[2]
		if name == "debounce-seconds" {
			fmt.Fprintf(w, "ddocs %s: error: argument --%s: invalid float value: '%s'\n", command, name, value)
			return
		}
		if name == "parent-link-folder-indexes" || name == "no-parent-link-folder-indexes" {
			fmt.Fprintf(w, "ddocs %s: error: argument --parent-link-folder-indexes/--no-parent-link-folder-indexes: ignored explicit argument '%s'\n", command, value)
			return
		}
		if name == "parent-link-indexed-files" || name == "no-parent-link-indexed-files" {
			fmt.Fprintf(w, "ddocs %s: error: argument --parent-link-indexed-files/--no-parent-link-indexed-files: ignored explicit argument '%s'\n", command, value)
			return
		}
	}
	fmt.Fprintf(w, "ddocs %s: error: %s\n", command, message)
}

func runTree(ctx context.Context, command string, args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("ddocs "+command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var flags commonFlags
	addCommon(fs, &flags)
	once := false
	debounce := -1.0
	if command == "watch" {
		fs.BoolVar(&once, "once", false, "run one reconciliation pass and exit")
		fs.Float64Var(&debounce, "debounce-seconds", -1, "override the watcher debounce interval")
	}
	if helpRequested(args) {
		treeHelp(out, command)
		return 0
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		writeTreeParseError(errOut, command, err)
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(errOut, topUsageLine)
		fmt.Fprintf(errOut, "ddocs: error: unrecognized arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}
	c, path, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	applyOverrides(&c, flags)
	scope, err := resolveScope(flags.root, c.Root, path)
	if err != nil {
		return fail(errOut, err)
	}
	features := selectedFeatures(flags, c)
	if (features.Indexes || features.Reverse) && !repository.DocsRootExists(scope) {
		fmt.Fprintf(errOut, "ddocs error: docs root does not exist: %s\n", scope.DocsRoot)
		return 2
	}
	reverseOptions := reverseOptions{}
	if features.Reverse {
		reverseOptions, err = resolveReverseOptions(flags, c, scope)
		if err != nil {
			return fail(errOut, err)
		}
	}
	if command == "watch" {
		var d *float64
		if debounce >= 0 {
			d = &debounce
		}
		if err := runSelectedWatch(ctx, scope, c, features, reverseOptions, d, once, out); err != nil {
			return fail(errOut, err)
		}
		return 0
	}

	indexResult := model.ReconcileResult{}
	if features.Indexes {
		indexResult, err = reconcile.TreeWithIgnoreRoot(scope.DocsRoot, scope.RepositoryRoot, c)
		if err != nil {
			return fail(errOut, err)
		}
	}
	linkPlan := links.Plan{}
	reversePlan := reverseindex.Plan{}
	if features.Reverse {
		reversePlan, err = reverseindex.Build(scope.RepositoryRoot, scope.DocsRoot, reverseOptions.roots, c, reverseOptions.format)
		if err != nil {
			return fail(errOut, err)
		}
	}
	if command == "fix" {
		changed := 0
		if features.Indexes {
			count, err := reconcile.ApplyWithin(indexResult, scope.DocsRoot)
			if err != nil {
				return fail(errOut, err)
			}
			changed += count
		}
		if features.Links {
			linkPlan, err = links.Reconcile(scope.RepositoryRoot)
			if err != nil {
				return fail(errOut, err)
			}
			count, err := links.ApplyAndSave(&linkPlan)
			if err != nil {
				return fail(errOut, err)
			}
			changed += count
		}
		if features.Reverse {
			count, err := reverseindex.Apply(scope.RepositoryRoot, reversePlan)
			if err != nil {
				return fail(errOut, err)
			}
			changed += count
		}
		fmt.Fprintf(out, "ddocs fix updated %d file(s)\n", changed)
		writeMessages(out, indexResult.Messages)
		writeMessages(out, linkPlan.Messages)
		writeReverseIndexDiagnostics(out, reversePlan.Diagnostics)
		if linkPlan.Unresolved > 0 {
			fmt.Fprintf(out, "ddocs fix unresolved %d link(s)\n", linkPlan.Unresolved)
			return 1
		}
		return 0
	}

	if features.Links {
		linkPlan, err = links.Reconcile(scope.RepositoryRoot)
		if err != nil {
			return fail(errOut, err)
		}
	}
	orphanDocuments := []string{}
	if features.Links && repository.DocsRootExists(scope) {
		orphanDocuments, err = findOrphanDocuments(scope, c, linkPlan)
		if err != nil {
			return fail(errOut, err)
		}
	}
	failed := len(indexResult.Updates) > 0 || features.Links && linkPlan.Failed() || features.Reverse && reversePlan.Failed() || len(orphanDocuments) > 0
	if failed {
		fmt.Fprintln(out, "ddocs check failed")
		for _, update := range indexResult.Updates {
			fmt.Fprintln(out, update.Path)
		}
		for _, update := range linkPlan.Updates {
			fmt.Fprintln(out, update.Path)
		}
		for _, update := range reversePlan.Updates {
			fmt.Fprintln(out, update.Path)
		}
		writeMessages(out, indexResult.Messages)
		writeMessages(out, linkPlan.Messages)
		for _, path := range orphanDocuments {
			fmt.Fprintf(out, "message: Orphan document: %s\n", path)
		}
		writeReverseIndexDiagnostics(out, reversePlan.Diagnostics)
		return 1
	}
	fmt.Fprintln(out, "ddocs check passed")
	return 0
}

func selectedFeatures(flags commonFlags, c config.Config) watch.Features {
	if !flags.docsOnly && !flags.linksOnly && !flags.reverseOnly {
		reverseConfigured := len(c.ReverseIndex.Roots) > 0 || len(flags.reverseRoots.values) > 0 || len(flags.codemapHeadings.values) > 0
		return watch.Features{Indexes: true, Links: true, Reverse: reverseConfigured}
	}
	return watch.Features{Indexes: flags.docsOnly, Links: flags.linksOnly, Reverse: flags.reverseOnly}
}

func writeMessages(out io.Writer, messages []string) {
	for _, message := range messages {
		fmt.Fprintf(out, "message: %s\n", message)
	}
}
func load(f commonFlags, errOut io.Writer) (config.Config, string, int) {
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	explicit := ""
	if f.config.set {
		explicit = f.config.value
	}
	p := ""
	if f.config.set && explicit == "" {
		p = filepath.Clean(cwd)
	} else {
		p = config.Select(cwd, explicit, f.noLocal, f.noGlobal, os.Getenv, home)
	}
	if p == "" {
		return config.Default(), "", 0
	}
	c, err := config.Load(p)
	if err != nil {
		fmt.Fprintf(errOut, "ddocs error: %v\n", err)
		return c, p, 2
	}
	return c, p, 0
}
func applyOverrides(c *config.Config, f commonFlags) {
	if f.index.set {
		c.IndexFile = f.index.value
		c.Files.IndexFile = f.index.value
	}
	if f.draft.set {
		c.Draft.Folder = f.draft.value
	}
	if f.prefix.set {
		c.Draft.DescriptionPrefix = f.prefix.value
	}
	if f.includes.values != nil {
		c.Files.IncludePatterns = f.includes.values
	}
	if f.excludes.values != nil {
		c.Files.ExcludePatterns = f.excludes.values
	}
	if f.marker.set {
		c.Markers.Prefix = f.marker.value
	}
	if f.parent.set {
		c.ParentLink.Label = f.parent.value
	}
	if f.folderLinks.set {
		c.ParentLink.FolderIndexes = f.folderLinks.value
	}
	if f.fileLinks.set {
		c.ParentLink.IndexedFiles = f.fileLinks.value
	}
}
func resolveScope(arg optionalString, configured, configPath string) (repository.Scope, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return repository.Scope{}, err
	}
	return repository.ResolveScope(repository.ScopeOptions{
		WorkingDirectory: cwd,
		ConfigPath:       configPath,
		ConfiguredRoot:   configured,
		RootOverride:     arg.value,
		HasRootOverride:  arg.set,
	})
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "ddocs error: %v\n", err); return 2 }

func codemapHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs codemap [-h] {export,benchmark,precision} ...\n\nExtract authored code-map relationships and benchmark missing-link suggestions.\n\npositional arguments:\n  {export,benchmark,precision}\n    export              write the deterministic codemap dataset as JSON\n    benchmark           run a deterministic missing-link benchmark\n    precision           generate, sample, or evaluate precision suggestions\n\noptions:\n  -h, --help            show this help message and exit")
}

func codemapExportHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs codemap export [-h] [--root PATH] [--config PATH]\n                             [--no-local-config] [--no-global-config]\n                             [--heading TEXT] [--target-base BASE]\n                             [--target-root PATH] [--output PATH]\n\nScan Markdown documents and export normalized code-map links, diagnostics, target resolution, and content hashes. JSON is written to stdout unless --output is provided.\n\noptions:\n  -h, --help          show this help message and exit\n  --root PATH         override the configured docs root\n  --config PATH       explicit ddocs config file\n  --no-local-config   skip current-directory local config\n  --no-global-config  skip the global user config\n  --heading TEXT      accepted code-map heading; repeat to replace configured headings\n  --target-base BASE  resolve targets from repository or document (default repository)\n  --target-root PATH  repository-relative component root; repeat as needed\n  --output PATH       write JSON to a file instead of stdout")
}

func runCodemap(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, "usage: ddocs codemap [-h] {export,benchmark,precision} ...")
		fmt.Fprintln(errOut, "ddocs codemap: error: the following arguments are required: codemap_command")
		return 2
	}
	if args[0] == "-h" || args[0] == "--help" {
		codemapHelp(out)
		return 0
	}
	if args[0] == "benchmark" {
		return runCodemapBenchmark(ctx, args[1:], out, errOut)
	}
	if args[0] == "precision" {
		return runCodemapPrecision(ctx, args[1:], out, errOut)
	}
	if args[0] != "export" {
		fmt.Fprintln(errOut, "usage: ddocs codemap [-h] {export,benchmark,precision} ...")
		fmt.Fprintf(errOut, "ddocs codemap: error: argument codemap_command: invalid choice: '%s' (choose from export, benchmark, precision)\n", args[0])
		return 2
	}
	if helpRequested(args[1:]) {
		codemapExportHelp(out)
		return 0
	}

	fs := flag.NewFlagSet("ddocs codemap export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var flags commonFlags
	var headings stringsFlag
	var targetRoots stringsFlag
	var output optionalString
	targetBase := string(codemap.TargetBaseRepository)
	fs.Var(&flags.root, "root", "override the configured docs root")
	fs.Var(&flags.config, "config", "explicit ddocs config file")
	fs.BoolVar(&flags.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&flags.noGlobal, "no-global-config", false, "skip the global user config")
	fs.Var(&headings, "heading", "accepted code-map heading")
	fs.StringVar(&targetBase, "target-base", targetBase, "repository or document")
	fs.Var(&targetRoots, "target-root", "repository-relative component root")
	fs.Var(&output, "output", "write JSON to a file")
	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(errOut, "ddocs codemap export: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 0 {
		writeUnrecognized(errOut, fs.Args())
		return 2
	}
	if targetBase != string(codemap.TargetBaseRepository) && targetBase != string(codemap.TargetBaseDocument) {
		fmt.Fprintf(errOut, "ddocs codemap export: error: invalid --target-base %q; expected repository or document\n", targetBase)
		return 2
	}
	resolved, configPath, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	scope, err := resolveScope(flags.root, resolved.Root, configPath)
	if err != nil {
		return fail(errOut, err)
	}
	if !repository.DocsRootExists(scope) {
		fmt.Fprintf(errOut, "ddocs error: docs root does not exist: %s\n", scope.DocsRoot)
		return 2
	}
	format := codemap.DefaultFormat()
	format.SectionHeadings = append([]string(nil), resolved.Codemap.Headings...)
	if len(headings.values) > 0 {
		format.SectionHeadings = headings.values
	}
	format.TargetBase = codemap.TargetBase(targetBase)
	format.TargetRoots = targetRoots.values
	dataset, err := codemap.BuildDataset(scope.RepositoryRoot, scope.DocsRoot, format)
	if err != nil {
		return fail(errOut, err)
	}
	if output.set {
		if err := codemap.ExportDataset(output.value, dataset); err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "exported %d codemap link(s) from %d document(s) to %s\n", len(dataset.Entries), len(dataset.Documents), output.value)
		return 0
	}
	encoded, err := codemap.MarshalDataset(dataset)
	if err != nil {
		return fail(errOut, err)
	}
	if _, err := out.Write(encoded); err != nil {
		return fail(errOut, err)
	}
	return 0
}

func configHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs config [-h] {paths,show,init} ...\n\nInspect config discovery and show the resolved selected config.\n\npositional arguments:\n  {paths,show,init}\n    paths            show config path candidates\n    show             show the resolved selected config\n    init             write a legacy standalone config file\n\noptions:\n  -h, --help         show this help message and exit\n\nRepository config is discovered by searching upward for .ddocs/config.toml.\nLegacy local config lookup remains current-directory only.\nLocal and global configs are not merged.\nCLI flags override the selected config.\n\nSubcommands:\n  paths  print repository, local, global, and selected config paths\n  show   print the resolved selected config\n  init   write a legacy standalone local or global config")
}

func configPathsHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs config paths [-h]\n\nPrint repository, local, global, and selected config paths.\n\noptions:\n  -h, --help  show this help message and exit\n\nRepository config:\n  nearest .ddocs/config.toml found by searching upward\n\nLegacy local config candidates:\n  ./.demon-docs.toml\n  ./demon-docs.toml\n  ./.doc-ledger.toml\n  ./doc-ledger.toml\n\nGlobal config candidates:\n  demon-docs/config.toml\n  doc-ledger/config.toml")
}

func configShowHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs config show [-h] [--config PATH] [--no-local-config]\n                              [--no-global-config]\n\nPrint the resolved selected config after config-file selection and before CLI\noverrides.\n\noptions:\n  -h, --help          show this help message and exit\n  --config PATH       explicit ddocs config file\n  --no-local-config   skip current-directory local config\n  --no-global-config  skip the global user config")
}

func configInitHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs config init [-h] (--local | --global) [--force]\n\nWrite a starter config file in the current directory or the global user config location.\n\noptions:\n  -h, --help  show this help message and exit\n  --local     write .demon-docs.toml in the current directory\n  --global    write the global config file\n  --force     overwrite an existing config file\n\nOptions:\n  --local   write .demon-docs.toml in the current directory\n  --global  write the global config file\n  --force   overwrite an existing config file")
}

const (
	topUsageLine    = "usage: ddocs [-h] [-v] {init,status,mv,fix,check,watch,codemap,suggestions,changes,config,demon} ..."
	configUsageLine = "usage: ddocs config [-h] {paths,show,init} ..."
	configShowUsage = "usage: ddocs config show [-h] [--config PATH] [--no-local-config]\n                              [--no-global-config]"
	configInitUsage = "usage: ddocs config init [-h] (--local | --global) [--force]"
)

func writeUnrecognized(w io.Writer, args []string) {
	fmt.Fprintln(w, topUsageLine)
	fmt.Fprintf(w, "ddocs: error: unrecognized arguments: %s\n", strings.Join(args, " "))
}

func writeConfigFlagError(w io.Writer, command string, err error) {
	message := err.Error()
	if match := unknownFlagPattern.FindStringSubmatch(message); match != nil {
		writeUnrecognized(w, []string{"--" + match[1]})
		return
	}
	if match := missingValuePattern.FindStringSubmatch(message); match != nil {
		if command == "show" {
			fmt.Fprintln(w, configShowUsage)
		}
		fmt.Fprintf(w, "ddocs config %s: error: argument --%s: expected one argument\n", command, match[1])
		return
	}
	fmt.Fprintf(w, "ddocs config %s: error: %s\n", command, message)
}

func configChoiceList() string {
	if runtime.GOOS == "windows" {
		return "'paths', 'show', 'init'"
	}
	return "paths, show, init"
}

func runConfig(args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, configUsageLine)
		fmt.Fprintln(errOut, "ddocs config: error: the following arguments are required: config_command")
		return 2
	}
	if args[0] == "--help" || args[0] == "-h" {
		configHelp(out)
		return 0
	}
	switch args[0] {
	case "paths":
		if helpRequested(args[1:]) {
			configPathsHelp(out)
			return 0
		}
		fs := flag.NewFlagSet("ddocs config paths", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			writeConfigFlagError(errOut, "paths", err)
			return 2
		}
		if fs.NArg() != 0 {
			writeUnrecognized(errOut, fs.Args())
			return 2
		}
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		dot := filepath.Join(cwd, ".demon-docs.toml")
		plain := filepath.Join(cwd, "demon-docs.toml")
		legacyDot := filepath.Join(cwd, ".doc-ledger.toml")
		legacyPlain := filepath.Join(cwd, "doc-ledger.toml")
		local := config.LocalPath(cwd)
		repositoryRoot, repositoryConfig := "", ""
		if location, ok := repository.Discover(cwd); ok {
			repositoryRoot = location.Root
			repositoryConfig = location.ConfigPath
		}
		global := config.GlobalPath(os.Getenv, home)
		legacyGlobal := config.LegacyGlobalPath(os.Getenv, home)
		selected := config.Select(cwd, "", false, false, os.Getenv, home)
		fmt.Fprintf(out, "cwd = %s\nrepository root = %s\nrepository config = %s\nlocal dot config = %s exists=%s\nlocal plain config = %s exists=%s\nlegacy local dot config = %s exists=%s\nlegacy local plain config = %s exists=%s\nselected local config = %s\nglobal config = %s exists=%s\nlegacy global config = %s exists=%s\nselected config = %s\n", cwd, none(repositoryRoot), none(repositoryConfig), dot, existsText(dot), plain, existsText(plain), legacyDot, existsText(legacyDot), legacyPlain, existsText(legacyPlain), none(local), global, existsText(global), legacyGlobal, existsText(legacyGlobal), none(selected))
		return 0
	case "show":
		if helpRequested(args[1:]) {
			configShowHelp(out)
			return 0
		}
		fs := flag.NewFlagSet("ddocs config show", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		var f commonFlags
		fs.Var(&f.config, "config", "explicit ddocs config file")
		fs.BoolVar(&f.noLocal, "no-local-config", false, "skip current-directory local config")
		fs.BoolVar(&f.noGlobal, "no-global-config", false, "skip global user config")
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			writeConfigFlagError(errOut, "show", err)
			return 2
		}
		if fs.NArg() != 0 {
			writeUnrecognized(errOut, fs.Args())
			return 2
		}
		c, p, code := load(f, errOut)
		if code != 0 {
			return code
		}
		show(out, c, p)
		return 0
	case "init":
		if helpRequested(args[1:]) {
			configInitHelp(out)
			return 0
		}
		localAt, globalAt := -1, -1
		for index, arg := range args[1:] {
			switch arg {
			case "--local":
				localAt = index
			case "--global":
				globalAt = index
			}
		}
		if localAt < 0 && globalAt < 0 {
			fmt.Fprintln(errOut, configInitUsage)
			fmt.Fprintln(errOut, "ddocs config init: error: one of the arguments --local --global is required")
			return 2
		}
		if localAt >= 0 && globalAt >= 0 {
			fmt.Fprintln(errOut, configInitUsage)
			if localAt < globalAt {
				fmt.Fprintln(errOut, "ddocs config init: error: argument --global: not allowed with argument --local")
			} else {
				fmt.Fprintln(errOut, "ddocs config init: error: argument --local: not allowed with argument --global")
			}
			return 2
		}
		fs := flag.NewFlagSet("ddocs config init", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		fs.Bool("local", false, "write .demon-docs.toml in current directory")
		global := fs.Bool("global", false, "write global config")
		force := fs.Bool("force", false, "overwrite existing config")
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			writeConfigFlagError(errOut, "init", err)
			return 2
		}
		if fs.NArg() != 0 {
			writeUnrecognized(errOut, fs.Args())
			return 2
		}
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		target := filepath.Join(cwd, ".demon-docs.toml")
		if *global {
			target = config.GlobalPath(os.Getenv, home)
		}
		if _, err := os.Stat(target); err == nil && !*force {
			fmt.Fprintf(errOut, "ddocs error: config file already exists: %s\n", target)
			return 2
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fail(errOut, err)
		}
		if err := os.WriteFile(target, []byte(config.StarterText()), 0o644); err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintln(out, target)
		return 0
	default:
		fmt.Fprintln(errOut, configUsageLine)
		fmt.Fprintf(errOut, "ddocs config: error: argument config_command: invalid choice: '%s' (choose from %s)\n", args[0], configChoiceList())
		return 2
	}
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func existsText(p string) string {
	if pathExists(p) {
		return "True"
	}
	return "False"
}
func none(p string) string {
	if p == "" {
		return "<none>"
	}
	return p
}
func quote(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	return "'" + replacer.Replace(s) + "'"
}
func list(v []string) string {
	q := make([]string, len(v))
	for i, x := range v {
		q[i] = quote(x)
	}
	return "[" + strings.Join(q, ", ") + "]"
}
func show(w io.Writer, c config.Config, path string) {
	selected := path
	if selected == "" {
		selected = "<built-in defaults>"
	}
	fmt.Fprintf(w, "selected_config_path = %s\ndocs_root = %s\nindex_file = %s\n[reverse_index]\nroots = %s\n[codemap]\nheadings = %s\n[review]\nundo_depth = %d\nundo_max_age_days = %d\n[markers]\nprefix = %s\n[parent_link]\nlabel = %s\nfolder_indexes = %t\nindexed_files = %t\n[drafts]\nfolder = %s\ndescription_prefix = %s\n[files]\ninclude_patterns = %s\nexclude_patterns = %s\n", selected, quote(c.Root), quote(c.IndexFile), list(c.ReverseIndex.Roots), list(c.Codemap.Headings), c.Review.UndoDepth, c.Review.UndoMaxAgeDays, quote(c.Markers.Prefix), quote(c.ParentLink.Label), c.ParentLink.FolderIndexes, c.ParentLink.IndexedFiles, quote(c.Draft.Folder), quote(c.Draft.DescriptionPrefix), list(c.Files.IncludePatterns), list(c.Files.ExcludePatterns))
}
