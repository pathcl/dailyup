package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pathcl/dailyup/internal/azdevops"
)

// Render produces a Markdown summary of work done between from and to.
func Render(from, to time.Time, items []azdevops.WorkItem, prs []azdevops.PullRequest, commits []azdevops.Commit) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Work Summary — %s – %s\n\n", from.Format("Jan 2, 2006"), to.Format("Jan 2, 2006"))

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
