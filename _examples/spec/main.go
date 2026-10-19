package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/DeprecatedLuar/gohelp-luar"
)

//go:embed help.md
var helpSpec string

func main() {
	if err := gohelp.RunSpec(os.Args[1:], helpSpec); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
