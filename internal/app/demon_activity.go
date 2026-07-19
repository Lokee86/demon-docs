package app

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/demon"
)

func demonAcquire(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("demon acquire", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	client := fs.String("client", "", "external agent client name")
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: demon acquire [-h] --client NAME [PATH]\n       ddocs demon acquire [-h] --client NAME [PATH]\n\nRegister an MCP, Codex, Hermes, or other external agent feeder. The caller must refresh the returned token before feeder expiry and release it on every completion path. The returned line contains the feeder token and whether this call claimed a new owner.\n\noptions:\n  -h, --help     show this help message and exit\n  --client NAME  identify the external host in feeder state\n\nPATH defaults to the current directory. A linked worktree is bootstrapped on this mutating entry.")
		return 0
	}
	path, code := parseDemonPath(args, fs)
	if code != 0 || *client == "" {
		fmt.Fprintln(errOut, "usage: demon acquire --client NAME [PATH]")
		return 2
	}
	location, err := demonLocation(path, true)
	if err != nil {
		return fail(errOut, err)
	}
	resolved, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	if !resolved.Demon.Run {
		return fail(errOut, fmt.Errorf("demon disabled for %s", location.Root))
	}
	runtime := demon.New(location.Root)
	runtime.ClearShutdown()
	feeder, err := runtime.AddAgentFeeder(*client, os.Getpid(), parentPID())
	if err != nil {
		return fail(errOut, err)
	}
	claimed, err := ensureDemonOwner(runtime, location.Root)
	if err != nil {
		_ = runtime.RemoveFeeder(feeder.Token)
		return fail(errOut, err)
	}
	fmt.Fprintf(out, "token=%s claimed=%t\n", feeder.Token, claimed)
	return 0
}

func demonHeartbeat(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("demon heartbeat", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	token := fs.String("token", "", "external feeder token")
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: demon heartbeat [-h] --token TOKEN [PATH]\n       ddocs demon heartbeat [-h] --token TOKEN [PATH]\n\nRefresh an external agent feeder and recover the repository demon if its owner lease is missing or stale. Heartbeat fails when the feeder token is unknown or the repository demon is disabled.\n\noptions:\n  -h, --help    show this help message and exit\n  --token TOKEN token returned by `demon acquire`\n\nPATH defaults to the current directory and must resolve to the same initialized repository as the token.")
		return 0
	}
	path, code := parseDemonPath(args, fs)
	if code != 0 || *token == "" {
		fmt.Fprintln(errOut, "usage: demon heartbeat --token TOKEN [PATH]")
		return 2
	}
	location, err := demonLocation(path, false)
	if err != nil {
		return fail(errOut, err)
	}
	resolved, err := config.Load(location.ConfigPath)
	if err != nil {
		return fail(errOut, err)
	}
	if !resolved.Demon.Run {
		return fail(errOut, fmt.Errorf("demon disabled for %s", location.Root))
	}
	runtime := demon.New(location.Root)
	if _, err := runtime.HeartbeatFeeder(*token); err != nil {
		return fail(errOut, err)
	}
	runtime.ClearShutdown()
	if _, err := ensureDemonOwner(runtime, location.Root); err != nil {
		return fail(errOut, err)
	}
	_ = out
	return 0
}

func demonRelease(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("demon release", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	token := fs.String("token", "", "external feeder token")
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: demon release [-h] --token TOKEN [PATH]\n       ddocs demon release [-h] --token TOKEN [PATH]\n\nRelease one external agent feeder. The demon remains active while other feeders exist and otherwise stops after its grace period. Hosts should call release on success, failure, cancellation, and timeout paths.\n\noptions:\n  -h, --help    show this help message and exit\n  --token TOKEN token returned by `demon acquire`\n\nPATH defaults to the current directory and must resolve to the same initialized repository as the token.")
		return 0
	}
	path, code := parseDemonPath(args, fs)
	if code != 0 || *token == "" {
		fmt.Fprintln(errOut, "usage: demon release --token TOKEN [PATH]")
		return 2
	}
	location, err := demonLocation(path, false)
	if err != nil {
		return fail(errOut, err)
	}
	if err := demon.New(location.Root).RemoveFeeder(*token); err != nil {
		return fail(errOut, err)
	}
	_ = out
	return 0
}

func ensureDemonOwner(runtime *demon.Runtime, root string) (bool, error) {
	owner, claimed, err := runtime.Claim(os.Getpid())
	if err != nil || !claimed {
		return claimed, err
	}
	pid, err := startDetached("__serve", root, owner.Token)
	if err != nil {
		_ = runtime.Release(owner)
		return false, err
	}
	_ = runtime.SetPID(owner.Token, pid)
	return true, nil
}
