package main

import (
	"crypto/rand"
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
		// Check if it's worker daemon mode (no subcommand or -f flag) or worker CLI commands
		if len(os.Args) < 3 || os.Args[2] == "-f" || os.Args[2] == "-project_id" {
			// Worker daemon mode
			runWorker(os.Args[2:])
		} else {
			// Worker CLI commands (create, list, enable, disable, reset-password)
			runCLI(os.Args[1:])
		}
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
	fmt.Println("Note: Use 'sf worker' daemon for workers, worker CLI commands are under 'sf task'")
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
	password := fs.String("password", "", "Set password for all users (if not provided, random passwords are generated)")
	populate := fs.Bool("populate", false, "Populate database from scripts/initdb/*.md files")
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

	actualAdminPassword, userPassword, workerPassword, orchestratorPassword, err := database.InitializeDatabaseWithPasswords(*password)
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
	fmt.Println("User credentials:")
	fmt.Println("  Username: user")
	fmt.Printf("  Password: %s\n\n", userPassword)
	fmt.Println("Worker credentials:")
	fmt.Println("  Username: worker")
	fmt.Printf("  Password: %s\n\n", workerPassword)
	fmt.Println("Orchestrator credentials:")
	fmt.Println("  Username: orchestrator")
	fmt.Printf("  Password: %s\n\n", orchestratorPassword)
	fmt.Println("IMPORTANT: Save these credentials securely. They cannot be recovered.")

	// Populate from scripts if requested
	if *populate {
		fmt.Println("\nPopulating database from embedded scripts...")
		if err := database.PopulateFromScripts(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to populate database: %v\n", err)
		} else {
			fmt.Println("Database populated successfully!")
		}
	}
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
	case "worker":
		handleWorkerCommand(client, config, subArgs)
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
	fmt.Println("  sf login                          Login and save session token")
	fmt.Println("  sf register                       Register a new account")
	fmt.Println()
	fmt.Println("Projects:")
	fmt.Println("  sf project list (-project_id <id> -name <n> -description <d>)  List projects")
	fmt.Println("  sf project get -project_id <id>   Get project details")
	fmt.Println("  sf project create -project_id <id> -name <n>  Create new project")
	fmt.Println("  sf project update -project_id <id> (-name <n> -description <d>)  Update project")
	fmt.Println("  sf project delete -project_id <id>  Delete project")
	fmt.Println("  sf project set <name>             Set active project")
	fmt.Println("  sf project unset                  Unset active project")
	fmt.Println()
	fmt.Println("Tasks:")
	fmt.Println("  sf task list (-project_id <p> -status <s> -type <t> -owner <o>)  List tasks")
	fmt.Println("  sf task get -task_id <id>         Get task details")
	fmt.Println("  sf task create -title <t>         Create new task")
	fmt.Println("  sf task update -task_id <id> ...  Update task")
	fmt.Println("  sf task delete -task_id <id>      Delete task")
	fmt.Println("  sf task assign -task_id <id> -worker_id <w> -role <r>  Assign task")
	fmt.Println("  sf task unassign -task_id <id> -worker_id <w>  Unassign task")
	fmt.Println("  sf task request                   Request a task to work on")
	fmt.Println("  sf task return -task_id <id> ...  Return completed task")
	fmt.Println("  sf task comment -task_id <id> -comment <text>  Add comment to task")
	fmt.Println("  sf task history -task_id <id>     Show task history")
	fmt.Println()
	fmt.Println("Users:")
	fmt.Println("  sf user list                      List all users")
	fmt.Println("  sf user create -username <u> -password <p>  Create user")
	fmt.Println("  sf user enable -username <u>      Enable user")
	fmt.Println("  sf user disable -username <u>     Disable user")
	fmt.Println("  sf user reset-password -username <u> -password <p>  Reset user password")
	fmt.Println()
	fmt.Println("Workers:")
	fmt.Println("  sf worker list                    List all workers")
	fmt.Println("  sf worker create -worker_id <id>  Create worker")
	fmt.Println("  sf worker enable -worker_id <id>  Enable worker")
	fmt.Println("  sf worker disable -worker_id <id> Disable worker")
	fmt.Println("  sf worker reset-password -worker_id <id>  Reset worker password")
	fmt.Println()
	fmt.Println("Roles:")
	fmt.Println("  sf role list                      List all roles")
	fmt.Println("  sf role get -role_id <id>         Get role details")
	fmt.Println("  sf role create -title <t> -description <d> -goals <g>  Create role")
	fmt.Println("  sf role update -role_id <id> (-title <t> -description <d> -goals <g>)  Update role")
	fmt.Println("  sf role history -role_id <id>     Show role history")
	fmt.Println()
	fmt.Println("Config:")
	fmt.Println("  sf config list                    List all config")
	fmt.Println("  sf config set -key K -val V       Set config value")
	fmt.Println("  sf config delete -key K           Delete config value")
	fmt.Println()
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
		// Build query parameters
		queryParams := ""
		if projectID := extractFlag(args, "-project_id"); projectID != "" {
			if queryParams == "" {
				queryParams = "?project_id=" + projectID
			} else {
				queryParams += "&project_id=" + projectID
			}
		}
		if name := extractFlag(args, "-name"); name != "" {
			if queryParams == "" {
				queryParams = "?name=" + name
			} else {
				queryParams += "&name=" + name
			}
		}
		if desc := extractFlag(args, "-description"); desc != "" {
			if queryParams == "" {
				queryParams = "?description=" + desc
			} else {
				queryParams += "&description=" + desc
			}
		}
		data, err := client.Request("GET", "/api/v1/projects"+queryParams, nil)
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
		id := extractFlag(args, "-task_id")
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
		projectID := extractFlag(args, "-project_id")
		if projectID == "" {
			fmt.Fprintln(os.Stderr, "Error: -project_id flag required")
			os.Exit(1)
		}
		name := extractFlag(args, "-name")
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: -name flag required")
			os.Exit(1)
		}
		desc := extractFlag(args, "-description")
		body := map[string]string{"id": projectID, "name": name}
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
	case "update":
		projectID := extractFlag(args, "-project_id")
		if projectID == "" {
			fmt.Fprintln(os.Stderr, "Error: -project_id flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{}
		if name := extractFlag(args, "-name"); name != "" {
			body["name"] = name
		}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		}
		if len(body) == 0 {
			fmt.Fprintln(os.Stderr, "Error: at least one field to update required")
			os.Exit(1)
		}
		data, err := client.Request("PUT", "/api/v1/projects/"+projectID, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project updated:")
		fmt.Println(string(data))
	case "delete", "rm":
		projectID := extractFlag(args, "-project_id")
		if projectID == "" {
			fmt.Fprintln(os.Stderr, "Error: -project_id flag required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", "/api/v1/projects/"+projectID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project deleted")
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
	case "set-default":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: project ID required")
			os.Exit(1)
		}
		projectID := args[1]
		// Save as default in config
		body := map[string]string{"value": projectID}
		_, err := client.Request("PUT", "/api/v1/config/default_project_id", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default project set to: %s\n", projectID)
	case "get-default":
		data, err := client.Request("GET", "/api/v1/config/default_project_id", nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "No default project set")
			os.Exit(0)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var configVal map[string]interface{}
			if err := parseJSON(data, &configVal); err == nil {
				if val, ok := configVal["value"]; ok {
					fmt.Printf("Default project: %v\n", val)
				}
			}
		}
	case "unset-default":
		_, err := client.Request("DELETE", "/api/v1/config/default_project_id", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Default project cleared")
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
		// Build query parameters
		queryParams := ""
		if projectID := extractFlag(args, "-project_id"); projectID != "" {
			if queryParams == "" {
				queryParams = "?project_id=" + projectID
			} else {
				queryParams += "&project_id=" + projectID
			}
		}
		if status := extractFlag(args, "-status"); status != "" {
			if queryParams == "" {
				queryParams = "?status=" + status
			} else {
				queryParams += "&status=" + status
			}
		}
		if taskType := extractFlag(args, "-type"); taskType != "" {
			if queryParams == "" {
				queryParams = "?type=" + taskType
			} else {
				queryParams += "&type=" + taskType
			}
		}
		if owner := extractFlag(args, "-owner"); owner != "" {
			if queryParams == "" {
				queryParams = "?owner=" + owner
			} else {
				queryParams += "&owner=" + owner
			}
		}
		data, err := client.Request("GET", "/api/v1/tasks"+queryParams, nil)
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
		id := extractFlag(args, "-task_id")
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
		id := extractFlag(args, "-task_id")
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
		id := extractFlag(args, "-task_id")
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
	case "assign":
		id := extractFlag(args, "-task_id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -task_id flag required")
			os.Exit(1)
		}
		workerID := extractFlag(args, "-worker_id")
		role := extractFlag(args, "-role")
		body := map[string]interface{}{}
		if workerID != "" {
			body["worker_id"] = workerID
		}
		if role != "" {
			body["role"] = role
		}
		if len(body) == 0 {
			fmt.Fprintln(os.Stderr, "Error: -worker_id or -role flag required")
			os.Exit(1)
		}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/assign", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task assigned:")
		fmt.Println(string(data))
	case "unassign":
		id := extractFlag(args, "-task_id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -task_id flag required")
			os.Exit(1)
		}
		workerID := extractFlag(args, "-worker_id")
		if workerID == "" {
			fmt.Fprintln(os.Stderr, "Error: -worker_id flag required")
			os.Exit(1)
		}
		body := map[string]string{"worker_id": workerID}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/unassign", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task unassigned:")
		fmt.Println(string(data))
	case "request":
		body := map[string]interface{}{}
		if projectID := extractFlag(args, "-project_id"); projectID != "" {
			body["project_id"] = projectID
		}
		if taskType := extractFlag(args, "-type"); taskType != "" {
			body["task_type"] = taskType
		}
		data, err := client.Request("POST", "/api/v1/tasks/request", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var task map[string]interface{}
			if err := parseJSON(data, &task); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Task assigned to you:")
			fmt.Printf("ID:    %v\n", task["id"])
			fmt.Printf("Title: %v\n", task["title"])
			fmt.Printf("Type:  %v\n", task["type"])
		}
	case "return":
		id := extractFlag(args, "-task_id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -task_id flag required")
			os.Exit(1)
		}
		state := extractFlag(args, "-state")
		if state == "" {
			state = "completed"
		}
		body := map[string]interface{}{"state": state}
		if summary := extractFlag(args, "-summary"); summary != "" {
			body["summary"] = summary
		}
		if result := extractFlag(args, "-result"); result != "" {
			body["result"] = result
		}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/return", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task returned:")
		fmt.Println(string(data))
	case "comment":
		id := extractFlag(args, "-task_id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -task_id flag required")
			os.Exit(1)
		}
		text := extractFlag(args, "-comment")
		if text == "" {
			fmt.Fprintln(os.Stderr, "Error: -comment flag required")
			os.Exit(1)
		}
		body := map[string]string{"text": text}
		data, err := client.Request("POST", "/api/v1/tasks/"+id+"/comments", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Comment added:")
		fmt.Println(string(data))
	case "history":
		id := extractFlag(args, "-task_id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -task_id flag required")
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
	case "reset-password":
		username := extractFlag(args, "-username")
		if username == "" {
			fmt.Fprintln(os.Stderr, "Error: -username flag required")
			os.Exit(1)
		}
		newPassword := extractFlag(args, "-password")
		if newPassword == "" {
			fmt.Fprintln(os.Stderr, "Error: -password flag required")
			os.Exit(1)
		}
		body := map[string]string{"username": username, "new_password": newPassword}
		_, err := client.Request("POST", "/api/v1/users/reset-password", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Password reset for user '%s'\n", username)
	default:
		fmt.Fprintf(os.Stderr, "Unknown user subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleWorkerCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		data, err := client.Request("GET", "/api/v1/workers", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			var workers []map[string]interface{}
			if err := parseJSON(data, &workers); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d workers:\n\n", len(workers))
			for _, w := range workers {
				fmt.Printf("ID:       %v\n", w["id"])
				fmt.Printf("Username: %v\n", w["username"])
				fmt.Printf("Active:   %v\n", w["is_active"])
				if projectID, ok := w["project_id"]; ok && projectID != nil {
					fmt.Printf("Project:  %v\n", projectID)
				}
				fmt.Println()
			}
		}
	case "create":
		workerID := extractFlag(args, "-worker_id")
		if workerID == "" {
			fmt.Fprintln(os.Stderr, "Error: -worker_id flag required")
			os.Exit(1)
		}
		// Auto-generate password if not provided
		password := extractFlag(args, "-password")
		if password == "" {
			password = generatePassword(16)
		}
		body := map[string]string{
			"worker_id": workerID,
			"password":  password,
		}
		data, err := client.Request("POST", "/api/v1/workers", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			fmt.Println(string(data))
		} else {
			fmt.Println("Worker created:")
			fmt.Printf("Worker ID: %s\n", workerID)
			fmt.Printf("Password:  %s\n", password)
			fmt.Println("\nSave these credentials - password cannot be retrieved later.")
		}
	case "enable":
		workerID := extractFlag(args, "-worker_id")
		if workerID == "" && len(args) > 1 {
			workerID = args[1]
		}
		if workerID == "" {
			fmt.Fprintln(os.Stderr, "Error: -worker_id required")
			os.Exit(1)
		}
		_, err := client.Request("POST", "/api/v1/workers/"+workerID+"/enable", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Worker '%s' enabled\n", workerID)
	case "disable":
		workerID := extractFlag(args, "-worker_id")
		if workerID == "" && len(args) > 1 {
			workerID = args[1]
		}
		if workerID == "" {
			fmt.Fprintln(os.Stderr, "Error: -worker_id required")
			os.Exit(1)
		}
		_, err := client.Request("POST", "/api/v1/workers/"+workerID+"/disable", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Worker '%s' disabled\n", workerID)
	case "reset-password":
		workerID := extractFlag(args, "-worker_id")
		if workerID == "" {
			fmt.Fprintln(os.Stderr, "Error: -worker_id flag required")
			os.Exit(1)
		}
		// Auto-generate new password if not provided
		newPassword := extractFlag(args, "-password")
		if newPassword == "" {
			newPassword = generatePassword(16)
		}
		body := map[string]string{"new_password": newPassword}
		_, err := client.Request("POST", "/api/v1/workers/"+workerID+"/reset-password", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Password reset for worker '%s'\n", workerID)
		fmt.Printf("New password: %s\n", newPassword)
		fmt.Println("\nSave this password - it cannot be retrieved later.")
	default:
		fmt.Fprintf(os.Stderr, "Unknown worker subcommand: %s\n", subcommand)
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
		id := extractFlag(args, "-task_id")
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
		title := extractFlag(args, "-title")
		if title == "" {
			fmt.Fprintln(os.Stderr, "Error: -title flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{"title": title}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		}
		if goals := extractFlag(args, "-goals"); goals != "" {
			body["goals"] = goals
		}
		data, err := client.Request("POST", "/api/v1/roles", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Role created:")
		fmt.Println(string(data))
	case "update":
		roleID := extractFlag(args, "-role_id")
		if roleID == "" {
			fmt.Fprintln(os.Stderr, "Error: -role_id flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{}
		if title := extractFlag(args, "-title"); title != "" {
			body["title"] = title
		}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		}
		if goals := extractFlag(args, "-goals"); goals != "" {
			body["goals"] = goals
		}
		if len(body) == 0 {
			fmt.Fprintln(os.Stderr, "Error: at least one field to update required")
			os.Exit(1)
		}
		data, err := client.Request("PUT", "/api/v1/roles/"+roleID, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Role updated:")
		fmt.Println(string(data))
	case "history":
		roleID := extractFlag(args, "-role_id")
		if roleID == "" {
			fmt.Fprintln(os.Stderr, "Error: -role_id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/roles/"+roleID+"/history", nil)
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
			fmt.Printf("Role history (%d entries):\n\n", len(history))
			for _, h := range history {
				fmt.Printf("Version:     %v\n", h["version"])
				fmt.Printf("Modified:    %v\n", h["modified_at"])
				if title, ok := h["title"]; ok && title != nil {
					fmt.Printf("Title:       %v\n", title)
				}
				if desc, ok := h["description"]; ok && desc != nil {
					fmt.Printf("Description: %v\n", desc)
				}
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

func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		// Use crypto/rand for secure random selection
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			// Fallback to less secure but still usable method
			password[i] = charset[i%len(charset)]
		} else {
			password[i] = charset[int(randomByte[0])%len(charset)]
		}
	}
	return string(password)
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

	// Call login endpoint to get session token
	client := cli.NewClient(config)
	loginBody := map[string]string{
		"username": config.Username,
		"password": config.Password,
	}
	data, err := client.Request("POST", "/api/v1/auth/login", loginBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}

	// Parse login response
	var loginResp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    string `json:"expires_at"`
		User         struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Type     string `json:"type"`
		} `json:"user"`
	}
	if err := parseJSON(data, &loginResp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse login response: %v\n", err)
		os.Exit(1)
	}

	// Save session token
	if err := cli.SaveSessionToken(loginResp.Token, loginResp.RefreshToken, loginResp.ExpiresAt, config.ServerURL, config.Username); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save session token: %v\n", err)
	}

	// Also save credentials as fallback
	if err := cli.SaveCredentials(config.ServerURL, config.Username, config.Password); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save credentials: %v\n", err)
	}

	if config.JSON {
		fmt.Println(string(data))
	} else {
		fmt.Printf("Logged in as: %v\n", loginResp.User.Username)
		fmt.Printf("Type: %v\n", loginResp.User.Type)
		fmt.Printf("\nSession token saved to ~/.config/sf/session.json\n")
		fmt.Printf("Token expires at: %s\n", loginResp.ExpiresAt)
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
