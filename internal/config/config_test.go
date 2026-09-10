package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pathcl/dailyup/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `
organization = "myorg"
project      = "myproject"
tags         = ["sprint-23", "backend"]
weeks        = 3
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Organization != "myorg" {
		t.Errorf("organization: got %q, want %q", cfg.Organization, "myorg")
	}
	if cfg.Project != "myproject" {
		t.Errorf("project: got %q, want %q", cfg.Project, "myproject")
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "sprint-23" || cfg.Tags[1] != "backend" {
		t.Errorf("tags: got %v, want [sprint-23 backend]", cfg.Tags)
	}
	if cfg.Weeks != 3 {
		t.Errorf("weeks: got %d, want 3", cfg.Weeks)
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, `
organization = "myorg"
project      = "myproject"
tags         = ["sprint-23"]
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Weeks != 2 {
		t.Errorf("default weeks: got %d, want 2", cfg.Weeks)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_BadTOML(t *testing.T) {
	path := writeTempConfig(t, `organization = [broken`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for bad TOML, got nil")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	path := writeTempConfig(t, `weeks = 1`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when organization/project missing, got nil")
	}
}

func TestLoad_IterationBase(t *testing.T) {
	path := writeTempConfig(t, `
organization   = "myorg"
project        = "myproject"
iteration_base = 'Company\Team'
sprint         = "Sprint1"
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IterationBase != `Company\Team` {
		t.Errorf("iteration_base: got %q, want %q", cfg.IterationBase, `Company\Team`)
	}
	if cfg.Sprint != "Sprint1" {
		t.Errorf("sprint: got %q, want %q", cfg.Sprint, "Sprint1")
	}
}

func TestBuildIterationPath(t *testing.T) {
	cases := []struct {
		base, leaf string
		want       string
	}{
		{`Company\Team`, "Sprint1", `Company\Team\Sprint1`},
		{`Company\Team`, "", `Company\Team`},
		{"", `Team\Sprint1`, `Team\Sprint1`},
		{"", "", ""},
	}
	for _, tc := range cases {
		got := config.BuildIterationPath(tc.base, tc.leaf)
		if got != tc.want {
			t.Errorf("BuildIterationPath(%q, %q) = %q, want %q", tc.base, tc.leaf, got, tc.want)
		}
	}
}
