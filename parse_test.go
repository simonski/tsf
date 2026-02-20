package main

import (
	"strings"
	"testing"
)

func TestParseEntitiesMarkdownValid(t *testing.T) {
	input := `
epic: my epic
desc: the description of the epic

task: my task
desc: it does a thing
ac: my acceptance criteria
is such that
- things occur
- then they dont

task: my other task
desc: is a thing that
ac:
`

	entities, issues := parseEntitiesMarkdown(input)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
	if len(entities) != 3 {
		t.Fatalf("expected 3 entities, got %d", len(entities))
	}

	cmds := renderParseCommands(entities)
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if !strings.Contains(cmds[0], `sf epic create -title "my epic"`) {
		t.Fatalf("unexpected first command: %s", cmds[0])
	}
	if !strings.Contains(cmds[1], `sf task create -title "my task"`) {
		t.Fatalf("unexpected second command: %s", cmds[1])
	}
	if !strings.Contains(cmds[1], `-acceptance_criteria "my acceptance criteria`) {
		t.Fatalf("expected multiline acceptance criteria in second command: %s", cmds[1])
	}
}

func TestParseEntitiesMarkdownMalformed(t *testing.T) {
	input := `
epic: my epic
desc: the description of the epic

task: my task
desc: it does a thing
ac: my acceptance criteria
is such that
- things occcur

taks: my other task
desc: is a thing that
ac:
`

	entities, issues := parseEntitiesMarkdown(input)
	if len(entities) != 2 {
		t.Fatalf("expected 2 valid entities, got %d", len(entities))
	}
	if len(issues) == 0 {
		t.Fatalf("expected parse issues, got none")
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "unrecognised key 'taks'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unrecognised key issue for taks, got: %+v", issues)
	}
}
