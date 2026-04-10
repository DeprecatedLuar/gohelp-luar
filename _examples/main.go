package main

import (
	"os"

	"github.com/DeprecatedLuar/gohelp-luar"
)

func main() {
	root := gohelp.NewPage("deploy", "zero-downtime deployment tool").
		Usage("deploy <command> [flags]").
		Section("Commands",
			gohelp.Cmd("up", "Deploy the application to the target environment").
				Example("deploy up --env staging"),
			gohelp.Cmd("down", "Tear down the deployment and release all resources"),
			gohelp.Cmd("rollback [n]", "Roll back to a previous release; defaults to the last stable release if n is omitted").
				Example("deploy rollback 2 --env prod"),
			gohelp.Cmd("status", "Show current deployment status, uptime, and active instances"),
		).
		Section("Flags",
			gohelp.Cmd("--env ENV", "Target environment: dev, staging, or prod (required)"),
			gohelp.Cmd("--dry-run", "Print the actions that would be taken without executing them"),
			gohelp.Cmd("--timeout DURATION", "Maximum time to wait for the deployment to complete before aborting (e.g. 2m, 90s)"),
			gohelp.Cmd("--yes", "Skip confirmation prompts"),
		).
		Text("Credentials are read from the environment. Set DEPLOY_TOKEN or run 'deploy auth login' to authenticate.")

	releases := gohelp.NewPage("releases", "list and inspect past deployments").
		Text("Each deployment creates a numbered release. Releases are immutable and retained for 30 days.").
		Section("Commands",
			gohelp.Cmd("releases list", "List all releases for the current environment").
				Example("deploy releases list --env prod"),
			gohelp.Cmd("releases inspect <n>", "Show the full metadata, config diff, and log output for release n").
				Example("deploy releases inspect 14 --env prod"),
		)

	auth := gohelp.NewPage("auth", "manage authentication credentials").
		Usage("deploy auth <command>").
		Text("Tokens are stored in ~/.config/deploy/credentials and never logged or transmitted in plaintext.").
		Section("Commands",
			gohelp.Cmd("login", "Authenticate and store credentials locally"),
			gohelp.Cmd("logout", "Remove stored credentials"),
			gohelp.Cmd("status", "Show the currently authenticated account and token expiry"),
		)

	gohelp.Run(os.Args[1:], root, releases, auth)
}
