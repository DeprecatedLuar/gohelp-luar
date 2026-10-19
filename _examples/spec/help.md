# notes | a tiny note-taking tool
@ header=cyan desc=#5fafff descalt= example=dim
$ notes <command> [flags]
> Notes live in ~/.notes. Run 'notes help sync' for remote setup.

## Commands
- add <text> | Save a new note | notes add "buy milk"
- list | Show every saved note
- sync | Push notes to the remote store

## Flags
- --tag TAG | Attach a tag to the note
- --json | Print machine-readable output

# sync | push and pull notes
$ notes sync [--force]
> Sync needs NOTES_REMOTE to be set.
## Commands
- sync push | Upload local notes | notes sync push
- sync pull | Download remote notes
