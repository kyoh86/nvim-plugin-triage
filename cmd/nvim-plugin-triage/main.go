package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/kyoh86/nvim-plugin-triage/internal/dirscan"
	"github.com/kyoh86/nvim-plugin-triage/internal/github"
	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
	"github.com/kyoh86/nvim-plugin-triage/internal/report"
	"github.com/kyoh86/nvim-plugin-triage/internal/rules"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "nvim-plugin-triage:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "scan":
		return scan(ctx, args[1:])
	case "help", "-h", "--help":
		return usage()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func scan(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var dirs multiFlag
	fs.Var(&dirs, "dir", "directory containing plugin repository checkouts; repeatable")
	format := fs.String("format", "json", "output format: json or markdown")
	includeClean := fs.Bool("include-clean", false, "include plugins with no flags in markdown output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("scan requires --dir")
	}
	src := dirscan.Source{Dirs: dirs}
	plugins, err := src.List(ctx)
	if err != nil {
		return err
	}
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Name < plugins[j].Name })
	client := github.Client{}
	now := time.Now()
	rep := plugin.Report{GeneratedAt: now, Results: make([]plugin.Result, 0, len(plugins))}
	for _, p := range plugins {
		res := plugin.Result{Plugin: p}
		if p.Repo == "" {
			res.Error = fmt.Sprintf("github repo not found for %q", p.Name)
			res.Flags = append(res.Flags, plugin.Flag{
				ID:       "repo_url_missing",
				Severity: "warn",
				Evidence: "could not resolve GitHub repo from git remote",
			})
			rep.Results = append(rep.Results, res)
			continue
		}
		facts, err := client.Facts(ctx, github.NormalizeRepo(p.Repo))
		if err != nil {
			res.Error = err.Error()
			res.Flags = append(res.Flags, plugin.Flag{
				ID:       "github_facts_unavailable",
				Severity: "warn",
				Evidence: err.Error(),
			})
			rep.Results = append(rep.Results, res)
			continue
		}
		res.Facts = facts
		res.Flags = rules.Evaluate(facts, rules.DefaultConfig(now))
		rep.Results = append(rep.Results, res)
	}
	switch *format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	case "markdown":
		return report.WriteMarkdown(os.Stdout, rep, *includeClean)
	default:
		return fmt.Errorf("unknown format %q", *format)
	}
}

type multiFlag []string

func (m *multiFlag) String() string {
	return fmt.Sprint([]string(*m))
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func usage() error {
	fmt.Fprintln(os.Stderr, `Usage:
  nvim-plugin-triage scan --dir ~/.local/share/nvim/lazy [--format json|markdown]

Environment:
  GITHUB_TOKEN  optional token for higher GitHub API rate limits`)
	return nil
}
