package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

type Client struct {
	HTTP *http.Client
}

func (c Client) Facts(ctx context.Context, repo string) (*plugin.Facts, error) {
	var r repoResponse
	if err := c.get(ctx, "https://api.github.com/repos/"+repo, &r); err != nil {
		return nil, err
	}
	facts := &plugin.Facts{
		Archived:        r.Archived,
		Disabled:        r.Disabled,
		OpenIssuesTotal: r.OpenIssuesCount,
	}
	if r.PushedAt != "" {
		if t, err := time.Parse(time.RFC3339, r.PushedAt); err == nil {
			facts.PushedAt = &t
		}
	}
	if rel, err := c.latestRelease(ctx, repo); err == nil {
		facts.LatestReleaseAt = rel
	}
	facts.OpenPRs = c.openPRCount(ctx, repo)
	facts.OpenIssues = max(r.OpenIssuesCount-facts.OpenPRs, 0)
	facts.RecentCI = c.recentCI(ctx, repo)
	return facts, nil
}

func (c Client) latestRelease(ctx context.Context, repo string) (*time.Time, error) {
	var r releaseResponse
	if err := c.get(ctx, "https://api.github.com/repos/"+repo+"/releases/latest", &r); err != nil {
		return nil, err
	}
	if r.PublishedAt == "" {
		return nil, errors.New("latest release has no published_at")
	}
	t, err := time.Parse(time.RFC3339, r.PublishedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (c Client) openPRCount(ctx context.Context, repo string) int {
	var prs []struct{}
	if err := c.get(ctx, "https://api.github.com/repos/"+repo+"/pulls?state=open&per_page=100", &prs); err != nil {
		return 0
	}
	return len(prs)
}

func (c Client) recentCI(ctx context.Context, repo string) []plugin.CIrun {
	var runs workflowRunsResponse
	if err := c.get(ctx, "https://api.github.com/repos/"+repo+"/actions/runs?per_page=5", &runs); err != nil {
		return nil
	}
	out := make([]plugin.CIrun, 0, len(runs.WorkflowRuns))
	for _, run := range runs.WorkflowRuns {
		out = append(out, plugin.CIrun{
			Name:       run.Name,
			Status:     run.Status,
			Conclusion: run.Conclusion,
			HTMLURL:    run.HTMLURL,
		})
	}
	return out
}

func (c Client) get(ctx context.Context, endpoint string, out any) error {
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("github resource not found: %s", endpoint)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		remain := resp.Header.Get("X-RateLimit-Remaining")
		reset := resp.Header.Get("X-RateLimit-Reset")
		return fmt.Errorf("github api status %d for %s%s", resp.StatusCode, endpoint, rateLimitHint(remain, reset))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func rateLimitHint(remain, reset string) string {
	if remain != "0" || reset == "" {
		return ""
	}
	sec, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		return " (rate limited)"
	}
	return " (rate limited until " + time.Unix(sec, 0).Format(time.RFC3339) + ")"
}

type repoResponse struct {
	PushedAt        string `json:"pushed_at"`
	Archived        bool   `json:"archived"`
	Disabled        bool   `json:"disabled"`
	OpenIssuesCount int    `json:"open_issues_count"`
}

type releaseResponse struct {
	PublishedAt string `json:"published_at"`
}

type workflowRunsResponse struct {
	WorkflowRuns []struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HTMLURL    string `json:"html_url"`
	} `json:"workflow_runs"`
}

func NormalizeRepo(repo string) string {
	return strings.Trim(strings.TrimSuffix(repo, ".git"), "/")
}
