// Package config loads and validates the TOML configuration file for dailyup.
// The default file path is ~/.config/dailyup/config.toml. Boolean fields
// (pull_requests, commits) use opt-out semantics: an absent key means enabled,
// and explicit false means disabled. Tilde expansion is performed on the path.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds all settings loaded from the TOML file.
type Config struct {
	Organization string   `toml:"organization"`
	Project      string   `toml:"project"`
	Tags         []string `toml:"tags"`          // optional — narrows sprint results
	AssignedTo   string   `toml:"assigned_to"`   // optional — e.g. "@Me" or display name
	Weeks        int      `toml:"weeks"`         // used for PRs/commits date window
	Email        string   `toml:"email"`         // used for commit author filtering
	PullRequests bool     `toml:"pull_requests"` // fetch pull requests (default: true)
	Commits      bool     `toml:"commits"`       // fetch commits (default: true)
}

// DefaultPath returns ~/.config/dailyup/config.toml.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "dailyup", "config.toml")
}

// Load reads and validates the config file at path.
func Load(path string) (*Config, error) {
	path = expandTilde(path)

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s — create it with organization, project, and tags fields", path)
		}
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if cfg.Organization == "" || cfg.Project == "" {
		return nil, fmt.Errorf("config must specify both organization and project")
	}

	if cfg.Weeks <= 0 {
		cfg.Weeks = 2
	}

	// TOML omitted booleans decode as false; treat absence as enabled.
	// Users must explicitly set pull_requests = false to disable.
	var raw map[string]interface{}
	if _, err := toml.DecodeFile(path, &raw); err == nil {
		if _, set := raw["pull_requests"]; !set {
			cfg.PullRequests = true
		}
		if _, set := raw["commits"]; !set {
			cfg.Commits = true
		}
	}

	return &cfg, nil
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
