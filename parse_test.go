package gohelp

import (
	"reflect"
	"strings"
	"testing"
)

const basicSpec = `# deploy | zero-downtime deployment tool
$ deploy history [list|deleted|<id>]
> First line
> second line.

> Separate paragraph.
## Commands
- up, u | Deploy it | deploy up --env staging
- view | View raw | deploy view abc --raw | pbcopy
- down | Tear down
## Empty
# auth | manage credentials
`

func kinds(p *Page) []elementKind {
	var out []elementKind
	for _, e := range p.elements {
		out = append(out, e.kind)
	}
	return out
}

func mustParse(t *testing.T, in string) *Spec {
	t.Helper()
	s, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return s
}

func TestParseStructure(t *testing.T) {
	s := mustParse(t, basicSpec)
	if s.Root.binary != "deploy" || s.Root.description != "zero-downtime deployment tool" {
		t.Fatalf("root = %q / %q", s.Root.binary, s.Root.description)
	}
	if len(s.Pages) != 1 || s.Pages[0].binary != "auth" || s.Pages[0].description != "manage credentials" {
		t.Fatalf("pages = %+v", s.Pages)
	}
	wantKinds := []elementKind{kindUsage, kindText, kindText, kindSection, kindSection}
	if got := kinds(s.Root); !reflect.DeepEqual(got, wantKinds) {
		t.Fatalf("kinds = %v, want %v", got, wantKinds)
	}
	el := s.Root.elements
	if el[0].pairs[0] != "deploy history [list|deleted|<id>]" {
		t.Errorf("usage = %q", el[0].pairs[0])
	}
	if el[1].pairs[0] != "First line second line." {
		t.Errorf("text 1 = %q", el[1].pairs[0])
	}
	if el[2].pairs[0] != "Separate paragraph." {
		t.Errorf("text 2 = %q", el[2].pairs[0])
	}
	wantEntries := []Entry{
		{cmd: "up, u", desc: "Deploy it", example: "deploy up --env staging"},
		{cmd: "view", desc: "View raw", example: "deploy view abc --raw | pbcopy"},
		{cmd: "down", desc: "Tear down"},
	}
	if !reflect.DeepEqual(el[3].entries, wantEntries) {
		t.Errorf("entries = %+v", el[3].entries)
	}
	if len(el[4].entries) != 0 {
		t.Errorf("empty section has %d entries", len(el[4].entries))
	}
	if s.Theme != Colors {
		t.Errorf("theme = %+v, want Colors", s.Theme)
	}
}

func TestParseOrder(t *testing.T) {
	s := mustParse(t, "# x\n$ x run\n> a\n## S\n- c | d\n> b\n")
	want := []elementKind{kindUsage, kindText, kindSection}
	if got := kinds(s.Root); !reflect.DeepEqual(got, want) {
		t.Fatalf("kinds = %v, want %v", got, want)
	}
}

func TestParseSectionScope(t *testing.T) {
	s := mustParse(t, "# x\n$ x run\n## S\n$ x s <a>\n> one\n> two\n- c | d\n> after\n## T\n- e | f\n")
	want := []elementKind{kindUsage, kindSection, kindSection}
	if got := kinds(s.Root); !reflect.DeepEqual(got, want) {
		t.Fatalf("kinds = %v, want %v", got, want)
	}
	wantS := []Entry{Usage("x s <a>"), Text("one two"), Item("c", "d"), Text("after")}
	if got := s.Root.elements[1].entries; !reflect.DeepEqual(got, wantS) {
		t.Errorf("S entries = %+v, want %+v", got, wantS)
	}
	wantT := []Entry{Item("e", "f")}
	if got := s.Root.elements[2].entries; !reflect.DeepEqual(got, wantT) {
		t.Errorf("T entries = %+v, want %+v", got, wantT)
	}
}

func TestParseIndentation(t *testing.T) {
	var b strings.Builder
	for _, l := range strings.Split(basicSpec, "\n") {
		b.WriteString("    \t" + l + "\n")
	}
	if got, want := mustParse(t, b.String()), mustParse(t, basicSpec); !reflect.DeepEqual(got, want) {
		t.Fatalf("indented parse differs")
	}
}

func TestParseCRLF(t *testing.T) {
	crlf := strings.ReplaceAll(basicSpec, "\n", "\r\n")
	if got, want := mustParse(t, crlf), mustParse(t, basicSpec); !reflect.DeepEqual(got, want) {
		t.Fatalf("CRLF parse differs")
	}
}

