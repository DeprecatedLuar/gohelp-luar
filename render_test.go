package gohelp

import (
	"strings"
	"testing"
)

func TestSectionLinesRender(t *testing.T) {
	restoreColors(t)
	Colors = Theme{}
	p := NewPage("luxos", "").Section("module",
		Usage("luxos module <verb> <name>"),
		Text("Manage modules."),
		Item("list", "List modules"),
		Item("add", "Create a module"),
	)
	got := capture(t, func() { Print(p) })
	want := strings.Join([]string{
		"",
		separator("luxos"),
		"",
		separator("module"),
		"",
		"  luxos module <verb> <name>",
		"",
		"  Manage modules.",
		"",
		"  list  List modules",
		"  add   Create a module",
		"",
		"",
	}, "\n")
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSectionLinesAlternateItemsOnly(t *testing.T) {
	restoreColors(t)
	p := NewPage("x", "").Section("S", Text("intro"), Item("a", "first"), Item("b", "second"))
	got := capture(t, func() { Print(p) })
	if !strings.Contains(got, paint(Colors.Desc, "first")) || !strings.Contains(got, paint(Colors.DescAlt, "second")) {
		t.Errorf("alternation should count items only:\n%q", got)
	}
}
