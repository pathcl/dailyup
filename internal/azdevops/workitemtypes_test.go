package azdevops_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchWorkItemTypes(t *testing.T) {
	resp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"name": "Epic", "description": "Epics"},
			{"name": "Feature", "description": "Features"},
			{"name": "User Story", "description": "User Stories"},
			{"name": "Task", "description": "Tasks"},
			{"name": "Bug", "description": "Bugs"},
		},
	}
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	c := newClient(t, srv)

	types, err := azdevops.FetchWorkItemTypes(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 5 {
		t.Fatalf("expected 5 types, got %d", len(types))
	}
	if types[1].Name != "Feature" {
		t.Errorf("expected Feature at index 1, got %q", types[1].Name)
	}
}