func TestParseBOM(t *testing.T) {
	if got, want := mustParse(t, "\ufeff"+basicSpec), mustParse(t, basicSpec); !reflect.DeepEqual(got, want) {
		t.Fatalf("BOM parse differs")
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct{ name, in, wantLine string }{
		{"no page", "\n\n", ""},
		{"text before page", "> hi\n# x\n", "line 1:"},
		{"section before", "## S\n# x\n", "line 1:"},
		{"empty name", "# | desc\n", "line 1:"},
		{"item no section", "# x\n- a | b\n", "line 2:"},
		{"one-field item", "# x\n## S\n- a\n", "line 3:"},
		{"empty cmd", "# x\n## S\n- | b\n", "line 3:"},
		{"empty desc", "# x\n## S\n- a |\n", "line 3:"},
		{"empty section", "# x\n##\n", "line 2:"},
		{"empty usage", "# x\n$\n", "line 2:"},
		{"unknown marker", "# x\n* foo\n", "line 2:"},
		{"glued marker", "# x\n##Items\n", "line 2:"},
		{"duplicate page", "# x\n# a\n# a\n", "line 3:"},
		{"partial", "# x\n## S\n- a | b\n# y\n$ y run\n# z\n* bad\n", "line 7:"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := Parse(c.in)
			if s != nil || err == nil {
				t.Fatalf("got (%v, %v), want nil spec and error", s, err)
			}
			if !strings.HasPrefix(err.Error(), "gohelp: ") {
				t.Errorf("error %q lacks gohelp: prefix", err)
			}
			if !strings.Contains(err.Error(), c.wantLine) {
				t.Errorf("error %q lacks %q", err, c.wantLine)
			}
		})
	}
}

func TestParseTheme(t *testing.T) {
	saved := Colors
	defer func() { Colors = saved }()

	with := func(mut func(*Theme)) Theme {
		th := Colors
		mut(&th)
		return th
	}
	cases := []struct {
		name, in string
		want     Theme
	}{
		{"default", "# x\n", Colors},
		{"hex", "# x\n@ desc=#5fafff\n", with(func(th *Theme) { th.Desc = Hex("#5fafff") })},
		{"empty", "# x\n@ descalt=\n", with(func(th *Theme) { th.DescAlt = "" })},
		{"palette", "# x\n@ example=dim\n", with(func(th *Theme) { th.Example = Dim })},
		{"palette case", "# x\n@ example=DIM\n", with(func(th *Theme) { th.Example = Dim })},
		{"before page", "@ header=cyan\n# x\n", with(func(th *Theme) { th.Header = Cyan })},
		{"one line", "# x\n@ header=red command=green\n", with(func(th *Theme) { th.Header, th.Command = Red, Green })},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := mustParse(t, c.in).Theme; got != c.want {
				t.Errorf("theme = %+v, want %+v", got, c.want)
			}
		})
	}
	if Colors != saved {
		t.Errorf("Parse mutated Colors")
	}
}

func TestParseThemeErrors(t *testing.T) {
	cases := []struct{ name, in, wantLine, wantText string }{
		{"theme no =", "# x\n@ desc\n", "line 2:", ""},
		{"unknown theme key", "# x\n@ commands=red\n", "line 2:", "valid: header, command"},
		{"duplicate key", "# x\n@ desc=red\n@ desc=blue\n", "line 3:", "line 2"},
		{"duplicate key same line", "# x\n@ desc=red desc=blue\n", "line 2:", ""},
		{"bad hex", "# x\n@ desc=#0000000\n", "line 2:", ""},
		{"unknown palette", "# x\n@ desc=orange\n", "line 2:", "palette: dim, black"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := Parse(c.in)
			if s != nil || err == nil {
				t.Fatalf("got (%v, %v), want nil spec and error", s, err)
			}
			msg := err.Error()
			if !strings.HasPrefix(msg, "gohelp: ") || strings.Count(msg, "gohelp:") != 1 ||
				!strings.Contains(msg, c.wantLine) || !strings.Contains(msg, c.wantText) {
				t.Errorf("error %q does not match %q / %q", msg, c.wantLine, c.wantText)
			}
		})
	}
}

func TestThemeKeysMatchTheme(t *testing.T) {
	keys := strings.Split(validThemeKeys, ", ")
	if n := reflect.TypeOf(Theme{}).NumField(); len(keys) != n {
		t.Fatalf("validThemeKeys lists %d keys, Theme has %d fields", len(keys), n)
	}
	var th Theme
	seen := map[*string]bool{}
	for _, k := range keys {
		dst := themeField(&th, k)
		if dst == nil || seen[dst] {
			t.Errorf("key %q does not resolve to a distinct Theme field", k)
		}
		seen[dst] = true
	}
}
