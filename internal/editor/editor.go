// Package editor opens the user's preferred text editor (via $VISUAL, $EDITOR,
// or vi as a fallback) with a pre-filled template and returns the edited content
// with comment lines stripped.
package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Open writes template to a temp file, opens the user's editor, waits for it
// to exit, then returns the file content with comment lines (starting with '#')
// removed. Returns an error if the editor exits non-zero or the file is empty
// after stripping comments.
func Open(template string) (string, error) {
	f, err := os.CreateTemp("", "dailyup-*.txt")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(template); err != nil {
		f.Close()
		return "", fmt.Errorf("write template: %w", err)
	}
	f.Close()

	if err := openInEditor(f.Name()); err != nil {
		return "", err
	}

	raw, err := os.ReadFile(f.Name())
	if err != nil {
		return "", fmt.Errorf("read edited file: %w", err)
	}
	return StripComments(string(raw)), nil
}

func openInEditor(path string) error {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// StripComments removes lines beginning with '#' and trims surrounding whitespace.
func StripComments(s string) string {
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			kept = append(kept, line)
		}
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}
