package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Lokee86/demon-docs/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(app.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
