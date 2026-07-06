package plugin

import "time"

type Plugin struct {
	Name      string `json:"name"`
	Manager   string `json:"manager"`
	Repo      string `json:"repo,omitempty"`
	URL       string `json:"url,omitempty"`
	Branch    string `json:"branch,omitempty"`
	LockedRev string `json:"locked_rev,omitempty"`
	Path      string `json:"path,omitempty"`
}

type Facts struct {
	PushedAt        *time.Time `json:"pushed_at,omitempty"`
	LatestReleaseAt *time.Time `json:"latest_release_at,omitempty"`
	Archived        bool       `json:"archived"`
	Disabled        bool       `json:"disabled"`
	OpenIssues      int        `json:"open_issues"`
	OpenIssuesTotal int        `json:"open_issues_including_prs"`
	OpenPRs         int        `json:"open_prs"`
	RecentCI        []CIrun    `json:"recent_ci,omitempty"`
}

type CIrun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url,omitempty"`
}

type Flag struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Evidence string `json:"evidence"`
}

type Result struct {
	Plugin Plugin `json:"plugin"`
	Facts  *Facts `json:"facts,omitempty"`
	Flags  []Flag `json:"flags,omitempty"`
	Error  string `json:"error,omitempty"`
}

type Report struct {
	GeneratedAt      time.Time         `json:"generated_at"`
	Summary          Summary           `json:"summary"`
	ReviewCandidates []ReviewCandidate `json:"review_candidates"`
	Results          []Result          `json:"results"`
}

type Summary struct {
	Total            int `json:"total"`
	Critical         int `json:"critical"`
	Warn             int `json:"warn"`
	ReviewCandidates int `json:"review_candidates"`
	ContextOnly      int `json:"context_only"`
	Clean            int `json:"clean"`
}

type ReviewCandidate struct {
	Name     string   `json:"name"`
	Repo     string   `json:"repo,omitempty"`
	Severity string   `json:"severity"`
	Flags    []string `json:"flags"`
}
