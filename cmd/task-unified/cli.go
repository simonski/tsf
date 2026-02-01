package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/simonski/task/internal/cli"
)

func cliMain() {
	if len(os.Args) < 2 {
		printCLIUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	// Parse config from all remaining args
	config, err := cli.NewConfig(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "Error: Username and password required")
		fmt.Fprintln(os.Stderr, "Set TASK_USERNAME and TASK_PASSWORD environment variables")
		fmt.Fprintln(os.Stderr, "Or use -username and -password flags")
		os.Exit(1)
	}

	client := cli.NewClient(config)

	// Extract subcommand and remaining args (skip flags)
	subArgs := filterNonFlags(os.Args[2:])

	switch command {
	case "project":
		handleProjectCommand(client, config, subArgs)
	case "task":
		handleTaskCommand(client, config, subArgs)
	case "user":
		handleUserCommand(client, config, subArgs)
	case "role":
		handleRoleCommand(client, config, subArgs)
	case "config":
		handleConfigCommand(client, config, subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printCLIUsage()
		os.Exit(1)
	}
}

func filterNonFlags(args []string) []string {
	var result []string
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-url" || arg == "-username" || arg == "-password" {
			skipNext = true
			continue
		}
		if arg == "-json" {
			continue
		}
		result = append(result, arg)
	}
	return result
}

func printCLIUsage() {
	fmt.Println("Task CLI Commands:")
	fmt.Println()
	fmt.Println("  task project list              List all projects")
	fmt.Println("  task project get -id <id>      Get project details")
	fmt.Println()
	fmt.Println("  task task list                 List all tasks")
	fmt.Println("  task task get -id <id>         Get task details")
	fmt.Println()
	fmt.Println("  task user list                 List all users")
	fmt.Println()
	fmt.Println("  task role list                 List all roles")
	fmt.Println()
	fmt.Println("  task config list               List all config")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  -url <url>        Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>  Username (or set TASK_USERNAME)")
	fmt.Println("  -password <pass>  Password (or set TASK_PASSWORD)")
	fmt.Println("  -json             Output as JSON")
}

func handleProjectCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: project subcommand required (list, get)")
		os.Exit(1)
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		data, err := client.Request("GET", "/api/v1/projects", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var projects []map[string]interface{}
			if err := parseJSON(data, &projects); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d projects:\n\n", len(projects))
			for _, p := range projects {
				fmt.Printf("ID:   %v\n", p["id"])
				fmt.Printf("Name: %v\n", p["name"])
				if desc, ok := p["description"]; ok && desc != nil {
					fmt.Printf("Desc: %v\n", desc)
				}
				fmt.Println()
			}
		}
	case "get":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/projects/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	default:
		fmt.Fprintf(os.Stderr, "Unknown project subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleTaskCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: task subcommand required (list, get)")
		os.Exit(1)
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		data, err := client.Request("GET", "/api/v1/tasks", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var tasks []map[string]interface{}
			if err := parseJSON(data, &tasks); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d tasks:\n\n", len(tasks))
			for _, t := range tasks {
				fmt.Printf("ID:     %v\n", t["id"])
				fmt.Printf("Title:  %v\n", t["title"])
				fmt.Printf("Status: %v\n", t["status"])
				if priority, ok := t["priority"]; ok && priority != nil {
					fmt.Printf("Priority: %v\n", priority)
				}
				fmt.Println()
			}
		}
	case "get":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/tasks/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	default:
		fmt.Fprintf(os.Stderr, "Unknown task subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleUserCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		data, err := client.Request("GET", "/api/v1/users", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var users []map[string]interface{}
			if err := parseJSON(data, &users); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d users:\n\n", len(users))
			for _, u := range users {
				fmt.Printf("Username: %v\n", u["username"])
				fmt.Printf("Type:     %v\n", u["type"])
				fmt.Printf("Active:   %v\n", u["is_active"])
				fmt.Println()
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown user subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleRoleCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		data, err := client.Request("GET", "/api/v1/roles", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var roles []map[string]interface{}
			if err := parseJSON(data, &roles); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d roles:\n\n", len(roles))
			for _, r := range roles {
				fmt.Printf("ID:   %v\n", r["id"])
				fmt.Printf("Name: %v\n", r["name"])
				fmt.Println()
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown role subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleConfigCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		data, err := client.Request("GET", "/api/v1/config", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	default:
		fmt.Fprintf(os.Stderr, "Unknown config subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func extractFlag(args []string, flag string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

