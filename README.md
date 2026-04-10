# gohelp

Go library for rendering formatted CLI help screens. Fluent builder API, automatic terminal-width wrapping, ANSI colors, fuzzy topic matching.

## Install

```bash
go get github.com/DeprecatedLuar/gohelp
```

## Usage

```go
root := gohelp.NewPage("mytool", "does something useful").
    Usage("mytool <command> [flags]").
    Section("Commands",
        gohelp.Cmd("start", "Start the service").Example("mytool start --env prod"),
        gohelp.Cmd("stop", "Gracefully stop the service"),
    ).
    Section("Flags",
        gohelp.Cmd("--verbose", "Enable debug output"),
        gohelp.Cmd("--help", "Show this help message"),
    ).
    Text("All commands support --help for detailed usage.")

config := gohelp.NewPage("config", "manage configuration").
    Usage("mytool config <command>").
    Section("Commands",
        gohelp.Cmd("show", "Print current config"),
        gohelp.Cmd("edit", "Open config in $EDITOR").Example("mytool config edit"),
    )

gohelp.Run(os.Args[1:], root, config)
```

## API

### Pages

```go
gohelp.NewPage(binary, description string) *Page
```

### Builder methods

| Method | Description |
|--------|-------------|
| `.Usage(s string)` | Adds a `Usage:` section with a single line |
| `.Section(title string, entries ...Entry)` | Adds a labeled block of command/description pairs |
| `.Text(s string)` | Adds a plain paragraph |

### Entries

```go
gohelp.Cmd(cmd, desc string) Entry       // create an entry
entry.Example(s string) Entry            // add a dim example line below the description
```

### Rendering

```go
gohelp.Run(args []string, root *Page, pages ...*Page)   // route and print (pass os.Args[1:])
gohelp.Print(p *Page, pages ...*Page)                   // print a specific page directly
```

`Run` routing:
- no args / `help` → root page
- `help <topic>` → named sub-page
- `help --all` → all pages
- `help <typo>` → fuzzy suggestion, exit 1

## Output

```
──[mytool - does something useful]──────────────────────────────────────────

Usage:
  mytool <command> [flags]

Commands:
  start  Start the service
         # mytool start --env prod

  stop   Gracefully stop the service

──[topics - mytool help <topic>]─────────────────────────────────────────────

  config  manage configuration
```

See `_examples/main.go` for a complete working example.
