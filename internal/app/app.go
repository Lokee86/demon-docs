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

	"github.com/Lokee86/doc-ledger/internal/config"
	"github.com/Lokee86/doc-ledger/internal/pathutil"
	"github.com/Lokee86/doc-ledger/internal/reconcile"
	"github.com/Lokee86/doc-ledger/internal/watch"
)

const Version = "0.1.1"

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
	includes, excludes                                 stringsFlag
	folderLinks, fileLinks                             optionalBool
}

func addCommon(fs *flag.FlagSet, c *commonFlags) {
	fs.Var(&c.root, "root", "docs root directory to reconcile")
	fs.Var(&c.config, "config", "explicit doc-ledger config file")
	fs.BoolVar(&c.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&c.noGlobal, "no-global-config", false, "skip the global user config")
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
		fmt.Fprintln(errOut, "usage: doc-ledger [-h] [-v] {fix,check,watch,config} ...")
		fmt.Fprintln(errOut, "doc-ledger: error: the following arguments are required: command")
		return 2
	}
	if args[0] == "--help" || args[0] == "-h" {
		topHelp(out)
		return 0
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Fprintf(out, "doc-ledger %s\n", Version)
		return 0
	}
	switch args[0] {
	case "fix", "check", "watch":
		return runTree(ctx, args[0], args[1:], out, errOut)
	case "config":
		return runConfig(args[1:], out, errOut)
	default:
		fmt.Fprintln(errOut, "usage: doc-ledger [-h] [-v] {fix,check,watch,config} ...")
		choices := "fix, check, watch, config"
		if runtime.GOOS == "windows" {
			choices = "'fix', 'check', 'watch', 'config'"
		}
		fmt.Fprintf(errOut, "doc-ledger: error: argument command: invalid choice: '%s' (choose from %s)\n", args[0], choices)
		return 2
	}
}
func topHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: doc-ledger [-h] [-v] {fix,check,watch,config} ...\n\ndoc-ledger reconciles local index files with a file tree.\n\npositional arguments:\n  {fix,check,watch,config}\n    fix                 reconcile and write updated files\n    check               reconcile without writing files\n    watch               watch the tree and rerun reconciliation\n    config              inspect config path selection and resolved config\n\noptions:\n  -h, --help            show this help message and exit\n  -v, --version         show program's version number and exit\n\nExamples:\n  doc-ledger fix\n  doc-ledger check\n  doc-ledger watch\n  doc-ledger config paths\n  doc-ledger config show\n  doc-ledger fix --root docs\n  doc-ledger check --root docs\n  doc-ledger watch --root docs\n  doc-ledger fix --config .doc-ledger.toml\n  doc-ledger --version")
}

func treeUsage(command string) string {
	return map[string]string{
		"fix": "usage: doc-ledger fix [-h] [--root PATH] [--config PATH] [--no-local-config]\n" +
			"                      [--no-global-config] [--index-file NAME]\n" +
			"                      [--draft-folder NAME] [--draft-description-prefix TEXT]\n" +
			"                      [--include PATTERN] [--exclude PATTERN]\n" +
			"                      [--marker-prefix TEXT] [--parent-label TEXT]\n" +
			"                      [--parent-link-folder-indexes | --no-parent-link-folder-indexes]\n" +
			"                      [--parent-link-indexed-files | --no-parent-link-indexed-files]",
		"check": "usage: doc-ledger check [-h] [--root PATH] [--config PATH] [--no-local-config]\n" +
			"                        [--no-global-config] [--index-file NAME]\n" +
			"                        [--draft-folder NAME]\n" +
			"                        [--draft-description-prefix TEXT] [--include PATTERN]\n" +
			"                        [--exclude PATTERN] [--marker-prefix TEXT]\n" +
			"                        [--parent-label TEXT]\n" +
			"                        [--parent-link-folder-indexes | --no-parent-link-folder-indexes]\n" +
			"                        [--parent-link-indexed-files | --no-parent-link-indexed-files]",
		"watch": "usage: doc-ledger watch [-h] [--root PATH] [--config PATH] [--no-local-config]\n" +
			"                        [--no-global-config] [--index-file NAME]\n" +
			"                        [--draft-folder NAME]\n" +
			"                        [--draft-description-prefix TEXT] [--include PATTERN]\n" +
			"                        [--exclude PATTERN] [--marker-prefix TEXT]\n" +
			"                        [--parent-label TEXT]\n" +
			"                        [--parent-link-folder-indexes | --no-parent-link-folder-indexes]\n" +
			"                        [--parent-link-indexed-files | --no-parent-link-indexed-files]\n" +
			"                        [--once] [--debounce-seconds FLOAT]",
	}[command]
}

