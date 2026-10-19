# gohelp

Go library for rendering formatted CLI help screens. Fluent builder API, automatic terminal-width wrapping, ANSI colors, fuzzy topic matching.

**One dependency:** [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) for cross-platform terminal width detection.

## Install

```bash
go get github.com/DeprecatedLuar/gohelp-luar
```

## Usage

Pages can also be built in Go:

```go
import (
    "fmt"
    "os"
    gohelp "github.com/DeprecatedLuar/gohelp-luar"
)

root := gohelp.NewPage("mytool", "does something useful").
    Usage("mytool <command> [flags]").
    Section("Commands",
        gohelp.Item("start", "Start the service", "mytool start --env prod"),
        gohelp.Item("stop", "Gracefully stop the service"),
    ).
    Section("Flags",
        gohelp.Item("--verbose", "Enable debug output"),
        gohelp.Item("--help", "Show this help message"),
    ).
    Text("All commands support --help for detailed usage.")

config := gohelp.NewPage("config", "manage configuration").
    Usage("mytool config <command>").
    Section("Commands",
        gohelp.Item("show", "Print current config"),
        gohelp.Item("edit", "Open config in $EDITOR", "mytool config edit"),
    )

if err := gohelp.Run(os.Args[1:], root, config); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}
```

## Help files

The preferred way to define help is a markdown-flavoured file, rendered with one call:

```go
import (
    _ "embed"
    "fmt"
    "os"
    gohelp "github.com/DeprecatedLuar/gohelp-luar"
)

//go:embed help.md
var helpSpec string

func main() {
    if err := gohelp.RunSpec(os.Args[1:], helpSpec); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

A Go string constant works the same way. `//go:embed` only reaches files in the package's own
directory or below. Name the file `help.md`, or `<name>.help.md` when a tool has more than one.

Each line starts with a marker, followed by a space:

| Marker | Meaning |
|--------|---------|
| `#`  | page header, `name \| description`. The first is the root page; each later one is a topic |
| `##` | section title |
| `-`  | item, `command \| description \| example` (example is optional) |
| `>`  | prose line; consecutive `>` lines join into one paragraph, a blank line ends it |
| `$`  | usage line |
| `@`  | theme assignments, `key=value ...` |

```markdown
# mytool | does something useful
@ header=cyan desc=#5fafff
$ mytool <command> [flags]

## Commands
- start | Start the service | mytool start --env prod
- stop | Gracefully stop the service

> All commands support --help for detailed usage.

# config | manage configuration
## Commands
- show | Print current config
```

Indentation carries no meaning, blank lines only end a `>` paragraph, and a section may hold zero items. An `-` item must
sit under a `##` section, and a `>` or `$` line closes the open section. Topic names must be unique.

`@` lines may appear anywhere. Keys: `header`, `command`, `desc`, `descalt`, `text`, `example`,
`syntax`. Values: a palette name (`dim`, `black`, `red`, `green`, `yellow`, `blue`, `magenta`,
`cyan`, `white`, case-insensitive) or a hex color (`#rrggbb` or `rrggbb`). Omitting a key keeps
the default; `key=` with no value means no color. Each key may be assigned once.

Mistakes surface as `gohelp: line N: ...` the first time the tool's help runs, and nothing is printed.

## API

### Pages

```go
gohelp.NewPage(binary, description string) *Page
```

### Builder methods

| Method | Description |
|--------|-------------|
| `.Usage(s string)` | Adds an indented usage line (no section bar) |
| `.Section(title string, entries ...Entry)` | Adds a labeled section with a `──[Title]──` bar |
| `.Text(s string)` | Adds a plain paragraph |

### Items

```go
gohelp.Item(cmd, desc string, example ...string) Entry  // optional third arg is a dim example line
```

### Rendering

```go
gohelp.Run(args []string, root *Page, pages ...*Page) error  // route and print (pass os.Args[1:])
gohelp.Print(p *Page, pages ...*Page)                         // print a specific page directly
```

### Help files

```go
gohelp.Parse(spec string) (*Spec, error)                      // help-file text -> Spec{Root, Pages, Theme}
gohelp.RunSpec(args []string, spec string) error              // Parse, apply the theme to Colors, then Run
```

`Parse` is all-or-nothing: on error it returns `nil` and no pages. `RunSpec` sets `Colors` only after
a successful parse; fields without an `@` assignment keep their current value.

`Run` routing:
- no args / `help` → root page, `nil`
- `help <topic>` → named sub-page, `nil`
- `help --all` → all pages, `nil`
- `help <typo>` → non-nil error naming the topic, a fuzzy suggestion when
  found, and every known topic — the caller decides how to print it and
  exit

## Theming

Colors are set once, globally, before printing:

```go
gohelp.Colors.Header  = gohelp.Cyan          // ──[Title]──── bars and the plain divider
gohelp.Colors.Command = ""                   // command column, topic names
gohelp.Colors.Desc    = gohelp.Hex("#5fafff") // descriptions
gohelp.Colors.DescAlt = ""                   // alternating shade; empty falls back to Desc
gohelp.Colors.Text    = ""                   // page description, paragraphs, usage line
gohelp.Colors.Example = gohelp.Dim           // (e.g. ...) spans
gohelp.Colors.Syntax  = gohelp.Dim           // <arg>, [opt], --flag tokens
```

An empty field emits no escape sequence: the terminal's own foreground is used for base colors,
and an empty `Syntax` disables highlighting entirely.

Built-in palette: `Dim`, `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`.
`gohelp.Hex("#rrggbb")` builds a 24-bit color from a hex string (panics on malformed input — it's
meant to be called with a literal at startup).

## Output

```
──[deploy - zero-downtime deployment tool]──────────────────────────────────

  deploy <command> [flags]

──[Commands]────────────────────────────────────────────────────────────────

  up            Deploy the application to the target environment  (e.g. deploy
                up --env staging)
  down          Tear down the deployment and release all resources
  rollback [n]  Roll back to a previous release; defaults to the last stable
                release if n is omitted  (e.g. deploy rollback 2 --env prod)
  status        Show current deployment status, uptime, and active instances

──[Flags]───────────────────────────────────────────────────────────────────

  --env ENV           Target environment: dev, staging, or prod (required)
  --dry-run           Print the actions that would be taken without executing
                      them
  --timeout DURATION  Maximum time to wait for the deployment to complete
                      before aborting (e.g. 2m, 90s)
  --yes               Skip confirmation prompts


  Credentials are read from the environment. Set DEPLOY_TOKEN or run
  'deploy auth login' to authenticate.


────────────────────────────────────────────────────────────────────────────

Topics:
  releases  list and inspect past deployments
  auth      manage authentication credentials

Run 'deploy help <topic>' for details.
```

See `_examples/main.go` (builder) and `_examples/spec` (help file) for complete working examples.
