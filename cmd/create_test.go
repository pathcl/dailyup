package cmd_test

import (
	"testing"

	"github.com/pathcl/dailyup/cmd"
)

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