func treeHelp(w io.Writer, command string) {
	usage := treeUsage(command)
	description := map[string]string{
		"fix":   "Reconcile the docs tree and write any needed updates.",
		"check": "Verify that the docs tree is already reconciled.",
		"watch": "Watch runs in the foreground by default, runs one reconciliation immediately, and then watches for relevant filesystem changes.",
	}[command]
	watchOptions := ""
	if command == "watch" {
		watchOptions = "  --once                run one reconciliation pass and exit\n  --debounce-seconds FLOAT\n                        override the watcher debounce interval\n"
	}
	fmt.Fprintf(w, "%s\n\n%s\n\noptions:\n  -h, --help            show this help message and exit\n  --root PATH           docs root directory to reconcile\n  --config PATH         explicit doc-ledger config file\n  --no-local-config     skip current-directory local config\n  --no-global-config    skip the global user config\n  --index-file NAME     override the folder index filename\n  --draft-folder NAME   override the draft folder name\n  --draft-description-prefix TEXT\n                        override the draft file description prefix\n  --include PATTERN     add an include pattern for indexed files\n  --exclude PATTERN     add an exclude pattern for indexed files\n  --marker-prefix TEXT  override the managed marker prefix\n  --parent-label TEXT   override the parent link label\n  --parent-link-folder-indexes, --no-parent-link-folder-indexes\n                        enable parent links in folder indexes\n  --parent-link-indexed-files, --no-parent-link-indexed-files\n                        enable parent links in indexed files\n%s\nConfig selection order:\n  1. --config PATH\n  2. ./.doc-ledger.toml\n  3. ./doc-ledger.toml\n  4. global user config\n  5. built-in defaults\n\nConfig rules:\n  - local config is current-directory only\n  - there is no upward parent-directory search\n  - local and global configs are not merged\n  - CLI flags override the selected config\n", usage, description, watchOptions)
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
		fmt.Fprintln(w, "usage: doc-ledger [-h] [-v] {fix,check,watch,config} ...")
		fmt.Fprintf(w, "doc-ledger: error: unrecognized arguments: --%s\n", match[1])
		return
	}
	fmt.Fprintln(w, treeUsage(command))
	if match := missingValuePattern.FindStringSubmatch(message); match != nil {
		fmt.Fprintf(w, "doc-ledger %s: error: argument --%s: expected one argument\n", command, match[1])
		return
	}
	match := invalidValuePattern.FindStringSubmatch(message)
	if match == nil {
		match = invalidBoolPattern.FindStringSubmatch(message)
	}
	if match != nil {
		value, name := match[1], match[2]
		if name == "debounce-seconds" {
			fmt.Fprintf(w, "doc-ledger %s: error: argument --%s: invalid float value: '%s'\n", command, name, value)
			return
		}
		if name == "parent-link-folder-indexes" || name == "no-parent-link-folder-indexes" {
			fmt.Fprintf(w, "doc-ledger %s: error: argument --parent-link-folder-indexes/--no-parent-link-folder-indexes: ignored explicit argument '%s'\n", command, value)
			return
		}
		if name == "parent-link-indexed-files" || name == "no-parent-link-indexed-files" {
			fmt.Fprintf(w, "doc-ledger %s: error: argument --parent-link-indexed-files/--no-parent-link-indexed-files: ignored explicit argument '%s'\n", command, value)
			return
		}
	}
	fmt.Fprintf(w, "doc-ledger %s: error: %s\n", command, message)
}

