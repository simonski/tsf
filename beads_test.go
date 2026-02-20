package main

import "testing"

func TestParseBeadsMarkdownParsesBlocksAndMultiline(t *testing.T) {
	input := `--------------------------------------------------------------------------------
id: bd-1
title: Epic One
desc: top line
  detail line
priority: 1
type: epic
ac: first check
second check
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
title: Task Two
description: work item
priority: 2
type: task
acceptance_criteria: done
--------------------------------------------------------------------------------
`

	records, err := parseBeadsMarkdown(input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].ID != "bd-1" {
		t.Fatalf("unexpected first id: %q", records[0].ID)
	}
	if records[0].Type != "epic" {
		t.Fatalf("unexpected first type: %q", records[0].Type)
	}
	if records[0].Desc != "top line\n  detail line" {
		t.Fatalf("unexpected first desc: %q", records[0].Desc)
	}
	if records[0].AC != "first check\nsecond check" {
		t.Fatalf("unexpected first ac: %q", records[0].AC)
	}
	if records[1].Title != "Task Two" {
		t.Fatalf("unexpected second title: %q", records[1].Title)
	}
}

func TestValidateBeadRecords(t *testing.T) {
	records := []beadRecord{
		{Title: "", Desc: "", Priority: "5", Type: "bug", AC: ""},
		{Title: "ok", Desc: "d", Priority: "2", Type: "task", AC: "a"},
	}

	errs := validateBeadRecords(records)
	if len(errs) != 1 {
		t.Fatalf("expected 1 validation error entry, got %d", len(errs))
	}
	if errs[0].Index != 0 {
		t.Fatalf("unexpected index: %d", errs[0].Index)
	}
	if len(errs[0].Issues) < 5 {
		t.Fatalf("expected multiple issues, got %v", errs[0].Issues)
	}
}

func TestContainsEpic(t *testing.T) {
	if containsEpic([]beadRecord{{Type: "task"}}) {
		t.Fatalf("expected false when no epic present")
	}
	if !containsEpic([]beadRecord{{Type: "task"}, {Type: "epic"}}) {
		t.Fatalf("expected true when epic present")
	}
}
