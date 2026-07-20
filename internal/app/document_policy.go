package app

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/documentpolicy"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func newHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs new [-h] [--force] [--root PATH] DOCUMENT_TYPE PATH\n\nCreate a Markdown document from the TOML schema whose name is the document type. Existing files require confirmation in an interactive terminal or --force in noninteractive use.\n\noptions:\n  -h, --help   show this help message and exit\n  --force      overwrite an existing file without prompting\n  --root PATH  override the configured docs root")
}

func runNew(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		newHelp(out)
		return 0
	}
	fs := flag.NewFlagSet("ddocs new", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var flags commonFlags
	force := false
	fs.BoolVar(&force, "force", false, "overwrite existing file")
	fs.Var(&flags.root, "root", "override configured docs root")
	fs.Var(&flags.config, "config", "explicit ddocs config file")
	fs.BoolVar(&flags.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&flags.noGlobal, "no-global-config", false, "skip global user config")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs new: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 2 {
		fmt.Fprintln(errOut, "usage: ddocs new [-h] [--force] [--root PATH] DOCUMENT_TYPE PATH")
		fmt.Fprintln(errOut, "ddocs new: error: DOCUMENT_TYPE and PATH are required")
		return 2
	}
	cfg, configPath, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	scope, err := resolveScope(flags.root, cfg.Root, configPath)
	if err != nil {
		return fail(errOut, err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	target := fs.Arg(1)
	if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}
	if _, err := os.Stat(target); err == nil && !force {
		confirmed, err := confirmOverwrite(target, errOut)
		if err != nil {
			return fail(errOut, err)
		}
		if !confirmed {
			fmt.Fprintf(errOut, "ddocs new: target exists; use --force for noninteractive overwrite: %s\n", target)
			return 2
		}
		force = true
	}
	created, err := documentpolicy.Create(scope.RepositoryRoot, scope.DocsRoot, cfg, fs.Arg(0), target, force, time.Now())
	if err != nil {
		return fail(errOut, err)
	}
	fmt.Fprintln(out, created)
	return 0
}

func confirmOverwrite(path string, out io.Writer) (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return false, nil
	}
	fmt.Fprintf(out, "WARNING: %s already exists. Overwrite it? [y/N] ", path)
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(answer) == 0 {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func formatHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs format [-h] {ignore,merge,delete} ...\n\nResolve document-body format conflicts explicitly.\n\ncommands:\n  ignore --heading HEADING FILE\n          add a section or duplicate allowance to the document-specific schema\n  merge --heading HEADING FILE\n          merge duplicate sibling sections without rewriting prose\n  delete --heading HEADING --occurrence N FILE\n          delete one explicit section occurrence")
}

func runFormat(args []string, out, errOut io.Writer) int {
	if len(args) == 0 || helpRequested(args) {
		formatHelp(out)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	action := args[0]
	fs := flag.NewFlagSet("ddocs format "+action, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var flags commonFlags
	heading := ""
	occurrence := 0
	fs.StringVar(&heading, "heading", "", "section heading")
	fs.IntVar(&occurrence, "occurrence", 0, "one-based occurrence to delete")
	fs.Var(&flags.config, "config", "explicit ddocs config file")
	fs.BoolVar(&flags.noLocal, "no-local-config", false, "skip current-directory local config")
	fs.BoolVar(&flags.noGlobal, "no-global-config", false, "skip global user config")
	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(errOut, "ddocs format %s: error: %v\n", action, err)
		return 2
	}
	if fs.NArg() != 1 || strings.TrimSpace(heading) == "" {
		fmt.Fprintf(errOut, "usage: ddocs format %s --heading HEADING", action)
		if action == "delete" {
			fmt.Fprint(errOut, " --occurrence N")
		}
		fmt.Fprint(errOut, " FILE")
		fmt.Fprintln(errOut)
		return 2
	}
	cfg, configPath, code := load(flags, errOut)
	if code != 0 {
		return code
	}
	scope, err := resolveScope(optionalString{}, cfg.Root, configPath)
	if err != nil {
		return fail(errOut, err)
	}
	path, err := absoluteDocumentPath(fs.Arg(0))
	if err != nil {
		return fail(errOut, err)
	}
	if !repository.Contains(scope.DocsRoot, path) {
		return fail(errOut, fmt.Errorf("document is outside docs root: %s", path))
	}
	switch action {
	case "ignore":
		schemaPath, err := documentpolicy.IgnoreSection(scope.RepositoryRoot, cfg, path, heading)
		if err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "updated document-specific schema: %s\n", schemaPath)
	case "merge":
		count, err := documentpolicy.MergeSections(path, heading, cfg.Frontmatter.AllowedFormats)
		if err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "merged %d section occurrence(s)\n", count)
	case "delete":
		if occurrence < 1 {
			fmt.Fprintln(errOut, "ddocs format delete: --occurrence must be at least 1")
			return 2
		}
		if err := documentpolicy.DeleteSection(path, heading, occurrence, cfg.Frontmatter.AllowedFormats); err != nil {
			return fail(errOut, err)
		}
		fmt.Fprintf(out, "deleted %s occurrence %s\n", strconv.Quote(heading), strconv.Itoa(occurrence))
	default:
		fmt.Fprintf(errOut, "ddocs format: invalid action %q; expected ignore, merge, or delete\n", action)
		return 2
	}
	return 0
}

func schemaHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs schema init [--force]\n\nWrite the provided Space Rocks-derived starter TOML schemas to the configured human schema directory.")
}

func runSchema(args []string, out, errOut io.Writer) int {
	if len(args) == 0 || helpRequested(args) {
		schemaHelp(out)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	if args[0] != "init" {
		fmt.Fprintf(errOut, "ddocs schema: invalid action %q; expected init\n", args[0])
		return 2
	}
	fs := flag.NewFlagSet("ddocs schema init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "overwrite existing starter schemas")
	if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(errOut, "usage: ddocs schema init [--force]")
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
	cfg, err := configForLocation(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	paths, err := documentpolicy.WriteBuiltinSchemas(location.Root, cfg.Format, *force)
	if err != nil {
		return fail(errOut, err)
	}
	for _, path := range paths {
		fmt.Fprintln(out, path)
	}
	return 0
}

func configForLocation(path string) (config.Config, error) {
	return config.Load(path)
}

func absoluteDocumentPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Abs(path)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(cwd, path))
}
