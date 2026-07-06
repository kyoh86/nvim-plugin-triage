# nvim-plugin-triage

Static triage for Neovim plugin repositories.

This tool collects mechanical repository facts and emits flags. It does not try
to make a final maintenance decision. The intended workflow is:

1. scan all plugins cheaply with this CLI
2. pass flagged repositories to a human or LLM reviewer
3. decide whether to keep, watch, replace, fork, or remove each plugin

## Usage

```sh
go run ./cmd/nvim-plugin-triage scan \
  --lock /path/to/nvim/lazy-lock.json \
  --lazy-dir ~/.local/share/nvim/lazy \
  --format markdown
```

Set `GITHUB_TOKEN` to raise GitHub API rate limits.

## Current Input

The first input implementation is intentionally `lazy.nvim`-specific:

- `lazy-lock.json` provides plugin names and locked revisions
- `--lazy-dir` plugin checkouts provide Git remote URLs

The internal model is plugin-manager independent. Additional inventory sources
should implement `internal/inventory.Source` and return normalized plugin data.

Planned sources:

- `nvim-pack-lock.json`
- `vim.pack.get()` JSON dump
- plain repository list
- package directory scanner

## Flags

Flags are rule-based and evidence-oriented. Examples:

- `repo_archived`
- `pushed_at_older_than_threshold`
- `latest_release_older_than_threshold`
- `open_issues_over_threshold`
- `ci_recent_runs_not_successful`
- `repo_url_missing`
