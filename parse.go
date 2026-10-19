package gohelp

import (
	"errors"
	"fmt"
	"strings"
)

const (
	bom = "\ufeff"

	markerPage    = "#"
	markerSection = "##"
	markerItem    = "-"
	markerProse   = ">"
	markerUsage   = "$"

	fieldSep   = "|"
	itemFields = 3
	proseJoin  = " "
	blanks     = " \t"

	errPrefix = "gohelp: "
	errNoPage = errPrefix + "no page: a help file needs at least one '# name' line"
)

// Spec is a parsed help file.
type Spec struct {
	Root  *Page   // first # page
	Pages []*Page // every later # page, in document order
	Theme Theme   // Colors with any @ assignments applied
}

// parser holds the state of a single pass over a help file.
type parser struct {
	pages   []*Page
	names   map[string]bool
	title   string
	entries []Entry
	open    bool
	prose   []string
}

// Parse builds pages from help-file text using only the public builder.
func Parse(spec string) (*Spec, error) {
	p := &parser{names: map[string]bool{}}
	lines := strings.Split(strings.TrimPrefix(spec, bom), "\n")
	for i, raw := range lines {
		if err := p.line(strings.TrimSpace(raw)); err != nil {
			return nil, fmt.Errorf("%sline %d: %w", errPrefix, i+1, err)
		}
	}
	if len(p.pages) == 0 {
		return nil, errors.New(errNoPage)
	}
	p.flushProse()
	p.flushSection()
	return &Spec{Root: p.pages[0], Pages: p.pages[1:], Theme: Colors}, nil
}

func splitMarker(line string) (marker, rest string) {
	i := strings.IndexAny(line, blanks)
	if i < 0 {
		return line, ""
	}
	return line[:i], strings.TrimSpace(line[i:])
}

func (p *parser) line(line string) error {
	if line == "" {
		p.flushProse()
		return nil
	}
	marker, rest := splitMarker(line)
	if marker != markerProse {
		p.flushProse()
	}
	switch marker {
	case markerPage:
		return p.page(rest)
	case markerSection:
		return p.section(rest)
	case markerItem:
		return p.item(rest)
	case markerProse:
		return p.proseLine(rest)
	case markerUsage:
		return p.usage(rest)
	}
	return fmt.Errorf("unknown marker %q", marker)
}

func (p *parser) current() *Page {
	return p.pages[len(p.pages)-1]
}

func (p *parser) requirePage(marker string) error {
	if len(p.pages) == 0 {
		return fmt.Errorf("%q before the first '# name' line", marker)
	}
	return nil
}

func (p *parser) flushProse() {
	if len(p.prose) == 0 {
		return
	}
	p.current().Text(strings.Join(p.prose, proseJoin))
	p.prose = nil
}

func (p *parser) flushSection() {
	if !p.open {
		return
	}
	p.current().Section(p.title, p.entries...)
	p.title, p.entries, p.open = "", nil, false
}

func (p *parser) page(rest string) error {
	name, desc, _ := strings.Cut(rest, fieldSep)
	name, desc = strings.TrimSpace(name), strings.TrimSpace(desc)
	if name == "" {
		return fmt.Errorf("page needs a name")
	}
	if len(p.pages) > 0 {
		if p.names[name] {
			return fmt.Errorf("duplicate page %q", name)
		}
		p.names[name] = true
	}
	p.flushSection()
	p.pages = append(p.pages, NewPage(name, desc))
	return nil
}

func (p *parser) section(rest string) error {
	if err := p.requirePage(markerSection); err != nil {
		return err
	}
	if rest == "" {
		return fmt.Errorf("section needs a title")
	}
	p.flushSection()
	p.title, p.open = rest, true
	return nil
}

func (p *parser) item(rest string) error {
	if err := p.requirePage(markerItem); err != nil {
		return err
	}
	if !p.open {
		return fmt.Errorf("item outside a section: add a '## Title' line above it")
	}
	fields := strings.SplitN(rest, fieldSep, itemFields)
	if len(fields) < 2 {
		return fmt.Errorf("item needs 'command | description'")
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	if fields[0] == "" || fields[1] == "" {
		return fmt.Errorf("item needs a non-empty command and description")
	}
	p.entries = append(p.entries, Item(fields[0], fields[1], fields[2:]...))
	return nil
}

func (p *parser) proseLine(rest string) error {
	if err := p.requirePage(markerProse); err != nil {
		return err
	}
	p.flushSection()
	if rest != "" {
		p.prose = append(p.prose, rest)
	}
	return nil
}

func (p *parser) usage(rest string) error {
	if err := p.requirePage(markerUsage); err != nil {
		return err
	}
	if rest == "" {
		return fmt.Errorf("usage needs text")
	}
	p.flushSection()
	p.current().Usage(rest)
	return nil
}
