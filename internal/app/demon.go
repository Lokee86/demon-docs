package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/demon"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/watch"
)

func demonHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs demon [-h] {run,--status,--logs} ...\n\nManage the repository-local self-managing Demon Docs watcher.\n\ncommands:\n  run [--true|--false] [PATH]  enable/check and feed the repository demon\n  --status [PATH]              show demon ownership and feeder status\n  --logs [PATH]                print repository-specific demon logs")
}

func runDemon(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		demonHelp(out)
		return 0
	}
	switch args[0] {
	case "run":
		return demonRun(ctx, args[1:], out, errOut)
	case "--status":
		return demonStatus(args[1:], out, errOut)
	case "--logs":
		return demonLogs(args[1:], out, errOut)
	case "__serve":
		return demonServe(ctx, args[1:], out, errOut)
	case "__feed":
		return demonFeed(ctx, args[1:], out, errOut)
	case "__shutdown":
		return demonShutdown(args[1:], out, errOut)
	case "__enter":
		return demonEnter(args[1:], out, errOut)
	case "__leave":
		return demonLeave(args[1:], out, errOut)
	case "__shell-hook":
		return demonShellHook(args[1:], out, errOut)
	default:
		fmt.Fprintf(errOut, "ddocs demon: error: invalid command %q\n", args[0])
		return 2
	}
}

func demonLocation(argument string, allowBootstrap bool) (repository.Location, error) {
	if argument == "" {
		argument, _ = os.Getwd()
	}
	if !filepath.IsAbs(argument) {
		cwd, err := os.Getwd()
		if err != nil {
			return repository.Location{}, err
		}
		argument = filepath.Join(cwd, argument)
	}
	location, ok := repository.Discover(argument)
	if !ok {
		if allowBootstrap {
			if bootstrapped, detected, bootstrapErr := repository.BootstrapLinkedWorktree(argument); detected {
				if bootstrapErr != nil {
					return repository.Location{}, bootstrapErr
				}
				return bootstrapped, nil
			}
		} else if detected, linked, detectErr := repository.DetectLinkedWorktree(argument); linked {
			if detectErr != nil {
				return repository.Location{}, detectErr
			}
			return detected, nil
		}
		return repository.Location{}, fmt.Errorf("no initialized Demon Docs repository found from %s", argument)
	}
	return location, nil
}

func parseDemonPath(args []string, fs *flag.FlagSet) (string, int) {
	if err := fs.Parse(args); err != nil {
		return "", 2
	}
	if fs.NArg() > 1 {
		return "", 2
	}
	if fs.NArg() == 1 {
		return fs.Arg(0), 0
	}
	return "", 0
}

