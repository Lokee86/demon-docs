package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNestedCommandHelpContract(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   []string
		absent []string
	}{
		{
			name: "precision source",
			args: []string{"codemap", "precision", "source", "--help"},
			want: []string{"usage: ddocs codemaps precision source", "--exclude-prefix PATH", "default current directory", "does not edit authored codemap sections"},
		},
		{
			name: "precision sample",
			args: []string{"codemap", "precision", "sample", "--help"},
			want: []string{"usage: ddocs codemaps precision sample", "--suggestions PATH", "default 150", "--repository TEXT", "--revision TEXT"},
		},
		{
			name: "precision evaluate",
			args: []string{"codemap", "precision", "evaluate", "--help"},
			want: []string{"usage: ddocs codemaps precision evaluate", "--benchmark PATH", "--suggestions PATH", "default text"},
		},
		{
			name:   "suggestions select",
			args:   []string{"suggestions", "select", "--help"},
			want:   []string{"usage: ddocs suggestions select", "displayed number or target path", "exactly one candidate", "hash-guarded repair"},
			absent: []string{"ddocs suggestions {declined,log,show,select,decline,reconsider}"},
		},
		{
			name: "suggestions decline",
			args: []string{"suggestions", "decline", "--help"},
			want: []string{"usage: ddocs suggestions decline", "--reason TEXT", "evidence fingerprint"},
		},
		{
			name: "suggestions reconsider",
			args: []string{"suggestions", "reconsider", "--help"},
			want: []string{"usage: ddocs suggestions reconsider", "current or historical suggestion"},
		},
		{
			name:   "changes undo",
			args:   []string{"changes", "undo", "--help"},
			want:   []string{"usage: ddocs changes undo", "--repair ID", "--block", "recorded after hash", "undo limits"},
			absent: []string{"ddocs changes {related,show,log,undo,undo-run,block,unblock}"},
		},
		{
			name: "changes undo run",
			args: []string{"changes", "undo-run", "--help"},
			want: []string{"usage: ddocs changes undo-run", "preflight", "No partial run undo"},
		},
		{
			name: "changes block",
			args: []string{"changes", "block", "--help"},
			want: []string{"usage: ddocs changes block", "--repair ID", "relationship fingerprint"},
		},
		{
			name: "demon acquire",
			args: []string{"demon", "acquire", "--help"},
			want: []string{"usage: demon acquire", "--client NAME", "refresh the returned token", "linked worktree"},
		},
		{
			name: "demon heartbeat",
			args: []string{"demon", "heartbeat", "--help"},
			want: []string{"usage: demon heartbeat", "--token TOKEN", "owner lease", "same initialized repository"},
		},
		{
			name: "demon release",
			args: []string{"demon", "release", "--help"},
			want: []string{"usage: demon release", "--token TOKEN", "cancellation", "grace period"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(context.Background(), test.args, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%q", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("help wrote stderr: %q", stderr.String())
			}
			for _, text := range test.want {
				if !strings.Contains(stdout.String(), text) {
					t.Errorf("help missing %q:\n%s", text, stdout.String())
				}
			}
			for _, text := range test.absent {
				if strings.Contains(stdout.String(), text) {
					t.Errorf("help unexpectedly contained parent text %q:\n%s", text, stdout.String())
				}
			}
		})
	}
}

func TestEveryReviewSubcommandHasScopedHelp(t *testing.T) {
	commands := map[string][]string{
		"suggestions declined":   {"suggestions", "declined", "--help"},
		"suggestions log":        {"suggestions", "log", "--help"},
		"suggestions show":       {"suggestions", "show", "--help"},
		"suggestions select":     {"suggestions", "select", "--help"},
		"suggestions decline":    {"suggestions", "decline", "--help"},
		"suggestions reconsider": {"suggestions", "reconsider", "--help"},
		"changes related":        {"changes", "related", "--help"},
		"changes show":           {"changes", "show", "--help"},
		"changes log":            {"changes", "log", "--help"},
		"changes undo":           {"changes", "undo", "--help"},
		"changes undo-run":       {"changes", "undo-run", "--help"},
		"changes block":          {"changes", "block", "--help"},
		"changes unblock":        {"changes", "unblock", "--help"},
	}
	for name, args := range commands {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(context.Background(), args, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), "usage: ddocs "+name) {
				t.Fatalf("scoped usage missing:\n%s", stdout.String())
			}
			if !strings.Contains(stdout.String(), "-h, --help") {
				t.Fatalf("help option missing:\n%s", stdout.String())
			}
		})
	}
}
