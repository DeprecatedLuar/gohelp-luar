package gohelp

type elementKind int

const (
	kindText    elementKind = iota
	kindSection
)

type Entry struct {
	cmd     string
	desc    string
	example string
}

// Cmd creates a section entry with a command and description.
func Cmd(cmd, desc string) Entry {
	return Entry{cmd: cmd, desc: desc}
}

// Example adds a dim example line rendered below the description.
func (e Entry) Example(s string) Entry {
	e.example = s
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

// Usage appends a "Usage:" section with a single line of content.
func (p *Page) Usage(usage string) *Page {
	return p.Section("Usage", Cmd(usage, ""))
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