func demonRun(ctx context.Context, args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("ddocs demon run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	trueFlag := fs.Bool("true", false, "enable the repository demon")
	falseFlag := fs.Bool("false", false, "disable the repository demon")
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: ddocs demon run [--true|--false] [PATH]\n\nEnsure the repository demon is active and feed it from this shell. --false persists disablement; --true persists enablement.")
		return 0
	}
	path, code := parseDemonPath(args, fs)
	if code != 0 || *trueFlag && *falseFlag {
		fmt.Fprintln(errOut, "usage: ddocs demon run [--true|--false] [PATH]")
		return 2
	}
	location, err := demonLocation(path, true)
	if err != nil {
		return fail(errOut, err)
	}
	c, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	if *trueFlag || *falseFlag {
		enabled := *trueFlag
		if err := config.SetDemonRun(location.ConfigPath, enabled); err != nil {
			return fail(errOut, err)
		}
		c.Demon.Run = enabled
		if enabled {
			demon.New(location.Root).ClearShutdown()
		}
		if !enabled {
			r := demon.New(location.Root)
			if err := r.RemoveAllFeeders(); err != nil {
				return fail(errOut, err)
			}
			_ = r.RequestShutdown()
			fmt.Fprintf(out, "disabled demon for %s\n", location.Root)
			return 0
		}
	}
	if !c.Demon.Run {
		fmt.Fprintf(out, "demon disabled for %s\n", location.Root)
		return 0
	}
	r := demon.New(location.Root)
	r.ClearShutdown()
	feeder, feederExists := r.FindFeeder("shell", parentPID())
	if !feederExists {
		feeder, err = r.AddFeeder("shell", os.Getpid(), parentPID())
		if err != nil {
			return fail(errOut, err)
		}
	}
	owner, claimed, err := r.Claim(os.Getpid())
	if err != nil {
		if !feederExists {
			_ = r.RemoveFeeder(feeder.Token)
		}
		return fail(errOut, err)
	}
	if claimed {
		pid, err := startDetached("__serve", location.Root, owner.Token)
		if err != nil {
			_ = r.Release(owner)
			if !feederExists {
				_ = r.RemoveFeeder(feeder.Token)
			}
			return fail(errOut, err)
		}
		_ = r.SetPID(owner.Token, pid)
		fmt.Fprintf(out, "document demon summoned for %s\n", location.Root)
	}
	if !feederExists {
		if _, err := startDetached("__feed", location.Root, feeder.Token); err != nil {
			_ = r.RemoveFeeder(feeder.Token)
			return fail(errOut, err)
		}
	}
	feeders, _ := r.ListFeeders()
	shells, _ := demon.CountKinds(feeders)
	if shells == 1 {
		fmt.Fprintln(out, "1 active shell feeding the demon")
	} else {
		fmt.Fprintf(out, "%d active shells feeding the demon\n", shells)
	}
	_ = ctx
	return 0
}

func demonStatus(args []string, out, errOut io.Writer) int {
	if len(args) > 1 || helpRequested(args) {
		if helpRequested(args) {
			fmt.Fprintln(out, "usage: ddocs demon --status [PATH]")
			return 0
		}
		fmt.Fprintln(errOut, "usage: ddocs demon --status [PATH]")
		return 2
	}
	location, err := demonLocation(firstArg(args), false)
	if err != nil {
		return fail(errOut, err)
	}
	c, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	r := demon.New(location.Root)
	owner, fresh := r.OwnerFresh()
	state := "stopped"
	if _, err := r.ReadOwner(); err == nil {
		state = "stale"
		if fresh {
			state = "running"
		}
	}
	feeders, err := r.SnapshotFeeders()
	if err != nil {
		return fail(errOut, err)
	}
	shells, agents := demon.CountKinds(feeders)
	last := "none"
	if !owner.Heartbeat.IsZero() {
		last = owner.Heartbeat.Format(time.RFC3339)
	}
	docsRoot := filepath.Join(location.Root, c.Root)
	if abs, _, resolveErr := repository.ResolveDocsRoot(location.Root, c.Root); resolveErr == nil {
		docsRoot = filepath.Join(location.Root, filepath.FromSlash(abs))
	}
	pid := "none"
	if owner.PID != 0 {
		pid = strconv.Itoa(owner.PID)
	}
	fmt.Fprintf(out, "repository: %s\nenabled: %t\ndemon: %s\npid: %s\nactive shells: %d\nactive agents: %d\nlast demon heartbeat: %s\nwatching: %s\n", location.Root, c.Demon.Run, state, pid, shells, agents, last, docsRoot)
	return 0
}

func demonLogs(args []string, out, errOut io.Writer) int {
	if len(args) > 1 || helpRequested(args) {
		if helpRequested(args) {
			fmt.Fprintln(out, "usage: ddocs demon --logs [PATH]")
			return 0
		}
		fmt.Fprintln(errOut, "usage: ddocs demon --logs [PATH]")
		return 2
	}
	location, err := demonLocation(firstArg(args), false)
	if err != nil {
		return fail(errOut, err)
	}
	r := demon.New(location.Root)
	for i := demon.LogFiles - 1; i >= 0; i-- {
		path := r.Paths.Log
		if i > 0 {
			path += "." + strconv.Itoa(i)
		}
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fail(errOut, err)
		}
		if _, err := out.Write(data); err != nil {
			return fail(errOut, err)
		}
	}
	return 0
}

