package azdevops_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchCommits_MultiRepo(t *testing.T) {
	reposResp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"id": "repo-a", "name": "repo-a"},
			{"id": "repo-b", "name": "repo-b"},
		},
	}
	commitsResp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"commitId": "abc1234567890", "comment": "Fix bug", "author": map[string]interface{}{
				"name": "Test User", "date": time.Now().Add(-2 * 24 * time.Hour).Format(time.RFC3339),
			}},
		},
	}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/commits") {
			json.NewEncoder(w).Encode(commitsResp)
		} else {
			json.NewEncoder(w).Encode(reposResp)
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	since := time.Now().Add(-14 * 24 * time.Hour)
	commits, err := azdevops.FetchCommits(c, "test@example.com", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// one commit per repo = 2 total
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits (one per repo), got %d", len(commits))
	}
	if !strings.HasPrefix(commits[0].ShortID, "abc1234") {
		t.Errorf("unexpected commit id: %s", commits[0].ShortID)
	}
}

func TestFetchCommits_NoRepos(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
	})
	defer srv.Close()
	c := newClient(t, srv)

	since := time.Now().Add(-14 * 24 * time.Hour)
	commits, err := azdevops.FetchCommits(c, "test@example.com", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected 0 commits, got %d", len(commits))
	}
}
