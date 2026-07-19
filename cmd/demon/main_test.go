package main

import (
	"reflect"
	"testing"
)

func TestNormalizeDemonArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "empty", want: []string{"demon", "--help"}},
		{name: "help", args: []string{"--help"}, want: []string{"demon", "--help"}},
		{name: "short help", args: []string{"-h"}, want: []string{"demon", "--help"}},
		{name: "run", args: []string{"run", "--true"}, want: []string{"demon", "run", "--true"}},
		{name: "status", args: []string{"--status"}, want: []string{"demon", "--status"}},
		{name: "version", args: []string{"--version"}, want: []string{"--version"}},
		{name: "already qualified", args: []string{"demon", "run"}, want: []string{"demon", "run"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeDemonArgs(test.args); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("normalizeDemonArgs(%v) = %v, want %v", test.args, got, test.want)
			}
		})
	}
}
