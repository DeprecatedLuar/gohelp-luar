package gohelp

import (
	"bytes"
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

func TestPageTitle(t *testing.T) {
	restoreColors(t)
	Colors = Theme{}
	bare := capture(t, func() { Print(NewPage("x", "")) })
	if want := "\n" + separator("x") + "\n\n"; bare != want {
		t.Errorf("empty description: got %q, want %q", bare, want)
	}
	described := capture(t, func() { Print(NewPage("x", "does y")) })
	if want := "\n" + separator("x") + "\n\n  does y\n\n"; described != want {
		t.Errorf("with description: got %q, want %q", described, want)
	}
	if !strings.HasPrefix(separator("x"), "──[x]─") {
		t.Errorf("bar = %q", separator("x"))
	}
}

func TestFprintSection(t *testing.T) {
	restoreColors(t)
	p := NewPage("x", "d").
		Usage("x <cmd>").
		Section("A", Item("a", "first")).
		Section("B", Usage("x b <n>"), Item("bb", "second", "x b 1"), Item("bc", "third")).
		Text("tail")
	full := capture(t, func() { Print(p) })
	start := strings.Index(full, separator("B"))
	end := strings.Index(full, "\n\n  "+paint(Colors.Text, "tail"))
	if start < 0 || end < 0 {
		t.Fatalf("cannot locate section B in:\n%q", full)
	}
	want := full[start:end] + "\n"
	var buf bytes.Buffer
	if err := p.FprintSection(&buf, "B"); err != nil {
		t.Fatal(err)
	}
	if buf.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestFprintSectionMissing(t *testing.T) {
	p := NewPage("x", "").Section("A", Item("a", "b"))
	var buf bytes.Buffer
	err := p.FprintSection(&buf, "nope")
	if err == nil || !strings.HasPrefix(err.Error(), "gohelp: ") || !strings.Contains(err.Error(), `"nope"`) {
		t.Fatalf("err = %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("wrote %q on error", buf.String())
	}
}
