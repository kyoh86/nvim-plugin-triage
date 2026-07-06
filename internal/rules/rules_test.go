package rules

import (
	"testing"
	"time"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

func TestEvaluateFlagsStaleRepo(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	pushedAt := now.AddDate(-2, 0, 0)
	releasedAt := now.AddDate(-1, -1, 0)
	flags := Evaluate(&plugin.Facts{
		PushedAt:        &pushedAt,
		LatestReleaseAt: &releasedAt,
		OpenIssues:      78,
		OpenPRs:         12,
		RecentCI: []plugin.CIrun{
			{Conclusion: "cancelled"},
		},
	}, DefaultConfig(now))
	want := map[string]bool{
		"pushed_at_older_than_threshold":      true,
		"inactive_with_backlog":               true,
		"latest_release_older_than_threshold": true,
		"open_issues_over_threshold":          true,
		"open_prs_over_threshold":             true,
		"ci_recent_runs_not_successful":       true,
	}
	for _, flag := range flags {
		delete(want, flag.ID)
	}
	for id := range want {
		t.Fatalf("missing flag %q in %#v", id, flags)
	}
}

func TestEvaluateDoesNotFlagMissingReleaseOrCI(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	pushedAt := now
	flags := Evaluate(&plugin.Facts{
		PushedAt: &pushedAt,
	}, DefaultConfig(now))
	if len(flags) != 0 {
		t.Fatalf("unexpected flags: %#v", flags)
	}
}

func TestEvaluateSuccessfulCI(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	pushedAt := now
	flags := Evaluate(&plugin.Facts{
		PushedAt: &pushedAt,
		RecentCI: []plugin.CIrun{
			{Conclusion: "success"},
			{Conclusion: "failure"},
		},
	}, DefaultConfig(now))
	for _, flag := range flags {
		if flag.ID == "ci_recent_runs_not_successful" {
			t.Fatalf("unexpected ci flag: %#v", flags)
		}
	}
}
