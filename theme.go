package gohelp

import (
	"fmt"
	"strconv"
	"strings"
)

// ANSI sequences used as defaults and exposed as the built-in palette.
const (
	reset = "\033[0m"

	Dim     = "\033[2m"
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

// blueAlt is the second shade used for alternating descriptions.
const blueAlt = "\033[38;5;75m"

const (
	hexPrefix = "#"
	hexDigits = 6
	truecolor = "\033[38;2;%d;%d;%dm"
)

// Theme holds one color per visual element. Every field is a raw ANSI sequence,
// so the built-in palette constants, Hex, and a hand-written escape are all
// valid values. An empty field emits no escape at all: base colors fall through
// to the terminal's own foreground, and an empty Syntax disables highlighting.
type Theme struct {
	Header  string // ──[Title]──── bars and the plain divider
	Command string // command column, topic names
	Desc    string // descriptions
	DescAlt string // alternating description shade; empty falls back to Desc
	Text    string // page description, paragraphs, usage line
	Example string // (e.g. ...) spans
	Syntax  string // <arg>, [opt] and --flag tokens, layered over Command/Text
}

// Colors is the theme used by every page. Set its fields before rendering.
var Colors = Theme{
	Desc:    Blue,
	DescAlt: blueAlt,
	Example: Dim,
	Syntax:  Dim,
}

// Hex converts "#rrggbb" (or "rrggbb") to a 24-bit foreground sequence.
// It panics on malformed input: callers pass a literal at startup, so a bad
// value is a programming error, not something to paper over at render time.
func Hex(s string) string {
	seq, err := parseHex(s)
	if err != nil {
		panic("gohelp: " + err.Error())
	}
	return seq
}

// parseHex is the error-returning form of Hex.
func parseHex(s string) (string, error) {
	h := strings.TrimPrefix(s, hexPrefix)
	if len(h) != hexDigits {
		return "", fmt.Errorf("invalid hex color %q: want %s followed by %d hex digits", s, hexPrefix, hexDigits)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return "", fmt.Errorf("invalid hex color %q: %v", s, err)
	}
	return fmt.Sprintf(truecolor, v>>16&0xff, v>>8&0xff, v&0xff), nil
}

// descColor returns the description color for entry index i, alternating with
// DescAlt only when one is set.
func descColor(i int) string {
	if i%2 == 1 && Colors.DescAlt != "" {
		return Colors.DescAlt
	}
	return Colors.Desc
}

// paint wraps s in color, emitting nothing when color is empty.
func paint(color, s string) string {
	if color == "" {
		return s
	}
	return color + s + reset
}
