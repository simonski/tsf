package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/simonski/task/internal/cli"
	"github.com/simonski/task/internal/db"
	"github.com/simonski/task/internal/orchestrator"
	"github.com/simonski/task/internal/server"
	"github.com/simonski/task/internal/tui"
	"github.com/simonski/task/internal/web"
	"github.com/simonski/task/internal/worker"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "server":
		runServer(os.Args[2:])
	case "orchestrator":
		runOrchestrator(os.Args[2:])
	case "worker":
		runWorker(os.Args[2:])
	case "initdb":
		runInitDB(os.Args[2:])
	case "tui", "-tui":
		runTUI(os.Args[2:])
	case "login", "register", "project", "task", "user", "role", "config":
		runCLI(os.Args[1:])
	case "version", "-v", "--version":
		fmt.Printf("sf version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("sf - Software Factory")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage: sf <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  server        Start the HTTP server")
	fmt.Println("  orchestrator  Start the orchestrator daemon")
	fmt.Println("  worker        Start a worker daemon")
	fmt.Println("  initdb        Initialize the database")
	fmt.Println("  tui           Start the Terminal User Interface")
	fmt.Println()
	fmt.Println("  login         Login and save credentials")
	fmt.Println("  register      Register a new account")
	fmt.Println()
	fmt.Println("  project       Manage projects")
	fmt.Println("  task          Manage tasks")
	fmt.Println("  user          Manage users")
	fmt.Println("  role          Manage roles")
	fmt.Println("  config        Manage configuration")
	fmt.Println()
	fmt.Println("  version       Show version information")
	fmt.Println("  help          Show this help message")
	fmt.Println()
	fmt.Println("Use 'sf <command> -h' for more information about a command.")
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	dbPath := fs.String("f", "", "Path to database file (default: ~/.config/sf/sf.db)")
	port := fs.Int("port", 8080, "Server port")
	fs.Parse(args)

	finalPath := *dbPath
	if finalPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		finalPath = filepath.Join(home, ".config", "sf", "sf.db")
	}

	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		log.Fatalf("Database does not exist at %s. Please run 'sf initdb' first.", finalPath)
	}

	database, err := db.Open(finalPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	srv := server.New(database)
	srv.SetWebFS(web.FS)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Database: %s", finalPath)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runOrchestrator(args []string) {
	fs := flag.NewFlagSet("orchestrator", flag.ExitOnError)
	serverURL := fs.String("url", "http://localhost:8080", "Server URL")
	username := fs.String("username", os.Getenv("SF_USERNAME"), "Username")
	password := fs.String("password", os.Getenv("SF_PASSWORD"), "Password")
	fs.Parse(args)

	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: SF_USERNAME and SF_PASSWORD required")
		os.Exit(1)
	}

	orch := orchestrator.New(*serverURL, *username, *password)
	if err := orch.Start(); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	orch.Stop()
	log.Println("Orchestrator stopped")
}

func runWorker(args []string) {
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	serverURL := fs.String("url", "http://localhost:8080", "Server URL")
	username := fs.String("username", os.Getenv("SF_USERNAME"), "Username")
	password := fs.String("password", os.Getenv("SF_PASSWORD"), "Password")
	fs.Parse(args)

	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: SF_USERNAME and SF_PASSWORD required")
		os.Exit(1)
	}

	w := worker.New(*serverURL, *username, *password)
	if err := w.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	w.Stop()
	log.Println("Worker stopped")
}

