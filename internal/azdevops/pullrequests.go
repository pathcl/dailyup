package azdevops

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// PullRequest represents an Azure DevOps pull request.
type PullRequest struct {
	ID           int
	Title        string
	Status       string
	CreationDate time.Time
	RepoName     string
}

type prListResponse struct {
	Value []struct {
		ID    int    `json:"pullRequestId"`
		Title string `json:"title"`
		// completed | active | abandoned
		Status       string `json:"status"`
		CreationDate string `json:"creationDate"`
		Repository   struct {
			Name string `json:"name"`
		} `json:"repository"`
	} `json:"value"`
}

// FetchPullRequests returns PRs created by the authenticated user since `since`.
func FetchPullRequests(c *Client, since time.Time) ([]PullRequest, error) {
	url := fmt.Sprintf(
		"%s/_apis/git/pullrequests?searchCriteria.status=all&searchCriteria.queryTimeRangeType=created&api-version=7.1",
		c.ProjectURL(),
	)
	resp, err := c.HTTP().Get(url)
	if err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pull requests HTTP %d: %s", resp.StatusCode, b)
	}

	var result prListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode pull requests: %w", err)
	}

	var prs []PullRequest
	for _, v := range result.Value {
		created, err := time.Parse(time.RFC3339, v.CreationDate)
		if err != nil {
			created, err = time.Parse("2006-01-02T15:04:05.999999999Z", v.CreationDate)
			if err != nil {
				slog.Warn("skipping PR with unparseable date", "id", v.ID, "title", v.Title, "date", v.CreationDate)
				continue
			}
		}
		if created.Before(since) {
			continue
		}
		prs = append(prs, PullRequest{
			ID:           v.ID,
			Title:        v.Title,
			Status:       v.Status,
			CreationDate: created,
			RepoName:     v.Repository.Name,
		})
	}
	return prs, nil
}