func demonServe(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(errOut, "usage: ddocs demon __serve PATH OWNER_TOKEN")
		return 2
	}
	location, err := demonLocation(args[0], false)
	if err != nil {
		return fail(errOut, err)
	}
	r := demon.New(location.Root)
	owner, err := r.ReadOwner()
	if err != nil || owner.Token != args[1] {
		return 2
	}
	log, err := demon.OpenLog(r.Paths)
	if err != nil {
		return fail(errOut, err)
	}
	defer log.Close()
	logger := io.Writer(log)
	_, _ = fmt.Fprintf(logger, "%s demon started pid=%d repository=%s\n", time.Now().UTC().Format(time.RFC3339), os.Getpid(), location.Root)
	c, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	err = r.Serve(ctx, owner, func() (bool, error) {
		fresh, loadErr := config.Load(location.ConfigPath)
		return loadErr == nil && fresh.Demon.Run, loadErr
	}, func(wctx context.Context, w io.Writer) error {
		_, docsRoot, resolveErr := repository.ResolveDocsRoot(location.Root, c.Root)
		if resolveErr != nil {
			return resolveErr
		}
		scope := repository.Scope{RepositoryRoot: location.Root, DocsRoot: docsRoot, ConfigPath: location.ConfigPath, IgnorePath: filepath.Join(location.Root, ".docignore"), Initialized: true}
		features := watch.Features{Indexes: true, Links: true, Reverse: len(c.ReverseIndex.Roots) > 0}
		reverse := reverseOptions{}
		if features.Reverse {
			reverse, resolveErr = resolveReverseOptions(commonFlags{}, c, scope)
			if resolveErr != nil {
				return resolveErr
			}
		}
		return runSelectedWatch(wctx, scope, c, features, reverse, nil, false, w)
	}, logger)
	if err != nil {
		_, _ = fmt.Fprintf(logger, "%s demon stopped: %v\n", time.Now().UTC().Format(time.RFC3339), err)
		return fail(errOut, err)
	}
	_, _ = fmt.Fprintf(logger, "%s demon stopped\n", time.Now().UTC().Format(time.RFC3339))
	_ = out
	return 0
}

func demonFeed(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(errOut, "usage: ddocs demon __feed PATH FEEDER_TOKEN")
		return 2
	}
	location, err := demonLocation(args[0], false)
	if err != nil {
		return fail(errOut, err)
	}
	r := demon.New(location.Root)
	feeders, err := r.ListFeeders()
	if err != nil {
		return fail(errOut, err)
	}
	var feeder demon.Feeder
	for _, candidate := range feeders {
		if candidate.Token == args[1] {
			feeder = candidate
			break
		}
	}
	if feeder.Token == "" {
		return 2
	}
	defer r.RemoveFeeder(feeder.Token)
	for {
		c, loadErr := config.Load(location.ConfigPath)
		if loadErr != nil {
			return fail(errOut, loadErr)
		}
		if !c.Demon.Run || ctx.Err() != nil {
			return 0
		}
		if feeder.ParentPID > 0 && !parentAlive(feeder.ParentPID) {
			return 0
		}
		if _, fresh := r.OwnerFresh(); !fresh {
			owner, claimed, claimErr := r.Claim(os.Getpid())
			if claimErr != nil {
				return fail(errOut, claimErr)
			}
			if claimed {
				pid, startErr := startDetached("__serve", location.Root, owner.Token)
				if startErr != nil {
					_ = r.Release(owner)
					return fail(errOut, startErr)
				}
				_ = r.SetPID(owner.Token, pid)
			}
		}
		if err := r.FeedHeartbeat(feeder); err != nil {
			return 0
		}
		timer := time.NewTimer(r.Timing.FeederHeartbeat)
		select {
		case <-ctx.Done():
			timer.Stop()
			return 0
		case <-timer.C:
		}
	}
}

