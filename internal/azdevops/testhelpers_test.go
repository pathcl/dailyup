package azdevops_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pathcl/dailyup/internal/azdevops"
)

func newMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func newClient(t *testing.T, srv *httptest.Server) *azdevops.Client {
	t.Helper()
	return azdevops.NewClientWithToken(srv.URL, "org", "proj", "tok")
}
