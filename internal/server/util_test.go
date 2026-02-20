package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendErrorLogsServerErrors(t *testing.T) {
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	rr := httptest.NewRecorder()
	sendError(rr, http.StatusInternalServerError, "boom")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(buf.String(), "internal server error") {
		t.Fatalf("expected log output for 5xx, got: %s", buf.String())
	}
}
