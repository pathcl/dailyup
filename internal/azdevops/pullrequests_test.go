package azdevops_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchPullRequests_FiltersByDate(t *testing.T) {
	now := time.Now().UTC()
	old := now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)
	recent := now.Add(-3 * 24 * time.Hour).Format(time.RFC3339)

	resp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"pullRequestId": 10, "title": "Recent PR", "status": "completed", "creationDate": recent},
			{"pullRequestId": 11, "title": "Old PR", "status": "completed", "creationDate": old},
		},
	}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	c := newClient(t, srv)

	since := now.Add(-14 * 24 * time.Hour)
	prs, err := azdevops.FetchPullRequests(c, since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 PR after date filter, got %d", len(prs))
	}
	if prs[0].ID != 10 {
		t.Errorf("expected PR #10, got #%d", prs[0].ID)
	}
}

func TestFetchPullRequests_Empty(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
	})
	defer srv.Close()
	c := newClient(t, srv)

	since := time.Now().Add(-14 * 24 * time.Hour)
	prs, err := azdevops.FetchPullRequests(c, since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}
