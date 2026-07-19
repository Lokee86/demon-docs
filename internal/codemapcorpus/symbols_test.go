package codemapcorpus

import (
	"reflect"
	"testing"
)

func TestGoDeclaredSymbolsIncludesTypesFunctionsAndQualifiedMethods(t *testing.T) {
	got := goDeclaredSymbols("runtime.go", []byte(`package playerdata

type Runtime struct{}
func NewRuntime() *Runtime { return nil }
func (r *Runtime) LoadStats() {}
`))
	want := []string{"NewRuntime", "Runtime.LoadStats"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols = %v, want %v", got, want)
	}
}

func TestGDScriptDeclaredSymbolsIncludesClassAndQualifiedMethods(t *testing.T) {
	got := gdscriptDeclaredSymbols([]byte(`class_name ToolingPacketRouter
extends RefCounted

func dispatch(packet):
	pass
`))
	want := []string{"ToolingPacketRouter", "ToolingPacketRouter.dispatch"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols = %v, want %v", got, want)
	}
}
