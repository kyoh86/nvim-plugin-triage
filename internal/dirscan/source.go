package dirscan

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kyoh86/nvim-plugin-triage/internal/gitrepo"
	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

type Source struct {
	Dirs     []string
	Progress io.Writer
}

func (s Source) List(ctx context.Context) ([]plugin.Plugin, error) {
	var plugins []plugin.Plugin
	for _, dir := range s.Dirs {
		s.progressf("scan: reading directory %s\n", dir)
		info, err := os.Stat(dir)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", dir)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			info, err := os.Stat(path)
			if err != nil || !info.IsDir() {
				continue
			}
			s.progressf("scan: inspecting %s\n", path)
			remote, err := gitrepo.RemoteURL(ctx, path)
			if err != nil {
				continue
			}
			plugins = append(plugins, plugin.Plugin{
				Name:      gitrepo.NameFromPath(path),
				Manager:   "directory",
				Repo:      gitrepo.GitHubRepo(remote),
				URL:       remote,
				LockedRev: gitrepo.HeadRev(ctx, path),
				Path:      path,
			})
		}
	}
	return plugins, nil
}

func (s Source) progressf(format string, args ...any) {
	if s.Progress == nil {
		return
	}
	fmt.Fprintf(s.Progress, format, args...)
}
