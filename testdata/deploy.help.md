# deploy | zero-downtime deployment tool
$ deploy <command> [flags]

## Commands
- up | Deploy the application to the target environment | deploy up --env staging
- down | Tear down the deployment and release all resources
- rollback [n] | Roll back to a previous release; defaults to the last stable release if n is omitted | deploy rollback 2 --env prod
- status | Show current deployment status, uptime, and active instances

## Flags
- --env ENV | Target environment: dev, staging, or prod (required)
- --dry-run | Print the actions that would be taken without executing them
- --timeout DURATION | Maximum time to wait for the deployment to complete before aborting (e.g. 2m, 90s)
- --strategy STRATEGY | Rollout strategy to use for this deployment | deploy up --env prod --strategy=canary-incremental-with-healthcheck
- --yes | Skip confirmation prompts

> Credentials are read from the environment. Set DEPLOY_TOKEN or run 'deploy auth login' to authenticate.

# releases | list and inspect past deployments
> Each deployment creates a numbered release. Releases are immutable and retained for 30 days.
## Commands
- releases list | List all releases for the current environment | deploy releases list --env prod
- releases inspect <n> | Show the full metadata, config diff, and log output for release n | deploy releases inspect 14 --env prod

# auth | manage authentication credentials
$ deploy auth <command>
> Tokens are stored in ~/.config/deploy/credentials and never logged or transmitted in plaintext.
## Commands
- login | Authenticate and store credentials locally
- logout | Remove stored credentials
- status | Show the currently authenticated account and token expiry
