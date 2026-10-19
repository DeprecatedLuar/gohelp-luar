package gohelp

import (
	"io"
	"os"
	"strings"
	"testing"
)

const deploySpecPath = "testdata/deploy.help.md"

// capture swaps os.Stdout for a pipe around fn and returns what was written.
func capture(t *testing.T, fn func()) string {
	t.Helper()
	if w := termWidth(); w != defaultTermWidth {
		t.Fatalf("termWidth() = %d, want %d for deterministic wrapping", w, defaultTermWidth)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	func() {
		defer func() {
			os.Stdout = orig
			w.Close()
		}()
		fn()
	}()
	return <-done
}

func restoreColors(t *testing.T) {
	saved := Colors
	t.Cleanup(func() { Colors = saved })
}

func readDeploySpec(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(deploySpecPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRunSpecEquivalence(t *testing.T) {
	restoreColors(t)
	root := NewPage("deploy", "zero-downtime deployment tool").
		Usage("deploy <command> [flags]").
		Section("Commands",
			Item("up", "Deploy the application to the target environment", "deploy up --env staging"),
			Item("down", "Tear down the deployment and release all resources"),
			Item("rollback [n]", "Roll back to a previous release; defaults to the last stable release if n is omitted", "deploy rollback 2 --env prod"),
			Item("status", "Show current deployment status, uptime, and active instances"),
		).
		Section("Flags",
			Item("--env ENV", "Target environment: dev, staging, or prod (required)"),
			Item("--dry-run", "Print the actions that would be taken without executing them"),
			Item("--timeout DURATION", "Maximum time to wait for the deployment to complete before aborting (e.g. 2m, 90s)"),
			Item("--strategy STRATEGY", "Rollout strategy to use for this deployment", "deploy up --env prod --strategy=canary-incremental-with-healthcheck"),
			Item("--yes", "Skip confirmation prompts"),
		).
		Text("Credentials are read from the environment. Set DEPLOY_TOKEN or run 'deploy auth login' to authenticate.")
	releases := NewPage("releases", "list and inspect past deployments").
		Text("Each deployment creates a numbered release. Releases are immutable and retained for 30 days.").
		Section("Commands",
			Item("releases list", "List all releases for the current environment", "deploy releases list --env prod"),
			Item("releases inspect <n>", "Show the full metadata, config diff, and log output for release n", "deploy releases inspect 14 --env prod"),
		)
	auth := NewPage("auth", "manage authentication credentials").
		Usage("deploy auth <command>").
		Text("Tokens are stored in ~/.config/deploy/credentials and never logged or transmitted in plaintext.").
		Section("Commands",
			Item("login", "Authenticate and store credentials locally"),
			Item("logout", "Remove stored credentials"),
			Item("status", "Show the currently authenticated account and token expiry"),
		)
	spec := readDeploySpec(t)

	for _, args := range [][]string{{"help"}, {"help", "releases"}, {"help", "auth"}, {"help", "--all"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var wantErr, gotErr error
			want := capture(t, func() { wantErr = Run(args, root, releases, auth) })
			got := capture(t, func() { gotErr = RunSpec(args, spec) })
			if wantErr != nil || gotErr != nil {
				t.Fatalf("errors: builder %v, spec %v", wantErr, gotErr)
			}
			if got != want {
				t.Errorf("output differs\nwant:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

func TestRunSpecUnknownTopic(t *testing.T) {
	restoreColors(t)
	err := RunSpec([]string{"help", "relases"}, readDeploySpec(t))
	if err == nil || !strings.Contains(err.Error(), "did you mean: releases") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunSpecParseError(t *testing.T) {
	restoreColors(t)
	sentinel := Theme{Header: Red}
	Colors = sentinel
	var err error
	out := capture(t, func() { err = RunSpec(nil, "# x\n* bad\n") })
	if err == nil || !strings.Contains(err.Error(), "line 2:") {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if Colors != sentinel {
		t.Errorf("Colors changed on parse error")
	}
}

func TestRunSpecKeepsCallerColors(t *testing.T) {
	restoreColors(t)
	Colors.Header = Cyan
	capture(t, func() {
		if err := RunSpec(nil, "# x\n"); err != nil {
			t.Error(err)
		}
	})
	if Colors.Header != Cyan {
		t.Errorf("Colors.Header = %q, want Cyan", Colors.Header)
	}
}
