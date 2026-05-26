package azdevops_test

import (
	"net/http"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func TestNewClient_WithToken(t *testing.T) {
	srv := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Errorf("expected Authorization header, got none")
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	c := azdevops.NewClientWithToken(srv.URL, "myorg", "myproject", "test-token")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.Organization() != "myorg" {
		t.Errorf("organization: got %q, want %q", c.Organization(), "myorg")
	}
	if c.Project() != "myproject" {
		t.Errorf("project: got %q, want %q", c.Project(), "myproject")
	}
}

func TestNewClient_BaseURL(t *testing.T) {
	c := azdevops.NewClientWithToken("https://dev.azure.com", "myorg", "myproject", "tok")
	if c.BaseURL() != "https://dev.azure.com" {
		t.Errorf("base url: got %q", c.BaseURL())
	}
}
