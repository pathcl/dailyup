// Package azdevops provides a thin client for the Azure DevOps REST API (v7.1).
// Authentication uses the Azure CLI credential (az login); no PAT is required.
// The package is split into focused files by resource type: workitems, pullrequests,
// commits, and revisions — all sharing the single *Client defined here.
package azdevops

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
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
	debug        bool
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
func NewClientFromAzCLI(organization, project string, debug bool) (*Client, error) {
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
	var inner http.RoundTripper = &authTransport{
		wrapped: http.DefaultTransport,
		header:  "Bearer " + tok.Token,
	}
	if debug {
		inner = &debugTransport{wrapped: inner}
	}
	return &Client{
		httpClient:   &http.Client{Transport: inner},
		baseURL:      "https://dev.azure.com",
		organization: organization,
		project:      project,
		debug:        debug,
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
	if clone.Header.Get("Content-Type") == "" {
		clone.Header.Set("Content-Type", "application/json")
	}
	return t.wrapped.RoundTrip(clone)
}

type debugTransport struct {
	wrapped http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	slog.Debug("http request", "method", req.Method, "url", req.URL.String())

	resp, err := t.wrapped.RoundTrip(req)
	if err != nil {
		slog.Debug("http error", "error", err)
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	slog.Debug("http response", "status", resp.StatusCode, "body", string(body))

	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}
