package gohelp

import (
	"fmt"
	"os"
	"strings"

	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
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

var blues = [2]string{blue, blueAlt}

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

func separator(title string) string {
	width := termWidth() - 4
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
		printTopics(rootBinary, pages, width)
	}
}

func printSection(title string, entries []Entry, width int) {
	alignAt := 0
	for _, e := range entries {
		l := ansi.PrintableRuneWidth("├ " + e.cmd)
		if l > alignAt {
			alignAt = l
		}
	}
	alignAt += 2

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

		visibleCmdLen := ansi.PrintableRuneWidth("├ " + e.cmd) // ├ and ╰ are same width
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
	wrapWidth := max(width-alignAt, 20)

	full := desc
	if example != "" {
		full += egPrefix + example + ")"
	}
	if full == "" {
		fmt.Println(prefix)
		return
	}

	w := wordwrap.NewWriter(wrapWidth)
	w.Breakpoints = []rune{}
	fmt.Fprint(w, full)
	w.Close()
	wrapped := w.String()
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

func printTopics(binary string, pages []*Page, width int) {
	alignAt := 0
	for _, p := range pages {
		l := ansi.PrintableRuneWidth("├ " + p.binary)
		if l > alignAt {
			alignAt = l
		}
	}
	alignAt += 2

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
		visibleLen := ansi.PrintableRuneWidth("├ " + p.binary)
		if visibleLen < alignAt {
			mainPart += strings.Repeat(" ", alignAt-visibleLen)
		}
		line := mainPart + blue + p.description + reset
		fmt.Println(truncate.StringWithTail(line, uint(width), ">"))
	}

	fmt.Println()
	fmt.Printf("Run '%s help <topic>' for details.\n", binary)
	fmt.Println()
}
