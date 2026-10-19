package gohelp

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	defaultTermWidth = 80 // fallback when terminal size cannot be detected
	separatorMargin  = 4  // chars consumed by "──[" + "]" decorators
	alignPad         = 2  // extra spaces added after the longest command to form the description column
	minWrapWidth     = 20 // minimum description wrap width regardless of terminal size
	minDescWidth     = 23 // minimum columns guaranteed to the description before the command column gives way
	ellipsis         = "…"
)

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return defaultTermWidth
	}
	return w
}

// maxAlignAt caps the description column so it can never eat more than
// width-minDescWidth columns, guaranteeing the description always keeps at
// least minDescWidth columns to wrap in.
func maxAlignAt(width int) int {
	return max(width-minDescWidth, minWrapWidth)
}

func separator(title string) string {
	width := termWidth() - separatorMargin
	if title == "" {
		return paint(Colors.Header, strings.Repeat("─", width))
	}
	prefix := "──["
	suffix := "]"
	displayWidth := 3 + len(title) + 1 // "──[" + title + "]"

	if displayWidth >= width {
		return paint(Colors.Header, prefix+title+suffix)
	}
	return paint(Colors.Header, prefix+title+suffix+strings.Repeat("─", width-displayWidth))
}

// Print renders a page to stdout. If pages are provided, a topics block is appended.
// The topics footer uses p.binary as the root binary name. When printing a sub-page
// directly (outside of Run), call printPage with the root binary explicitly.
func Print(p *Page, pages ...*Page) {
	printPage(p, p.binary, false, pages...)
}

func printPage(p *Page, rootBinary string, noTruncate bool, pages ...*Page) {
	width := termWidth()

	fmt.Println()
	fmt.Println(separator(p.binary))
	fmt.Println()

	if p.description != "" {
		fmt.Println("  " + paint(Colors.Text, p.description))
		fmt.Println()
	}

	for _, el := range p.elements {
		switch el.kind {
		case kindUsage:
			fmt.Println("  " + highlightSyntax(el.pairs[0], Colors.Text))
			fmt.Println()

		case kindText:
			fmt.Println()
			fmt.Println("  " + paint(Colors.Text, el.pairs[0]))
			fmt.Println()

		case kindSection:
			printSection(el.title, el.entries, width, noTruncate)
		}
	}

	if len(pages) > 0 {
		printTopics(rootBinary, pages)
	}
}

func printSection(title string, entries []Entry, width int, noTruncate bool) {
	alignAt := 0
	for _, e := range entries {
		l := ansiWidth("  " + e.cmd)
		if l > alignAt {
			alignAt = l
		}
	}
	alignAt += alignPad
	alignAt = min(alignAt, maxAlignAt(width))

	fmt.Println(separator(title))
	fmt.Println()

	contIndent := strings.Repeat(" ", alignAt)
	for i, e := range entries {
		entryColor := descColor(i)
		cmdLine := "  " + e.cmd
		visibleCmdLen := ansiWidth(cmdLine)

		var firstPrefix string
		overflowed := false
		switch {
		case visibleCmdLen < alignAt:
			firstPrefix = highlightSyntax(cmdLine, Colors.Command) + strings.Repeat(" ", alignAt-visibleCmdLen)
		case noTruncate:
			fmt.Println(highlightSyntax(cmdLine, Colors.Command))
			firstPrefix = contIndent
			overflowed = true
		default:
			keep := max(alignAt-alignPad-1, 0)
			truncated, _, _ := takeVisible(cmdLine, keep)
			firstPrefix = highlightSyntax(truncated, Colors.Command) + ellipsis + strings.Repeat(" ", alignPad)
		}

		printWrappedDesc(firstPrefix, e.desc, e.example, contIndent, entryColor, alignAt, width)

		// an overflowed entry spans multiple lines; a blank line keeps it
		// from running into the next entry
		if overflowed && i != len(entries)-1 {
			fmt.Println()
		}
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
			if idx := strings.Index(line, "(e.g."); idx != -1 {
				inExample = true
				fmt.Printf("%s%s%s\n", indent, paint(color, line[:idx]), paint(Colors.Example, line[idx:]))
			} else {
				fmt.Printf("%s%s\n", indent, paint(color, line))
			}
		} else {
			fmt.Printf("%s%s\n", indent, paint(Colors.Example, line))
		}
	}
}

