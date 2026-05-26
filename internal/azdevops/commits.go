package azdevops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Commit represents a single git commit.
type Commit struct {
	ShortID  string
	Message  string
	RepoName string
	Date     time.Time
}

type repoListResponse struct {
	Value []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"value"`
}

type commitListResponse struct {
	Value []struct {
		CommitID string `json:"commitId"`
		Comment  string `json:"comment"`
		Author   struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"value"`
}

// FetchCommits returns commits authored by `authorEmail` in all repos since `since`.
func FetchCommits(c *Client, authorEmail string, since time.Time) ([]Commit, error) {
	repos, err := listRepos(c)
	if err != nil {
		return nil, err
	}

	var all []Commit
	for _, repo := range repos {
		commits, err := listCommitsForRepo(c, repo.id, repo.name, authorEmail, since)
		if err != nil {
			return nil, err
		}
		all = append(all, commits...)
	}
	return all, nil
}

type repoInfo struct {
	id   string
	name string
}

func listRepos(c *Client) ([]repoInfo, error) {
	apiURL := fmt.Sprintf("%s/_apis/git/repositories?api-version=7.1", c.ProjectURL())
	resp, err := c.HTTP().Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("list repos: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list repos HTTP %d: %s", resp.StatusCode, b)
	}
	var result repoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	repos := make([]repoInfo, len(result.Value))
	for i, v := range result.Value {
		repos[i] = repoInfo{id: v.ID, name: v.Name}
	}
	return repos, nil
}

func listCommitsForRepo(c *Client, repoID, repoName, authorEmail string, since time.Time) ([]Commit, error) {
	params := url.Values{}
	params.Set("searchCriteria.fromDate", since.Format(time.RFC3339))
	if authorEmail != "" {
		params.Set("searchCriteria.author", authorEmail)
	}
	params.Set("api-version", "7.1")

	apiURL := fmt.Sprintf("%s/_apis/git/repositories/%s/commits?%s", c.ProjectURL(), repoID, params.Encode())
	resp, err := c.HTTP().Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("list commits for repo %s: %w", repoName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("commits HTTP %d for repo %s: %s", resp.StatusCode, repoName, b)
	}
	var result commitListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	commits := make([]Commit, 0, len(result.Value))
	for _, v := range result.Value {
		d, _ := time.Parse(time.RFC3339, v.Author.Date)
		short := v.CommitID
		if len(short) > 7 {
			short = short[:7]
		}
		commits = append(commits, Commit{
			ShortID:  short,
			Message:  firstLine(v.Comment),
			RepoName: repoName,
			Date:     d,
		})
	}
	return commits, nil
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}
