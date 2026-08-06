package azdevops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// WorkItem represents a single Azure DevOps work item.
type WorkItem struct {
	ID          int
	Title       string
	State       string
	Type        string
	Tags        string
	AssignedTo  string
	SprintCount int // number of distinct sprints this item has been in
}

// WorkItemOpts controls how work items are queried.
// Exactly one of Sprints or Since should be set; if neither is set,
// @CurrentIteration is used.
type WorkItemOpts struct {
	// Sprints filters by iteration path. One sprint = UNDER, multiple = IN.
	// Empty means @CurrentIteration.
	Sprints []string
	// Since filters by changed date instead of sprint. Mutually exclusive with Sprints.
	Since *time.Time
	// Tags optionally narrows results to items with at least one matching tag.
	Tags []string
	// AssignedTo optionally filters by assignee. Use "@Me" for the current user,
	// or a display name / email for a specific person. Empty = no filter.
	AssignedTo string
	// Types optionally limits results to specific work item types,
	// e.g. []string{"Feature", "User Story", "Task"}. Empty = all types.
	Types []string
}

type wiqlRequest struct {
	Query string `json:"query"`
}

type wiqlResponse struct {
	WorkItems []struct {
		ID int `json:"id"`
	} `json:"workItems"`
}

type batchRequest struct {
	IDs    []int    `json:"ids"`
	Fields []string `json:"fields"`
}

// adoIdentity matches the object ADO returns for person fields like System.AssignedTo.
type adoIdentity struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

type batchResponse struct {
	Value []struct {
		ID     int `json:"id"`
		Fields struct {
			Title      string      `json:"System.Title"`
			State      string      `json:"System.State"`
			Type       string      `json:"System.WorkItemType"`
			Tags       string      `json:"System.Tags"`
			AssignedTo adoIdentity `json:"System.AssignedTo"`
		} `json:"fields"`
	} `json:"value"`
}

