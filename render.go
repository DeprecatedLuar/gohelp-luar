package gohelp

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	bold      = "\033[1m"
	dim       = "\033[2m"
	purple    = "\033[35m"
	blue      = "\033[34m"
	blueAlt   = "\033[38;5;75m"
	reset     = "\033[0m"
)

const (
	defaultTermWidth = 80 // fallback when terminal size cannot be detected
	separatorMargin  = 4  // chars consumed by "──[" + "]" decorators
	alignPad         = 2  // extra spaces added after the longest command to form the description column
	minWrapWidth     = 20 // minimum description wrap width regardless of terminal size
)

var blues = [2]string{blue, blueAlt}

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return defaultTermWidth
	}
	return w
}

func separator(title string) string {
	width := termWidth() - separatorMargin
	if title == "" {
		return dim + strings.Repeat("─", width) + reset
	}
	prefix := "──["
	suffix := "]"
	displayWidth := 3 + len(title) + 1 // "──[" + title + "]"

	if displayWidth >= width {
		return dim + prefix + reset + blue + title + reset + dim + suffix + reset
	}
	return dim + prefix + reset + blue + title + reset + dim + suffix + strings.Repeat("─", width-displayWidth) + reset
}

func printTitle(title string) {
	fmt.Printf("%s%s%s:%s\n", bold, purple, title, reset)
}

// Print renders a page to stdout. If pages are provided, a topics block is appended.
// The topics footer uses p.binary as the root binary name. When printing a sub-page
// directly (outside of Run), call printPage with the root binary explicitly.
func Print(p *Page, pages ...*Page) {
	printPage(p, p.binary, pages...)
}

func printPage(p *Page, rootBinary string, pages ...*Page) {
	width := termWidth()

	fmt.Println()
	fmt.Println(separator(p.binary + " - " + p.description))
	fmt.Println()

	for _, el := range p.elements {
		switch el.kind {
		case kindText:
			fmt.Println()
			fmt.Println(el.pairs[0])
			fmt.Println()

		case kindSection:
			printSection(el.title, el.entries, width)
		}
	}

	if len(pages) > 0 {
		printTopics(rootBinary, pages)
	}
}

func printSection(title string, entries []Entry, width int) {
	alignAt := 0
	for _, e := range entries {
		l := ansiWidth("├ " + e.cmd)
		if l > alignAt {
			alignAt = l
		}
	}
	alignAt += alignPad

	printTitle(title)

	for i, e := range entries {
		last := i == len(entries)-1

		var branch, contIndent string
		if last {
			branch = dim + "╰ " + reset
			contIndent = strings.Repeat(" ", alignAt)
		} else {
			branch = dim + "├ " + reset
			contIndent = dim + "│" + reset + strings.Repeat(" ", alignAt-1)
		}

		entryBlue := blues[i%2]

		visibleCmdLen := ansiWidth("├ " + e.cmd) // ├ and ╰ are same width
		var firstPrefix string
		if visibleCmdLen < alignAt {
			firstPrefix = branch + e.cmd + strings.Repeat(" ", alignAt-visibleCmdLen)
		} else {
			fmt.Println(branch + e.cmd)
			firstPrefix = contIndent
		}

		printWrappedDesc(firstPrefix, e.desc, e.example, contIndent, entryBlue, alignAt, width)
	}
	fmt.Println()
}

const egPrefix = "  (e.g. "

func printWrappedDesc(prefix, desc, example, contIndent, color string, alignAt, width int) {
	wrapWidth := max(width-alignAt, minWrapWidth)

	full := desc
	if example != "" {
		full += egPrefix + example + ")"
	}
	if full == "" {
		fmt.Println(prefix)
		return
	}

	wrapped := ansiWordWrap(full, wrapWidth)
	lines := strings.Split(strings.TrimRight(wrapped, "\n"), "\n")

	inExample := false
	for i, line := range lines {
		var indent string
		if i == 0 {
			indent = prefix
		} else {
			indent = contIndent
		}

		if !inExample {
			if idx := strings.Index(line, egPrefix); idx != -1 {
				inExample = true
				fmt.Printf("%s%s%s%s%s%s%s\n", indent, color, line[:idx], reset, dim, line[idx:], reset)
			} else {
				fmt.Printf("%s%s%s%s\n", indent, color, line, reset)
			}
		} else {
			fmt.Printf("%s%s%s%s\n", indent, dim, line, reset)
		}
	}
}

func printTopics(binary string, pages []*Page) {
	alignAt := 0
	for _, p := range pages {
		l := ansiWidth("├ " + p.binary)
		if l > alignAt {
			alignAt = l
		}
	}
	alignAt += alignPad

	fmt.Println()
	fmt.Println(separator(""))
	fmt.Println()
	printTitle("Topics")

	for i, p := range pages {
		last := i == len(pages)-1
		var branch string
		if last {
			branch = dim + "╰ " + reset
		} else {
			branch = dim + "├ " + reset
		}

		mainPart := branch + p.binary
		visibleLen := ansiWidth("├ " + p.binary)
		if visibleLen < alignAt {
			mainPart += strings.Repeat(" ", alignAt-visibleLen)
		}
		line := mainPart + blue + p.description + reset
		fmt.Println(line)
	}

	fmt.Println()
	fmt.Printf("Run '%s help <topic>' for details.\n", binary)
	fmt.Println()
}

// ansiWidth returns the visible (printable) width of s, ignoring ANSI escape sequences.
// Rune width is assumed to be 1 for all printable characters (ASCII + Latin, no CJK).
func ansiWidth(s string) int {
	w := 0
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // past 'm'
			}
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		w++
		i += size
	}
	return w
}

// ansiWordWrap wraps s at space boundaries so each line's visible width stays ≤ limit.
// ANSI escape sequences are preserved but do not count toward width.
func ansiWordWrap(s string, limit int) string {
	if limit <= 0 {
		return s
	}
	var out strings.Builder
	lineW := 0
	pendingSpaces := 0

	i := 0
	n := len(s)
	for i < n {
		if s[i] == ' ' {
			pendingSpaces++
			i++
			continue
		}

		// collect word: non-space run (may include embedded ANSI codes)
		wordStart := i
		wordVis := 0
		j := i
		for j < n && s[j] != ' ' {
			if s[j] == '\033' && j+1 < n && s[j+1] == '[' {
				j += 2
				for j < n && s[j] != 'm' {
					j++
				}
				if j < n {
					j++
				}
				continue
			}
			_, size := utf8.DecodeRuneInString(s[j:])
			wordVis++
			j += size
		}
		word := s[wordStart:j]
		i = j

		if lineW == 0 {
			pendingSpaces = 0
			out.WriteString(word)
			lineW += wordVis
		} else if lineW+pendingSpaces+wordVis > limit {
			out.WriteByte('\n')
			lineW = 0
			pendingSpaces = 0
			out.WriteString(word)
			lineW += wordVis
		} else {
			for k := 0; k < pendingSpaces; k++ {
				out.WriteByte(' ')
			}
			lineW += pendingSpaces
			pendingSpaces = 0
			out.WriteString(word)
			lineW += wordVis
		}
	}
	// trailing spaces (rare but preserve them)
	for k := 0; k < pendingSpaces; k++ {
		out.WriteByte(' ')
	}
	return out.String()
}
