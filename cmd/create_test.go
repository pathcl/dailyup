package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pathcl/dailyup/cmd"
)

func TestItemTypeFromFlag(t *testing.T) {
	cases := []struct {
		typeFlag string
		taskFlag bool
		want     string
		wantErr  bool
	}{
		{"story", false, "User Story", false},
		{"", false, "User Story", false},
		{"task", false, "Task", false},
		{"feature", false, "Feature", false},
		{"FEATURE", false, "Feature", false},
		{"", true, "Task", false},
		{"feature", true, "Feature", false}, // --type wins over --task
		{"epic", false, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.typeFlag+"/task="+func() string {
			if tc.taskFlag {
				return "true"
			}
			return "false"
		}(), func(t *testing.T) {
			got, err := cmd.ItemTypeFromFlag(tc.typeFlag, tc.taskFlag)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadTemplate_FallsBackToDefault(t *testing.T) {
	got := cmd.LoadTemplate("story", "", t.TempDir())
	if !strings.Contains(got, "Title:") {
		t.Error("default story template should contain 'Title:'")
	}
	if !strings.Contains(got, "Description:") {
		t.Error("default story template should contain 'Description:'")
	}
}

func TestLoadTemplate_ReadsCustomPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.md")
	content := "Title:\n\nDescription:\ncustom template\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	got := cmd.LoadTemplate("story", path, t.TempDir())
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLoadTemplate_ReadsFromTemplatesDir(t *testing.T) {
	dir := t.TempDir()
	content := "Title:\n\nDescription:\nfrom templates dir\n"
	if err := os.WriteFile(filepath.Join(dir, "feature.md"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	got := cmd.LoadTemplate("feature", "", dir)
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLoadTemplate_CustomPathTakesPrecedenceOverDir(t *testing.T) {
	dir := t.TempDir()
	dirContent := "Title:\n\nDescription:\nfrom dir\n"
	os.WriteFile(filepath.Join(dir, "story.md"), []byte(dirContent), 0600)

	customDir := t.TempDir()
	customContent := "Title:\n\nDescription:\ncustom\n"
	customPath := filepath.Join(customDir, "custom.md")
	os.WriteFile(customPath, []byte(customContent), 0600)

	got := cmd.LoadTemplate("story", customPath, dir)
	if got != customContent {
		t.Errorf("custom path should win; got %q, want %q", got, customContent)
	}
}

func TestMarkdownToHTML(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string // substrings that must appear in output
	}{
		{
			name:  "heading",
			input: "## Acceptance Criteria",
			want:  []string{"<h2>", "Acceptance Criteria", "</h2>"},
		},
		{
			name:  "bullet list",
			input: "- item one\n- item two",
			want:  []string{"<ul>", "<li>", "item one", "</li>"},
		},
		{
			name:  "paragraph",
			input: "As a platform engineer,\nI want capability,\nso that outcome.",
			want:  []string{"<p>", "As a platform engineer", "</p>"},
		},
		{
			name:  "empty",
			input: "",
			want:  []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cmd.MarkdownToHTML(tc.input)
			for _, substr := range tc.want {
				if !strings.Contains(got, substr) {
					t.Errorf("output missing %q\nfull output: %s", substr, got)
				}
			}
		})
	}
}

func TestParseCreateContent(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		title   string
		desc    string
		wantErr bool
	}{
		{
			name:  "title only",
			input: "Title: My Story\nDescription:",
			title: "My Story",
			desc:  "",
		},
		{
			name:  "title and description",
			input: "Title: My Story\nDescription: Some details here",
			title: "My Story",
			desc:  "Some details here",
		},
		{
			name:  "multi-line description",
			input: "Title: My Story\nDescription:\nLine one\nLine two",
			title: "My Story",
			desc:  "Line one\nLine two",
		},
		{
			name:    "missing title",
			input:   "Title:\nDescription: something",
			wantErr: true,
		},
		{
			name:    "no title line at all",
			input:   "Description: something",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			title, desc, err := cmd.ParseCreateContent(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if title != tc.title {
				t.Errorf("title: got %q, want %q", title, tc.title)
			}
			if desc != tc.desc {
				t.Errorf("desc: got %q, want %q", desc, tc.desc)
			}
		})
	}
}
