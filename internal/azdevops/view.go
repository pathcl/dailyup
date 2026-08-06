package azdevops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WorkItemDetail holds all displayable fields for a single work item.
type WorkItemDetail struct {
	ID            int
	Title         string
	Type          string
	State         string
	AreaPath      string
	IterationPath string
	AssignedTo    string
	CreatedBy     string
	CreatedDate   string
	ChangedDate   string
	Priority      int
	Tags          string
	Description   string
}

// FetchWorkItemDetail fetches all displayable fields for a single work item by ID.
func FetchWorkItemDetail(c *Client, id int) (*WorkItemDetail, error) {
	body, err := json.Marshal(batchRequest{
		IDs: []int{id},
		Fields: []string{
			"System.Id",
			"System.Title",
			"System.WorkItemType",
			"System.State",
			"System.AreaPath",
			"System.IterationPath",
			"System.AssignedTo",
			"System.CreatedBy",
			"System.CreatedDate",
			"System.ChangedDate",
			"Microsoft.VSTS.Common.Priority",
			"System.Tags",
			"System.Description",
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/_apis/wit/workitemsbatch?api-version=7.1", c.OrgURL())
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
				Title         string      `json:"System.Title"`
				Type          string      `json:"System.WorkItemType"`
				State         string      `json:"System.State"`
				AreaPath      string      `json:"System.AreaPath"`
				IterationPath string      `json:"System.IterationPath"`
				AssignedTo    adoIdentity `json:"System.AssignedTo"`
				CreatedBy     adoIdentity `json:"System.CreatedBy"`
				CreatedDate   string      `json:"System.CreatedDate"`
				ChangedDate   string      `json:"System.ChangedDate"`
				Priority      int         `json:"Microsoft.VSTS.Common.Priority"`
				Tags          string      `json:"System.Tags"`
				Description   string      `json:"System.Description"`
			} `json:"fields"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Value) == 0 {
		return nil, fmt.Errorf("work item #%d not found", id)
	}

	v := raw.Value[0]
	return &WorkItemDetail{
		ID:            v.ID,
		Title:         v.Fields.Title,
		Type:          v.Fields.Type,
		State:         v.Fields.State,
		AreaPath:      v.Fields.AreaPath,
		IterationPath: v.Fields.IterationPath,
		AssignedTo:    v.Fields.AssignedTo.DisplayName,
		CreatedBy:     v.Fields.CreatedBy.DisplayName,
		CreatedDate:   v.Fields.CreatedDate,
		ChangedDate:   v.Fields.ChangedDate,
		Priority:      v.Fields.Priority,
		Tags:          v.Fields.Tags,
		Description:   v.Fields.Description,
	}, nil
}
