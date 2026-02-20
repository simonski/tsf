package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type parseEntity struct {
	Kind     string
	Title    string
	Desc     string
	Priority string
	Labels   string
	AC       string
	Line     int
}

type parseIssue struct {
	Line    int
	Message string
}

func handleParseCommand(args []string) {
	filename := extractFlag(args, "-f")
	if strings.TrimSpace(filename) == "" {
		fmt.Fprintln(os.Stderr, "Error: -f <filename> is required")
		os.Exit(1)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	entities, issues := parseEntitiesMarkdown(string(data))
	printParseSummary(entities, issues)

	if len(issues) > 0 {
		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "line %d: %s\n", issue.Line, issue.Message)
		}
		os.Exit(1)
	}

	commands := renderParseCommands(entities)
	if len(commands) > 0 {
		fmt.Println()
		for _, cmd := range commands {
			fmt.Println(cmd)
			fmt.Println()
		}
	}
}

func parseEntitiesMarkdown(content string) ([]parseEntity, []parseIssue) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	entityKeys := map[string]bool{
		"project": true,
		"epic":    true,
		"task":    true,
		"feature": true,
		"bug":     true,
		"chore":   true,
	}

	var entities []parseEntity
	var issues []parseIssue
	var current *parseEntity
	currentField := ""

	finalize := func() {
		if current == nil {
			return
		}
		if strings.TrimSpace(current.Title) == "" {
			issues = append(issues, parseIssue{Line: current.Line, Message: fmt.Sprintf("%s title is required", current.Kind)})
		} else {
			entities = append(entities, *current)
		}
		current = nil
		currentField = ""
	}

	for idx, raw := range lines {
		lineNo := idx + 1
		line := strings.TrimRight(raw, " \t")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if current != nil && currentField == "ac" {
				if current.AC == "" {
					current.AC = "\n"
				} else {
					current.AC += "\n"
				}
			}
			continue
		}

		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "--") {
			continue
		}

		if sep := strings.Index(line, ":"); sep >= 0 {
			key := strings.ToLower(strings.TrimSpace(line[:sep]))
			val := strings.TrimSpace(line[sep+1:])

			if entityKeys[key] {
				finalize()
				current = &parseEntity{
					Kind:  key,
					Title: val,
					Line:  lineNo,
				}
				currentField = ""
				continue
			}

			if current == nil {
				issues = append(issues, parseIssue{Line: lineNo, Message: fmt.Sprintf("unrecognised key '%s'", key)})
				currentField = ""
				continue
			}

			switch key {
			case "desc", "description":
				current.Desc = val
				currentField = "desc"
			case "p", "priority":
				current.Priority = val
				currentField = ""
			case "label", "labels":
				current.Labels = val
				currentField = ""
			case "ac", "acceptance":
				current.AC = val
				currentField = "ac"
			default:
				issues = append(issues, parseIssue{Line: lineNo, Message: fmt.Sprintf("unrecognised key '%s'", key)})
				currentField = ""
			}
			continue
		}

		if current == nil {
			issues = append(issues, parseIssue{Line: lineNo, Message: "unexpected content outside an entity"})
			continue
		}

		switch currentField {
		case "desc":
			if current.Desc == "" {
				current.Desc = trimmed
			} else {
				current.Desc += "\n" + trimmed
			}
		case "ac":
			if current.AC == "" {
				current.AC = trimmed
			} else {
				current.AC += "\n" + trimmed
			}
		default:
			issues = append(issues, parseIssue{Line: lineNo, Message: "unexpected content; expected key:value"})
		}
	}

	finalize()
	return entities, issues
}

func printParseSummary(entities []parseEntity, issues []parseIssue) {
	countByKind := map[string]int{}
	for _, e := range entities {
		countByKind[e.Kind]++
	}

	order := []string{"project", "epic", "task", "feature", "bug", "chore"}
	display := map[string]string{
		"project": "Project",
		"epic":    "Epic",
		"task":    "Task",
		"feature": "Feature",
		"bug":     "Bug",
		"chore":   "Chore",
	}
	for _, kind := range order {
		if countByKind[kind] > 0 {
			fmt.Printf("%d %s\n", countByKind[kind], display[kind])
		}
	}
	if len(issues) > 0 {
		sort.SliceStable(issues, func(i, j int) bool { return issues[i].Line < issues[j].Line })
		fmt.Printf("%d failed entities\n", len(issues))
		for _, issue := range issues {
			fmt.Printf("  [line %d %s]\n", issue.Line, issue.Message)
		}
	}
}

func renderParseCommands(entities []parseEntity) []string {
	var commands []string
	currentProjectID := ""

	for _, e := range entities {
		switch e.Kind {
		case "project":
			projectID := slugifyIdentifier(e.Title)
			currentProjectID = projectID
			cmd := fmt.Sprintf("sf project create -project_id %q -name %q", projectID, e.Title)
			if strings.TrimSpace(e.Desc) != "" {
				cmd += fmt.Sprintf(" -description %q", e.Desc)
			}
			commands = append(commands, cmd)
		case "epic", "task", "feature", "bug", "chore":
			base := "sf task create"
			switch e.Kind {
			case "epic":
				base = "sf epic create"
			case "bug":
				base = "sf bug create"
			case "chore":
				base = "sf chore create"
			case "feature":
				// feature maps to a normal task in current CLI entity types.
				base = "sf task create"
			}

			cmd := fmt.Sprintf("%s -title %q", base, e.Title)
			if strings.TrimSpace(e.Desc) != "" {
				cmd += fmt.Sprintf(" -description %q", e.Desc)
			}
			if strings.TrimSpace(e.AC) != "" {
				cmd += fmt.Sprintf(" -acceptance_criteria %q", e.AC)
			}
			if strings.TrimSpace(e.Priority) != "" {
				cmd += fmt.Sprintf(" -priority %q", e.Priority)
			}
			if strings.TrimSpace(e.Labels) != "" {
				cmd += fmt.Sprintf(" -labels %q", e.Labels)
			}
			if currentProjectID != "" {
				cmd += fmt.Sprintf(" -project_id %q", currentProjectID)
			}
			commands = append(commands, cmd)
		}
	}

	return commands
}
