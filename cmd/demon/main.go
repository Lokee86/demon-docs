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
	if len(args) > 0 && (args[0] == "run" || args[0] == "--status" || args[0] == "--logs" || strings.HasPrefix(args[0], "__")) {
		args = append([]string{"demon"}, args...)
	}
	os.Exit(app.Run(ctx, args, os.Stdout, os.Stderr))
}
