package azdevops_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchWorkItems_SprintCountPopulated(t *testing.T) {
	wiqlResp := map[string]interface{}{
		"workItems": []map[string]interface{}{
			{"id": 1}, {"id": 2},
		},
	}
	batchResp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"id": 1, "fields": map[string]interface{}{
				"System.Title": "Carry-over task", "System.State": "Active",
				"System.WorkItemType": "Task", "System.Tags": "",
			}},
			{"id": 2, "fields": map[string]interface{}{
				"System.Title": "New task", "System.State": "Active",
				"System.WorkItemType": "Task", "System.Tags": "",
			}},
		},
	}
	// item 1 has been in 3 sprints, item 2 in 1 sprint
	revisions := map[int][]string{
		1: {"Project\\Sprint 66", "Project\\Sprint 67", "Project\\Sprint 68"},
		2: {"Project\\Sprint 68"},
	}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "wiql"):
			json.NewEncoder(w).Encode(wiqlResp)
		case strings.Contains(r.URL.Path, "workitemsbatch"):
			json.NewEncoder(w).Encode(batchResp)
		case strings.Contains(r.URL.Path, "/revisions"):
			// path: /org/_apis/wit/workitems/{id}/revisions
			parts := strings.Split(r.URL.Path, "/")
			var id int
			fmt.Sscanf(parts[len(parts)-2], "%d", &id)
			paths := revisions[id]
			value := make([]map[string]interface{}, len(paths))
			for i, p := range paths {
				value[i] = map[string]interface{}{
					"fields": map[string]interface{}{"System.IterationPath": p},
				}
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"value": value})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	items, err := azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: []string{"Sprint 68"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	byTitle := make(map[string]azdevops.WorkItem)
	for _, it := range items {
		byTitle[it.Title] = it
	}

	if got := byTitle["Carry-over task"].SprintCount; got != 3 {
		t.Errorf("carry-over task: expected SprintCount 3, got %d", got)
	}
	if got := byTitle["New task"].SprintCount; got != 1 {
		t.Errorf("new task: expected SprintCount 1, got %d", got)
	}
}
