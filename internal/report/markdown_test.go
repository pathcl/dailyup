package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/report"
)

func date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestRender_ContainsHeadings(t *testing.T) {
	items := []azdevops.WorkItem{
		{ID: 1, Title: "Story One", State: "Active", Type: "User Story", Tags: "sprint-23"},
		{ID: 2, Title: "Task Two", State: "Closed", Type: "Task", Tags: "sprint-23"},
	}
	prs := []azdevops.PullRequest{
		{ID: 42, Title: "feat: Add endpoint", Status: "completed", CreationDate: date("2026-05-20")},
	}
	commits := []azdevops.Commit{
		{ShortID: "abc1234", Message: "Fix bug", RepoName: "my-repo", Date: date("2026-05-21")},
	}

	out := report.Render("Team A\\Sprint 68", items, prs, commits)

	checks := []string{
		"# Work Summary — Team A\\Sprint 68",
		"## Work Items",
		"### User Story",
		"### Task",
		"[#1]", "Story One", "Active",
		"[#2]", "Task Two", "Closed",
		"## Pull Requests",
		"[#42]", "feat: Add endpoint",
		"## Commits",
		"`abc1234`", "Fix bug", "my-repo",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n\nFull output:\n%s", want, out)
		}
	}
}

func TestRender_EmptySections(t *testing.T) {
	out := report.Render("Team A\\Sprint 68", nil, nil, nil)

	if !strings.Contains(out, "# Work Summary") {
		t.Error("missing title")
	}
	if !strings.Contains(out, "_No items found") {
		t.Errorf("expected empty-state placeholder, got:\n%s", out)
	}
}

func TestRender_GroupsByType(t *testing.T) {
	items := []azdevops.WorkItem{
		{ID: 1, Title: "Bug One", State: "Active", Type: "Bug"},
		{ID: 2, Title: "Story One", State: "Active", Type: "User Story"},
		{ID: 3, Title: "Task One", State: "Active", Type: "Task"},
	}

	out := report.Render("Team A\\Sprint 68", items, nil, nil)

	bugIdx := strings.Index(out, "### Bug")
	storyIdx := strings.Index(out, "### User Story")
	taskIdx := strings.Index(out, "### Task")

	if bugIdx == -1 || storyIdx == -1 || taskIdx == -1 {
		t.Fatalf("missing type headings in:\n%s", out)
	}
}

func TestRender_TitleReflectsSprints(t *testing.T) {
	out := report.Render("Team A\\Sprint 67, Team A\\Sprint 68", nil, nil, nil)
	if !strings.Contains(out, "Team A\\Sprint 67, Team A\\Sprint 68") {
		t.Errorf("expected sprint names in title, got:\n%s", out)
	}
}

func TestRender_TitleReflectsDateRange(t *testing.T) {
	out := report.Render("Nov 26, 2025 – May 26, 2026", nil, nil, nil)
	if !strings.Contains(out, "Nov 26, 2025 – May 26, 2026") {
		t.Errorf("expected date range in title, got:\n%s", out)
	}
}