// FetchWorkItems queries work items in the given sprint, optionally narrowed by tags and assignee.
func FetchWorkItems(c *Client, opts WorkItemOpts) ([]WorkItem, error) {
	if len(opts.Tags) == 0 {
		return fetchByIteration(c, opts)
	}

	// Run one WIQL per tag, deduplicate results.
	seen := make(map[int]struct{})
	var ids []int
	for _, tag := range opts.Tags {
		found, err := runWIQL(c, buildQuery(c, opts, tag))
		if err != nil {
			return nil, fmt.Errorf("wiql for tag %q: %w", tag, err)
		}
		for _, id := range found {
			if _, dup := seen[id]; !dup {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return fetchBatch(c, ids)
}

// fetchByIteration runs a single query with no tag filter.
func fetchByIteration(c *Client, opts WorkItemOpts) ([]WorkItem, error) {
	ids, err := runWIQL(c, buildQuery(c, opts, ""))
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return fetchBatch(c, ids)
}

// buildQuery constructs the WIQL query string.
// tag is optional; pass "" to omit the tag condition.
func buildQuery(c *Client, opts WorkItemOpts, tag string) string {
	var conditions []string
	conditions = append(conditions, fmt.Sprintf("[System.TeamProject] = '%s'", c.Project()))

	// Sprint / iteration path / date range
	switch {
	case len(opts.Sprints) == 1:
		iterPath := fmt.Sprintf("%s\\%s", c.Project(), opts.Sprints[0])
		conditions = append(conditions, fmt.Sprintf("[System.IterationPath] UNDER '%s'", iterPath))
	case len(opts.Sprints) > 1:
		paths := make([]string, len(opts.Sprints))
		for i, s := range opts.Sprints {
			paths[i] = fmt.Sprintf("'%s\\%s'", c.Project(), s)
		}
		conditions = append(conditions, fmt.Sprintf("[System.IterationPath] IN (%s)", strings.Join(paths, ", ")))
	case opts.Since != nil:
		conditions = append(conditions, fmt.Sprintf("[System.ChangedDate] >= '%s'", opts.Since.Format("2006-01-02")))
	default:
		conditions = append(conditions, "[System.IterationPath] UNDER @CurrentIteration")
	}

	if tag != "" {
		conditions = append(conditions, fmt.Sprintf("[System.Tags] CONTAINS '%s'", tag))
	}

	if opts.AssignedTo != "" {
		if opts.AssignedTo == "@Me" {
			conditions = append(conditions, "[System.AssignedTo] = @Me")
		} else {
			conditions = append(conditions, fmt.Sprintf("[System.AssignedTo] = '%s'", opts.AssignedTo))
		}
	}

	if len(opts.Types) == 1 {
		conditions = append(conditions, fmt.Sprintf("[System.WorkItemType] = '%s'", opts.Types[0]))
	} else if len(opts.Types) > 1 {
		quoted := make([]string, len(opts.Types))
		for i, t := range opts.Types {
			quoted[i] = fmt.Sprintf("'%s'", t)
		}
		conditions = append(conditions, fmt.Sprintf("[System.WorkItemType] IN (%s)", strings.Join(quoted, ", ")))
	}

	return fmt.Sprintf(
		"SELECT [System.Id] FROM WorkItems WHERE %s ORDER BY [System.ChangedDate] DESC",
		strings.Join(conditions, " AND "),
	)
}

func runWIQL(c *Client, query string) ([]int, error) {
	slog.Debug("running wiql", "query", query)
	url := fmt.Sprintf("%s/_apis/wit/wiql?api-version=7.1", c.ProjectURL())
	body, err := json.Marshal(wiqlRequest{Query: query})
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP().Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wiql HTTP %d: %s", resp.StatusCode, b)
	}
	var result wiqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	ids := make([]int, len(result.WorkItems))
	for i, wi := range result.WorkItems {
		ids[i] = wi.ID
	}
	return ids, nil
}

// CopyableWorkItem holds the fields needed to duplicate a work item to a new area/sprint.
type CopyableWorkItem struct {
	ID          int
	Title       string
	Type        string
	Tags        string
	Description string
	AreaPath    string
}

// FetchWorkItemsByIDs fetches a richer set of fields for the given IDs, suitable for copying.
func FetchWorkItemsByIDs(c *Client, ids []int) ([]CopyableWorkItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	endpoint := fmt.Sprintf("%s/_apis/wit/workitemsbatch?api-version=7.1", c.OrgURL())
	body, err := json.Marshal(batchRequest{
		IDs: ids,
		Fields: []string{
			"System.Id", "System.Title", "System.WorkItemType",
			"System.Tags", "System.Description", "System.AreaPath",
		},
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP().Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("workitemsbatch HTTP %d: %s", resp.StatusCode, b)
	}
	var raw struct {
		Value []struct {
			ID     int `json:"id"`
			Fields struct {
				Title       string `json:"System.Title"`
				Type        string `json:"System.WorkItemType"`
				Tags        string `json:"System.Tags"`
				Description string `json:"System.Description"`
				AreaPath    string `json:"System.AreaPath"`
			} `json:"fields"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	items := make([]CopyableWorkItem, len(raw.Value))
	for i, v := range raw.Value {
		items[i] = CopyableWorkItem{
			ID:          v.ID,
			Title:       v.Fields.Title,
			Type:        v.Fields.Type,
			Tags:        v.Fields.Tags,
			Description: v.Fields.Description,
			AreaPath:    v.Fields.AreaPath,
		}
	}
	return items, nil
}

type patchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// CreateWorkItem creates a new work item of the same type as src, overriding area and iteration.
// If parentID > 0, the new item is linked as a child of that work item.
// Returns the ID of the newly created work item.
func CreateWorkItem(c *Client, src CopyableWorkItem, toArea, toIteration string, parentID int) (int, error) {
	ops := []patchOp{
		{Op: "add", Path: "/fields/System.Title", Value: src.Title},
		{Op: "add", Path: "/fields/System.AreaPath", Value: qualifyPath(c.Project(), toArea)},
		{Op: "add", Path: "/fields/System.IterationPath", Value: qualifyPath(c.Project(), toIteration)},
	}
	if src.Tags != "" {
		ops = append(ops, patchOp{Op: "add", Path: "/fields/System.Tags", Value: src.Tags})
	}
	if src.Description != "" {
		ops = append(ops, patchOp{Op: "add", Path: "/fields/System.Description", Value: src.Description})
	}
	if parentID > 0 {
		ops = append(ops, patchOp{
			Op:   "add",
			Path: "/relations/-",
			Value: relationValue{
				Rel:        "System.LinkTypes.Hierarchy-Reverse",
				URL:        fmt.Sprintf("%s/_apis/wit/workItems/%d", c.OrgURL(), parentID),
				Attributes: map[string]string{"comment": ""},
			},
		})
	}

	body, err := json.Marshal(ops)
	if err != nil {
		return 0, err
	}

	endpoint := fmt.Sprintf("%s/_apis/wit/workitems/$%s?api-version=7.1",
		c.ProjectURL(), src.Type)
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

func fetchBatch(c *Client, ids []int) ([]WorkItem, error) {
	url := fmt.Sprintf("%s/_apis/wit/workitemsbatch?api-version=7.1", c.OrgURL())
	body, err := json.Marshal(batchRequest{
		IDs:    ids,
		Fields: []string{"System.Id", "System.Title", "System.State", "System.WorkItemType", "System.Tags", "System.AssignedTo"},
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP().Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("workitemsbatch HTTP %d: %s", resp.StatusCode, b)
	}
	var result batchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	items := make([]WorkItem, len(result.Value))
	allIDs := make([]int, len(result.Value))
	for i, v := range result.Value {
		items[i] = WorkItem{
			ID:         v.ID,
			Title:      v.Fields.Title,
			State:      v.Fields.State,
			Type:       v.Fields.Type,
			Tags:       v.Fields.Tags,
			AssignedTo: v.Fields.AssignedTo.DisplayName,
		}
		allIDs[i] = v.ID
	}

	slog.Debug("fetching iteration history", "items", len(allIDs))
	counts := fetchIterationCounts(c, allIDs)
	for i := range items {
		items[i].SprintCount = counts[items[i].ID]
	}

	return items, nil
}