func printTopics(binary string, pages []*Page) {
	alignAt := 0
	for _, p := range pages {
		l := ansiWidth("  " + p.binary)
		if l > alignAt {
			alignAt = l
		}
	}
	alignAt += alignPad

	fmt.Println(separator(""))
	fmt.Println()
	fmt.Println("Topics:")
	fmt.Println()

	for _, p := range pages {
		mainPart := paint(Colors.Command, "  "+p.binary)
		visibleLen := ansiWidth("  " + p.binary)
		if visibleLen < alignAt {
			mainPart += strings.Repeat(" ", alignAt-visibleLen)
		}
		fmt.Println(mainPart + paint(Colors.Desc, p.description))
	}

	fmt.Println()
	fmt.Printf("Run '%s help <topic>' for details.\n", binary)
	fmt.Println()
}

// highlightSyntax colors CLI notation tokens (<arg>, [opt], and -x/--flag runs)
// in Colors.Syntax, leaving the rest of s in base. Existing ANSI escapes in s
// pass through untouched. Returns s unchanged when Colors.Syntax is empty.
func highlightSyntax(s, base string) string {
	if Colors.Syntax == "" {
		return s
	}

	var out strings.Builder
	i, n := 0, len(s)
	atWordStart := true

	flush := func(text string) {
		if text != "" {
			out.WriteString(paint(base, text))
		}
	}

	plainStart := 0
	for i < n {
		if s[i] == '\033' && i+1 < n && s[i+1] == '[' {
			// pass an existing escape through untouched, as part of the plain run
			j := i + 2
			for j < n && s[j] != 'm' {
				j++
			}
			if j < n {
				j++
			}
			i = j
			continue
		}

		switch {
		case s[i] == '<' || s[i] == '[':
			close := byte('>')
			if s[i] == '[' {
				close = ']'
			}
			end := strings.IndexByte(s[i:], close)
			if end == -1 {
				i++
				atWordStart = false
				continue
			}
			flush(s[plainStart:i])
			end += i + 1
			out.WriteString(paint(Colors.Syntax, s[i:end]))
			i = end
			plainStart = i
			atWordStart = false

		case s[i] == '-' && atWordStart:
			end := i
			for end < n && s[end] != ' ' {
				end++
			}
			flush(s[plainStart:i])
			out.WriteString(paint(Colors.Syntax, s[i:end]))
			i = end
			plainStart = i
			atWordStart = false

		case s[i] == ' ':
			i++
			atWordStart = true

		default:
			i++
			atWordStart = false
		}
	}
	flush(s[plainStart:])
	return out.String()
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

// takeVisible returns the first n visible characters of s as chunk, the remainder as rest,
// and the actual visible width of chunk. ANSI escapes pass through without counting.
func takeVisible(s string, n int) (chunk, rest string, vis int) {
	i, w := 0, 0
	for i < len(s) && w < n {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		w++
		i += size
	}
	return s[:i], s[i:], w
}

// writeWord writes word to out, hard-breaking at limit if the word alone exceeds it.
func writeWord(out *strings.Builder, word string, wordVis int, lineW *int, limit int) {
	if wordVis <= limit {
		out.WriteString(word)
		*lineW += wordVis
		return
	}
	// word is longer than limit: hard-break it
	rem := word
	for rem != "" {
		chunk, rest, cv := takeVisible(rem, limit)
		out.WriteString(chunk)
		*lineW = cv
		rem = rest
		if rem != "" {
			out.WriteByte('\n')
			*lineW = 0
		}
	}
}

// ansiWordWrap wraps s at space boundaries so each line's visible width stays ≤ limit.
// ANSI escape sequences are preserved but do not count toward width.
// Words longer than limit are hard-broken to prevent terminal-controlled overflow.
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
			writeWord(&out, word, wordVis, &lineW, limit)
		} else if lineW+pendingSpaces+wordVis > limit {
			out.WriteByte('\n')
			lineW = 0
			pendingSpaces = 0
			writeWord(&out, word, wordVis, &lineW, limit)
		} else {
			for k := 0; k < pendingSpaces; k++ {
				out.WriteByte(' ')
			}
			lineW += pendingSpaces
			pendingSpaces = 0
			writeWord(&out, word, wordVis, &lineW, limit)
		}
	}
	// trailing spaces (rare but preserve them)
	for k := 0; k < pendingSpaces; k++ {
		out.WriteByte(' ')
	}
	return out.String()
}
