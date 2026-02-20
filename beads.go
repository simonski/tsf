package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var validBeadTypes = map[string]bool{
	"epic":  true,
	"task":  true,
	"chore": true,
}

const defaultBeadsFile = "TODO.md"

type beadRecord struct {
	ID       string
	Title    string
	Desc     string
	Priority string
	Type     string
	AC       string
}

type beadValidationError struct {
	Index  int
	Issues []string
}

func handleBeadsCommand(args []string) {
	if len(args) == 0 {
		runExternalCommand("bd", []string{"list", "--pretty"}, false)
		return
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "list", "ls":
		filename := getBeadsFilename(args)
		records, err := parseBeadsMarkdownFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printBeadRecords(records)
	case "format", "fmt", "tidy", "fix":
		filename := getBeadsFilename(args)
		records, err := parseBeadsMarkdownFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := writeBeadsMarkdownFile(filename, records); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Formatted %d bead entries in %s\n", len(records), filename)
	case "test", "validate":
		filename := getBeadsFilename(args)
		records, err := parseBeadsMarkdownFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		validationErrors := validateBeadRecords(records)
		if len(validationErrors) > 0 {
			for _, issue := range validationErrors {
				fmt.Fprintf(os.Stderr, "Entry %d invalid:\n", issue.Index+1)
				for _, msg := range issue.Issues {
					fmt.Fprintf(os.Stderr, "  - %s\n", msg)
				}
			}
			os.Exit(1)
		}
		fmt.Printf("Validation passed: %d bead entries\n", len(records))
	case "import":
		filename := getBeadsFilename(args)
		apply := hasFlag(args, "-apply")
		records, err := parseBeadsMarkdownFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		validationErrors := validateBeadRecords(records)
		if len(validationErrors) > 0 {
			for _, issue := range validationErrors {
				fmt.Fprintf(os.Stderr, "Entry %d invalid:\n", issue.Index+1)
				for _, msg := range issue.Issues {
					fmt.Fprintf(os.Stderr, "  - %s\n", msg)
				}
			}
			os.Exit(1)
		}
		if !containsEpic(records) {
			fmt.Fprintln(os.Stderr, "Error: import requires at least one epic entry")
			os.Exit(1)
		}

		commands := renderImportCommands(records)
		for _, cmd := range commands {
			fmt.Println(cmd)
		}

		if apply {
			createdOrUpdated := 0
			for i := range records {
				id, wasApplied, err := applyBeadRecord(&records[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error applying entry %d: %v\n", i+1, err)
					os.Exit(1)
				}
				if wasApplied {
					createdOrUpdated++
				}
				if id != "" {
					records[i].ID = id
				}
			}
			if err := writeBeadsMarkdownFile(filename, records); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing IDs back to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Applied %d bead entries and updated %s\n", createdOrUpdated, filename)
		}
	case "export":
		filename := getBeadsFilename(args)
		records, err := exportBeadRecordsFromBD()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := writeBeadsMarkdownFile(filename, records); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported %d bead entries to %s\n", len(records), filename)
	default:
		fmt.Fprintf(os.Stderr, "Unknown beads subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func getBeadsFilename(args []string) string {
	if filename := extractFlag(args, "-f"); filename != "" {
		return filename
	}
	return defaultBeadsFile
}

func parseBeadsMarkdownFile(filename string) ([]beadRecord, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseBeadsMarkdown(string(data))
}

func parseBeadsMarkdown(content string) ([]beadRecord, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	var records []beadRecord
	inBlock := false
	var block []string
	for _, line := range lines {
		if isBeadDelimiter(line) {
			if inBlock {
				records = append(records, parseBeadBlock(block))
				block = nil
				inBlock = false
			} else {
				inBlock = true
				block = nil
			}
			continue
		}
		if inBlock {
			block = append(block, line)
		}
	}

	if inBlock {
		return nil, fmt.Errorf("unterminated bead block; expected closing delimiter")
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no bead blocks found")
	}

	return records, nil
}

func isBeadDelimiter(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	for _, r := range trimmed {
		if r != '-' {
			return false
		}
	}
	return true
}

func parseBeadBlock(lines []string) beadRecord {
	record := beadRecord{}
	currentMultiline := ""

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\t ")
		if strings.TrimSpace(line) == "" {
			if currentMultiline == "desc" {
				record.Desc += "\n"
			} else if currentMultiline == "ac" {
				record.AC += "\n"
			}
			continue
		}

		if idx := strings.Index(line, ":"); idx >= 0 {
			key := strings.ToLower(strings.TrimSpace(line[:idx]))
			value := strings.TrimSpace(line[idx+1:])
			switch key {
			case "id":
				record.ID = value
				currentMultiline = ""
			case "title":
				record.Title = value
				currentMultiline = ""
			case "desc", "description":
				record.Desc = value
				currentMultiline = "desc"
			case "priority":
				record.Priority = value
				currentMultiline = ""
			case "type":
				record.Type = strings.ToLower(value)
				currentMultiline = ""
			case "ac", "acceptance", "acceptance_criteria":
				record.AC = value
				currentMultiline = "ac"
			default:
				currentMultiline = ""
			}
			continue
		}

		if currentMultiline == "desc" {
			if record.Desc == "" {
				record.Desc = line
			} else {
				record.Desc += "\n" + line
			}
		}
		if currentMultiline == "ac" {
			if record.AC == "" {
				record.AC = line
			} else {
				record.AC += "\n" + line
			}
		}
	}

	record.Desc = strings.TrimSuffix(record.Desc, "\n")
	record.AC = strings.TrimSuffix(record.AC, "\n")
	return record
}

func validateBeadRecords(records []beadRecord) []beadValidationError {
	var errs []beadValidationError
	for i, record := range records {
		var issues []string
		if record.Title == "" {
			issues = append(issues, "missing title")
		}
		if record.Desc == "" {
			issues = append(issues, "missing desc")
		}
		if record.Priority == "" {
			issues = append(issues, "missing priority")
		}
		if record.Type == "" {
			issues = append(issues, "missing type")
		}
		if record.AC == "" {
			issues = append(issues, "missing ac")
		}
		if record.Priority != "" && record.Priority != "1" && record.Priority != "2" && record.Priority != "3" {
			issues = append(issues, "priority must be one of 1, 2, 3")
		}
		if record.Type != "" && !validBeadTypes[record.Type] {
			issues = append(issues, "type must be one of epic, task, chore")
		}
		if len(issues) > 0 {
			errs = append(errs, beadValidationError{Index: i, Issues: issues})
		}
	}
	return errs
}

func containsEpic(records []beadRecord) bool {
	for _, record := range records {
		if record.Type == "epic" {
			return true
		}
	}
	return false
}

func printBeadRecords(records []beadRecord) {
	fmt.Printf("Found %d bead entries:\n\n", len(records))
	for i, record := range records {
		fmt.Printf("%d. [%s] %s\n", i+1, record.Type, record.Title)
		if record.ID != "" {
			fmt.Printf("   ID: %s\n", record.ID)
		}
		fmt.Printf("   Priority: %s\n", record.Priority)
		fmt.Printf("   Desc: %s\n", truncateSingleLine(record.Desc, 80))
		fmt.Printf("   AC: %s\n", truncateSingleLine(record.AC, 80))
		fmt.Println()
	}
}

func truncateSingleLine(value string, max int) string {
	single := strings.ReplaceAll(value, "\n", " ")
	single = strings.TrimSpace(single)
	if len(single) <= max {
		return single
	}
	if max < 4 {
		return single[:max]
	}
	return single[:max-3] + "..."
}

func renderImportCommands(records []beadRecord) []string {
	commands := make([]string, 0, len(records))
	for _, record := range records {
		if record.ID == "" {
			commands = append(commands,
				"bd create "+shellQuote(record.Title)+
					" -t "+record.Type+
					" -p "+record.Priority+
					" -d "+shellQuote(record.Desc)+
					" --acceptance "+shellQuote(record.AC),
			)
			continue
		}
		commands = append(commands,
			"bd update "+shellQuote(record.ID)+
				" --title "+shellQuote(record.Title)+
				" -t "+record.Type+
				" -p "+record.Priority+
				" -d "+shellQuote(record.Desc)+
				" --acceptance "+shellQuote(record.AC),
		)
	}
	return commands
}

func applyBeadRecord(record *beadRecord) (string, bool, error) {
	if record.ID != "" {
		_, _, err := runExternalCommandCapture("bd", []string{"show", record.ID}, true)
		if err == nil {
			_, stderr, updateErr := runExternalCommandCapture(
				"bd",
				[]string{
					"update", record.ID,
					"--title", record.Title,
					"-t", record.Type,
					"-p", record.Priority,
					"-d", record.Desc,
					"--acceptance", record.AC,
				},
				false,
			)
			if updateErr != nil {
				return "", false, fmt.Errorf("bd update failed: %s", strings.TrimSpace(stderr))
			}
			return record.ID, true, nil
		}
	}

	cmdArgs := []string{
		"create", record.Title,
		"-t", record.Type,
		"-p", record.Priority,
		"-d", record.Desc,
		"--acceptance", record.AC,
		"--silent",
	}
	if record.ID != "" {
		cmdArgs = append(cmdArgs, "--id", record.ID)
	}

	stdout, stderr, err := runExternalCommandCapture("bd", cmdArgs, false)
	if err != nil {
		return "", false, fmt.Errorf("bd create failed: %s", strings.TrimSpace(stderr))
	}
	createdID := strings.TrimSpace(stdout)
	if createdID == "" {
		return "", false, fmt.Errorf("bd create returned empty ID")
	}
	return createdID, true, nil
}

func exportBeadRecordsFromBD() ([]beadRecord, error) {
	stdout, stderr, err := runExternalCommandCapture("bd", []string{"export", "--format", "jsonl"}, false)
	if err != nil {
		return nil, fmt.Errorf("bd export failed: %s", strings.TrimSpace(stderr))
	}

	scanner := bufio.NewScanner(strings.NewReader(stdout))
	records := make([]beadRecord, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			return nil, fmt.Errorf("failed to parse export line as JSON: %w", err)
		}
		record := beadRecord{
			ID:       extractAnyString(payload, "id"),
			Title:    extractAnyString(payload, "title", "name"),
			Desc:     extractAnyString(payload, "description", "desc"),
			Priority: extractAnyString(payload, "priority"),
			Type:     strings.ToLower(extractAnyString(payload, "type")),
			AC:       extractAnyString(payload, "acceptance", "acceptance_criteria", "ac"),
		}
		if !validBeadTypes[record.Type] {
			continue
		}
		if record.Priority == "" {
			record.Priority = "2"
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func extractAnyString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			return v
		case float64:
			if float64(int(v)) == v {
				return strconv.Itoa(int(v))
			}
			return fmt.Sprintf("%v", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func writeBeadsMarkdownFile(filename string, records []beadRecord) error {
	var buf bytes.Buffer
	for i, record := range records {
		buf.WriteString("--------------------------------------------------------------------------------\n")
		writeBeadField(&buf, "id", record.ID)
		writeBeadField(&buf, "title", record.Title)
		writeBeadField(&buf, "desc", record.Desc)
		writeBeadField(&buf, "priority", record.Priority)
		writeBeadField(&buf, "type", record.Type)
		writeBeadField(&buf, "ac", record.AC)
		buf.WriteString("--------------------------------------------------------------------------------\n")
		if i < len(records)-1 {
			buf.WriteString("\n")
		}
	}
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

func writeBeadField(buf *bytes.Buffer, key, value string) {
	if strings.Contains(value, "\n") {
		parts := strings.Split(value, "\n")
		buf.WriteString(key + ": " + parts[0] + "\n")
		for _, part := range parts[1:] {
			buf.WriteString(part + "\n")
		}
		return
	}
	if value == "" {
		buf.WriteString(key + ":\n")
		return
	}
	buf.WriteString(key + ": " + value + "\n")
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func runExternalCommand(name string, args []string, silentErrors bool) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if silentErrors {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error running %s: %v\n", name, err)
		os.Exit(1)
	}
}

func runExternalCommandCapture(name string, args []string, silentErrors bool) (string, string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil && !silentErrors {
		return stdout.String(), stderr.String(), err
	}
	if err != nil {
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
