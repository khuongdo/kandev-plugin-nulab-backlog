package ci

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Run is the cmd/ci entry point. It returns 0 on success, 1 for findings or
// refusals and 2 for usage errors, like pkgverify.Run.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, "ci: subcommand required: secrets, preflight, marketplace, workflows, contract or changes")
		return 2
	}
	cmd, ok := commands[args[0]]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "ci: unknown subcommand %q\n", args[0])
		return 2
	}
	fs := flag.NewFlagSet("ci "+args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	exec := cmd(fs)
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	return exec(stdout, stderr)
}

// A command declares its flags on fs and returns the function that runs it.
type command func(fs *flag.FlagSet) func(stdout, stderr io.Writer) int

var commands = map[string]command{
	"secrets":     secretsCommand,
	"preflight":   preflightCommand,
	"marketplace": marketplaceCommand,
	"workflows":   workflowsCommand,
	"contract":    contractCommand,
	"changes":     changesCommand,
}

func usage(stderr io.Writer, name, flags string) int {
	_, _ = fmt.Fprintf(stderr, "ci %s: %s required\n", name, flags)
	return 2
}

func fail(stderr io.Writer, name string, err error) int {
	_, _ = fmt.Fprintf(stderr, "ci %s: %v\n", name, err)
	return 1
}

func secretsCommand(fs *flag.FlagSet) func(io.Writer, io.Writer) int {
	root := fs.String("root", "", "repository root to scan")
	return func(stdout, stderr io.Writer) int {
		if *root == "" {
			return usage(stderr, "secrets", "-root")
		}
		found, err := ScanSecrets(*root)
		if err != nil {
			return fail(stderr, "secrets", err)
		}
		for _, f := range found {
			_, _ = fmt.Fprintln(stderr, f) // path, line and rule only, never the match
		}
		if len(found) > 0 {
			_, _ = fmt.Fprintf(stderr, "ci secrets: %d credential-shaped string(s); fake secrets must start with %s\n", len(found), allowedPrefix)
			return 1
		}
		_, _ = fmt.Fprintln(stdout, "ci secrets: OK")
		return 0
	}
}

func preflightCommand(fs *flag.FlagSet) func(io.Writer, io.Writer) int {
	var f Facts
	var releases string
	fs.StringVar(&f.Tag, "tag", "", "pushed tag, for example v0.1.0")
	fs.StringVar(&f.ManifestVersion, "version", "", "manifest.yaml version")
	fs.BoolVar(&f.OnMain, "on-main", false, "the tagged commit is on origin/main")
	fs.StringVar(&releases, "releases", "", "gh release list --json tagName,isDraft,isPrerelease output; empty when gh failed")
	fs.StringVar(&f.ManualChecksDir, "manual-checks", "docs/manual-checks", "manual-check records")
	return func(stdout, stderr io.Writer) int {
		if f.Tag == "" || f.ManifestVersion == "" {
			return usage(stderr, "preflight", "-tag and -version")
		}
		var err error
		if f.Releases, err = ParseReleases(releases); err != nil {
			return fail(stderr, "preflight", err)
		}
		if err := CheckRelease(f); err != nil {
			return fail(stderr, "preflight", err)
		}
		_, _ = fmt.Fprintf(stdout, "ci preflight: OK %s\n", f.Tag)
		return 0
	}
}

func marketplaceCommand(fs *flag.FlagSet) func(io.Writer, io.Writer) int {
	manifest := fs.String("manifest", "manifest.yaml", "plugin manifest")
	repo := fs.String("repo", "", "GitHub owner/name")
	categories := fs.String("categories", "", "comma-separated registry categories")
	registry := fs.String("registry", "", "optional plugin-registry/plugins.yaml to check against")
	return func(stdout, stderr io.Writer) int {
		if *repo == "" {
			return usage(stderr, "marketplace", "-repo")
		}
		m, err := ReadManifest(*manifest)
		if err != nil {
			return fail(stderr, "marketplace", err)
		}
		var cats []string
		if *categories != "" {
			cats = strings.Split(*categories, ",")
		}
		entry, err := MarketplaceEntry(m, *repo, cats)
		if err != nil {
			return fail(stderr, "marketplace", err)
		}
		if *registry != "" {
			raw, err := os.ReadFile(*registry)
			if err != nil {
				return fail(stderr, "marketplace", err)
			}
			listed, err := CheckRegistry(raw, m.ID, *repo)
			if err != nil {
				return fail(stderr, "marketplace", err)
			}
			_, _ = fmt.Fprintf(stdout, "# registry: listed=%t\n", listed)
		}
		_, _ = fmt.Fprint(stdout, entry)
		return 0
	}
}

func workflowsCommand(fs *flag.FlagSet) func(io.Writer, io.Writer) int {
	dir := fs.String("dir", "", "workflow directory")
	return func(stdout, stderr io.Writer) int {
		if *dir == "" {
			return usage(stderr, "workflows", "-dir")
		}
		if err := CheckWorkflows(*dir); err != nil {
			return fail(stderr, "workflows", err)
		}
		_, _ = fmt.Fprintln(stdout, "ci workflows: OK")
		return 0
	}
}

func contractCommand(fs *flag.FlagSet) func(io.Writer, io.Writer) int {
	cfg := Config{}
	fs.StringVar(&cfg.BaseURL, "base-url", "", "throwaway Kandev base URL")
	fs.StringVar(&cfg.PackagePath, "package", "", "plugin package (.tar.gz)")
	fs.StringVar(&cfg.PluginID, "plugin-id", "", "manifest id")
	fs.StringVar(&cfg.HostVersion, "host-version", "", "exact host version, for example v0.96.0")
	fs.DurationVar(&cfg.ReadyTimeout, "ready-timeout", 3*time.Minute, "time limit for /ready")
	fs.DurationVar(&cfg.ActiveTimeout, "active-timeout", time.Minute, "time limit for the active status")
	fs.DurationVar(&cfg.PollInterval, "poll", time.Second, "poll interval")
	return func(stdout, stderr io.Writer) int {
		if cfg.BaseURL == "" || cfg.PackagePath == "" || cfg.PluginID == "" || cfg.HostVersion == "" {
			return usage(stderr, "contract", "-base-url, -package, -plugin-id and -host-version")
		}
		if err := RunContract(context.Background(), cfg); err != nil {
			return fail(stderr, "contract", err)
		}
		_, _ = fmt.Fprintf(stdout, "ci contract: OK %s on Kandev %s\n", cfg.PluginID, cfg.HostVersion)
		return 0
	}
}
