package azdevops_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchWorkItems_DeduplicatesAndGroups(t *testing.T) {
	wiqlResp := map[string]interface{}{
		"workItems": []map[string]interface{}{
			{"id": 1}, {"id": 2}, {"id": 1}, // duplicate
		},
	}
	batchResp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"id": 1, "fields": map[string]interface{}{
				"System.Title": "Story One", "System.State": "Active",
				"System.WorkItemType": "User Story", "System.Tags": "sprint-23",
			}},
			{"id": 2, "fields": map[string]interface{}{
				"System.Title": "Task Two", "System.State": "Closed",
				"System.WorkItemType": "Task", "System.Tags": "sprint-23",
			}},
		},
	}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			json.NewEncoder(w).Encode(wiqlResp)
		} else {
			json.NewEncoder(w).Encode(batchResp)
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	items, err := azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{
		Sprints: []string{"Sprint 68"},
		Tags:    []string{"sprint-23"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 deduplicated items, got %d", len(items))
	}
	byType := make(map[string]int)
	for _, it := range items {
		byType[it.Type]++
	}
	if byType["User Story"] != 1 {
		t.Errorf("expected 1 User Story, got %d", byType["User Story"])
	}
	if byType["Task"] != 1 {
		t.Errorf("expected 1 Task, got %d", byType["Task"])
	}
}

func TestFetchWorkItems_EmptyResult(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	items, err := azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: []string{"Sprint 68"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestFetchWorkItems_CurrentIteration(t *testing.T) {
	var capturedQuery string
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			var req struct{ Query string }
			json.NewDecoder(r.Body).Decode(&req)
			capturedQuery = req.Query
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: nil})
	if !strings.Contains(capturedQuery, "@CurrentIteration") {
		t.Errorf("expected @CurrentIteration in query, got: %s", capturedQuery)
	}
}

func TestFetchWorkItems_NamedSprint(t *testing.T) {
	var capturedQuery string
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			var req struct{ Query string }
			json.NewDecoder(r.Body).Decode(&req)
			capturedQuery = req.Query
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: []string{"Sprint 68"}})
	if !strings.Contains(capturedQuery, "Sprint 68") {
		t.Errorf("expected 'Sprint 68' in query, got: %s", capturedQuery)
	}
	if strings.Contains(capturedQuery, "@CurrentIteration") {
		t.Errorf("should not use @CurrentIteration for named sprint, got: %s", capturedQuery)
	}
}

func TestFetchWorkItems_AssignedToMe(t *testing.T) {
	var capturedQuery string
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			var req struct{ Query string }
			json.NewDecoder(r.Body).Decode(&req)
			capturedQuery = req.Query
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: []string{"Sprint 68"}, AssignedTo: "@Me"})
	if !strings.Contains(capturedQuery, "[System.AssignedTo] = @Me") {
		t.Errorf("expected AssignedTo @Me in query, got: %s", capturedQuery)
	}
}

func TestFetchWorkItems_MultiSprint(t *testing.T) {
	var capturedQuery string
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			var req struct{ Query string }
			json.NewDecoder(r.Body).Decode(&req)
			capturedQuery = req.Query
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: []string{"Sprint 67", "Sprint 68"}})
	if !strings.Contains(capturedQuery, "IN") {
		t.Errorf("expected IN operator for multiple sprints, got: %s", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "Sprint 67") || !strings.Contains(capturedQuery, "Sprint 68") {
		t.Errorf("expected both sprint names in query, got: %s", capturedQuery)
	}
	if strings.Contains(capturedQuery, "UNDER") {
		t.Errorf("should not use UNDER for multiple sprints, got: %s", capturedQuery)
	}
}

func TestFetchWorkItems_DateFallback(t *testing.T) {
	var capturedQuery string
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			var req struct{ Query string }
			json.NewDecoder(r.Body).Decode(&req)
			capturedQuery = req.Query
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	since := time.Date(2025, 11, 26, 0, 0, 0, 0, time.UTC)
	azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Since: &since})
	if !strings.Contains(capturedQuery, "ChangedDate") {
		t.Errorf("expected ChangedDate in date-based query, got: %s", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "2025-11-26") {
		t.Errorf("expected date 2025-11-26 in query, got: %s", capturedQuery)
	}
	if strings.Contains(capturedQuery, "IterationPath") {
		t.Errorf("should not use IterationPath in date-based query, got: %s", capturedQuery)
	}
}

func TestFetchWorkItems_NoTagsQueriesAll(t *testing.T) {
	callCount := 0
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "wiql") {
			callCount++
			json.NewEncoder(w).Encode(map[string]interface{}{"workItems": []interface{}{}})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
		}
	})
	defer srv.Close()
	c := newClient(t, srv)

	azdevops.FetchWorkItems(c, azdevops.WorkItemOpts{Sprints: []string{"Sprint 68"}})
	if callCount != 1 {
		t.Errorf("expected exactly 1 WIQL call when no tags, got %d", callCount)
	}
}
