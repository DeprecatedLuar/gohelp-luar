# gohelp

Go library for rendering formatted CLI help screens. Fluent builder API, automatic terminal-width wrapping, ANSI colors, fuzzy topic matching.

**One dependency:** [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) for cross-platform terminal width detection.

## Install

```bash
go get github.com/DeprecatedLuar/gohelp-luar
```

## Usage

```go
import (
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
| `.Usage(s string)` | Adds an indented usage line (no section bar) |
| `.Section(title string, entries ...Entry)` | Adds a labeled section with a `──[Title]──` bar |
| `.Text(s string)` | Adds a plain paragraph |

### Items

```go
gohelp.Item(cmd, desc string, example ...string) Entry  // optional third arg is a dim example line
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

See `_examples/main.go` for a complete working example.
