package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	root, config, index, draft, prefix, marker, parent string
	noLocal, noGlobal                                  bool
	includes, excludes                                 stringsFlag
	folderLinks, fileLinks                             optionalBool
}

func addCommon(fs *flag.FlagSet, c *commonFlags) {
	fs.StringVar(&c.root, "root", "", "docs root directory to reconcile")
	fs.StringVar(&c.config, "config", "", "explicit doc-ledger config file")
	fs.BoolVar(&c.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&c.noGlobal, "no-global-config", false, "skip the global user config")
	fs.StringVar(&c.index, "index-file", "", "override the folder index filename")
	fs.StringVar(&c.draft, "draft-folder", "", "override the draft folder name")
	fs.StringVar(&c.prefix, "draft-description-prefix", "", "override the draft file description prefix")
	fs.Var(&c.includes, "include", "add an include pattern for indexed files")
	fs.Var(&c.excludes, "exclude", "add an exclude pattern for indexed files")
	fs.StringVar(&c.marker, "marker-prefix", "", "override the managed marker prefix")
	fs.StringVar(&c.parent, "parent-label", "", "override the parent link label")
	fs.Var(&c.folderLinks, "parent-link-folder-indexes", "enable parent links in folder indexes")
	fs.Var(boolNeg{&c.folderLinks}, "no-parent-link-folder-indexes", "disable parent links in folder indexes")
	fs.Var(&c.fileLinks, "parent-link-indexed-files", "enable parent links in indexed files")
	fs.Var(boolNeg{&c.fileLinks}, "no-parent-link-indexed-files", "disable parent links in indexed files")
}

type boolNeg struct{ b *optionalBool }

func (n boolNeg) String() string   { return strconv.FormatBool(!n.b.value) }
func (n boolNeg) Set(string) error { n.b.set = true; n.b.value = false; return nil }
func (n boolNeg) IsBoolFlag() bool { return true }

func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		topHelp(errOut)
		fmt.Fprintln(errOut, "doc-ledger error: command is required")
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
		fmt.Fprintf(errOut, "doc-ledger error: unknown command: %s\n", args[0])
		topHelp(errOut)
		return 2
	}
}
func topHelp(w io.Writer) {
	fmt.Fprintln(w, "doc-ledger reconciles local index files with a file tree.\n\nUsage: doc-ledger <command> [options]\n\nCommands:\n  fix     reconcile and write updated files\n  check   reconcile without writing files\n  watch   watch the tree and rerun reconciliation\n  config  inspect config path selection and resolved config\n\nExamples:\n  doc-ledger fix\n  doc-ledger check\n  doc-ledger watch\n  doc-ledger config paths\n  doc-ledger config show\n  doc-ledger fix --root docs\n  doc-ledger check --root docs\n  doc-ledger watch --root docs\n  doc-ledger fix --config .doc-ledger.toml\n  doc-ledger --version")
}
func runTree(ctx context.Context, command string, args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("doc-ledger "+command, flag.ContinueOnError)
	fs.SetOutput(errOut)
	var flags commonFlags
	addCommon(fs, &flags)
	once := false
	debounce := -1.0
	if command == "watch" {
		fs.BoolVar(&once, "once", false, "run one reconciliation pass and exit")
		fs.Float64Var(&debounce, "debounce-seconds", -1, "override the watcher debounce interval")
	}
	if helpRequested(args) {
		fs.SetOutput(out)
	}
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: doc-ledger %s [options]\n", command)
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "\nConfig selection order:\n  1. --config PATH\n  2. ./.doc-ledger.toml\n  3. ./doc-ledger.toml\n  4. global user config\n  5. built-in defaults\n\nConfig rules:\n  - local config is current-directory only\n  - there is no upward parent-directory search\n  - local and global configs are not merged\n  - CLI flags override the selected config")
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(errOut, "doc-ledger error: unexpected argument: %s\n", fs.Arg(0))
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
	p := config.Select(cwd, f.config, f.noLocal, f.noGlobal, os.Getenv, home)
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
	if f.index != "" {
		c.IndexFile = f.index
		c.Files.IndexFile = f.index
	}
	if f.draft != "" {
		c.Draft.Folder = f.draft
	}
	if f.prefix != "" {
		c.Draft.DescriptionPrefix = f.prefix
	}
	if f.includes.values != nil {
		c.Files.IncludePatterns = f.includes.values
	}
	if f.excludes.values != nil {
		c.Files.ExcludePatterns = f.excludes.values
	}
	if f.marker != "" {
		c.Markers.Prefix = f.marker
	}
	if f.parent != "" {
		c.ParentLink.Label = f.parent
	}
	if f.folderLinks.set {
		c.ParentLink.FolderIndexes = f.folderLinks.value
	}
	if f.fileLinks.set {
		c.ParentLink.IndexedFiles = f.fileLinks.value
	}
}
func resolveRoot(arg, configured, configPath string) (string, error) {
	if arg != "" {
		return pathutil.Resolve(arg, "")
	}
	if configPath != "" {
		return pathutil.Resolve(configured, filepath.Dir(configPath))
	}
	return pathutil.Resolve(configured, "")
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "doc-ledger error: %v\n", err); return 2 }

func runConfig(args []string, out, errOut io.Writer) int {
	if len(args) == 0 || args[0] == "--help" {
		fmt.Fprintln(out, "Inspect current-directory local config lookup and show the resolved selected config.\n\nSubcommands:\n  paths  print local, global, and selected config paths\n  show   print the resolved selected config\n  init   write a starter local or global config")
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	switch args[0] {
	case "paths":
		fs := flag.NewFlagSet("doc-ledger config paths", flag.ContinueOnError)
		fs.SetOutput(errOut)
		if helpRequested(args[1:]) {
			fs.SetOutput(out)
		}
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
		fs := flag.NewFlagSet("doc-ledger config show", flag.ContinueOnError)
		fs.SetOutput(errOut)
		var f commonFlags
		fs.StringVar(&f.config, "config", "", "explicit doc-ledger config file")
		fs.BoolVar(&f.noLocal, "no-local-config", false, "skip current-directory local config")
		fs.BoolVar(&f.noGlobal, "no-global-config", false, "skip global user config")
		if helpRequested(args[1:]) {
			fs.SetOutput(out)
		}
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
		fs := flag.NewFlagSet("doc-ledger config init", flag.ContinueOnError)
		fs.SetOutput(errOut)
		local := fs.Bool("local", false, "write .doc-ledger.toml in current directory")
		global := fs.Bool("global", false, "write global config")
		force := fs.Bool("force", false, "overwrite existing config")
		if helpRequested(args[1:]) {
			fs.SetOutput(out)
		}
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
