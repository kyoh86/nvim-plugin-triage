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
  --dir ~/.local/share/nvim/lazy \
  --format markdown
```

Set `GITHUB_TOKEN` to raise GitHub API rate limits.

## Current Input

The primary input is a directory containing plugin repository checkouts:

- `--dir ~/.local/share/nvim/lazy`
- repeat `--dir` to scan multiple roots

The scanner reads each child repository's Git remote URL and HEAD revision.
This works with `lazy.nvim`, Vim packages, `vim.pack`, or any plugin manager that
keeps plugins as Git checkout directories.

The internal model is plugin-manager independent. Additional inventory sources
should implement `internal/inventory.Source` and return normalized plugin data.

Planned sources:

- `nvim-pack-lock.json`, if its schema remains self-contained enough to read
- `vim.pack.get()` JSON dump
- plain repository list

## Flags

Flags are rule-based and evidence-oriented. Examples:

- `repo_archived`
- `pushed_at_older_than_threshold`
- `latest_release_older_than_threshold`
- `open_issues_over_threshold`
- `ci_recent_runs_not_successful`
- `repo_url_missing`