func runTree(ctx context.Context, command string, args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("doc-ledger "+command, flag.ContinueOnError)
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
		fmt.Fprintln(errOut, "usage: doc-ledger [-h] [-v] {fix,check,watch,config} ...")
		fmt.Fprintf(errOut, "doc-ledger: error: unrecognized arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}
	c, path, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	applyOverrides(&c, flags)
	root, err := resolveRoot(flags.root, c.Root, path)
	if err != nil {
		return fail(errOut, err)
	}
	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		fmt.Fprintf(errOut, "doc-ledger error: docs root does not exist: %s\n", root)
		return 2
	}
	if command == "watch" {
		var d *float64
		if debounce >= 0 {
			d = &debounce
		}
		if err := watch.Root(ctx, root, c, d, once, out); err != nil {
			return fail(errOut, err)
		}
		return 0
	}
	result, err := reconcile.Tree(root, c)
	if err != nil {
		return fail(errOut, err)
	}
	if command == "fix" {
		changed, err := reconcile.Apply(result)
		if err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "doc-ledger fix updated %d file(s)\n", changed)
		for _, m := range result.Messages {
			fmt.Fprintf(out, "message: %s\n", m)
		}
		return 0
	}
	if len(result.Updates) > 0 {
		fmt.Fprintln(out, "doc-ledger check failed")
		for _, u := range result.Updates {
			fmt.Fprintln(out, u.Path)
		}
		for _, m := range result.Messages {
			fmt.Fprintf(out, "message: %s\n", m)
		}
		return 1
	}
	fmt.Fprintln(out, "doc-ledger check passed")
	return 0
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
		fmt.Fprintf(errOut, "doc-ledger error: %v\n", err)
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
func resolveRoot(arg optionalString, configured, configPath string) (string, error) {
	if arg.set {
		return pathutil.Resolve(arg.value, "")
	}
	if configPath != "" {
		return pathutil.Resolve(configured, filepath.Dir(configPath))
	}
	return pathutil.Resolve(configured, "")
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "doc-ledger error: %v\n", err); return 2 }

func configHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: doc-ledger config [-h] {paths,show,init} ...\n\nInspect current-directory local config lookup and show the resolved selected config.\n\npositional arguments:\n  {paths,show,init}\n    paths            show config path candidates\n    show             show the resolved selected config\n    init             write a starter config file\n\noptions:\n  -h, --help         show this help message and exit\n\nLocal config lookup is current-directory only.\nThere is no upward parent-directory search.\nLocal and global configs are not merged.\nCLI flags override the selected config.\n\nSubcommands:\n  paths  print local, global, and selected config paths\n  show   print the resolved selected config\n  init   write a starter local or global config")
}

func configPathsHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: doc-ledger config paths [-h]\n\nPrint the current-directory local config, global user config, and selected config paths.\n\noptions:\n  -h, --help  show this help message and exit\n\nThis reports the current-directory local config candidates, the global user config path, and the selected config path.")
}

func configShowHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: doc-ledger config show [-h] [--config PATH] [--no-local-config]\n                              [--no-global-config]\n\nPrint the resolved selected config after config-file selection and before CLI\noverrides.\n\noptions:\n  -h, --help          show this help message and exit\n  --config PATH       explicit doc-ledger config file\n  --no-local-config   skip current-directory local config\n  --no-global-config  skip the global user config")
}

func configInitHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: doc-ledger config init [-h] (--local | --global) [--force]\n\nWrite a starter config file in the current directory or the global user config location.\n\noptions:\n  -h, --help  show this help message and exit\n  --local     write .doc-ledger.toml in the current directory\n  --global    write the global config file\n  --force     overwrite an existing config file\n\nOptions:\n  --local   write .doc-ledger.toml in the current directory\n  --global  write the global config file\n  --force   overwrite an existing config file")
}

