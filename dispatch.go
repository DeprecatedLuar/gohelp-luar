package gohelp

import (
	"fmt"
	"sort"
	"strings"
)

// Run routes help output based on args (pass os.Args[1:] at the call site).
//
// Routing:
//   - no args or "help"        → print root page, return nil
//   - "help <topic>"           → print named sub-page, return nil
//   - "help --all"             → print all pages sequentially, return nil
//   - "help <unknown>"         → return an error naming the topic, a
//     fuzzy suggestion when one is found, and every known topic (sorted)
//
// Note: passing os.Args instead of os.Args[1:] will route on the binary path
// as a topic name. This is a call-site concern, not defended against here.
const noTruncateFlag = "--no-truncate"

func Run(args []string, root *Page, pages ...*Page) error {
	isHelp := func(s string) bool { return s == "help" || s == "-h" || s == "--help" }

	noTruncate := false
	filtered := make([]string, 0, len(args))
	for _, a := range args {
		if a == noTruncateFlag {
			noTruncate = true
			continue
		}
		filtered = append(filtered, a)
	}
	args = filtered

	if len(args) == 0 || (len(args) == 1 && isHelp(args[0])) {
		printPage(root, root.binary, noTruncate, pages...)
		return nil
	}

	if !isHelp(args[0]) {
		printPage(root, root.binary, noTruncate, pages...)
		return nil
	}

	topic := args[1]

	if topic == "--all" {
		printPage(root, root.binary, noTruncate, pages...)
		for _, p := range pages {
			printPage(p, root.binary, noTruncate, pages...)
		}
		return nil
	}

	pageMap := make(map[string]*Page, len(pages))
	for _, p := range pages {
		pageMap[p.binary] = p
	}

	if p, ok := pageMap[topic]; ok {
		printPage(p, root.binary, noTruncate, pages...)
		return nil
	}

	topics := make([]string, 0, len(pageMap))
	for name := range pageMap {
		topics = append(topics, name)
	}
	sort.Strings(topics)

	if suggest := fuzzyMatch(topic, pageMap); suggest != "" {
		return fmt.Errorf("unknown topic %q — did you mean: %s? (topics: %s)", topic, suggest, strings.Join(topics, ", "))
	}
	return fmt.Errorf("unknown topic %q (topics: %s)", topic, strings.Join(topics, ", "))
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
