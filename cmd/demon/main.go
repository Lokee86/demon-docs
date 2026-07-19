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
	args := os.Args[1:]
	if len(args) > 0 && args[0] != "demon" && isDemonCommand(args[0]) {
		args = append([]string{"demon"}, args...)
	}
	os.Exit(app.Run(ctx, args, os.Stdout, os.Stderr))
}

func isDemonCommand(command string) bool {
	switch command {
	case "run", "acquire", "heartbeat", "release", "--status", "--logs":
		return true
	default:
		return strings.HasPrefix(command, "__")
	}
}