func runConfig(args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		configHelp(errOut)
		fmt.Fprintln(errOut, "doc-ledger config: error: the following arguments are required: config_command")
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
		fs := flag.NewFlagSet("doc-ledger config paths", flag.ContinueOnError)
		fs.SetOutput(errOut)
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 2
		}
		if fs.NArg() != 0 {
			fmt.Fprintf(errOut, "doc-ledger error: unexpected argument: %s\n", fs.Arg(0))
			return 2
		}
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		dot := filepath.Join(cwd, ".doc-ledger.toml")
		plain := filepath.Join(cwd, "doc-ledger.toml")
		local := config.LocalPath(cwd)
		global := config.GlobalPath(os.Getenv, home)
		selected := config.Select(cwd, "", false, false, os.Getenv, home)
		fmt.Fprintf(out, "cwd = %s\nlocal dot config = %s exists=%s\nlocal plain config = %s exists=%s\nselected local config = %s\nglobal config = %s exists=%s\nselected config = %s\n", cwd, dot, existsText(dot), plain, existsText(plain), none(local), global, existsText(global), none(selected))
		return 0
	case "show":
		if helpRequested(args[1:]) {
			configShowHelp(out)
			return 0
		}
		fs := flag.NewFlagSet("doc-ledger config show", flag.ContinueOnError)
		fs.SetOutput(errOut)
		var f commonFlags
		fs.Var(&f.config, "config", "explicit doc-ledger config file")
		fs.BoolVar(&f.noLocal, "no-local-config", false, "skip current-directory local config")
		fs.BoolVar(&f.noGlobal, "no-global-config", false, "skip global user config")
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 2
		}
		if fs.NArg() != 0 {
			fmt.Fprintf(errOut, "doc-ledger error: unexpected argument: %s\n", fs.Arg(0))
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
		fs := flag.NewFlagSet("doc-ledger config init", flag.ContinueOnError)
		fs.SetOutput(errOut)
		local := fs.Bool("local", false, "write .doc-ledger.toml in current directory")
		global := fs.Bool("global", false, "write global config")
		force := fs.Bool("force", false, "overwrite existing config")
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 2
		}
		if fs.NArg() != 0 {
			fmt.Fprintf(errOut, "doc-ledger error: unexpected argument: %s\n", fs.Arg(0))
			return 2
		}
		if *local == *global {
			fmt.Fprintln(errOut, "doc-ledger error: exactly one of --local or --global is required")
			return 2
		}
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		target := filepath.Join(cwd, ".doc-ledger.toml")
		if *global {
			target = config.GlobalPath(os.Getenv, home)
		}
		if _, err := os.Stat(target); err == nil && !*force {
			fmt.Fprintf(errOut, "doc-ledger error: config file already exists: %s\n", target)
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
		return fail(errOut, fmt.Errorf("unknown config command: %s", args[0]))
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

func existsText(p string) string {
	_, err := os.Stat(p)
	if err == nil {
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
	fmt.Fprintf(w, "selected_config_path = %s\nroot = %s\nindex_file = %s\n[markers]\nprefix = %s\n[parent_link]\nlabel = %s\nfolder_indexes = %t\nindexed_files = %t\n[drafts]\nfolder = %s\ndescription_prefix = %s\n[files]\ninclude_patterns = %s\nexclude_patterns = %s\n", selected, quote(c.Root), quote(c.IndexFile), quote(c.Markers.Prefix), quote(c.ParentLink.Label), c.ParentLink.FolderIndexes, c.ParentLink.IndexedFiles, quote(c.Draft.Folder), quote(c.Draft.DescriptionPrefix), list(c.Files.IncludePatterns), list(c.Files.ExcludePatterns))
}
