package azdevops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// qualifyPath ensures path starts with "project\", as required by ADO for
// System.AreaPath and System.IterationPath field values.
func qualifyPath(project, path string) string {
	prefix := project + `\`
	if strings.HasPrefix(path, prefix) {
		return path
	}
	return prefix + path
}

type createPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type relationValue struct {
	Rel        string            `json:"rel"`
	URL        string            `json:"url"`
	Attributes map[string]string `json:"attributes"`
}

// CreateNewWorkItem creates a work item of itemType under the given parent,
// area, and iteration. Returns the ID of the newly created item.
func CreateNewWorkItem(c *Client, itemType, title, description, areaPath, iterationPath string, parentID int) (int, error) {
	ops := []createPatchOp{
		{Op: "add", Path: "/fields/System.Title", Value: title},
		{Op: "add", Path: "/fields/System.AreaPath", Value: qualifyPath(c.Project(), areaPath)},
		{Op: "add", Path: "/fields/System.IterationPath", Value: qualifyPath(c.Project(), iterationPath)},
	}
	if description != "" {
		ops = append(ops, createPatchOp{Op: "add", Path: "/fields/System.Description", Value: description})
	}
	if parentID > 0 {
		parentURL := fmt.Sprintf("%s/_apis/wit/workItems/%d", c.OrgURL(), parentID)
		ops = append(ops, createPatchOp{
			Op:   "add",
			Path: "/relations/-",
			Value: relationValue{
				Rel:        "System.LinkTypes.Hierarchy-Reverse",
				URL:        parentURL,
				Attributes: map[string]string{"comment": ""},
			},
		})
	}

	body, err := json.Marshal(ops)
	if err != nil {
		return 0, err
	}

	endpoint := fmt.Sprintf("%s/_apis/wit/workitems/$%s?api-version=7.1",
		c.ProjectURL(), url.PathEscape(itemType))
	req, err := http.NewRequest(http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json-patch+json")

	resp, err := c.HTTP().Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("create work item HTTP %d: %s", resp.StatusCode, b)
	}
	var result struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.ID, nil
}
