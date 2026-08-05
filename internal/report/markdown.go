// Package report renders Azure DevOps data as Markdown.
// Render is the single entry point; it groups work items by type, formats pull
// requests with status and repo name, and lists commits with short SHA. The
// report title is passed as a pre-formatted string by the caller so the package
// has no dependency on query mode or date arithmetic.
package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pathcl/dailyup/internal/azdevops"
)

// Render produces a Markdown summary. title describes what was queried
// (sprint names or a date range) and is used verbatim in the heading.
func Render(title string, items []azdevops.WorkItem, prs []azdevops.PullRequest, commits []azdevops.Commit) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Work Summary — %s\n\n", title)

	sb.WriteString("## Work Items\n\n")
	if len(items) == 0 {
		sb.WriteString("_No items found._\n\n")
	} else {
		writeWorkItems(&sb, items)
	}

	sb.WriteString("## Pull Requests\n\n")
	if len(prs) == 0 {
		sb.WriteString("_No items found._\n\n")
	} else {
		writePRs(&sb, prs)
	}

	sb.WriteString("## Commits\n\n")
	if len(commits) == 0 {
		sb.WriteString("_No items found._\n\n")
	} else {
		writeCommits(&sb, commits)
	}

	return sb.String()
}

func writeWorkItems(sb *strings.Builder, items []azdevops.WorkItem) {
	// group by type
	grouped := make(map[string][]azdevops.WorkItem)
	var types []string
	for _, item := range items {
		if _, ok := grouped[item.Type]; !ok {
			types = append(types, item.Type)
		}
		grouped[item.Type] = append(grouped[item.Type], item)
	}
	sort.Strings(types)

	for _, t := range types {
		fmt.Fprintf(sb, "### %s\n\n", t)
		for _, item := range grouped[t] {
			fmt.Fprintf(sb, "- [#%d] %s (%s)", item.ID, item.Title, item.State)
			if item.Tags != "" {
				fmt.Fprintf(sb, " — tags: %s", item.Tags)
			}
			if item.SprintCount > 1 {
				fmt.Fprintf(sb, " — carried over %d sprints", item.SprintCount)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
}

func writePRs(sb *strings.Builder, prs []azdevops.PullRequest) {
	for _, pr := range prs {
		status := pr.Status
		if status == "completed" {
			status = "Merged"
		} else if status == "active" {
			status = "Active"
		}
		repo := ""
		if pr.RepoName != "" {
			repo = fmt.Sprintf(" — %s", pr.RepoName)
		}
		fmt.Fprintf(sb, "- [#%d] %s — **%s**%s — %s\n",
			pr.ID, pr.Title, status, repo, pr.CreationDate.Format("2006-01-02"))
	}
	sb.WriteString("\n")
}

func writeCommits(sb *strings.Builder, commits []azdevops.Commit) {
	for _, c := range commits {
		fmt.Fprintf(sb, "- `%s` %s — %s — %s\n",
			c.ShortID, c.Message, c.RepoName, c.Date.Format("2006-01-02"))
	}
	sb.WriteString("\n")
}