func demonShutdown(args []string, out, errOut io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintln(errOut, "usage: ddocs demon __shutdown [PATH]")
		return 2
	}
	location, err := demonLocation(firstArg(args), false)
	if err != nil {
		return fail(errOut, err)
	}
	if err := demon.New(location.Root).RequestShutdown(); err != nil {
		return fail(errOut, err)
	}
	_ = out
	return 0
}

func demonEnter(args []string, out, errOut io.Writer) int {
	if len(args) != 2 || (args[1] != "shell" && args[1] != "agent") {
		fmt.Fprintln(errOut, "usage: ddocs demon __enter PATH {shell|agent}")
		return 2
	}
	location, err := demonLocation(args[0], true)
	if err != nil {
		return fail(errOut, err)
	}
	c, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	if !c.Demon.Run {
		return fail(errOut, fmt.Errorf("demon disabled for %s", location.Root))
	}
	r := demon.New(location.Root)
	r.ClearShutdown()
	feeder, feederExists := demon.Feeder{}, false
	if args[1] == "shell" {
		feeder, feederExists = r.FindFeeder(args[1], parentPID())
	}
	if !feederExists {
		feeder, err = r.AddFeeder(args[1], os.Getpid(), parentPID())
		if err != nil {
			return fail(errOut, err)
		}
	}
	owner, claimed, err := r.Claim(os.Getpid())
	if err != nil {
		if !feederExists {
			_ = r.RemoveFeeder(feeder.Token)
		}
		return fail(errOut, err)
	}
	if claimed {
		pid, err := startDetached("__serve", location.Root, owner.Token)
		if err != nil {
			_ = r.Release(owner)
			if !feederExists {
				_ = r.RemoveFeeder(feeder.Token)
			}
			return fail(errOut, err)
		}
		_ = r.SetPID(owner.Token, pid)
	}
	if !feederExists {
		if _, err := startDetached("__feed", location.Root, feeder.Token); err != nil {
			if claimed {
				_ = r.Release(owner)
			}
			_ = r.RemoveFeeder(feeder.Token)
			return fail(errOut, err)
		}
	}
	fmt.Fprintf(out, "token=%s claimed=%t\n", feeder.Token, claimed)
	return 0
}

func demonLeave(args []string, out, errOut io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(errOut, "usage: ddocs demon __leave PATH FEEDER_TOKEN")
		return 2
	}
	location, err := demonLocation(args[0], false)
	if err != nil {
		return fail(errOut, err)
	}
	if err := demon.New(location.Root).RemoveFeeder(args[1]); err != nil {
		return fail(errOut, err)
	}
	_ = out
	return 0
}

