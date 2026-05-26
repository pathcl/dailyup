package azdevops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WorkItemType describes a work item type available in the project.
type WorkItemType struct {
	Name        string
	Description string
}

type workItemTypesResponse struct {
	Value []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"value"`
}

// FetchWorkItemTypes returns all work item types configured for the project.
func FetchWorkItemTypes(c *Client) ([]WorkItemType, error) {
	url := fmt.Sprintf("%s/_apis/wit/workitemtypes?api-version=7.1", c.ProjectURL())
	resp, err := c.HTTP().Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch work item types: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("work item types HTTP %d: %s", resp.StatusCode, b)
	}
	var result workItemTypesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode work item types: %w", err)
	}
	types := make([]WorkItemType, len(result.Value))
	for i, v := range result.Value {
		types[i] = WorkItemType{Name: v.Name, Description: v.Description}
	}
	return types, nil
}
