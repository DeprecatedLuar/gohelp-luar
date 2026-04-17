package main

import (
	"os"

	"github.com/DeprecatedLuar/gohelp-luar"
)

func main() {
	root := gohelp.NewPage("deploy", "zero-downtime deployment tool").
		Usage("deploy <command> [flags]").
		Section("Commands",
			gohelp.Item("up", "Deploy the application to the target environment", "deploy up --env staging"),
			gohelp.Item("down", "Tear down the deployment and release all resources"),
			gohelp.Item("rollback [n]", "Roll back to a previous release; defaults to the last stable release if n is omitted", "deploy rollback 2 --env prod"),
			gohelp.Item("status", "Show current deployment status, uptime, and active instances"),
		).
		Section("Flags",
			gohelp.Item("--env ENV", "Target environment: dev, staging, or prod (required)"),
			gohelp.Item("--dry-run", "Print the actions that would be taken without executing them"),
			gohelp.Item("--timeout DURATION", "Maximum time to wait for the deployment to complete before aborting (e.g. 2m, 90s)"),
			gohelp.Item("--strategy STRATEGY", "Rollout strategy to use for this deployment", "deploy up --env prod --strategy=canary-incremental-with-healthcheck"),
			gohelp.Item("--yes", "Skip confirmation prompts"),
		).
		Text("Credentials are read from the environment. Set DEPLOY_TOKEN or run 'deploy auth login' to authenticate.")

	releases := gohelp.NewPage("releases", "list and inspect past deployments").
		Text("Each deployment creates a numbered release. Releases are immutable and retained for 30 days.").
		Section("Commands",
			gohelp.Item("releases list", "List all releases for the current environment", "deploy releases list --env prod"),
			gohelp.Item("releases inspect <n>", "Show the full metadata, config diff, and log output for release n", "deploy releases inspect 14 --env prod"),
		)

	auth := gohelp.NewPage("auth", "manage authentication credentials").
		Usage("deploy auth <command>").
		Text("Tokens are stored in ~/.config/deploy/credentials and never logged or transmitted in plaintext.").
		Section("Commands",
			gohelp.Item("login", "Authenticate and store credentials locally"),
			gohelp.Item("logout", "Remove stored credentials"),
			gohelp.Item("status", "Show the currently authenticated account and token expiry"),
		)

	gohelp.Run(os.Args[1:], root, releases, auth)
}
