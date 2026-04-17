package gohelp

type elementKind int

const (
	kindText    elementKind = iota
	kindSection
	kindUsage
)

type Entry struct {
	cmd     string
	desc    string
	example string
}

// Item creates a section entry. The optional third argument is a dim example line.
func Item(cmd, desc string, example ...string) Entry {
	e := Entry{cmd: cmd, desc: desc}
	if len(example) > 0 {
		e.example = example[0]
	}
	return e
}

type element struct {
	kind    elementKind
	title   string
	pairs   []string // kindText only
	entries []Entry  // kindSection only
}

// Page holds the content and metadata for a help page.
type Page struct {
	binary      string
	description string
	elements    []element
}

// NewPage creates a new help page for the given binary and one-line description.
func NewPage(binary, description string) *Page {
	return &Page{binary: binary, description: description}
}

// Usage appends an indented usage line (no section bar).
func (p *Page) Usage(usage string) *Page {
	p.elements = append(p.elements, element{kind: kindUsage, pairs: []string{usage}})
	return p
}

// Text appends a plain paragraph.
func (p *Page) Text(s string) *Page {
	p.elements = append(p.elements, element{kind: kindText, pairs: []string{s}})
	return p
}

// Section appends a labeled command/description section.
func (p *Page) Section(title string, entries ...Entry) *Page {
	p.elements = append(p.elements, element{kind: kindSection, title: title, entries: entries})
	return p
}

