package rules

import (
	"fmt"
	"time"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

type Config struct {
	Now              time.Time
	StalePushDays    int
	StaleReleaseDays int
	OpenIssuesWarnAt int
	OpenPRsWarnAt    int
}

func DefaultConfig(now time.Time) Config {
	return Config{
		Now:              now,
		StalePushDays:    365,
		StaleReleaseDays: 365,
		OpenIssuesWarnAt: 50,
		OpenPRsWarnAt:    10,
	}
}

func Evaluate(facts *plugin.Facts, cfg Config) []plugin.Flag {
	var flags []plugin.Flag
	if facts == nil {
		return flags
	}
	if facts.Archived {
		flags = append(flags, flag("repo_archived", "critical", "repository is archived"))
	}
	if facts.Disabled {
		flags = append(flags, flag("repo_disabled", "critical", "repository is disabled"))
	}
	if facts.PushedAt == nil {
		flags = append(flags, flag("pushed_at_missing", "warn", "GitHub API did not provide pushed_at"))
	} else if days := int(cfg.Now.Sub(*facts.PushedAt).Hours() / 24); days >= cfg.StalePushDays {
		flags = append(flags, flag("pushed_at_older_than_threshold", "warn", fmt.Sprintf("last push was %d days ago", days)))
	}
	if facts.LatestReleaseAt == nil {
		flags = append(flags, flag("no_latest_release", "info", "latest release was not found"))
	} else if days := int(cfg.Now.Sub(*facts.LatestReleaseAt).Hours() / 24); days >= cfg.StaleReleaseDays {
		flags = append(flags, flag("latest_release_older_than_threshold", "info", fmt.Sprintf("latest release was %d days ago", days)))
	}
	if facts.OpenIssues >= cfg.OpenIssuesWarnAt {
		flags = append(flags, flag("open_issues_over_threshold", "warn", fmt.Sprintf("%d open issues including PRs", facts.OpenIssues)))
	}
	if facts.OpenPRs >= cfg.OpenPRsWarnAt {
		flags = append(flags, flag("open_prs_over_threshold", "warn", fmt.Sprintf("%d open pull requests", facts.OpenPRs)))
	}
	if len(facts.RecentCI) == 0 {
		flags = append(flags, flag("ci_missing_or_unavailable", "info", "no recent GitHub Actions runs found"))
	} else if !hasSuccessfulCI(facts.RecentCI) {
		flags = append(flags, flag("ci_recent_runs_not_successful", "warn", "no success conclusion in recent GitHub Actions runs"))
	}
	return flags
}

func hasSuccessfulCI(runs []plugin.CIrun) bool {
	for _, run := range runs {
		if run.Conclusion == "success" {
			return true
		}
	}
	return false
}

func flag(id, severity, evidence string) plugin.Flag {
	return plugin.Flag{ID: id, Severity: severity, Evidence: evidence}
}
