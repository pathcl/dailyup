package azdevops

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// azureDevOpsResource is the resource ID used when requesting Azure DevOps tokens.
const azureDevOpsResource = "499b84ac-1321-427f-aa17-267ca6975798"

// Client wraps an HTTP client pre-configured for Azure DevOps REST calls.
type Client struct {
	httpClient   *http.Client
	baseURL      string
	organization string
	project      string
}

// NewClientWithToken creates a Client using a pre-existing token (PAT or Bearer).
func NewClientWithToken(baseURL, organization, project, token string) *Client {
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + token))
	transport := &authTransport{
		wrapped: http.DefaultTransport,
		header:  "Basic " + encoded,
	}
	return &Client{
		httpClient:   &http.Client{Transport: transport},
		baseURL:      baseURL,
		organization: organization,
		project:      project,
	}
}

// NewClientFromAzCLI creates a Client by fetching a token via the Azure CLI credential.
func NewClientFromAzCLI(organization, project string) (*Client, error) {
	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure cli credential: %w", err)
	}
	tok, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{azureDevOpsResource + "/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("get azure devops token (run 'az login' first): %w", err)
	}
	transport := &authTransport{
		wrapped: http.DefaultTransport,
		header:  "Bearer " + tok.Token,
	}
	return &Client{
		httpClient:   &http.Client{Transport: transport},
		baseURL:      "https://dev.azure.com",
		organization: organization,
		project:      project,
	}, nil
}

func (c *Client) Organization() string { return c.organization }
func (c *Client) Project() string      { return c.project }
func (c *Client) BaseURL() string      { return c.baseURL }
func (c *Client) HTTP() *http.Client   { return c.httpClient }

// OrgURL returns the base URL for the organization.
func (c *Client) OrgURL() string {
	return fmt.Sprintf("%s/%s", c.baseURL, c.organization)
}

// ProjectURL returns the base URL for the project.
func (c *Client) ProjectURL() string {
	return fmt.Sprintf("%s/%s/%s", c.baseURL, c.organization, c.project)
}

type authTransport struct {
	wrapped http.RoundTripper
	header  string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", t.header)
	clone.Header.Set("Content-Type", "application/json")
	return t.wrapped.RoundTrip(clone)
}
