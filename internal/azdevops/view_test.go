package azdevops_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchWorkItemDetail_ReturnsAllFields(t *testing.T) {
	resp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"id": 42, "fields": map[string]interface{}{
				"System.Title":                        "My Story",
				"System.WorkItemType":                 "User Story",
				"System.State":                        "Active",
				"System.AreaPath":                     "MyProject\\Area A",
				"System.IterationPath":                "MyProject\\Sprint 1",
				"System.AssignedTo":                   map[string]string{"displayName": "Jane Doe", "uniqueName": "jane@example.com"},
				"System.CreatedBy":                    map[string]string{"displayName": "Bob Smith", "uniqueName": "bob@example.com"},
				"System.CreatedDate":                  "2026-05-01T10:00:00Z",
				"System.ChangedDate":                  "2026-05-10T12:00:00Z",
				"Microsoft.VSTS.Common.Priority":      2,
				"System.Tags":                         "backend; api",
				"System.Description":                  "<p>Some details</p>",
			}},
		},
	}
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	c := newClient(t, srv)

	item, err := azdevops.FetchWorkItemDetail(c, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != 42 {
		t.Errorf("ID: want 42, got %d", item.ID)
	}
	if item.Title != "My Story" {
		t.Errorf("Title: want %q, got %q", "My Story", item.Title)
	}
	if item.State != "Active" {
		t.Errorf("State: want %q, got %q", "Active", item.State)
	}
	if item.AssignedTo != "Jane Doe" {
		t.Errorf("AssignedTo: want %q, got %q", "Jane Doe", item.AssignedTo)
	}
	if item.Priority != 2 {
		t.Errorf("Priority: want 2, got %d", item.Priority)
	}
	if item.Description != "<p>Some details</p>" {
		t.Errorf("Description: want %q, got %q", "<p>Some details</p>", item.Description)
	}
}

func TestFetchWorkItemDetail_NotFound(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"value": []interface{}{}})
	})
	defer srv.Close()
	c := newClient(t, srv)

	_, err := azdevops.FetchWorkItemDetail(c, 999)
	if err == nil {
		t.Fatal("expected error for missing work item, got nil")
	}
}
