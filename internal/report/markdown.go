package report

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

func WriteMarkdown(w io.Writer, r plugin.Report, includeClean bool) error {
	if _, err := fmt.Fprintf(w, "# nvim-plugin-triage report\n\nGenerated: `%s`\n\n", r.GeneratedAt.Format("2006-01-02 15:04:05 MST")); err != nil {
		return err
	}
	results := append([]plugin.Result(nil), r.Results...)
	summary, candidates := Analyze(results)
	if err := writeSummary(w, summary); err != nil {
		return err
	}
	sort.SliceStable(results, func(i, j int) bool {
		return resultWeight(results[i]) > resultWeight(results[j])
	})
	if err := writeReviewCandidates(w, candidates); err != nil {
		return err
	}
	for _, res := range results {
		if !includeClean && !needsReview(res) {
			continue
		}
		if _, err := fmt.Fprintf(w, "## %s\n\n", res.Plugin.Name); err != nil {
			return err
		}
		if res.Plugin.Repo != "" {
			if _, err := fmt.Fprintf(w, "- Repo: `%s`\n", res.Plugin.Repo); err != nil {
				return err
			}
		}
		if res.Plugin.LockedRev != "" {
			if _, err := fmt.Fprintf(w, "- Revision: `%s`\n", shortSHA(res.Plugin.LockedRev)); err != nil {
				return err
			}
		}
		if res.Error != "" {
			if _, err := fmt.Fprintf(w, "- Error: `%s`\n", res.Error); err != nil {
				return err
			}
		}
		for _, f := range res.Flags {
			if _, err := fmt.Fprintf(w, "- `%s` [%s]: %s\n", f.ID, f.Severity, f.Evidence); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}

func Analyze(results []plugin.Result) (plugin.Summary, []plugin.ReviewCandidate) {
	var summary plugin.Summary
	for _, res := range results {
		summary.Total++
		switch {
		case hasSeverity(res, "critical"):
			summary.Critical++
		case hasSeverity(res, "warn") || res.Error != "":
			summary.Warn++
		case len(res.Flags) > 0:
			summary.ContextOnly++
		default:
			summary.Clean++
		}
	}
	candidates := make([]plugin.ReviewCandidate, 0)
	for _, res := range results {
		if !needsReview(res) {
			continue
		}
		var severity string
		switch {
		case hasSeverity(res, "critical"):
			severity = "critical"
		default:
			severity = "warn"
		}
		candidates = append(candidates, plugin.ReviewCandidate{
			Name:     res.Plugin.Name,
			Repo:     res.Plugin.Repo,
			Severity: severity,
			Flags:    reviewFlagIDs(res),
		})
	}
	summary.ReviewCandidates = len(candidates)
	return summary, candidates
}

func writeSummary(w io.Writer, summary plugin.Summary) error {
	if _, err := fmt.Fprintln(w, "## Summary"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\n- Total: %d\n", summary.Total); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "- Critical: %d\n", summary.Critical); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "- Warn: %d\n", summary.Warn); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "- Review candidates: %d\n", summary.ReviewCandidates); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "- Context-only: %d\n", summary.ContextOnly); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "- Clean: %d\n\n", summary.Clean)
	return err
}

func writeReviewCandidates(w io.Writer, candidates []plugin.ReviewCandidate) error {
	if len(candidates) == 0 {
		if _, err := fmt.Fprintln(w, "## Review Candidates"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "No review candidates."); err != nil {
			return err
		}
		_, err := fmt.Fprintln(w)
		return err
	}
	if _, err := fmt.Fprintln(w, "## Review Candidates"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	for _, candidate := range candidates {
		if _, err := fmt.Fprintf(w, "- `%s`: %s\n", candidate.Name, strings.Join(candidate.Flags, ", ")); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func needsReview(res plugin.Result) bool {
	return res.Error != "" || hasSeverity(res, "critical") || hasSeverity(res, "warn")
}

func hasSeverity(res plugin.Result, severity string) bool {
	for _, flag := range res.Flags {
		if flag.Severity == severity {
			return true
		}
	}
	return false
}

func resultWeight(res plugin.Result) int {
	weight := 0
	for _, flag := range res.Flags {
		switch flag.Severity {
		case "critical":
			weight += 100
		case "warn":
			weight += 10
		case "info":
			weight++
		}
	}
	if res.Error != "" {
		weight += 10
	}
	return weight
}

func reviewFlagIDs(res plugin.Result) []string {
	if res.Error != "" {
		return []string{"error"}
	}
	var ids []string
	for _, flag := range res.Flags {
		if flag.Severity == "critical" || flag.Severity == "warn" {
			ids = append(ids, flag.ID)
		}
	}
	return ids
}

func shortSHA(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12]
}
