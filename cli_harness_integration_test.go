package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/simonski/task/internal/cliharness"
	"github.com/simonski/task/internal/db"
	"github.com/simonski/task/internal/server"
)

func TestCLIHarnessFromMarkdown(t *testing.T) {
	repoRoot, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sf")
	dbPath := filepath.Join(tmpDir, "cli-harness.db")
	specPath := filepath.Join(repoRoot, "docs", "CLI_HARNESS.md")
	reportPath := filepath.Join(repoRoot, "docs", "generated", "CLI_HARNESS_REPORT.md")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	_, _, _, _, err = database.InitializeDatabaseWithPasswords("admin")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	srv := server.New(database)
	ts := newTestServer(t, srv)

	if err := buildBinary(repoRoot, binaryPath); err != nil {
		t.Fatalf("build sf binary: %v", err)
	}

	scenarios, err := cliharness.ParseSpec(specPath)
	if err != nil {
		t.Fatalf("parse harness spec: %v", err)
	}

	runner := cliharness.Runner{
		BinaryPath: binaryPath,
		Timeout:    5 * time.Second,
		Env: map[string]string{
			"SF_URL":      ts.URL,
			"SF_USERNAME": "admin",
			"SF_PASSWORD": "admin",
			"HOME":        tmpDir,
		},
	}

	results, _ := runner.Run(scenarios)
	if err := cliharness.WriteReport(reportPath, specPath, results); err != nil {
		t.Fatalf("write report: %v", err)
	}

	failed := 0
	for _, result := range results {
		if !result.Passed {
			failed++
			t.Logf("FAILED: %s (%s): %s", result.Scenario.Name, result.Expanded, result.Failure)
		}
	}
	if failed > 0 {
		t.Fatalf("cli harness failed: %d/%d scenarios failed (see %s)", failed, len(results), reportPath)
	}
}