func runInitDB(args []string) {
	fs := flag.NewFlagSet("initdb", flag.ExitOnError)
	dbPath := fs.String("f", "", "Path to database file (default: ~/.config/sf/sf.db)")
	force := fs.Bool("force", false, "Force rebuild database (removes existing database)")
	adminPassword := fs.String("password", "", "Set admin password (if not provided, a random password is generated)")
	fs.Parse(args)

	finalPath := *dbPath
	if finalPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		finalPath = filepath.Join(home, ".config", "sf", "sf.db")
	}

	// Check if database exists
	if _, err := os.Stat(finalPath); err == nil {
		if !*force {
			fmt.Fprintf(os.Stderr, "Error: Database already exists at %s\n", finalPath)
			fmt.Fprintf(os.Stderr, "Please remove it first or use --force to rebuild\n")
			os.Exit(1)
		}
		// Remove existing database
		if err := os.Remove(finalPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to remove existing database: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed existing database at %s\n", finalPath)
	}

	database, err := db.Open(finalPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	actualAdminPassword, orchestratorPassword, err := database.InitializeDatabaseWithPasswords(*adminPassword, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
		os.Remove(finalPath)
		os.Exit(1)
	}

	fmt.Println("Database initialized successfully!")
	fmt.Printf("Location: %s\n\n", finalPath)
	fmt.Println("Admin credentials:")
	fmt.Println("  Username: admin")
	fmt.Printf("  Password: %s\n\n", actualAdminPassword)
	fmt.Println("Orchestrator credentials:")
	fmt.Println("  Username: orchestrator")
	fmt.Printf("  Password: %s\n\n", orchestratorPassword)
	fmt.Println("IMPORTANT: Save these credentials securely. They cannot be recovered.")
}

func runTUI(args []string) {
	// Parse command line args for config
	config, err := cli.NewConfig(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Run the TUI
	if err := tui.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(args []string) {
	if len(args) < 1 {
		printCLIUsage()
		os.Exit(0)
	}

	command := args[0]

	// Parse config from all remaining args
	config, err := cli.NewConfig(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Login and register don't require prior authentication
	if command == "login" {
		handleLoginCommand(config)
		return
	}
	if command == "register" {
		handleRegisterCommand(config)
		return
	}

	if config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "Error: Username and password required")
		fmt.Fprintln(os.Stderr, "Set SF_USERNAME and SF_PASSWORD environment variables")
		fmt.Fprintln(os.Stderr, "Or use -username and -password flags")
		fmt.Fprintln(os.Stderr, "Or run 'sf login' to authenticate")
		os.Exit(1)
	}

	client := cli.NewClient(config)

	// Extract subcommand and remaining args (skip flags)
	subArgs := filterNonFlags(args[1:])

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
	fmt.Println("SF CLI Commands:")
	fmt.Println()
	fmt.Println("Authentication:")
	fmt.Println("  sf login                     Login and save credentials")
	fmt.Println("  sf register                  Register a new account")
	fmt.Println()
	fmt.Println("  sf project list              List all projects")
	fmt.Println("  sf project get -id <id>      Get project details")
	fmt.Println("  sf project set <name>        Set active project")
	fmt.Println("  sf project unset             Unset active project")
	fmt.Println("  sf project create -name <n>  Create new project")
	fmt.Println()
	fmt.Println("  sf task list                 List all tasks")
	fmt.Println("  sf task get -id <id>         Get task details")
	fmt.Println("  sf task create -title <t>    Create new task")
	fmt.Println("  sf task update -id <id> ...  Update task")
	fmt.Println("  sf task delete -id <id>      Delete task")
	fmt.Println("  sf task claim -id <id>       Claim task")
	fmt.Println("  sf task free -id <id>        Free task")
	fmt.Println("  sf task assign -id <id> -u   Assign task to user")
	fmt.Println("  sf task complete -id <id>    Complete task")
	fmt.Println("  sf task block -id <id> -by B Block task by another task")
	fmt.Println("  sf task unblock -id <id>     Remove task blocking")
	fmt.Println("  sf task history -id <id>     Show task history")
	fmt.Println("  sf task deps -id <id>        Show task dependencies")
	fmt.Println()
	fmt.Println("  sf user list                 List all users")
	fmt.Println("  sf user create -u <name>     Create user")
	fmt.Println("  sf user enable -u <name>     Enable user")
	fmt.Println("  sf user disable -u <name>    Disable user")
	fmt.Println()
	fmt.Println("  sf role list                 List all roles")
	fmt.Println("  sf role get -id <id>         Get role details")
	fmt.Println("  sf role create -name <n>     Create role")
	fmt.Println()
	fmt.Println("  sf config list               List all config")
	fmt.Println("  sf config set -key K -val V  Set config value")
	fmt.Println("  sf config delete -key K      Delete config value")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  -url <url>        Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>  Username (or set SF_USERNAME)")
	fmt.Println("  -password <pass>  Password (or set SF_PASSWORD)")
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
	case "create":
		name := extractFlag(args, "-name")
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: -name flag required")
			os.Exit(1)
		}
		desc := extractFlag(args, "-description")
		body := map[string]string{"name": name}
		if desc != "" {
			body["description"] = desc
		}
		data, err := client.Request("POST", "/api/v1/projects", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project created:")
		fmt.Println(string(data))
	case "set":
		// Set active project context
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: project name required")
			os.Exit(1)
		}
		projectName := args[1]
		// Find project by name
		data, err := client.Request("GET", "/api/v1/projects", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		var projects []map[string]interface{}
		if err := parseJSON(data, &projects); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
			os.Exit(1)
		}
		var projectID string
		for _, p := range projects {
			if name, ok := p["name"].(string); ok && name == projectName {
				if id, ok := p["id"].(string); ok {
					projectID = id
					break
				}
			}
		}
		if projectID == "" {
			fmt.Fprintf(os.Stderr, "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}
		if err := cli.SaveProjectContext(projectID); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving project context: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Active project set to: %s (%s)\n", projectName, projectID)
	case "unset":
		if err := cli.SaveProjectContext(""); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing project context: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Active project cleared")
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
	case "create":
		title := extractFlag(args, "-title")
		if title == "" {
			fmt.Fprintln(os.Stderr, "Error: -title flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{"title": title}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		} else {
			body["description"] = title // Use title as default description
		}
		if projectID := extractFlag(args, "-project_id"); projectID != "" {
			body["project_id"] = projectID
		} else if config.ProjectID != "" {
			body["project_id"] = config.ProjectID
		}
		if taskType := extractFlag(args, "-type"); taskType != "" {
			body["type"] = taskType
		} else {
			body["type"] = "task" // Default type
		}
		if priority := extractFlag(args, "-priority"); priority != "" {
			body["priority"] = priority
		}
		if acceptance := extractFlag(args, "-acceptance"); acceptance != "" {
			body["acceptance_criteria"] = acceptance
		}
		data, err := client.Request("POST", "/api/v1/tasks", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task created:")
		fmt.Println(string(data))
	case "update":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{}
		if title := extractFlag(args, "-title"); title != "" {
			body["title"] = title
		}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		}
		if status := extractFlag(args, "-status"); status != "" {
			body["status"] = status
		}
		if priority := extractFlag(args, "-priority"); priority != "" {
			body["priority"] = priority
		}
		if acceptance := extractFlag(args, "-acceptance"); acceptance != "" {
			body["acceptance_criteria"] = acceptance
		}
		if len(body) == 0 {
			fmt.Fprintln(os.Stderr, "Error: at least one field to update required")
			os.Exit(1)
		}
		data, err := client.Request("PUT", "/api/v1/tasks/"+id, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task updated:")
		fmt.Println(string(data))
	case "delete", "rm":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", "/api/v1/tasks/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task deleted")
	case "claim":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/claim", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task claimed:")
		fmt.Println(string(data))
	case "free":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/free", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task freed:")
		fmt.Println(string(data))
	case "assign":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		username := extractFlag(args, "-username")
		if username == "" {
			username = extractFlag(args, "-u")
		}
		if username == "" {
			fmt.Fprintln(os.Stderr, "Error: -username flag required")
			os.Exit(1)
		}
		body := map[string]string{"assignee": username}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/assign", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task assigned:")
		fmt.Println(string(data))
	case "complete":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		notes := extractFlag(args, "-notes")
		var body interface{}
		if notes != "" {
			body = map[string]string{"notes": notes}
		}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/complete", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task completed:")
		fmt.Println(string(data))
	case "block":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		blockedBy := extractFlag(args, "-by")
		if blockedBy == "" {
			fmt.Fprintln(os.Stderr, "Error: -by flag required (task ID that blocks this task)")
			os.Exit(1)
		}
		body := map[string]interface{}{"depends_on_task_id": blockedBy}
		data, err := client.Request("PUT", "/api/v1/tasks/"+id, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Task %s is now blocked by %s\n", id, blockedBy)
		if config.JSON {
			fmt.Println(string(data))
		}
	case "unblock":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		// Send null to clear the dependency
		body := map[string]interface{}{"depends_on_task_id": nil}
		data, err := client.Request("PUT", "/api/v1/tasks/"+id, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Task %s is now unblocked\n", id)
		if config.JSON {
			fmt.Println(string(data))
		}
	case "history":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/tasks/"+id+"/history", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var history []map[string]interface{}
			if err := parseJSON(data, &history); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			if len(history) == 0 {
				fmt.Println("No history entries found")
				return
			}
			fmt.Printf("Task history (%d entries):\n\n", len(history))
			for _, h := range history {
				fmt.Printf("State:    %v\n", h["state"])
				fmt.Printf("Started:  %v\n", h["started_at"])
				if ended, ok := h["ended_at"]; ok && ended != nil {
					fmt.Printf("Ended:    %v\n", ended)
				}
				if worker, ok := h["worker_id"]; ok && worker != nil {
					fmt.Printf("Worker:   %v\n", worker)
				}
				if notes, ok := h["notes"]; ok && notes != nil {
					fmt.Printf("Notes:    %v\n", notes)
				}
				fmt.Println()
			}
		}
	case "deps", "dependencies":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/tasks/"+id+"/dependencies", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var deps map[string]interface{}
			if err := parseJSON(data, &deps); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			if blockedBy, ok := deps["blocked_by"]; ok && blockedBy != nil {
				fmt.Printf("Blocked by: %v\n", blockedBy)
			} else {
				fmt.Println("No dependencies")
			}
			if blocking, ok := deps["blocking"]; ok {
				if blockingArr, ok := blocking.([]interface{}); ok && len(blockingArr) > 0 {
					fmt.Printf("Blocking %d tasks:\n", len(blockingArr))
					for _, t := range blockingArr {
						if task, ok := t.(map[string]interface{}); ok {
							fmt.Printf("  - %v: %v\n", task["id"], task["title"])
						}
					}
				}
			}
		}
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
	case "create":
		username := extractFlag(args, "-username")
		if username == "" {
			username = extractFlag(args, "-u")
		}
		if username == "" {
			fmt.Fprintln(os.Stderr, "Error: -username flag required")
			os.Exit(1)
		}
		password := extractFlag(args, "-password")
		if password == "" {
			password = extractFlag(args, "-p")
		}
		if password == "" {
			fmt.Fprintln(os.Stderr, "Error: -password flag required")
			os.Exit(1)
		}
		userType := extractFlag(args, "-type")
		if userType == "" {
			userType = "human"
		}
		body := map[string]string{
			"username": username,
			"password": password,
			"type":     userType,
		}
		data, err := client.Request("POST", "/api/v1/auth/register", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("User created:")
		fmt.Println(string(data))
	case "enable":
		username := extractFlag(args, "-username")
		if username == "" {
			username = extractFlag(args, "-u")
		}
		if username == "" && len(args) > 1 {
			username = args[1]
		}
		if username == "" {
			fmt.Fprintln(os.Stderr, "Error: username required")
			os.Exit(1)
		}
		_, err := client.Request("POST", "/api/v1/users/"+username+"/enable", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' enabled\n", username)
	case "disable":
		username := extractFlag(args, "-username")
		if username == "" {
			username = extractFlag(args, "-u")
		}
		if username == "" && len(args) > 1 {
			username = args[1]
		}
		if username == "" {
			fmt.Fprintln(os.Stderr, "Error: username required")
			os.Exit(1)
		}
		_, err := client.Request("POST", "/api/v1/users/"+username+"/disable", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' disabled\n", username)
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
	case "get":
		id := extractFlag(args, "-id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/roles/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	case "create":
		name := extractFlag(args, "-name")
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: -name flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{"name": name}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		}
		if rules := extractFlag(args, "-rules"); rules != "" {
			body["rules"] = rules
		}
		if scope := extractFlag(args, "-scope"); scope != "" {
			body["scope"] = scope
		}
		data, err := client.Request("POST", "/api/v1/roles", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Role created:")
		fmt.Println(string(data))
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
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var configs []map[string]interface{}
			if err := parseJSON(data, &configs); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d config entries:\n\n", len(configs))
			for _, c := range configs {
				fmt.Printf("%v = %v\n", c["key"], c["value"])
			}
		}
	case "get":
		key := extractFlag(args, "-key")
		if key == "" && len(args) > 1 {
			key = args[1]
		}
		if key == "" {
			fmt.Fprintln(os.Stderr, "Error: -key flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/config/"+key, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	case "set":
		key := extractFlag(args, "-key")
		if key == "" {
			fmt.Fprintln(os.Stderr, "Error: -key flag required")
			os.Exit(1)
		}
		value := extractFlag(args, "-value")
		if value == "" {
			value = extractFlag(args, "-val")
		}
		if value == "" {
			fmt.Fprintln(os.Stderr, "Error: -value flag required")
			os.Exit(1)
		}
		body := map[string]string{"value": value}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		}
		data, err := client.Request("PUT", "/api/v1/config/"+key, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Config set:")
		fmt.Println(string(data))
	case "delete", "rm":
		key := extractFlag(args, "-key")
		if key == "" && len(args) > 1 {
			key = args[1]
		}
		if key == "" {
			fmt.Fprintln(os.Stderr, "Error: -key flag required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", "/api/v1/config/"+key, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config key '%s' deleted\n", key)
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

func handleLoginCommand(config *cli.Config) {
	// Prompt for credentials if not provided
	if config.Username == "" {
		fmt.Print("Username: ")
		fmt.Scanln(&config.Username)
	}
	if config.Password == "" {
		fmt.Print("Password: ")
		fmt.Scanln(&config.Password)
	}

	if config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "Error: Username and password are required")
		os.Exit(1)
	}

	// Test authentication
	client := cli.NewClient(config)
	data, err := client.Request("GET", "/api/v1/auth/me", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}

	// Save credentials
	if err := cli.SaveCredentials(config.ServerURL, config.Username, config.Password); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save credentials: %v\n", err)
	}

	if config.JSON {
		fmt.Println(string(data))
	} else {
		var user map[string]interface{}
		if err := parseJSON(data, &user); err == nil {
			fmt.Printf("Logged in as: %v\n", user["username"])
			fmt.Printf("Type: %v\n", user["type"])
			fmt.Println("\nCredentials saved to ~/.config/sf/credentials.json")
		} else {
			fmt.Println("Login successful")
		}
	}
}

func handleRegisterCommand(config *cli.Config) {
	// Prompt for credentials if not provided
	if config.Username == "" {
		fmt.Print("Username: ")
		fmt.Scanln(&config.Username)
	}
	if config.Password == "" {
		fmt.Print("Password: ")
		fmt.Scanln(&config.Password)
	}

	if config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "Error: Username and password are required")
		os.Exit(1)
	}

	// Register user
	client := cli.NewClient(config)
	body := map[string]string{
		"username": config.Username,
		"password": config.Password,
		"type":     "human",
	}
	data, err := client.Request("POST", "/api/v1/auth/register", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Registration failed: %v\n", err)
		os.Exit(1)
	}

	// Save credentials
	if err := cli.SaveCredentials(config.ServerURL, config.Username, config.Password); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save credentials: %v\n", err)
	}

	if config.JSON {
		fmt.Println(string(data))
	} else {
		var user map[string]interface{}
		if err := parseJSON(data, &user); err == nil {
			fmt.Printf("Account created: %v\n", user["username"])
			fmt.Printf("Type: %v\n", user["type"])
			fmt.Println("\nCredentials saved to ~/.config/sf/credentials.json")
			fmt.Println("You can now use 'sf' commands without providing credentials.")
		} else {
			fmt.Println("Registration successful")
		}
	}
}
