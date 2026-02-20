package cliharness

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const delimiter = "--------------------------------------------------------------------------------"

type CaptureRule struct {
	Name  string
	Regex string
}

type Scenario struct {
	Name           string
	Command        string
	ExpectedExit   int
	StdoutContains []string
	StderrContains []string
	CaptureRules   []CaptureRule
}

type Result struct {
	Scenario   Scenario
	Expanded   string
	ActualExit int
	Stdout     string
	Stderr     string
	Passed     bool
	Failure    string
}

type Runner struct {
	BinaryPath string
	Env        map[string]string
	Timeout    time.Duration
}

func ParseSpec(path string) ([]Scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inBlock := false
	current := map[string]string{}
	var scenarios []Scenario

	flush := func() error {
		if len(current) == 0 {
			return nil
		}
		s, err := scenarioFromMap(current)
		if err != nil {
			return err
		}
		scenarios = append(scenarios, s)
		current = map[string]string{}
		return nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == delimiter {
			if inBlock {
				if err := flush(); err != nil {
					return nil, err
				}
				inBlock = false
			} else {
				inBlock = true
				current = map[string]string{}
			}
			continue
		}
		if !inBlock || line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		current[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inBlock {
		return nil, fmt.Errorf("unterminated scenario block in %s", path)
	}
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("no scenarios found in %s", path)
	}
	return scenarios, nil
}

func scenarioFromMap(m map[string]string) (Scenario, error) {
	name := m["name"]
	if name == "" {
		return Scenario{}, fmt.Errorf("scenario missing name")
	}
	cmd := m["cmd"]
	if cmd == "" {
		return Scenario{}, fmt.Errorf("scenario %q missing cmd", name)
	}
	exit := 0
	if raw := m["exit"]; raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Scenario{}, fmt.Errorf("scenario %q has invalid exit %q", name, raw)
		}
		exit = parsed
	}
	return Scenario{
		Name:           name,
		Command:        cmd,
		ExpectedExit:   exit,
		StdoutContains: splitList(m["stdout"]),
		StderrContains: splitList(m["stderr"]),
		CaptureRules:   parseCaptureRules(m["capture"]),
	}, nil
}

func splitList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ";;")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseCaptureRules(raw string) []CaptureRule {
	items := splitList(raw)
	out := make([]CaptureRule, 0, len(items))
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		re := strings.TrimSpace(parts[1])
		if name == "" || re == "" {
			continue
		}
		out = append(out, CaptureRule{Name: name, Regex: re})
	}
	return out
}

func (r Runner) Run(scenarios []Scenario) ([]Result, map[string]string) {
	vars := map[string]string{}
	results := make([]Result, 0, len(scenarios))

	for _, scenario := range scenarios {
		expanded := replaceVars(scenario.Command, vars)
		stdout, stderr, code := r.exec(expanded)
		result := Result{
			Scenario:   scenario,
			Expanded:   expanded,
			ActualExit: code,
			Stdout:     stdout,
			Stderr:     stderr,
		}

		result.Passed, result.Failure = evaluate(scenario, stdout, stderr, code, vars)
		if result.Passed {
			for _, c := range scenario.CaptureRules {
				re := regexp.MustCompile(c.Regex)
				match := re.FindStringSubmatch(stdout + "\n" + stderr)
				if len(match) < 2 {
					result.Passed = false
					result.Failure = fmt.Sprintf("capture %q failed with regex %q", c.Name, c.Regex)
					break
				}
				vars[c.Name] = match[1]
			}
		}

		results = append(results, result)
	}

	return results, vars
}

func (r Runner) exec(command string) (string, string, int) {
	if strings.HasPrefix(command, "sf ") {
		command = shellQuote(r.BinaryPath) + " " + strings.TrimPrefix(command, "sf ")
	} else if command == "sf" {
		command = shellQuote(r.BinaryPath)
	}

	timeout := r.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	env := os.Environ()
	for k, v := range r.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		exitCode = 124
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func evaluate(s Scenario, stdout, stderr string, code int, vars map[string]string) (bool, string) {
	if code != s.ExpectedExit {
		return false, fmt.Sprintf("expected exit %d, got %d", s.ExpectedExit, code)
	}
	for _, expected := range s.StdoutContains {
		expected = replaceVars(expected, vars)
		if !strings.Contains(stdout, expected) {
			return false, fmt.Sprintf("stdout missing %q", expected)
		}
	}
	for _, expected := range s.StderrContains {
		expected = replaceVars(expected, vars)
		if !strings.Contains(stderr, expected) {
			return false, fmt.Sprintf("stderr missing %q", expected)
		}
	}
	return true, ""
}

func replaceVars(input string, vars map[string]string) string {
	out := input
	for k, v := range vars {
		out = strings.ReplaceAll(out, "${"+k+"}", v)
	}
	return out
}

func WriteReport(path, specPath string, results []Result) error {
	var b strings.Builder
	b.WriteString("# CLI Harness Report\n\n")
	b.WriteString("Source: `" + specPath + "`\n\n")
	b.WriteString("--------------------------------------------------------------------------------\n")
	b.WriteString("name: summary\n")
	b.WriteString(fmt.Sprintf("total: %d\n", len(results)))
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	b.WriteString(fmt.Sprintf("passed: %d\n", passed))
	b.WriteString(fmt.Sprintf("failed: %d\n", len(results)-passed))
	b.WriteString("--------------------------------------------------------------------------------\n\n")

	for _, r := range results {
		b.WriteString("--------------------------------------------------------------------------------\n")
		b.WriteString("name: " + r.Scenario.Name + "\n")
		b.WriteString("cmd: " + r.Scenario.Command + "\n")
		b.WriteString(fmt.Sprintf("exit: %d\n", r.Scenario.ExpectedExit))
		if len(r.Scenario.StdoutContains) > 0 {
			b.WriteString("stdout: " + strings.Join(r.Scenario.StdoutContains, ";;") + "\n")
		}
		if len(r.Scenario.StderrContains) > 0 {
			b.WriteString("stderr: " + strings.Join(r.Scenario.StderrContains, ";;") + "\n")
		}
		if len(r.Scenario.CaptureRules) > 0 {
			captures := make([]string, 0, len(r.Scenario.CaptureRules))
			for _, c := range r.Scenario.CaptureRules {
				captures = append(captures, c.Name+"="+c.Regex)
			}
			b.WriteString("capture: " + strings.Join(captures, ";;") + "\n")
		}
		b.WriteString("actual_cmd: " + r.Expanded + "\n")
		b.WriteString(fmt.Sprintf("actual_exit: %d\n", r.ActualExit))
		if r.Failure != "" {
			b.WriteString("failure: " + sanitizeLine(r.Failure) + "\n")
		}
		b.WriteString("pass: " + strconv.FormatBool(r.Passed) + "\n")
		if strings.TrimSpace(r.Stdout) != "" {
			b.WriteString("actual_stdout: " + sanitizeLine(snippet(r.Stdout)) + "\n")
		}
		if strings.TrimSpace(r.Stderr) != "" {
			b.WriteString("actual_stderr: " + sanitizeLine(snippet(r.Stderr)) + "\n")
		}
		b.WriteString("--------------------------------------------------------------------------------\n\n")
	}

	if err := os.MkdirAll(filepathDir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func snippet(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " ⏎ "))
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

func sanitizeLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func filepathDir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return "."
	}
	return path[:idx]
}
