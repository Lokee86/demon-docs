package app

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/demon"
)

func TestDemonAgentCommandLifecycle(t *testing.T) {
	root := initializedDemonRepo(t)
	withWorkingDirectory(t, root, func(string) {
		runtime := demon.New(root)
		owner, won, err := runtime.Claim(os.Getpid())
		if err != nil || !won {
			t.Fatalf("seed owner: won=%t err=%v", won, err)
		}
		defer runtime.Release(owner)

		var out, errOut bytes.Buffer
		code := Run(context.Background(), []string{"demon", "acquire", "--client", "mcp"}, &out, &errOut)
		if code != 0 {
			t.Fatalf("acquire code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		fields := strings.Fields(out.String())
		if len(fields) != 2 || !strings.HasPrefix(fields[0], "token=") || fields[1] != "claimed=false" {
			t.Fatalf("unexpected acquire output: %q", out.String())
		}
		token := strings.TrimPrefix(fields[0], "token=")
		feeder, err := runtime.ReadFeeder(token)
		if err != nil || feeder.Client != "mcp" || feeder.Kind != "agent" {
			t.Fatalf("feeder=%+v err=%v", feeder, err)
		}

		out.Reset()
		errOut.Reset()
		code = Run(context.Background(), []string{"demon", "heartbeat", "--token", token}, &out, &errOut)
		if code != 0 {
			t.Fatalf("heartbeat code=%d out=%q err=%q", code, out.String(), errOut.String())
		}

		out.Reset()
		errOut.Reset()
		code = Run(context.Background(), []string{"demon", "--status"}, &out, &errOut)
		if code != 0 || !strings.Contains(out.String(), "active agents: 1") {
			t.Fatalf("status code=%d out=%q err=%q", code, out.String(), errOut.String())
		}

		out.Reset()
		errOut.Reset()
		code = Run(context.Background(), []string{"demon", "release", "--token", token}, &out, &errOut)
		if code != 0 {
			t.Fatalf("release code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if _, err := runtime.ReadFeeder(token); !os.IsNotExist(err) {
			t.Fatalf("released feeder remained: %v", err)
		}
	})
}

func TestDemonAgentHelpShowsStandaloneAndDdocsForms(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run(context.Background(), []string{"demon", "acquire", "--help"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "usage: demon acquire") || !strings.Contains(out.String(), "ddocs demon acquire") {
		t.Fatalf("missing command forms: %q", out.String())
	}
}
