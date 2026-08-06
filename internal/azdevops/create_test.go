package azdevops_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestCreateNewWorkItem_SendsCorrectPatch(t *testing.T) {
	var capturedBody []map[string]interface{}
	var capturedContentType string

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 55})
	})
	defer srv.Close()
	c := newClient(t, srv)

	// Pass unqualified paths; the client project is "proj" so they get prefixed.
	newID, err := azdevops.CreateNewWorkItem(c, "User Story", "My Story", "Some details",
		"Area B", "Iteration 2", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newID != 55 {
		t.Errorf("newID: want 55, got %d", newID)
	}
	if !strings.Contains(capturedContentType, "application/json-patch+json") {
		t.Errorf("Content-Type: want json-patch+json, got %q", capturedContentType)
	}

	findField := func(field string) interface{} {
		for _, op := range capturedBody {
			if op["path"] == "/fields/"+field {
				return op["value"]
			}
		}
		return nil
	}
	if v, _ := findField("System.Title").(string); v != "My Story" {
		t.Errorf("Title: want %q, got %q", "My Story", v)
	}
	if v, _ := findField("System.AreaPath").(string); v != `proj\Area B` {
		t.Errorf("AreaPath: want %q, got %q", `proj\Area B`, v)
	}
	if v, _ := findField("System.IterationPath").(string); v != `proj\Iteration 2` {
		t.Errorf("IterationPath: want %q, got %q", `proj\Iteration 2`, v)
	}
	if v, _ := findField("System.Description").(string); v != "Some details" {
		t.Errorf("Description: want %q, got %q", "Some details", v)
	}

	// Verify parent relation op
	var relationOp map[string]interface{}
	for _, op := range capturedBody {
		if op["path"] == "/relations/-" {
			relationOp = op
			break
		}
	}
	if relationOp == nil {
		t.Fatal("expected a /relations/- op for parent link, got none")
	}
	val, _ := relationOp["value"].(map[string]interface{})
	if rel, _ := val["rel"].(string); rel != "System.LinkTypes.Hierarchy-Reverse" {
		t.Errorf("relation type: want Hierarchy-Reverse, got %q", rel)
	}
	if parentURL, _ := val["url"].(string); !strings.Contains(parentURL, "100") {
		t.Errorf("parent URL should contain parent ID 100, got %q", parentURL)
	}
}

func TestCreateNewWorkItem_QualifiesAreaAndIterationPaths(t *testing.T) {
	var capturedBody []map[string]interface{}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	})
	defer srv.Close()
	c := newClient(t, srv) // project = "proj"

	// Pass paths without the project prefix — they should be qualified automatically.
	azdevops.CreateNewWorkItem(c, "User Story", "T", "", "Area B", "Sprint 1", 0)

	findField := func(field string) string {
		for _, op := range capturedBody {
			if op["path"] == "/fields/"+field {
				v, _ := op["value"].(string)
				return v
			}
		}
		return ""
	}
	if v := findField("System.AreaPath"); v != `proj\Area B` {
		t.Errorf("AreaPath: want %q, got %q", `proj\Area B`, v)
	}
	if v := findField("System.IterationPath"); v != `proj\Sprint 1` {
		t.Errorf("IterationPath: want %q, got %q", `proj\Sprint 1`, v)
	}
}

func TestCreateNewWorkItem_AlreadyQualifiedPathUnchanged(t *testing.T) {
	var capturedBody []map[string]interface{}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	})
	defer srv.Close()
	c := newClient(t, srv) // project = "proj"

	azdevops.CreateNewWorkItem(c, "User Story", "T", "", `proj\Area B`, `proj\Sprint 1`, 0)

	findField := func(field string) string {
		for _, op := range capturedBody {
			if op["path"] == "/fields/"+field {
				v, _ := op["value"].(string)
				return v
			}
		}
		return ""
	}
	if v := findField("System.AreaPath"); v != `proj\Area B` {
		t.Errorf("AreaPath: want %q, got %q", `proj\Area B`, v)
	}
	if v := findField("System.IterationPath"); v != `proj\Sprint 1` {
		t.Errorf("IterationPath: want %q, got %q", `proj\Sprint 1`, v)
	}
}

func TestCreateNewWorkItem_NoDescriptionOmitted(t *testing.T) {
	var capturedBody []map[string]interface{}

	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	})
	defer srv.Close()
	c := newClient(t, srv)

	azdevops.CreateNewWorkItem(c, "Task", "T", "", "A", "B", 0)

	for _, op := range capturedBody {
		if op["path"] == "/fields/System.Description" {
			t.Error("Description op should be omitted when empty")
		}
		if op["path"] == "/relations/-" {
			t.Error("relation op should be omitted when parentID is 0")
		}
	}
}
