package main

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/Lokee86/demon-docs/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(app.Run(ctx, normalizeDemonArgs(os.Args[1:]), os.Stdout, os.Stderr))
}

func normalizeDemonArgs(args []string) []string {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		return []string{"demon", "--help"}
	}
	if args[0] != "demon" && isDemonCommand(args[0]) {
		return append([]string{"demon"}, args...)
	}
	return args
}

func isDemonCommand(command string) bool {
	switch command {
	case "run", "acquire", "heartbeat", "release", "--status", "--logs":
		return true
	default:
		return strings.HasPrefix(command, "__")
	}
}
