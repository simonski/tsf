package main

import (
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/simonski/task/internal/server"
)

func buildBinary(repoRoot, outPath string) error {
	cmd := exec.Command("go", "build", "-o", outPath, ".")
	cmd.Dir = repoRoot
	return cmd.Run()
}

func newTestServer(t *testing.T, srv *server.Server) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(srv.Router())
	t.Cleanup(ts.Close)
	return ts
}
