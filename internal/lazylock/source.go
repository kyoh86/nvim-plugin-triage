package lazylock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

type Source struct {
	LockPath string
	LazyDir  string
}

type lockEntry struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

func (s Source) List(ctx context.Context) ([]plugin.Plugin, error) {
	data, err := os.ReadFile(s.LockPath)
	if err != nil {
		return nil, err
	}
	var lock map[string]lockEntry
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}
	plugins := make([]plugin.Plugin, 0, len(lock))
	for name, entry := range lock {
		p := plugin.Plugin{
			Name:      name,
			Manager:   "lazy.nvim",
			Branch:    entry.Branch,
			LockedRev: entry.Commit,
		}
		if s.LazyDir != "" {
			p.Path = filepath.Join(s.LazyDir, name)
			if remote, err := gitRemoteURL(ctx, p.Path); err == nil {
				p.URL = remote
				p.Repo = githubRepo(remote)
			}
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func gitRemoteURL(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func githubRepo(raw string) string {
	raw = strings.TrimSuffix(raw, ".git")
	if strings.HasPrefix(raw, "git@github.com:") {
		return strings.TrimPrefix(raw, "git@github.com:")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host != "github.com" {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

func DefaultLazyDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "nvim", "lazy")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "nvim", "lazy")
}

func ValidateRepo(p plugin.Plugin) error {
	if p.Repo == "" {
		return fmt.Errorf("github repo not found for %q", p.Name)
	}
	return nil
}
