package codemapcorpus

import (
	"reflect"
	"testing"
)

func TestDeclaredSymbolsExcludeGenericAndUnexportedNames(t *testing.T) {
	goSymbols := goDeclaredSymbols("generic.go", []byte(`package sample

type Game struct{}
type PlayerDataStats struct{}
func project() {}
func (g *Game) emit() {}
func (g *Game) HandlePacket() {}
`))
	if want := []string{"Game.HandlePacket", "PlayerDataStats"}; !reflect.DeepEqual(goSymbols, want) {
		t.Fatalf("Go symbols = %v, want %v", goSymbols, want)
	}

	gdSymbols := gdscriptDeclaredSymbols([]byte(`class_name Player
extends Node
func update():
	pass
`))
	if len(gdSymbols) != 0 {
		t.Fatalf("generic GDScript class leaked into symbols: %v", gdSymbols)
	}
}