func demonShellHook(args []string, out, errOut io.Writer) int {
	if len(args) != 1 || (args[0] != "bash" && args[0] != "powershell") {
		fmt.Fprintln(errOut, "usage: ddocs demon __shell-hook {bash|powershell}")
		return 2
	}
	if args[0] == "bash" {
		_, _ = io.WriteString(out, `# Demon Docs shell integration. Add: eval "$(ddocs demon __shell-hook bash)"
__ddocs_demon_repo=""
__ddocs_demon_token=""
__ddocs_demon_leave() {
  if [ -n "$__ddocs_demon_repo" ] && [ -n "$__ddocs_demon_token" ]; then
    ddocs demon __leave "$__ddocs_demon_repo" "$__ddocs_demon_token" >/dev/null 2>&1 || true
  fi
  __ddocs_demon_repo=""
  __ddocs_demon_token=""
}
__ddocs_demon_tick() {
  local candidate="${PWD}"
  local marker enter token claimed count
  marker="$(ddocs demon --status "$candidate" 2>/dev/null | sed -n 's/^repository: //p')"
  if [ "$marker" = "$__ddocs_demon_repo" ]; then return; fi
  __ddocs_demon_leave
  if [ -z "$marker" ]; then return; fi
  enter="$(DDOCS_PARENT_PID=$$ ddocs demon __enter "$marker" shell 2>/dev/null | tail -n 1)"
  token="$(printf '%s\n' "$enter" | sed -n 's/.*token=\([^ ]*\).*/\1/p')"
  claimed="$(printf '%s\n' "$enter" | sed -n 's/.*claimed=\([^ ]*\).*/\1/p')"
  if [ -z "$token" ]; then return; fi
  __ddocs_demon_repo="$marker"
  __ddocs_demon_token="$token"
  count="$(ddocs demon --status "$marker" 2>/dev/null | sed -n 's/^active shells: //p')"
  if [ "$claimed" != "true" ]; then
    if [ "$count" = "1" ]; then echo "1 active shell feeding the demon"; else echo "$count active shells feeding the demon"; fi
  else
    echo "document demon summoned for $marker"
    if [ "$count" = "1" ]; then echo "1 active shell feeding the demon"; else echo "$count active shells feeding the demon"; fi
  fi
}
	case ";${PROMPT_COMMAND[*]}" in *"__ddocs_demon_tick"*) ;; *) PROMPT_COMMAND="__ddocs_demon_tick${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;; esac`+"\n")
		return 0
	}
	_, _ = io.WriteString(out, `# Demon Docs shell integration. Add: Invoke-Expression (& ddocs demon __shell-hook powershell)
$global:__DdocsDemonRepo = ""
$global:__DdocsDemonToken = ""
function Leave-DdocsDemon {
  if ($global:__DdocsDemonRepo -and $global:__DdocsDemonToken) {
    & ddocs demon __leave $global:__DdocsDemonRepo $global:__DdocsDemonToken *> $null
  }
  $global:__DdocsDemonRepo = ""
  $global:__DdocsDemonToken = ""
}
function Invoke-DdocsDemonHook {
  $candidate = (Get-Location).Path
  $status = @(& ddocs demon --status $candidate 2>$null)
  $repo = ($status | Where-Object { $_ -like "repository: *" } | ForEach-Object { $_.Substring(13) })
  if ($repo -eq $global:__DdocsDemonRepo) { return }
  Leave-DdocsDemon
  if (-not $repo) { return }
  $enter = (& ddocs demon __enter $repo shell 2>$null | Select-Object -Last 1).Trim()
  $token = if ($enter -match 'token=([^ ]+)') { $Matches[1] } else { "" }
  $claimed = if ($enter -match 'claimed=([^ ]+)') { $Matches[1] } else { "false" }
  if (-not $token) { return }
  $global:__DdocsDemonRepo = $repo
  $global:__DdocsDemonToken = $token
  $after = @(& ddocs demon --status $repo 2>$null)
  $count = ($after | Where-Object { $_ -like "active shells: *" } | ForEach-Object { $_.Substring(16) })
  if ($claimed -eq "true") { Write-Host "document demon summoned for $repo" }
  if ($count -eq "1") { Write-Host "1 active shell feeding the demon" } else { Write-Host "$count active shells feeding the demon" }
}
if (-not (Get-Variable __DdocsOriginalPrompt -Scope Global -ErrorAction SilentlyContinue)) {
  $global:__DdocsOriginalPrompt = $function:prompt
  function global:prompt { Invoke-DdocsDemonHook; & $global:__DdocsOriginalPrompt }
}`+"\n")
	return 0
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func parentPID() int {
	if value := os.Getenv("DDOCS_PARENT_PID"); value != "" {
		if pid, err := strconv.Atoi(value); err == nil {
			return pid
		}
	}
	return os.Getppid()
}
