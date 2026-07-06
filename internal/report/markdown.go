package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

func WriteMarkdown(w io.Writer, r plugin.Report, includeClean bool) error {
	if _, err := fmt.Fprintf(w, "# nvim-plugin-triage report\n\nGenerated: `%s`\n\n", r.GeneratedAt.Format("2006-01-02 15:04:05 MST")); err != nil {
		return err
	}
	results := append([]plugin.Result(nil), r.Results...)
	sort.SliceStable(results, func(i, j int) bool {
		return len(results[i].Flags) > len(results[j].Flags)
	})
	for _, res := range results {
		if !includeClean && len(res.Flags) == 0 && res.Error == "" {
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
			if _, err := fmt.Fprintf(w, "- Locked: `%s`\n", shortSHA(res.Plugin.LockedRev)); err != nil {
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

func shortSHA(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12]
}
