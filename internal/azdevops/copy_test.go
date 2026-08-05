package azdevops_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestFetchWorkItemsByIDs_ReturnsRichFields(t *testing.T) {
	batchResp := map[string]interface{}{
		"value": []map[string]interface{}{
			{"id": 42, "fields": map[string]interface{}{
				"System.Title":        "My Story",
				"System.WorkItemType": "User Story",
				"System.Tags":        "backend; api",
				"System.Description": "<p>Some details</p>",
				"System.AreaPath":    "MyProject\\Area A",
			}},
		},
	}
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(batchResp)
	})
	defer srv.Close()
	c := newClient(t, srv)

	items, err := azdevops.FetchWorkItemsByIDs(c, []int{42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	it := items[0]
	if it.ID != 42 {
		t.Errorf("ID: want 42, got %d", it.ID)
	}
	if it.Title != "My Story" {
		t.Errorf("Title: want %q, got %q", "My Story", it.Title)
	}
	if it.Description != "<p>Some details</p>" {
		t.Errorf("Description: want %q, got %q", "<p>Some details</p>", it.Description)
	}
	if it.AreaPath != "MyProject\\Area A" {
		t.Errorf("AreaPath: want %q, got %q", "MyProject\\Area A", it.AreaPath)
	}
	if it.Tags != "backend; api" {
		t.Errorf("Tags: want %q, got %q", "backend; api", it.Tags)
	}
}

func TestFetchWorkItemsByIDs_EmptyIDs(t *testing.T) {
	items, err := azdevops.FetchWorkItemsByIDs(nil, []int{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items != nil {
		t.Errorf("expected nil for empty IDs, got %v", items)
	}
}

func TestCreateWorkItem_SendsCorrectPatch(t *testing.T) {
	var capturedBody []map[string]string
	var capturedContentType string

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 99})
	})
	defer srv.Close()
	c := newClient(t, srv)

	src := azdevops.CopyableWorkItem{
		ID:          42,
		Title:       "My Story",
		Type:        "User Story",
		Tags:        "backend",
		Description: "<p>details</p>",
		AreaPath:    "MyProject\\Area A",
	}
	newID, err := azdevops.CreateWorkItem(c, src, "MyProject\\Area B", "Team B\\Iteration 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newID != 99 {
		t.Errorf("newID: want 99, got %d", newID)
	}
	if !strings.Contains(capturedContentType, "application/json-patch+json") {
		t.Errorf("Content-Type: want json-patch+json, got %q", capturedContentType)
	}

	findOp := func(path string) string {
		for _, op := range capturedBody {
			if op["path"] == "/fields/"+path {
				return op["value"]
			}
		}
		return ""
	}
	if v := findOp("System.Title"); v != "My Story" {
		t.Errorf("Title op: want %q, got %q", "My Story", v)
	}
	if v := findOp("System.AreaPath"); v != "MyProject\\Area B" {
		t.Errorf("AreaPath op: want %q, got %q", "MyProject\\Area B", v)
	}
	if v := findOp("System.IterationPath"); v != "Team B\\Iteration 2" {
		t.Errorf("IterationPath op: want %q, got %q", "Team B\\Iteration 2", v)
	}
	if v := findOp("System.Tags"); v != "backend" {
		t.Errorf("Tags op: want %q, got %q", "backend", v)
	}
}

func TestCreateWorkItem_OmitsEmptyTagsAndDescription(t *testing.T) {
	var capturedBody []map[string]string

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	})
	defer srv.Close()
	c := newClient(t, srv)

	src := azdevops.CopyableWorkItem{Title: "T", Type: "Task"} // no Tags, no Description
	azdevops.CreateWorkItem(c, src, "A", "B")

	for _, op := range capturedBody {
		if op["path"] == "/fields/System.Tags" {
			t.Error("Tags op should be omitted when empty")
		}
		if op["path"] == "/fields/System.Description" {
			t.Error("Description op should be omitted when empty")
		}
	}
}
