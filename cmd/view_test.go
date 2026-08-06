package cmd_test

import (
	"strings"
	"testing"

	"github.com/pathcl/dailyup/cmd"
	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestRenderDetail_ContainsAllFields(t *testing.T) {
	item := &azdevops.WorkItemDetail{
		ID:            42,
		Title:         "My Story",
		Type:          "User Story",
		State:         "Active",
		AreaPath:      "MyProject\\Area A",
		IterationPath: "MyProject\\Sprint 1",
		AssignedTo:    "Jane Doe",
		CreatedBy:     "Bob Smith",
		CreatedDate:   "2026-05-01T10:00:00Z",
		ChangedDate:   "2026-05-10T12:00:00Z",
		Priority:      2,
		Tags:          "backend; api",
		Description:   "<p>Some details</p>",
	}

	out := cmd.RenderDetail(item)

	checks := []string{
		"#42", "My Story",
		"User Story", "Active",
		"Jane Doe", "Bob Smith",
		"2026-05-01", "2026-05-10",
		"MyProject\\Area A", "MyProject\\Sprint 1",
		"backend; api",
		"Some details", // HTML stripped
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
	// Raw HTML tags should not appear
	if strings.Contains(out, "<p>") {
		t.Errorf("output should not contain raw HTML tags")
	}
}

func TestRenderDetail_EmptyFieldsOmitted(t *testing.T) {
	item := &azdevops.WorkItemDetail{
		ID:    1,
		Title: "Minimal",
		Type:  "Task",
		State: "New",
	}
	out := cmd.RenderDetail(item)
	if strings.Contains(out, "Assigned To") {
		t.Error("empty AssignedTo should be omitted")
	}
	if strings.Contains(out, "Tags") {
		t.Error("empty Tags should be omitted")
	}
	if strings.Contains(out, "Description") {
		t.Error("empty Description should be omitted")
	}
}
