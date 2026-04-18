package gohelp

import (
	"fmt"
	"os"
)

// Run routes help output based on args (pass os.Args[1:] at the call site).
//
// Routing:
//   - no args or "help"        → print root page
//   - "help <topic>"           → print named sub-page
//   - "help --all"             → print all pages sequentially
//   - "help <unknown>"         → fuzzy-suggest, list topics, exit 1
//
// Note: passing os.Args instead of os.Args[1:] will route on the binary path
// as a topic name. This is a call-site concern, not defended against here.
func Run(args []string, root *Page, pages ...*Page) {
	isHelp := func(s string) bool { return s == "help" || s == "-h" || s == "--help" }

	if len(args) == 0 || (len(args) == 1 && isHelp(args[0])) {
		Print(root, pages...)
		return
	}

	if !isHelp(args[0]) {
		Print(root, pages...)
		return
	}

	topic := args[1]

	if topic == "--all" {
		Print(root, pages...)
		for _, p := range pages {
			printPage(p, root.binary, pages...)
		}
		return
	}

	pageMap := make(map[string]*Page, len(pages))
	for _, p := range pages {
		pageMap[p.binary] = p
	}

	if p, ok := pageMap[topic]; ok {
		printPage(p, root.binary, pages...)
		return
	}

	if suggest := fuzzyMatch(topic, pageMap); suggest != "" {
		fmt.Fprintf(os.Stderr, "unknown topic %q — did you mean: %s?\n\n", topic, suggest)
	} else {
		fmt.Fprintf(os.Stderr, "unknown topic %q\n\n", topic)
	}
	fmt.Fprintln(os.Stderr, "Available topics:")
	for name := range pageMap {
		fmt.Fprintf(os.Stderr, "  %s\n", name)
	}
	os.Exit(1)
}

func fuzzyMatch(input string, pages map[string]*Page) string {
	best, bestDist := "", 3 // threshold is <= 2; init at 3 so any match wins
	for name := range pages {
		if d := levenshtein(input, name); d < bestDist {
			bestDist = d
			best = name
		}
	}
	return best
}

func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	row := make([]int, lb+1)
	for j := range row {
		row[j] = j
	}
	for i := 1; i <= la; i++ {
		prev := row[0]
		row[0] = i
		for j := 1; j <= lb; j++ {
			tmp := row[j]
			if a[i-1] == b[j-1] {
				row[j] = prev
			} else {
				row[j] = 1 + min3(prev, row[j], row[j-1])
			}
			prev = tmp
		}
	}
	return row[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
