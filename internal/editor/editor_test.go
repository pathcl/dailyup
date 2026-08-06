package editor_test

import (
	"testing"

	"github.com/pathcl/dailyup/internal/editor"
)

func TestStripComments(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips hash lines",
			input: "# comment\nTitle: hello\n# another comment",
			want:  "Title: hello",
		},
		{
			name:  "keeps non-comment lines",
			input: "Title: hello\nDescription: world",
			want:  "Title: hello\nDescription: world",
		},
		{
			name:  "trims surrounding whitespace",
			input: "\n\nTitle: hello\n\n",
			want:  "Title: hello",
		},
		{
			name:  "inline hash not treated as comment",
			input: "Title: foo # not a comment",
			want:  "Title: foo # not a comment",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := editor.StripComments(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
