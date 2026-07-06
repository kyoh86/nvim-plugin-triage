package gitrepo

import (
	"context"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

func RemoteURL(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func HeadRev(ctx context.Context, dir string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GitHubRepo(raw string) string {
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

func NameFromPath(path string) string {
	return filepath.Base(filepath.Clean(path))
}
