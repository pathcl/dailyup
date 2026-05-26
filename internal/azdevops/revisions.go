package azdevops

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

type revisionsResponse struct {
	Value []struct {
		Fields struct {
			IterationPath string `json:"System.IterationPath"`
		} `json:"fields"`
	} `json:"value"`
}

// fetchIterationCounts fetches revision history for each work item concurrently
// and returns a map of item ID → number of distinct sprints the item has been in.
func fetchIterationCounts(c *Client, ids []int) map[int]int {
	counts := make(map[int]int, len(ids))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			n, err := iterationCount(c, id)
			if err != nil {
				slog.Warn("could not fetch revisions", "id", id, "error", err)
				return
			}
			mu.Lock()
			counts[id] = n
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return counts
}

// iterationCount returns the number of distinct iteration paths a work item has had.
func iterationCount(c *Client, id int) (int, error) {
	url := fmt.Sprintf(
		"%s/_apis/wit/workitems/%d/revisions?$select=System.IterationPath&api-version=7.1",
		c.OrgURL(), id,
	)
	resp, err := c.HTTP().Get(url)
	if err != nil {
		return 0, fmt.Errorf("revisions for %d: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("revisions HTTP %d for item %d: %s", resp.StatusCode, id, b)
	}

	var result revisionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode revisions for %d: %w", id, err)
	}

	seen := make(map[string]struct{})
	for _, rev := range result.Value {
		if p := rev.Fields.IterationPath; p != "" {
			seen[p] = struct{}{}
		}
	}
	return len(seen), nil
}
