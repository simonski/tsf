package main

import (
	cryptorand "crypto/rand"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/simonski/task/internal/cli"
	"github.com/simonski/task/internal/db"
	"github.com/simonski/task/internal/orchestrator"
	"github.com/simonski/task/internal/server"
	"github.com/simonski/task/internal/tui"
	"github.com/simonski/task/internal/web"
	"github.com/simonski/task/internal/worker"
	"golang.org/x/term"
)

//go:embed VERSION
var version string

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
		// Check if it's worker daemon mode (no subcommand or has daemon flags) or worker CLI commands
		if len(os.Args) < 3 || os.Args[2] == "-worker_id" || os.Args[2] == "-url" || os.Args[2] == "-password" || os.Args[2] == "-h" {
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
	case "login", "register", "logout", "project", "task", "epic", "bug", "chore", "comment", "user", "role", "config", "status", "bd", "bead", "beads":
		runCLI(os.Args[1:])
	case "version", "-v", "--version":
		fmt.Println(strings.TrimSpace(version))
	case "help", "-h", "--help":
		if len(os.Args) >= 3 {
			// sf help <command>
			showCommandHelp(os.Args[2])
		} else {
			printUsage()
		}
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
	fmt.Println("  logout        Logout and clear saved credentials")
	fmt.Println()
	fmt.Println("  project       Manage projects")
	fmt.Println("  task          Manage tasks")
	fmt.Println("  epic          Manage epics (task alias with implied -type epic)")
	fmt.Println("  bug           Manage bugs (task alias with implied -type bug)")
	fmt.Println("  chore         Manage chores (task alias with implied -type chore)")
	fmt.Println("  comment       Manage comments on projects and tasks")
	fmt.Println("  user          Manage users")
	fmt.Println("  role          Manage roles")
	fmt.Println("  config        Manage configuration")
	fmt.Println("  status        Show task statistics by type and status")
	fmt.Println("  beads         Manage beads markdown import/export workflow (aliases: bead, bd)")
	fmt.Println()
	fmt.Println("Note: Use 'sf worker' daemon for workers, worker CLI commands are under 'sf task'")
	fmt.Println()
	fmt.Println("  version       Show version information")
	fmt.Println("  help          Show this help message")
	fmt.Println()
	fmt.Println("Use 'sf <command> -h' for more information about a command.")
}

func showCommandHelp(command string) {
	switch command {
	case "server":
		printServerHelp()
	case "orchestrator":
		printOrchestratorHelp()
	case "worker":
		printWorkerHelp()
	case "initdb":
		printInitDBHelp()
	case "tui":
		printTUIHelp()
	case "project":
		printProjectHelp()
	case "task":
		printTaskHelp()
	case "epic":
		printTaskHelp()
	case "bug":
		printTaskHelp()
	case "chore":
		printTaskHelp()
	case "user":
		printUserHelp()
	case "comment":
		printCommentHelp()
	case "role":
		printRoleHelp()
	case "config":
		printConfigHelp()
	case "beads", "bead", "bd":
		printBeadsHelp()
	default:
		fmt.Fprintf(os.Stderr, "No detailed help available for '%s'\n\n", command)
		printUsage()
	}
}

func printServerHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF SERVER - HTTP API Server & Web UI")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  The sf server provides a RESTful HTTP API and embedded web UI for")
	fmt.Println("  managing projects, tasks, roles, users, and workers. It serves as")
	fmt.Println("  the central hub for the Software Factory system.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf server [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("  -f <path>     Database file path")
	fmt.Println("                Default: ~/.config/sf/sf.db")
	fmt.Println("                Specify a custom location for the SQLite database")
	fmt.Println()
	fmt.Println("  -port <port>  Server port number")
	fmt.Println("                Default: 8080")
	fmt.Println("                The HTTP port the server will listen on")
	fmt.Println()
	fmt.Println("  -v            Enable verbose HTTP request/response logging")
	fmt.Println("                Logs request/response details to stdout")
	fmt.Println()
	fmt.Println("  -h            Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # Start server with default settings (port 8080, default database)")
	fmt.Println("  $ sf server")
	fmt.Println()
	fmt.Println("  # Start server on custom port")
	fmt.Println("  $ sf server -port 3000")
	fmt.Println()
	fmt.Println("  # Start server with custom database file")
	fmt.Println("  $ sf server -f /path/to/my-project.db")
	fmt.Println()
	fmt.Println("  # Use custom port AND custom database")
	fmt.Println("  $ sf server -f ./dev.db -port 10606")
	fmt.Println()
	fmt.Println("  # For testing/development on non-standard port")
	fmt.Println("  $ sf server -f ./test.db -port 9999")
	fmt.Println()
	fmt.Println("ENDPOINTS")
	fmt.Println()
	fmt.Println("  Web UI:")
	fmt.Println("    http://localhost:<port>/              Main web interface")
	fmt.Println()
	fmt.Println("  API:")
	fmt.Println("    http://localhost:<port>/api/v1/       RESTful API base URL")
	fmt.Println("    http://localhost:<port>/api/v1/health Health check endpoint")
	fmt.Println()
	fmt.Println("  Monitoring:")
	fmt.Println("    http://localhost:<port>/metrics       Prometheus metrics")
	fmt.Println()
	fmt.Println("AUTHENTICATION")
	fmt.Println()
	fmt.Println("  The server supports three authentication methods:")
	fmt.Println()
	fmt.Println("  1. Basic Auth - Use username:password in HTTP Basic Auth header")
	fmt.Println("  2. Session Token - Login via /api/v1/auth/login to get JWT token")
	fmt.Println("  3. Passkeys (WebAuthn) - Passwordless authentication via browser")
	fmt.Println()
	fmt.Println("BEFORE FIRST RUN")
	fmt.Println()
	fmt.Println("  You must initialize the database before starting the server:")
	fmt.Println()
	fmt.Println("  $ sf initdb                    # Initialize default database")
	fmt.Println("  $ sf initdb -f custom.db       # Initialize custom database")
	fmt.Println()
	fmt.Println("  See 'sf help initdb' for more database initialization options.")
	fmt.Println()
	fmt.Println("API DOCUMENTATION")
	fmt.Println()
	fmt.Println("  Full API documentation is available in the OpenAPI specification:")
	fmt.Println("    api-specification.yaml")
	fmt.Println()
	fmt.Println("  Key API endpoints include:")
	fmt.Println("    • Authentication: /api/v1/auth/*")
	fmt.Println("    • Projects:       /api/v1/projects")
	fmt.Println("    • Tasks:          /api/v1/tasks")
	fmt.Println("    • Roles:          /api/v1/roles")
	fmt.Println("    • Users:          /api/v1/users")
	fmt.Println("    • Workers:        /api/v1/workers")
	fmt.Println("    • Config:         /api/v1/config")
	fmt.Println()
	fmt.Println("CONNECTING COMPONENTS")
	fmt.Println()
	fmt.Println("  Once the server is running, connect other components:")
	fmt.Println()
	fmt.Println("  # Start orchestrator (task assignment engine)")
	fmt.Println("  $ export SF_USERNAME=orchestrator")
	fmt.Println("  $ export SF_PASSWORD=<password>")
	fmt.Println("  $ sf orchestrator -url http://localhost:8080")
	fmt.Println()
	fmt.Println("  # Start worker daemons")
	fmt.Println("  $ export SF_USERNAME=worker1")
	fmt.Println("  $ export SF_PASSWORD=<password>")
	fmt.Println("  $ sf worker -worker_id worker1 -url http://localhost:8080")
	fmt.Println()
	fmt.Println("  # Use CLI tools")
	fmt.Println("  $ sf login                              # Interactive login")
	fmt.Println("  $ sf project list                       # List projects")
	fmt.Println("  $ sf task create \"Fix bug\"             # Create task")
	fmt.Println()
	fmt.Println("LOGS AND MONITORING")
	fmt.Println()
	fmt.Println("  The server logs to stdout. Redirect for persistent logging:")
	fmt.Println()
	fmt.Println("  $ sf server >> server.log 2>&1 &")
	fmt.Println()
	fmt.Println("  Monitor via Prometheus metrics endpoint:")
	fmt.Println()
	fmt.Println("  $ curl http://localhost:8080/metrics")
	fmt.Println()
	fmt.Println("STOPPING THE SERVER")
	fmt.Println()
	fmt.Println("  Use Ctrl+C to gracefully stop the server, or send SIGTERM:")
	fmt.Println()
	fmt.Println("  $ kill <pid>")
	fmt.Println("  $ pkill -f \"sf server\"")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help initdb        Database initialization")
	fmt.Println("  sf help orchestrator  Orchestrator daemon")
	fmt.Println("  sf help worker        Worker daemon")
	fmt.Println("  sf help tui           Terminal User Interface")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printOrchestratorHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF ORCHESTRATOR - Task Assignment Engine")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  The orchestrator is a daemon that automatically assigns tasks to")
	fmt.Println("  available workers based on roles, skills, and workload. It continuously")
	fmt.Println("  monitors the task queue and worker availability, intelligently matching")
	fmt.Println("  work to the most suitable workers.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf orchestrator [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("  -url <url>        Server URL")
	fmt.Println("                    Default: http://localhost:8080")
	fmt.Println("                    The URL of the sf server to connect to")
	fmt.Println()
	fmt.Println("  -username <name>  Username for authentication")
	fmt.Println("                    Default: $SF_USERNAME environment variable")
	fmt.Println("                    Typically 'orchestrator' user")
	fmt.Println()
	fmt.Println("  -password <pass>  Password for authentication")
	fmt.Println("                    Default: $SF_PASSWORD environment variable")
	fmt.Println("                    Keep this secure and use environment variables")
	fmt.Println()
	fmt.Println("  -h                Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # Start orchestrator with environment variables (recommended)")
	fmt.Println("  $ export SF_USERNAME=orchestrator")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf orchestrator")
	fmt.Println()
	fmt.Println("  # Start orchestrator with explicit credentials")
	fmt.Println("  $ sf orchestrator -username orchestrator -password secret123")
	fmt.Println()
	fmt.Println("  # Connect to remote server")
	fmt.Println("  $ export SF_USERNAME=orchestrator")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf orchestrator -url https://sf.example.com")
	fmt.Println()
	fmt.Println("  # Connect to custom port")
	fmt.Println("  $ export SF_USERNAME=orchestrator")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf orchestrator -url http://localhost:10606")
	fmt.Println()
	fmt.Println("  # Run in background with nohup")
	fmt.Println("  $ nohup sf orchestrator >> orchestrator.log 2>&1 &")
	fmt.Println()
	fmt.Println("HOW IT WORKS")
	fmt.Println()
	fmt.Println("  1. Connects to the sf server with admin/orchestrator credentials")
	fmt.Println("  2. Registers itself and sends periodic heartbeats")
	fmt.Println("  3. Queries for unassigned tasks in 'idle' status")
	fmt.Println("  4. Queries for available workers")
	fmt.Println("  5. Matches tasks to workers based on:")
	fmt.Println("     • Role requirements (task.role_id matches worker capabilities)")
	fmt.Println("     • Worker availability (not already assigned tasks)")
	fmt.Println("     • Task dependencies (prerequisite tasks completed)")
	fmt.Println("     • Priority (higher priority tasks assigned first)")
	fmt.Println("  6. Assigns matched tasks via POST /api/v1/tasks/{id}/assign")
	fmt.Println("  7. Monitors worker heartbeats for timeouts")
	fmt.Println("  8. Repeats continuously until stopped")
	fmt.Println()
	fmt.Println("REQUIREMENTS")
	fmt.Println()
	fmt.Println("  • Server must be running: sf server")
	fmt.Println("  • Database must be initialized: sf initdb")
	fmt.Println("  • User account must exist with orchestrator privileges")
	fmt.Println("  • Network connectivity to server URL")
	fmt.Println()
	fmt.Println("CREATING ORCHESTRATOR USER")
	fmt.Println()
	fmt.Println("  During database initialization, an 'orchestrator' user is created:")
	fmt.Println()
	fmt.Println("  $ sf initdb                          # Creates orchestrator user")
	fmt.Println("  $ sf initdb -password mypass123      # Set specific password")
	fmt.Println()
	fmt.Println("  Or create manually via CLI:")
	fmt.Println()
	fmt.Println("  $ sf user create orchestrator --type orchestrator")
	fmt.Println()
	fmt.Println("MONITORING")
	fmt.Println()
	fmt.Println("  Monitor orchestrator status:")
	fmt.Println()
	fmt.Println("  $ curl http://localhost:8080/api/v1/heartbeats")
	fmt.Println()
	fmt.Println("  Check orchestrator logs:")
	fmt.Println()
	fmt.Println("  $ tail -f orchestrator.log")
	fmt.Println()
	fmt.Println("  The orchestrator logs each assignment and decision for debugging.")
	fmt.Println()
	fmt.Println("STOPPING THE ORCHESTRATOR")
	fmt.Println()
	fmt.Println("  Use Ctrl+C to gracefully stop the orchestrator:")
	fmt.Println()
	fmt.Println("  ^C")
	fmt.Println()
	fmt.Println("  Or send SIGTERM/SIGINT:")
	fmt.Println()
	fmt.Println("  $ kill <pid>")
	fmt.Println("  $ pkill -f 'sf orchestrator'")
	fmt.Println()
	fmt.Println("  The orchestrator will cleanly disconnect from the server.")
	fmt.Println()
	fmt.Println("TROUBLESHOOTING")
	fmt.Println()
	fmt.Println("  Connection refused:")
	fmt.Println("    • Ensure server is running: sf server")
	fmt.Println("    • Check URL matches server port: -url http://localhost:8080")
	fmt.Println()
	fmt.Println("  Authentication failed:")
	fmt.Println("    • Verify credentials: Check SF_USERNAME and SF_PASSWORD")
	fmt.Println("    • Ensure user exists: sf user list")
	fmt.Println("    • Check user is enabled and has orchestrator type")
	fmt.Println()
	fmt.Println("  No tasks assigned:")
	fmt.Println("    • Check tasks exist: sf task list")
	fmt.Println("    • Check workers available: sf task worker list")
	fmt.Println("    • Check task roles match worker capabilities")
	fmt.Println("    • Review orchestrator logs for decision logic")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help server    HTTP server")
	fmt.Println("  sf help worker    Worker daemon")
	fmt.Println("  sf help initdb    Database initialization")
	fmt.Println("  sf help task      Task management commands")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printWorkerHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF WORKER - Task Execution Daemon")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  The worker daemon connects to the sf server, requests task assignments,")
	fmt.Println("  and executes them. Workers identify themselves with a unique worker_id")
	fmt.Println("  and continuously poll for new work. Multiple workers can run in parallel")
	fmt.Println("  to distribute workload across teams or machines.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf worker [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("  -worker_id <id>   Unique worker identifier (required)")
	fmt.Println("                    Default: $SF_USERNAME environment variable")
	fmt.Println("                    Examples: 'worker1', 'dev-worker', 'ci-agent-01'")
	fmt.Println("                    Must be unique across all active workers")
	fmt.Println()
	fmt.Println("  -url <url>        Server URL")
	fmt.Println("                    Default: http://localhost:8080")
	fmt.Println("                    The URL of the sf server to connect to")
	fmt.Println()
	fmt.Println("  -password <pass>  Password for authentication")
	fmt.Println("                    Default: $SF_PASSWORD environment variable")
	fmt.Println("                    Keep this secure and use environment variables")
	fmt.Println()
	fmt.Println("  -h                Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # Start worker with environment variables (recommended)")
	fmt.Println("  $ export SF_USERNAME=worker1")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf worker -worker_id worker1")
	fmt.Println()
	fmt.Println("  # Start multiple workers for parallel execution")
	fmt.Println("  $ sf worker -worker_id worker1 -password pass1 &")
	fmt.Println("  $ sf worker -worker_id worker2 -password pass2 &")
	fmt.Println("  $ sf worker -worker_id worker3 -password pass3 &")
	fmt.Println()
	fmt.Println("  # Connect to remote server")
	fmt.Println("  $ export SF_USERNAME=worker1")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf worker -worker_id worker1 -url https://sf.example.com")
	fmt.Println()
	fmt.Println("  # Connect to custom port")
	fmt.Println("  $ export SF_USERNAME=worker1")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf worker -worker_id worker1 -url http://localhost:10606")
	fmt.Println()
	fmt.Println("  # Run in background with nohup")
	fmt.Println("  $ nohup sf worker -worker_id worker1 >> worker1.log 2>&1 &")
	fmt.Println()
	fmt.Println("  # Using systemd for production")
	fmt.Println("  $ systemctl start sf-worker@worker1")
	fmt.Println()
	fmt.Println("HOW IT WORKS")
	fmt.Println()
	fmt.Println("  1. Registers with server: POST /api/v1/workers/register")
	fmt.Println("  2. Sends periodic heartbeats: POST /api/v1/workers/heartbeat")
	fmt.Println("  3. Requests task assignment: POST /api/v1/workers/request")
	fmt.Println("  4. Receives task (if available) or waits")
	fmt.Println("  5. Claims the task: POST /api/v1/tasks/{id}/claim")
	fmt.Println("  6. Executes the task (implementation-specific)")
	fmt.Println("  7. Reports completion: POST /api/v1/tasks/{id}/complete")
	fmt.Println("     or returns task: POST /api/v1/tasks/{id}/return (on error)")
	fmt.Println("  8. Repeats from step 3 until stopped")
	fmt.Println()
	fmt.Println("REQUIREMENTS")
	fmt.Println()
	fmt.Println("  • Server must be running: sf server")
	fmt.Println("  • Database must be initialized: sf initdb")
	fmt.Println("  • Worker user account must exist with 'worker' type")
	fmt.Println("  • worker_id must be unique (not already registered/active)")
	fmt.Println("  • Network connectivity to server URL")
	fmt.Println()
	fmt.Println("CREATING WORKER USERS")
	fmt.Println()
	fmt.Println("  Workers are regular users with type 'worker':")
	fmt.Println()
	fmt.Println("  $ sf initdb -populate              # Creates sample workers")
	fmt.Println()
	fmt.Println("  Or create manually:")
	fmt.Println()
	fmt.Println("  $ sf user create worker1 --type worker")
	fmt.Println("  $ sf user create worker2 --type worker")
	fmt.Println()
	fmt.Println("  Assign roles/capabilities to workers for task matching:")
	fmt.Println()
	fmt.Println("  $ sf role create \"backend-developer\" --goals \"Build APIs\"")
	fmt.Println("  $ sf task worker assign worker1 backend-developer")
	fmt.Println()
	fmt.Println("WORKER LIFECYCLE")
	fmt.Println()
	fmt.Println("  Registration:")
	fmt.Println("    On startup, worker registers with unique worker_id.")
	fmt.Println("    Server rejects duplicate active worker_id registrations.")
	fmt.Println()
	fmt.Println("  Heartbeat:")
	fmt.Println("    Worker sends heartbeat every 30 seconds.")
	fmt.Println("    Server marks worker as inactive if heartbeat times out (>2 min).")
	fmt.Println()
	fmt.Println("  Task Execution:")
	fmt.Println("    Worker claims task, executes, then completes or returns it.")
	fmt.Println("    Failed tasks can be retried by other workers.")
	fmt.Println()
	fmt.Println("  Shutdown:")
	fmt.Println("    Graceful stop (Ctrl+C) allows current task to complete.")
	fmt.Println("    In-progress tasks are released back to queue on timeout.")
	fmt.Println()
	fmt.Println("MONITORING")
	fmt.Println()
	fmt.Println("  List active workers:")
	fmt.Println()
	fmt.Println("  $ sf task worker list")
	fmt.Println("  $ curl http://localhost:8080/api/v1/workers")
	fmt.Println()
	fmt.Println("  Check worker status:")
	fmt.Println()
	fmt.Println("  $ curl http://localhost:8080/api/v1/workers/worker1")
	fmt.Println()
	fmt.Println("  View worker tasks:")
	fmt.Println()
	fmt.Println("  $ curl http://localhost:8080/api/v1/workers/worker1/tasks")
	fmt.Println()
	fmt.Println("  Check worker logs:")
	fmt.Println()
	fmt.Println("  $ tail -f worker1.log")
	fmt.Println()
	fmt.Println("STOPPING A WORKER")
	fmt.Println()
	fmt.Println("  Use Ctrl+C for graceful shutdown:")
	fmt.Println()
	fmt.Println("  ^C")
	fmt.Println()
	fmt.Println("  Or send SIGTERM/SIGINT:")
	fmt.Println()
	fmt.Println("  $ kill <pid>")
	fmt.Println("  $ pkill -f 'sf worker -worker_id worker1'")
	fmt.Println()
	fmt.Println("  Force kill (not recommended - may leave task in bad state):")
	fmt.Println()
	fmt.Println("  $ kill -9 <pid>")
	fmt.Println()
	fmt.Println("TROUBLESHOOTING")
	fmt.Println()
	fmt.Println("  Connection refused:")
	fmt.Println("    • Ensure server is running: sf server")
	fmt.Println("    • Check URL matches server port: -url http://localhost:8080")
	fmt.Println()
	fmt.Println("  Authentication failed:")
	fmt.Println("    • Verify credentials: Check SF_USERNAME and SF_PASSWORD")
	fmt.Println("    • Ensure user exists: sf user list")
	fmt.Println("    • Check user is enabled and has worker type")
	fmt.Println()
	fmt.Println("  Worker already registered:")
	fmt.Println("    • worker_id must be unique across active workers")
	fmt.Println("    • Stop existing worker first, or use different worker_id")
	fmt.Println("    • Check active workers: sf task worker list")
	fmt.Println()
	fmt.Println("  No tasks received:")
	fmt.Println("    • Check tasks exist: sf task list")
	fmt.Println("    • Verify orchestrator is running: sf orchestrator")
	fmt.Println("    • Check worker roles match task requirements")
	fmt.Println("    • Review server logs for assignment logic")
	fmt.Println()
	fmt.Println("SCALING WORKERS")
	fmt.Println()
	fmt.Println("  Run multiple workers for horizontal scaling:")
	fmt.Println()
	fmt.Println("  $ for i in {1..5}; do")
	fmt.Println("      sf worker -worker_id worker$i -password pass$i &")
	fmt.Println("    done")
	fmt.Println()
	fmt.Println("  Use docker-compose for containerized workers:")
	fmt.Println()
	fmt.Println("  $ docker-compose up --scale worker=10")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help server        HTTP server")
	fmt.Println("  sf help orchestrator  Orchestrator daemon")
	fmt.Println("  sf help initdb        Database initialization")
	fmt.Println("  sf help task          Task management commands")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printInitDBHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF INITDB - Database Initialization")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Initializes the SQLite database for the sf system. Creates the schema,")
	fmt.Println("  tables, indexes, and optionally populates with default users, projects,")
	fmt.Println("  roles, and tasks. This must be run before starting the server.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf initdb [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("  -f <path>         Database file path")
	fmt.Println("                    Default: ~/.config/sf/sf.db")
	fmt.Println("                    Specify a custom location for the database")
	fmt.Println()
	fmt.Println("  -force            Force rebuild (removes existing database)")
	fmt.Println("                    Default: false")
	fmt.Println("                    WARNING: Destroys all existing data!")
	fmt.Println()
	fmt.Println("  -password <pass>  Set password for all default users")
	fmt.Println("                    Default: randomly generated per user")
	fmt.Println("                    Useful for development/testing environments")
	fmt.Println()
	fmt.Println("  -populate         Populate with sample data from scripts/initdb/")
	fmt.Println("                    Default: false")
	fmt.Println("                    Loads projects, roles, and tasks from markdown files")
	fmt.Println()
	fmt.Println("  -h                Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # Basic initialization (default location, random passwords)")
	fmt.Println("  $ sf initdb")
	fmt.Println()
	fmt.Println("  # Initialize with known password for all users")
	fmt.Println("  $ sf initdb -password dev123")
	fmt.Println()
	fmt.Println("  # Initialize and populate with sample data")
	fmt.Println("  $ sf initdb -populate")
	fmt.Println()
	fmt.Println("  # Initialize with custom database location")
	fmt.Println("  $ sf initdb -f /var/lib/sf/production.db")
	fmt.Println()
	fmt.Println("  # Initialize test database with known credentials")
	fmt.Println("  $ sf initdb -f ./test.db -password test123 -populate")
	fmt.Println()
	fmt.Println("  # Force rebuild (destroys existing data)")
	fmt.Println("  $ sf initdb -force")
	fmt.Println()
	fmt.Println("  # Complete setup for development environment")
	fmt.Println("  $ sf initdb -f ./dev.db -password dev123 -populate")
	fmt.Println()
	fmt.Println("WHAT GETS CREATED")
	fmt.Println()
	fmt.Println("  Database Schema:")
	fmt.Println("    • users           - User accounts (admin, human, worker, orchestrator)")
	fmt.Println("    • sessions        - Authentication sessions and JWT tokens")
	fmt.Println("    • passkeys        - WebAuthn passkey credentials")
	fmt.Println("    • projects        - Project workspaces")
	fmt.Println("    • tasks           - Task definitions and assignments")
	fmt.Println("    • roles           - Role definitions for workers")
	fmt.Println("    • task_history    - Task lifecycle events")
	fmt.Println("    • task_comments   - Task discussion threads")
	fmt.Println("    • task_dependencies - Task prerequisite relationships")
	fmt.Println("    • heartbeats      - Worker/orchestrator liveness tracking")
	fmt.Println("    • config          - System configuration key-value store")
	fmt.Println()
	fmt.Println("  Default Users Created:")
	fmt.Println("    • admin          - Administrator account (full access)")
	fmt.Println("    • orchestrator   - Orchestrator daemon account")
	fmt.Println("    • worker1        - Sample worker account (if -populate)")
	fmt.Println("    • worker2        - Sample worker account (if -populate)")
	fmt.Println()
	fmt.Println("  Passwords:")
	fmt.Println("    Without -password: Random 16-character passwords (printed to stdout)")
	fmt.Println("    With -password:    All users get the specified password")
	fmt.Println()
	fmt.Println("DEFAULT LOCATION")
	fmt.Println()
	fmt.Println("  The database is created at: ~/.config/sf/sf.db")
	fmt.Println()
	fmt.Println("  The directory is created automatically if it doesn't exist.")
	fmt.Println()
	fmt.Println("POPULATE MODE (-populate)")
	fmt.Println()
	fmt.Println("  Loads sample data from markdown files in scripts/initdb/:")
	fmt.Println()
	fmt.Println("    • projects.md     - Sample project definitions")
	fmt.Println("    • roles.md        - Role definitions with goals")
	fmt.Println("    • tasks.md        - Sample tasks with dependencies")
	fmt.Println()
	fmt.Println("  This creates a working demo environment for testing and development.")
	fmt.Println()
	fmt.Println("SECURITY NOTES")
	fmt.Println()
	fmt.Println("  ⚠️  Random passwords are printed ONCE during initialization")
	fmt.Println("      Save them immediately! They cannot be recovered.")
	fmt.Println()
	fmt.Println("  ⚠️  The -password flag should only be used for development")
	fmt.Println("      Never use simple passwords in production!")
	fmt.Println()
	fmt.Println("  ⚠️  The -force flag destroys ALL data without confirmation")
	fmt.Println("      Use with extreme caution!")
	fmt.Println()
	fmt.Println("  For production:")
	fmt.Println("    1. Use sf initdb without -password (random passwords)")
	fmt.Println("    2. Save passwords securely (password manager, secrets vault)")
	fmt.Println("    3. Use environment variables (SF_USERNAME, SF_PASSWORD)")
	fmt.Println("    4. Enable passkeys for passwordless authentication")
	fmt.Println()
	fmt.Println("AFTER INITIALIZATION")
	fmt.Println()
	fmt.Println("  1. Start the server:")
	fmt.Println("     $ sf server")
	fmt.Println()
	fmt.Println("  2. Login with admin credentials:")
	fmt.Println("     $ export SF_USERNAME=admin")
	fmt.Println("     $ export SF_PASSWORD=<generated_password>")
	fmt.Println("     $ sf login")
	fmt.Println()
	fmt.Println("  3. Create additional users:")
	fmt.Println("     $ sf user create developer --type human")
	fmt.Println("     $ sf user create worker3 --type worker")
	fmt.Println()
	fmt.Println("  4. Start orchestrator and workers:")
	fmt.Println("     $ export SF_USERNAME=orchestrator")
	fmt.Println("     $ export SF_PASSWORD=<orchestrator_password>")
	fmt.Println("     $ sf orchestrator &")
	fmt.Println()
	fmt.Println("     $ export SF_USERNAME=worker1")
	fmt.Println("     $ export SF_PASSWORD=<worker1_password>")
	fmt.Println("     $ sf worker -worker_id worker1 &")
	fmt.Println()
	fmt.Println("RESETTING THE DATABASE")
	fmt.Println()
	fmt.Println("  To completely reset (useful for testing):")
	fmt.Println()
	fmt.Println("  $ sf initdb -force -password dev123 -populate")
	fmt.Println()
	fmt.Println("  Or manually:")
	fmt.Println()
	fmt.Println("  $ rm ~/.config/sf/sf.db")
	fmt.Println("  $ sf initdb")
	fmt.Println()
	fmt.Println("TROUBLESHOOTING")
	fmt.Println()
	fmt.Println("  Database already exists:")
	fmt.Println("    • Use -force to rebuild, or remove manually first")
	fmt.Println("    • rm ~/.config/sf/sf.db")
	fmt.Println()
	fmt.Println("  Permission denied:")
	fmt.Println("    • Check write permissions on ~/.config/sf/")
	fmt.Println("    • mkdir -p ~/.config/sf && chmod 755 ~/.config/sf")
	fmt.Println()
	fmt.Println("  Schema errors:")
	fmt.Println("    • Database may be corrupted or from incompatible version")
	fmt.Println("    • Use -force to rebuild from scratch")
	fmt.Println()
	fmt.Println("BACKUP AND RESTORE")
	fmt.Println()
	fmt.Println("  Backup database:")
	fmt.Println("  $ cp ~/.config/sf/sf.db ~/.config/sf/sf.db.backup")
	fmt.Println()
	fmt.Println("  Restore from backup:")
	fmt.Println("  $ cp ~/.config/sf/sf.db.backup ~/.config/sf/sf.db")
	fmt.Println()
	fmt.Println("  SQLite backup:")
	fmt.Println("  $ sqlite3 ~/.config/sf/sf.db \".backup backup.db\"")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help server        Start the HTTP server")
	fmt.Println("  sf help orchestrator  Start the orchestrator daemon")
	fmt.Println("  sf help worker        Start a worker daemon")
	fmt.Println("  sf user               Manage users")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printTUIHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF TUI - Terminal User Interface")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Interactive terminal-based user interface for managing the sf system.")
	fmt.Println("  Provides full-screen menus for managing projects, tasks, roles, users,")
	fmt.Println("  workers, and configuration without requiring CLI commands or a web browser.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf tui [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("  -url <url>        Server URL")
	fmt.Println("                    Default: http://localhost:8080")
	fmt.Println("                    The URL of the sf server to connect to")
	fmt.Println()
	fmt.Println("  -username <name>  Username for authentication")
	fmt.Println("                    Default: $SF_USERNAME environment variable")
	fmt.Println("                    If not provided, interactive login will prompt")
	fmt.Println()
	fmt.Println("  -password <pass>  Password for authentication")
	fmt.Println("                    Default: $SF_PASSWORD environment variable")
	fmt.Println("                    If not provided, interactive login will prompt")
	fmt.Println()
	fmt.Println("  -h                Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # Launch TUI with interactive login")
	fmt.Println("  $ sf tui")
	fmt.Println()
	fmt.Println("  # Launch with environment variable credentials")
	fmt.Println("  $ export SF_USERNAME=admin")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf tui")
	fmt.Println()
	fmt.Println("  # Connect to remote server")
	fmt.Println("  $ export SF_USERNAME=admin")
	fmt.Println("  $ export SF_PASSWORD=secret123")
	fmt.Println("  $ sf tui -url https://sf.example.com")
	fmt.Println()
	fmt.Println("  # Launch with explicit credentials (not recommended for production)")
	fmt.Println("  $ sf tui -username admin -password secret123")
	fmt.Println()
	fmt.Println("  # Connect to custom port")
	fmt.Println("  $ sf tui -url http://localhost:10606")
	fmt.Println()
	fmt.Println("FEATURES")
	fmt.Println()
	fmt.Println("  Main Menu:")
	fmt.Println("    • Projects       - Create, view, edit, delete projects")
	fmt.Println("    • Tasks          - Manage tasks, assignments, dependencies")
	fmt.Println("    • Roles          - Define roles with goals and capabilities")
	fmt.Println("    • Users          - User management (admin only)")
	fmt.Println("    • Workers        - View and manage worker daemons")
	fmt.Println("    • Configuration  - System settings and defaults")
	fmt.Println("    • Logout         - Exit and clear session")
	fmt.Println()
	fmt.Println("  Navigation:")
	fmt.Println("    • Arrow keys     - Move between menu items")
	fmt.Println("    • Enter          - Select menu item")
	fmt.Println("    • Tab            - Navigate between form fields")
	fmt.Println("    • Esc            - Go back/cancel")
	fmt.Println("    • Ctrl+C         - Exit TUI")
	fmt.Println()
	fmt.Println("  Forms and Input:")
	fmt.Println("    • Text fields    - Type normally, backspace to delete")
	fmt.Println("    • Dropdowns      - Arrow keys to select, Enter to confirm")
	fmt.Println("    • Multi-line     - Enter for new line, Ctrl+D to finish")
	fmt.Println()
	fmt.Println("SCREEN LAYOUTS")
	fmt.Println()
	fmt.Println("  Login Screen:")
	fmt.Println("    If credentials not provided, shows login form with username")
	fmt.Println("    and password fields. Also offers registration option if enabled.")
	fmt.Println()
	fmt.Println("  Main Menu:")
	fmt.Println("    Full-screen menu with navigation to all major features.")
	fmt.Println("    Shows current user and server connection in status bar.")
	fmt.Println()
	fmt.Println("  List Views:")
	fmt.Println("    Tabular display of projects/tasks/roles/users with:")
	fmt.Println("    • Search/filter capabilities")
	fmt.Println("    • Sorting options")
	fmt.Println("    • Pagination for large datasets")
	fmt.Println("    • Action shortcuts (create, edit, delete)")
	fmt.Println()
	fmt.Println("  Detail Views:")
	fmt.Println("    Full details with related information:")
	fmt.Println("    • Task details show project, role, assignee, dependencies")
	fmt.Println("    • Project details show task counts and activity")
	fmt.Println("    • User details show roles and assigned tasks")
	fmt.Println()
	fmt.Println("  Form Views:")
	fmt.Println("    Create/edit forms with validation:")
	fmt.Println("    • Required field indicators")
	fmt.Println("    • Real-time validation feedback")
	fmt.Println("    • Help text for complex fields")
	fmt.Println()
	fmt.Println("REQUIREMENTS")
	fmt.Println()
	fmt.Println("  • Server must be running: sf server")
	fmt.Println("  • Database must be initialized: sf initdb")
	fmt.Println("  • Terminal with minimum 80x24 size (larger recommended)")
	fmt.Println("  • User account with appropriate permissions")
	fmt.Println()
	fmt.Println("KEYBOARD SHORTCUTS")
	fmt.Println()
	fmt.Println("  Global:")
	fmt.Println("    Ctrl+C           - Exit TUI immediately")
	fmt.Println("    Esc              - Go back to previous screen")
	fmt.Println("    ?                - Show context help")
	fmt.Println()
	fmt.Println("  Lists:")
	fmt.Println("    ↑/↓              - Navigate items")
	fmt.Println("    Enter            - View/edit item")
	fmt.Println("    n                - Create new item")
	fmt.Println("    d                - Delete item")
	fmt.Println("    /                - Search/filter")
	fmt.Println("    r                - Refresh list")
	fmt.Println()
	fmt.Println("  Forms:")
	fmt.Println("    Tab/Shift+Tab    - Next/previous field")
	fmt.Println("    Enter            - Submit form (when focus on submit button)")
	fmt.Println("    Esc              - Cancel and return")
	fmt.Println()
	fmt.Println("TIPS AND TRICKS")
	fmt.Println()
	fmt.Println("  • Use 'sf login' first to save credentials, then run 'sf tui'")
	fmt.Println("  • Increase terminal size for better experience (120x40 recommended)")
	fmt.Println("  • Use tmux/screen for persistent sessions")
	fmt.Println("  • TUI respects environment variables like CLI commands")
	fmt.Println("  • Configuration screen allows setting default project")
	fmt.Println()
	fmt.Println("COMMON WORKFLOWS")
	fmt.Println()
	fmt.Println("  Create a new project:")
	fmt.Println("    1. Main Menu → Projects → Create New")
	fmt.Println("    2. Fill in name and description")
	fmt.Println("    3. Submit and it becomes default project")
	fmt.Println()
	fmt.Println("  Create and assign a task:")
	fmt.Println("    1. Main Menu → Tasks → Create New")
	fmt.Println("    2. Enter title, description, select project and role")
	fmt.Println("    3. Set priority and status")
	fmt.Println("    4. Submit to create (orchestrator will assign to worker)")
	fmt.Println()
	fmt.Println("  Monitor worker activity:")
	fmt.Println("    1. Main Menu → Workers")
	fmt.Println("    2. View active workers and their tasks")
	fmt.Println("    3. Select worker to see detailed history")
	fmt.Println()
	fmt.Println("  Define a new role:")
	fmt.Println("    1. Main Menu → Roles → Create New")
	fmt.Println("    2. Enter role name")
	fmt.Println("    3. Define goals/capabilities")
	fmt.Println("    4. Submit to make available for task assignment")
	fmt.Println()
	fmt.Println("TROUBLESHOOTING")
	fmt.Println()
	fmt.Println("  Screen garbled or rendering issues:")
	fmt.Println("    • Increase terminal size to minimum 80x24")
	fmt.Println("    • Try Ctrl+L to refresh screen")
	fmt.Println("    • Check TERM environment variable (xterm-256color recommended)")
	fmt.Println()
	fmt.Println("  Cannot connect to server:")
	fmt.Println("    • Ensure server is running: sf server")
	fmt.Println("    • Check server URL: -url http://localhost:8080")
	fmt.Println("    • Verify network connectivity")
	fmt.Println()
	fmt.Println("  Authentication fails:")
	fmt.Println("    • Verify credentials with: sf login")
	fmt.Println("    • Check environment variables: echo $SF_USERNAME")
	fmt.Println("    • Ensure user account exists and is enabled")
	fmt.Println()
	fmt.Println("  TUI freezes or hangs:")
	fmt.Println("    • Use Ctrl+C to force exit")
	fmt.Println("    • Check server logs for errors")
	fmt.Println("    • Verify database is not locked")
	fmt.Println()
	fmt.Println("COMPARISON: TUI vs CLI vs Web UI")
	fmt.Println()
	fmt.Println("  Use TUI when:")
	fmt.Println("    • You prefer interactive, menu-driven interfaces")
	fmt.Println("    • Working over SSH without X11 forwarding")
	fmt.Println("    • Need to browse/manage multiple items")
	fmt.Println("    • Want visual feedback without scripting")
	fmt.Println()
	fmt.Println("  Use CLI when:")
	fmt.Println("    • Automating tasks with scripts")
	fmt.Println("    • Need to pipe output to other commands")
	fmt.Println("    • Working with CI/CD pipelines")
	fmt.Println("    • Prefer command-line efficiency")
	fmt.Println()
	fmt.Println("  Use Web UI when:")
	fmt.Println("    • Rich visualization needed (charts, graphs)")
	fmt.Println("    • Multiple users collaborating")
	fmt.Println("    • Drag-and-drop interactions desired")
	fmt.Println("    • Mobile/tablet access required")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help server    Start the HTTP server and Web UI")
	fmt.Println("  sf login          CLI authentication")
	fmt.Println("  sf project        CLI project management")
	fmt.Println("  sf task           CLI task management")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printProjectHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF PROJECT - Project Management")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Manage project workspaces. Projects organize related tasks and provide")
	fmt.Println("  isolation for different work streams. Each project can have its own set")
	fmt.Println("  of tasks, roles, and team members.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf project <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println("  list|ls               List all projects")
	fmt.Println("  get <id>              Get project details by ID")
	fmt.Println("  create <name>         Create a new project")
	fmt.Println("  update <id>           Update project details")
	fmt.Println("  delete|rm <id>        Delete a project")
	fmt.Println("  set-default <id>      Set default project for commands")
	fmt.Println("  get-default           Show current default project")
	fmt.Println("  unset-default         Clear default project")
	fmt.Println("  file <subcommand>     Manage project files (CRUD)")
	fmt.Println("  note <subcommand>     Manage project notes (CRUD)")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS")
	fmt.Println("  -url <url>            Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>      Username (default: $SF_USERNAME)")
	fmt.Println("  -password <pass>      Password (default: $SF_PASSWORD)")
	fmt.Println("  -json                 Output in JSON format")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # List all projects")
	fmt.Println("  $ sf project list")
	fmt.Println("  $ sf project ls")
	fmt.Println()
	fmt.Println("  # Create a new project")
	fmt.Println("  $ sf project create \"API Development\"")
	fmt.Println()
	fmt.Println("  # Create project with description")
	fmt.Println("  $ sf project create \"Mobile App\" -description \"iOS and Android app\"")
	fmt.Println()
	fmt.Println("  # Get project details")
	fmt.Println("  $ sf project get proj_abc123")
	fmt.Println()
	fmt.Println("  # Update project")
	fmt.Println("  $ sf project update proj_abc123 -name \"API v2\" -description \"REST API v2\"")
	fmt.Println()
	fmt.Println("  # Set as default project (for task creation)")
	fmt.Println("  $ sf project set-default proj_abc123")
	fmt.Println()
	fmt.Println("  # List projects in JSON format")
	fmt.Println("  $ sf project list -json")
	fmt.Println()
	fmt.Println("  # List files for default project")
	fmt.Println("  $ sf project file list")
	fmt.Println()
	fmt.Println("  # Create a project note")
	fmt.Println("  $ sf project note create \"Kickoff\" -content \"Architecture decisions\"")
	fmt.Println()
	fmt.Println("  # Delete a project (warning: deletes all tasks)")
	fmt.Println("  $ sf project delete proj_abc123")
	fmt.Println()
	fmt.Println("AUTHENTICATION")
	fmt.Println()
	fmt.Println("  Use environment variables (recommended):")
	fmt.Println("    $ export SF_USERNAME=admin")
	fmt.Println("    $ export SF_PASSWORD=secret123")
	fmt.Println("    $ sf project list")
	fmt.Println()
	fmt.Println("  Or use sf login to save credentials:")
	fmt.Println("    $ sf login")
	fmt.Println("    $ sf project list")
	fmt.Println()
	fmt.Println("DEFAULT PROJECT")
	fmt.Println()
	fmt.Println("  Setting a default project simplifies task creation:")
	fmt.Println()
	fmt.Println("  Without default:")
	fmt.Println("    $ sf task create \"Fix bug\" -project proj_abc123")
	fmt.Println()
	fmt.Println("  With default:")
	fmt.Println("    $ sf project set-default proj_abc123")
	fmt.Println("    $ sf task create \"Fix bug\"    # Uses default project")
	fmt.Println()
	fmt.Println("TIPS")
	fmt.Println()
	fmt.Println("  • Use descriptive project names")
	fmt.Println("  • Set default project to avoid repetitive -project flags")
	fmt.Println("  • List projects with -json for scripting")
	fmt.Println("  • Projects are workspace-level - all users see all projects")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help task          Task management")
	fmt.Println("  sf help role          Role management")
	fmt.Println("  sf login              Authentication")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printProjectFileHelp() {
	fmt.Println("Project File Commands:")
	fmt.Println("  sf project file list|ls (-project_id <id>)")
	fmt.Println("  sf project file get <file_id> (-project_id <id>)")
	fmt.Println("  sf project file create <name> (-project_id <id>) (-content <text>)")
	fmt.Println("  sf project file update <file_id> (-name <n> -content <text>) (-project_id <id>)")
	fmt.Println("  sf project file delete|rm <file_id> (-project_id <id>)")
}

func printProjectNoteHelp() {
	fmt.Println("Project Note Commands:")
	fmt.Println("  sf project note list|ls (-project_id <id>)")
	fmt.Println("  sf project note get <note_id> (-project_id <id>)")
	fmt.Println("  sf project note create <title> (-project_id <id>) (-content <text>)")
	fmt.Println("  sf project note update <note_id> (-title <t> -content <text>) (-project_id <id>)")
	fmt.Println("  sf project note delete|rm <note_id> (-project_id <id>)")
}

func printTaskHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF TASK - Task Management")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Manage tasks and work items. Tasks represent units of work that can be")
	fmt.Println("  assigned to workers, tracked through their lifecycle, and organized with")
	fmt.Println("  dependencies, comments, and history.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf task <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println()
	fmt.Println("  Task Operations:")
	fmt.Println("    list|ls             List all tasks")
	fmt.Println("    get <id>            Get task details")
	fmt.Println("    create <title>      Create a new task")
	fmt.Println("    update <id>         Update task details")
	fmt.Println("    delete|rm <id>      Delete a task")
	fmt.Println()
	fmt.Println("  Task Lifecycle:")
	fmt.Println("    assign <id>         Assign task to worker")
	fmt.Println("    unassign <id>       Unassign task from worker")
	fmt.Println("    complete <id>       Mark task as completed")
	fmt.Println("    history <id>        View task history")
	fmt.Println()
	fmt.Println("  Task Relationships:")
	fmt.Println("    dependencies <id>   View/manage task dependencies")
	fmt.Println("    comments <id>       View/add task comments")
	fmt.Println()
	fmt.Println("  Worker Management:")
	fmt.Println("    worker list         List all workers")
	fmt.Println("    worker get <id>     Get worker details")
	fmt.Println("    worker tasks <id>   Get tasks assigned to worker")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS")
	fmt.Println("  -url <url>            Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>      Username (default: $SF_USERNAME)")
	fmt.Println("  -password <pass>      Password (default: $SF_PASSWORD)")
	fmt.Println("  -json                 Output in JSON format")
	fmt.Println()
	fmt.Println("CREATE OPTIONS")
	fmt.Println("  -project <id>         Project ID (or use default project)")
	fmt.Println("  -role <id>            Required role for task")
	fmt.Println("  -description <text>   Task description")
	fmt.Println("  -priority <n>         Priority level (1-10, default: 5)")
	fmt.Println("  -status <status>      Initial status (default: idle)")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # List all tasks")
	fmt.Println("  $ sf task list")
	fmt.Println("  $ sf task ls")
	fmt.Println()
	fmt.Println("  # Create a simple task")
	fmt.Println("  $ sf task create \"Fix login bug\"")
	fmt.Println()
	fmt.Println("  # Create task with full details")
	fmt.Println("  $ sf task create \"Build API\" \\")
	fmt.Println("      -project proj_abc123 \\")
	fmt.Println("      -role role_backend \\")
	fmt.Println("      -description \"Build REST API endpoints\" \\")
	fmt.Println("      -priority 8")
	fmt.Println()
	fmt.Println("  # Get task details")
	fmt.Println("  $ sf task get task_xyz789")
	fmt.Println()
	fmt.Println("  # Update task")
	fmt.Println("  $ sf task update task_xyz789 -status active -priority 9")
	fmt.Println()
	fmt.Println("  # Assign task to worker (admin/orchestrator only)")
	fmt.Println("  $ sf task assign task_xyz789 -worker worker1 -role role_backend")
	fmt.Println()
	fmt.Println("  # Mark task as complete")
	fmt.Println("  $ sf task complete task_xyz789")
	fmt.Println()
	fmt.Println("  # View task history")
	fmt.Println("  $ sf task history task_xyz789")
	fmt.Println()
	fmt.Println("  # Add comment to task")
	fmt.Println("  $ sf task comments task_xyz789 -add \"Found the root cause\"")
	fmt.Println()
	fmt.Println("  # List all workers")
	fmt.Println("  $ sf task worker list")
	fmt.Println()
	fmt.Println("  # View worker's tasks")
	fmt.Println("  $ sf task worker tasks worker1")
	fmt.Println()
	fmt.Println("  # List tasks in JSON format")
	fmt.Println("  $ sf task list -json | jq '.[] | select(.status==\"active\")'")
	fmt.Println()
	fmt.Println("TASK LIFECYCLE")
	fmt.Println()
	fmt.Println("  States:")
	fmt.Println("    idle      - Created, waiting for assignment")
	fmt.Println("    active    - Assigned to worker, in progress")
	fmt.Println("    completed - Successfully finished")
	fmt.Println("    failed    - Could not be completed")
	fmt.Println()
	fmt.Println("  Flow:")
	fmt.Println("    1. Create task (status: idle)")
	fmt.Println("    2. Orchestrator assigns to worker (status: active)")
	fmt.Println("    3. Worker completes or fails task")
	fmt.Println("    4. History tracks all transitions")
	fmt.Println()
	fmt.Println("DEPENDENCIES")
	fmt.Println()
	fmt.Println("  Tasks can depend on other tasks being completed first:")
	fmt.Println()
	fmt.Println("  $ sf task dependencies task_xyz789 -add task_abc123")
	fmt.Println()
	fmt.Println("  The orchestrator won't assign task_xyz789 until task_abc123 completes.")
	fmt.Println()
	fmt.Println("TIPS")
	fmt.Println()
	fmt.Println("  • Set default project: sf project set-default <id>")
	fmt.Println("  • Use priorities (1-10) to influence orchestrator assignment")
	fmt.Println("  • Add comments for collaboration and debugging")
	fmt.Println("  • Check history to understand task lifecycle")
	fmt.Println("  • Use -json with jq for advanced filtering")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help project       Project management")
	fmt.Println("  sf help role          Role management")
	fmt.Println("  sf help worker        Worker daemon")
	fmt.Println("  sf help orchestrator  Task assignment engine")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printCommentHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF COMMENT - Social Comments")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Create and manage comments attached to projects or tasks.")
	fmt.Println("  Comments track owner, timestamps, edit history, and soft-delete state.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf comment <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println("  list|ls               List comments for an entity")
	fmt.Println("  get <comment_id>      Get comment details")
	fmt.Println("  create                Create a comment")
	fmt.Println("  update <comment_id>   Edit your own comment")
	fmt.Println("  delete|rm <comment_id> Soft-delete your own comment")
	fmt.Println("  history <comment_id>  Show comment edit/delete history")
	fmt.Println()
	fmt.Println("TARGET OPTIONS")
	fmt.Println("  -project_id <id>      Target a project comment thread")
	fmt.Println("  -task_id <id>         Target a task comment thread")
	fmt.Println("  -entity_type <type>   project|task (alternative)")
	fmt.Println("  -entity_id <id>       Entity ID (alternative)")
	fmt.Println()
	fmt.Println("COMMENT OPTIONS")
	fmt.Println("  -text <text>          Comment text")
	fmt.Println("  -include_deleted      Include soft-deleted comments in list")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println("  $ sf comment list -project_id proj_abc123")
	fmt.Println("  $ sf comment create -task_id task_xyz -text \"Looks good\"")
	fmt.Println("  $ sf comment update cmt_123 -text \"Updated note\"")
	fmt.Println("  $ sf comment history cmt_123")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printUserHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF USER - User Management")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Manage user accounts. Users can be admins (full access), humans")
	fmt.Println("  (interactive users), workers (daemon accounts), or orchestrators")
	fmt.Println("  (task assignment daemons). Admin privileges required.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf user <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println("  list|ls               List all users")
	fmt.Println("  get <id>              Get user details by ID")
	fmt.Println("  create <username>     Create a new user")
	fmt.Println("  enable <username>     Enable a user account")
	fmt.Println("  disable <username>    Disable a user account")
	fmt.Println("  delete|rm <username>  Soft-delete user account (disable)")
	fmt.Println("  reset-password <id>   Reset user password (admin only)")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS")
	fmt.Println("  -url <url>            Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>      Admin username (default: $SF_USERNAME)")
	fmt.Println("  -password <pass>      Admin password (default: $SF_PASSWORD)")
	fmt.Println("  -json                 Output in JSON format")
	fmt.Println()
	fmt.Println("CREATE OPTIONS")
	fmt.Println("  -type <type>          User type: admin, human, worker, orchestrator")
	fmt.Println("                        Default: human")
	fmt.Println("  -password <pass>      Set user password (if not provided, random)")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # List all users")
	fmt.Println("  $ sf user list")
	fmt.Println()
	fmt.Println("  # Create a human user (developer)")
	fmt.Println("  $ sf user create alice -type human")
	fmt.Println()
	fmt.Println("  # Create a worker account")
	fmt.Println("  $ sf user create worker3 -type worker -password worker123")
	fmt.Println()
	fmt.Println("  # Create admin user")
	fmt.Println("  $ sf user create bob -type admin -password admin456")
	fmt.Println()
	fmt.Println("  # Get user details")
	fmt.Println("  $ sf user get user_abc123")
	fmt.Println()
	fmt.Println("  # Disable a user")
	fmt.Println("  $ sf user disable alice")
	fmt.Println()
	fmt.Println("  # Re-enable a user")
	fmt.Println("  $ sf user enable alice")
	fmt.Println()
	fmt.Println("  # Reset user password")
	fmt.Println("  $ sf user reset-password user_abc123 -password newpass789")
	fmt.Println()
	fmt.Println("  # List users in JSON format")
	fmt.Println("  $ sf user list -json")
	fmt.Println()
	fmt.Println("USER TYPES")
	fmt.Println()
	fmt.Println("  admin:")
	fmt.Println("    • Full system access")
	fmt.Println("    • Can manage users, projects, tasks, roles")
	fmt.Println("    • Can manually assign/unassign tasks")
	fmt.Println("    • Intended for administrators")
	fmt.Println()
	fmt.Println("  human:")
	fmt.Println("    • Interactive user accounts")
	fmt.Println("    • Can view and create projects/tasks/roles")
	fmt.Println("    • Cannot manage other users")
	fmt.Println("    • Intended for developers and team members")
	fmt.Println()
	fmt.Println("  worker:")
	fmt.Println("    • Daemon accounts for worker processes")
	fmt.Println("    • Can request task assignments")
	fmt.Println("    • Can claim, complete, and return tasks")
	fmt.Println("    • Intended for automated worker daemons")
	fmt.Println()
	fmt.Println("  orchestrator:")
	fmt.Println("    • Daemon account for orchestrator process")
	fmt.Println("    • Can assign tasks to workers")
	fmt.Println("    • Monitors worker heartbeats")
	fmt.Println("    • Intended for orchestrator daemon")
	fmt.Println()
	fmt.Println("AUTHENTICATION")
	fmt.Println()
	fmt.Println("  User management requires admin privileges:")
	fmt.Println()
	fmt.Println("  $ export SF_USERNAME=admin")
	fmt.Println("  $ export SF_PASSWORD=admin_password")
	fmt.Println("  $ sf user create developer1 -type human")
	fmt.Println()
	fmt.Println("PASSWORD MANAGEMENT")
	fmt.Println()
	fmt.Println("  When creating users:")
	fmt.Println("    • Without -password: Random password generated and printed")
	fmt.Println("    • With -password: Use specified password")
	fmt.Println()
	fmt.Println("  Security recommendations:")
	fmt.Println("    • Use strong, random passwords for production")
	fmt.Println("    • Use -password only for development/testing")
	fmt.Println("    • Store passwords securely (password manager, vault)")
	fmt.Println("    • Enable passkeys (WebAuthn) for passwordless auth")
	fmt.Println()
	fmt.Println("DISABLING USERS")
	fmt.Println()
	fmt.Println("  Disabled users:")
	fmt.Println("    • Cannot authenticate")
	fmt.Println("    • Existing sessions remain valid until expiry")
	fmt.Println("    • Can be re-enabled with 'sf user enable'")
	fmt.Println("    • Data is preserved (not deleted)")
	fmt.Println()
	fmt.Println("TIPS")
	fmt.Println()
	fmt.Println("  • Create worker accounts before starting worker daemons")
	fmt.Println("  • Use descriptive usernames (alice, worker1, ci-agent-01)")
	fmt.Println("  • Disable unused accounts instead of deleting")
	fmt.Println("  • Use -json for scripting bulk operations")
	fmt.Println("  • Regular users can view user list but can't modify")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help worker        Worker daemon")
	fmt.Println("  sf help orchestrator  Orchestrator daemon")
	fmt.Println("  sf login              Authentication")
	fmt.Println("  sf register           Self-service registration")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printRoleHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF ROLE - Role Management")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Manage role definitions. Roles define capabilities, skills, and goals")
	fmt.Println("  that workers must have to be assigned specific tasks. The orchestrator")
	fmt.Println("  uses roles to match tasks with appropriate workers.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf role <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println("  list|ls               List all roles")
	fmt.Println("  get <id>              Get role details by ID")
	fmt.Println("  create <name>         Create a new role")
	fmt.Println("  update <id>           Update role details")
	fmt.Println("  delete|rm <id>        Delete a role")
	fmt.Println("  history <id>          View role change history")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS")
	fmt.Println("  -url <url>            Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>      Username (default: $SF_USERNAME)")
	fmt.Println("  -password <pass>      Password (default: $SF_PASSWORD)")
	fmt.Println("  -json                 Output in JSON format")
	fmt.Println()
	fmt.Println("CREATE/UPDATE OPTIONS")
	fmt.Println("  -goals <text>         Role goals and objectives")
	fmt.Println("  -description <text>   Role description")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # List all roles")
	fmt.Println("  $ sf role list")
	fmt.Println()
	fmt.Println("  # Create a role")
	fmt.Println("  $ sf role create \"Backend Developer\" \\")
	fmt.Println("      -goals \"Build and maintain REST APIs\"")
	fmt.Println()
	fmt.Println("  # Create role with full details")
	fmt.Println("  $ sf role create \"Frontend Developer\" \\")
	fmt.Println("      -goals \"Build responsive web interfaces\" \\")
	fmt.Println("      -description \"React, TypeScript, CSS expert\"")
	fmt.Println()
	fmt.Println("  # Get role details")
	fmt.Println("  $ sf role get role_abc123")
	fmt.Println()
	fmt.Println("  # Update role")
	fmt.Println("  $ sf role update role_abc123 \\")
	fmt.Println("      -goals \"Build APIs and microservices\"")
	fmt.Println()
	fmt.Println("  # Delete a role")
	fmt.Println("  $ sf role delete role_abc123")
	fmt.Println()
	fmt.Println("  # View role history")
	fmt.Println("  $ sf role history role_abc123")
	fmt.Println()
	fmt.Println("  # List roles in JSON format")
	fmt.Println("  $ sf role list -json")
	fmt.Println()
	fmt.Println("HOW ROLES WORK")
	fmt.Println()
	fmt.Println("  Task Assignment:")
	fmt.Println("    1. Task created with required role: role_backend")
	fmt.Println("    2. Orchestrator finds workers with role_backend capability")
	fmt.Println("    3. Orchestrator assigns task to available worker")
	fmt.Println("    4. Worker executes task based on role requirements")
	fmt.Println()
	fmt.Println("  Worker Configuration:")
	fmt.Println("    Workers must be configured with roles they can fulfill:")
	fmt.Println("    • Via database during initialization (sf initdb -populate)")
	fmt.Println("    • Via task worker assign command")
	fmt.Println("    • Via database user_roles table")
	fmt.Println()
	fmt.Println("ROLE DESIGN")
	fmt.Println()
	fmt.Println("  Good role definitions are:")
	fmt.Println("    • Specific - \"Backend API Developer\" vs \"Developer\"")
	fmt.Println("    • Skill-based - Define required capabilities clearly")
	fmt.Println("    • Goal-oriented - What outcomes should be achieved")
	fmt.Println("    • Matchable - Workers can self-identify capability")
	fmt.Println()
	fmt.Println("  Example roles:")
	fmt.Println("    • \"Python Developer\" - goals: \"Write Python code, unit tests\"")
	fmt.Println("    • \"DevOps Engineer\" - goals: \"Deploy, monitor, scale services\"")
	fmt.Println("    • \"QA Tester\" - goals: \"Test features, report bugs\"")
	fmt.Println("    • \"Technical Writer\" - goals: \"Write documentation\"")
	fmt.Println("    • \"Code Reviewer\" - goals: \"Review code, ensure quality\"")
	fmt.Println()
	fmt.Println("GOALS FIELD")
	fmt.Println()
	fmt.Println("  The goals field is critical for AI/LLM workers:")
	fmt.Println()
	fmt.Println("  • Describes what the role should accomplish")
	fmt.Println("  • Provides context for task execution")
	fmt.Println("  • Guides worker behavior and decision-making")
	fmt.Println("  • Can include success criteria and constraints")
	fmt.Println()
	fmt.Println("  Example goals:")
	fmt.Println("    \"Build REST APIs following OpenAPI 3.0 spec. Write unit tests.")
	fmt.Println("    Ensure proper error handling and input validation.\"")
	fmt.Println()
	fmt.Println("ROLE HIERARCHY")
	fmt.Println()
	fmt.Println("  Roles are flat (no hierarchy), but you can model specialization:")
	fmt.Println()
	fmt.Println("  • \"Developer\" - General coding")
	fmt.Println("  • \"Backend Developer\" - API/database work")
	fmt.Println("  • \"Python Developer\" - Python-specific work")
	fmt.Println()
	fmt.Println("  Workers can have multiple roles for flexibility.")
	fmt.Println()
	fmt.Println("TIPS")
	fmt.Println()
	fmt.Println("  • Define roles before creating tasks")
	fmt.Println("  • Use clear, descriptive role names")
	fmt.Println("  • Write detailed goals for better matching")
	fmt.Println("  • Review role history to track changes")
	fmt.Println("  • One role can be assigned to many workers")
	fmt.Println("  • Tasks require exactly one role")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help task          Task management")
	fmt.Println("  sf help worker        Worker daemon")
	fmt.Println("  sf help orchestrator  Task assignment engine")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printConfigHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF CONFIG - Configuration Management")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Manage system configuration. Configuration is stored as key-value pairs")
	fmt.Println("  in the database and can be accessed by all components. Admin privileges")
	fmt.Println("  required for modifications.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf config <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println("  list|ls               List all configuration settings")
	fmt.Println("  get <key>             Get configuration value by key")
	fmt.Println("  set <key> <value>     Set configuration value (admin only)")
	fmt.Println("  delete|rm <key>       Delete configuration key (admin only)")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS")
	fmt.Println("  -url <url>            Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>      Username (default: $SF_USERNAME)")
	fmt.Println("  -password <pass>      Password (default: $SF_PASSWORD)")
	fmt.Println("  -json                 Output in JSON format")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println()
	fmt.Println("  # List all configuration")
	fmt.Println("  $ sf config list")
	fmt.Println("  $ sf config ls")
	fmt.Println()
	fmt.Println("  # Get a specific value")
	fmt.Println("  $ sf config get default_project_id")
	fmt.Println()
	fmt.Println("  # Set a configuration value")
	fmt.Println("  $ sf config set max_workers 10")
	fmt.Println()
	fmt.Println("  # Set a string value")
	fmt.Println("  $ sf config set organization_name \"Acme Corp\"")
	fmt.Println()
	fmt.Println("  # Delete a configuration key")
	fmt.Println("  $ sf config delete deprecated_setting")
	fmt.Println("  $ sf config rm deprecated_setting")
	fmt.Println()
	fmt.Println("  # List in JSON format")
	fmt.Println("  $ sf config list -json")
	fmt.Println()
	fmt.Println("  # Use in scripts")
	fmt.Println("  $ DEFAULT_PROJECT=$(sf config get default_project_id -json | jq -r .value)")
	fmt.Println()
	fmt.Println("COMMON CONFIGURATION KEYS")
	fmt.Println()
	fmt.Println("  default_project_id:")
	fmt.Println("    Default project for task creation")
	fmt.Println("    Set via: sf project set-default <id>")
	fmt.Println()
	fmt.Println("  registration.enabled:")
	fmt.Println("    Allow self-service user registration")
	fmt.Println("    Values: true, false")
	fmt.Println()
	fmt.Println("  orchestrator.enabled:")
	fmt.Println("    Enable automatic task assignment")
	fmt.Println("    Values: true, false")
	fmt.Println()
	fmt.Println("  worker.heartbeat_interval:")
	fmt.Println("    How often workers send heartbeats (seconds)")
	fmt.Println("    Default: 30")
	fmt.Println()
	fmt.Println("  worker.heartbeat_timeout:")
	fmt.Println("    When to consider worker inactive (seconds)")
	fmt.Println("    Default: 120")
	fmt.Println()
	fmt.Println("  task.default_priority:")
	fmt.Println("    Default priority for new tasks")
	fmt.Println("    Values: 1-10")
	fmt.Println()
	fmt.Println("CONFIGURATION SCOPE")
	fmt.Println()
	fmt.Println("  • System-wide - All users and components see same config")
	fmt.Println("  • No user-specific settings (use user table fields instead)")
	fmt.Println("  • No per-project settings (use project table fields instead)")
	fmt.Println()
	fmt.Println("AUTHENTICATION")
	fmt.Println()
	fmt.Println("  Viewing config:")
	fmt.Println("    Any authenticated user can list and get config values")
	fmt.Println()
	fmt.Println("  Modifying config:")
	fmt.Println("    Only admin users can set or delete config values")
	fmt.Println()
	fmt.Println("  Example:")
	fmt.Println("    $ export SF_USERNAME=admin")
	fmt.Println("    $ export SF_PASSWORD=admin_password")
	fmt.Println("    $ sf config set feature.new_ui true")
	fmt.Println()
	fmt.Println("VALUE TYPES")
	fmt.Println()
	fmt.Println("  Configuration values are stored as strings but can represent:")
	fmt.Println()
	fmt.Println("  • Strings: \"hello world\"")
	fmt.Println("  • Numbers: \"42\", \"3.14\"")
	fmt.Println("  • Booleans: \"true\", \"false\"")
	fmt.Println("  • JSON: '{\"key\": \"value\"}'")
	fmt.Println()
	fmt.Println("  Your application must parse values appropriately.")
	fmt.Println()
	fmt.Println("TIPS")
	fmt.Println()
	fmt.Println("  • Use namespaced keys: feature.*, worker.*, orchestrator.*")
	fmt.Println("  • Document configuration in README or docs/")
	fmt.Println("  • Use -json with jq for scripting")
	fmt.Println("  • Set sensible defaults in application code")
	fmt.Println("  • Don't store secrets (use environment variables)")
	fmt.Println()
	fmt.Println("CUSTOM CONFIGURATION")
	fmt.Println()
	fmt.Println("  You can add custom configuration for your application:")
	fmt.Println()
	fmt.Println("  $ sf config set app.theme dark")
	fmt.Println("  $ sf config set app.max_retries 3")
	fmt.Println("  $ sf config set app.api_endpoint https://api.example.com")
	fmt.Println()
	fmt.Println("  Then read in your application:")
	fmt.Println()
	fmt.Println("  theme=$(sf config get app.theme -json | jq -r .value)")
	fmt.Println()
	fmt.Println("SEE ALSO")
	fmt.Println()
	fmt.Println("  sf help server        Server configuration")
	fmt.Println("  sf help project       Project management")
	fmt.Println("  sf help initdb        Database initialization")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printBeadsHelp() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  SF BEADS - Local Beads Markdown Workflow")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("  Manage bead descriptors through a local markdown file.")
	fmt.Println("  This command does not call the sf server or require authentication.")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("  sf beads <subcommand> [options]")
	fmt.Println("  sf bead  <subcommand> [options]")
	fmt.Println("  sf bd    <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS")
	fmt.Println("  list|ls                 List beads in markdown file (default: TODO.md)")
	fmt.Println("  format|fmt|tidy|fix     Add missing fields in markdown file")
	fmt.Println("  test|validate           Validate required fields in markdown file")
	fmt.Println("  import                  Print bd commands from markdown (-apply to execute)")
	fmt.Println("  export                  Export bd issues to markdown file")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("  -f <filename>           Markdown file path (default: TODO.md)")
	fmt.Println("  -apply                  Execute generated bd commands (import only)")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println("  sf beads list")
	fmt.Println("  sf beads format")
	fmt.Println("  sf beads validate")
	fmt.Println("  sf beads import")
	fmt.Println("  sf beads import -apply")
	fmt.Println("  sf beads export")
	fmt.Println("  sf beads export -f beads.md")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	dbPath := fs.String("f", "", "Path to database file (default: ~/.config/sf/sf.db)")
	port := fs.Int("port", 8080, "Server port")
	verbose := fs.Bool("v", false, "Verbose request/response logging to stdout")
	help := fs.Bool("h", false, "Show help")

	fs.Usage = func() {
		printServerHelp()
	}

	fs.Parse(args)

	if *help {
		printServerHelp()
		return
	}

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
	srv.SetVerbose(*verbose)

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
	help := fs.Bool("h", false, "Show help")

	fs.Usage = func() {
		printOrchestratorHelp()
	}

	fs.Parse(args)

	if *help {
		printOrchestratorHelp()
		return
	}

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
	workerID := fs.String("worker_id", os.Getenv("SF_USERNAME"), "Worker ID")
	password := fs.String("password", os.Getenv("SF_PASSWORD"), "Password")
	help := fs.Bool("h", false, "Show help")

	fs.Usage = func() {
		printWorkerHelp()
	}

	fs.Parse(args)

	if *help {
		printWorkerHelp()
		return
	}

	if *workerID == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: SF_USERNAME and SF_PASSWORD required (or use -worker_id and -password flags)")
		os.Exit(1)
	}

	w := worker.New(*serverURL, *workerID, *password)
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
	help := fs.Bool("h", false, "Show help")

	fs.Usage = func() {
		printInitDBHelp()
	}

	fs.Parse(args)

	if *help {
		printInitDBHelp()
		return
	}

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
	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printTUIHelp()
			return
		}
	}

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
	subArgs := filterNonFlags(args[1:])

	// Parse config from all remaining args
	config, err := cli.NewConfig(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Consistent no-arg UX for command groups: show command help/subcommands.
	if len(subArgs) == 0 {
		switch command {
		case "project":
			showCommandHelp(command)
			printCurrentDefaultProject(config)
			return
		case "epic":
			showCommandHelp(command)
			printCurrentDefaultProject(config)
			return
		case "bug":
			showCommandHelp(command)
			printCurrentDefaultProject(config)
			return
		case "chore":
			showCommandHelp(command)
			printCurrentDefaultProject(config)
			return
		case "task":
			showCommandHelp(command)
			printCurrentDefaultProject(config)
			return
		case "comment", "user", "worker", "role", "config", "beads", "bead", "bd":
			showCommandHelp(command)
			return
		}
	}

	// Check for -h flag before requiring authentication
	for _, arg := range args[1:] {
		if arg == "-h" {
			showCommandHelp(command)
			return
		}
	}

	// Login, register, logout and local beads commands don't require prior authentication
	if command == "login" {
		handleLoginCommand(config)
		return
	}
	if command == "register" {
		handleRegisterCommand(config)
		return
	}
	if command == "logout" {
		handleLogoutCommand()
		return
	}
	if command == "beads" || command == "bead" || command == "bd" {
		handleBeadsCommand(subArgs)
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

	switch command {
	case "project":
		handleProjectCommand(client, config, subArgs)
	case "task":
		handleTaskCommand(client, config, subArgs)
	case "epic":
		handleTaskCommand(client, config, withImpliedTaskType(subArgs, "epic"))
	case "bug":
		handleTaskCommand(client, config, withImpliedTaskType(subArgs, "bug"))
	case "chore":
		handleTaskCommand(client, config, withImpliedTaskType(subArgs, "chore"))
	case "comment":
		handleCommentCommand(client, config, subArgs)
	case "user":
		handleUserCommand(client, config, subArgs)
	case "worker":
		handleWorkerCommand(client, config, subArgs)
	case "role":
		handleRoleCommand(client, config, subArgs)
	case "config":
		handleConfigCommand(client, config, subArgs)
	case "status":
		handleStatusCommand(client, config)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printCLIUsage()
		os.Exit(1)
	}
}

func printCurrentDefaultProject(config *cli.Config) {
	current := config.ProjectID
	if strings.TrimSpace(current) == "" {
		current = "default"
	}
	fmt.Printf("\nCurrent default project: %s\n", current)
}

func filterNonFlags(args []string) []string {
	var result []string
	skipNext := false
	seenSubcommand := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-json" || arg == "-v" {
			continue
		}
		if !seenSubcommand {
			if arg == "-url" || arg == "-username" || arg == "-password" {
				skipNext = true
				continue
			}
			seenSubcommand = true
		}
		result = append(result, arg)
	}
	return result
}

func withImpliedTaskType(args []string, impliedType string) []string {
	for i := 0; i < len(args); i++ {
		if args[i] == "-type" && i+1 < len(args) {
			return args
		}
	}
	out := make([]string, 0, len(args)+2)
	out = append(out, args...)
	out = append(out, "-type", impliedType)
	return out
}

func printCLIUsage() {
	fmt.Println("SF CLI Commands:")
	fmt.Println()
	fmt.Println("Authentication:")
	fmt.Println("  sf login                          Login and save session token")
	fmt.Println("  sf register                       Register a new account")
	fmt.Println("  sf logout                         Logout and clear saved credentials")
	fmt.Println()
	fmt.Println("Projects:")
	fmt.Println("  sf project list|ls (-project_id <id> -name <n> -description <d>)  List projects")
	fmt.Println("  sf project get -project_id <id>   Get project details")
	fmt.Println("  sf project create -project_id <id> -name <n>  Create new project")
	fmt.Println("  sf project update -project_id <id> (-name <n> -description <d>)  Update project")
	fmt.Println("  sf project delete|rm -project_id <id>  Delete project")
	fmt.Println("  sf project set <name>             Set active project")
	fmt.Println("  sf project unset                  Unset active project")
	fmt.Println("  sf project file list|ls (-project_id <id>)")
	fmt.Println("  sf project file get <file_id> (-project_id <id>)")
	fmt.Println("  sf project file create <name> (-project_id <id>) (-content <text>)")
	fmt.Println("  sf project file update <file_id> (-name <n> -content <text>) (-project_id <id>)")
	fmt.Println("  sf project file delete|rm <file_id> (-project_id <id>)")
	fmt.Println("  sf project note list|ls (-project_id <id>)")
	fmt.Println("  sf project note get <note_id> (-project_id <id>)")
	fmt.Println("  sf project note create <title> (-project_id <id>) (-content <text>)")
	fmt.Println("  sf project note update <note_id> (-title <t> -content <text>) (-project_id <id>)")
	fmt.Println("  sf project note delete|rm <note_id> (-project_id <id>)")
	fmt.Println()
	fmt.Println("Tasks:")
	fmt.Println("  sf task list|ls (-project_id <p> -status <s> -type <t> -owner <o>)  List tasks")
	fmt.Println("  sf task get -task_id <id>         Get task details")
	fmt.Println("  sf task create -title <t>         Create new task")
	fmt.Println("  sf task update -task_id <id> ...  Update task")
	fmt.Println("  sf task delete|rm -task_id <id>   Delete task")
	fmt.Println("  sf task assign -task_id <id> -worker_id <w> -role <r>  Assign task")
	fmt.Println("  sf task unassign -task_id <id> -worker_id <w>  Unassign task")
	fmt.Println("  sf task request                   Request a task to work on")
	fmt.Println("  sf task return -task_id <id> ...  Return completed task")
	fmt.Println("  sf task comment -task_id <id> -comment <text>  Add comment to task")
	fmt.Println("  sf task history -task_id <id>     Show task history")
	fmt.Println()
	fmt.Println("Epics:")
	fmt.Println("  sf epic <task-subcommand>         Alias of sf task with implied -type epic")
	fmt.Println("  sf epic create <title>            Create epic task")
	fmt.Println("  sf epic list|ls                   List epic tasks")
	fmt.Println()
	fmt.Println("Bugs:")
	fmt.Println("  sf bug <task-subcommand>          Alias of sf task with implied -type bug")
	fmt.Println("  sf bug create <title>             Create bug task")
	fmt.Println("  sf bug list|ls                    List bug tasks")
	fmt.Println()
	fmt.Println("Chores:")
	fmt.Println("  sf chore <task-subcommand>        Alias of sf task with implied -type chore")
	fmt.Println("  sf chore create <title>           Create chore task")
	fmt.Println("  sf chore list|ls                  List chore tasks")
	fmt.Println()
	fmt.Println("Comments:")
	fmt.Println("  sf comment list|ls (-project_id <id> | -task_id <id>)")
	fmt.Println("  sf comment get <comment_id>")
	fmt.Println("  sf comment create (-project_id <id> | -task_id <id>) -text <text>")
	fmt.Println("  sf comment update <comment_id> -text <text>")
	fmt.Println("  sf comment delete|rm <comment_id>")
	fmt.Println("  sf comment history <comment_id>")
	fmt.Println()
	fmt.Println("Users:")
	fmt.Println("  sf user list|ls                   List all users")
	fmt.Println("  sf user get -user_id <id>         Get user details")
	fmt.Println("  sf user create -username <u> -password <p>  Create user")
	fmt.Println("  sf user enable -username <u>      Enable user")
	fmt.Println("  sf user disable -username <u>     Disable user")
	fmt.Println("  sf user delete|rm -username <u>   Soft-delete user (disable)")
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
	fmt.Println("  sf role list|ls                   List all roles")
	fmt.Println("  sf role get -role_id <id>         Get role details")
	fmt.Println("  sf role create -title <t> -description <d> -goals <g>  Create role")
	fmt.Println("  sf role update -role_id <id> (-title <t> -description <d> -goals <g>)  Update role")
	fmt.Println("  sf role delete|rm -role_id <id>   Delete role")
	fmt.Println("  sf role history -role_id <id>     Show role history")
	fmt.Println()
	fmt.Println("Config:")
	fmt.Println("  sf config list|ls                 List all config")
	fmt.Println("  sf config set -key K -val V       Set config value")
	fmt.Println("  sf config delete|rm -key K        Delete config value")
	fmt.Println()
	fmt.Println()
	fmt.Println("  sf config list|ls            List all config")
	fmt.Println("  sf config set -key K -val V  Set config value")
	fmt.Println("  sf config delete|rm -key K   Delete config value")
	fmt.Println()
	fmt.Println("Beads:")
	fmt.Println("  sf beads                         List beads (aliases: sf bead, sf bd)")
	fmt.Println("  sf beads list (-f <file>)       List beads from markdown file (default: TODO.md)")
	fmt.Println("  sf beads format (-f <file>)     Normalize markdown bead blocks")
	fmt.Println("  sf beads validate (-f <file>)   Validate markdown bead blocks")
	fmt.Println("  sf beads import (-f <file>)     Print bd commands from markdown")
	fmt.Println("  sf beads import (-f <file>) -apply  Apply commands via bd and write IDs")
	fmt.Println("  sf beads export (-f <file>)     Export bd issues into markdown format")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  -url <url>        Server URL (default: http://localhost:8080)")
	fmt.Println("  -username <name>  Username (or set SF_USERNAME)")
	fmt.Println("  -password <pass>  Password (or set SF_PASSWORD)")
	fmt.Println("  -json             Output as JSON")
	fmt.Println("  -v                Show verbose HTTP request/response")
}

func handleProjectCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		printProjectHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "file", "files":
		handleProjectFileCommand(client, config, args[1:])
	case "note", "notes":
		handleProjectNoteCommand(client, config, args[1:])
	case "list", "ls":
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
			printResponseData(data, config)
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
		id := extractFlag(args, "-project_id")
		if id == "" && len(args) > 1 {
			id = args[1]
		}
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -project_id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/projects/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
	case "create":
		projectID := extractFlag(args, "-project_id")
		name := extractFlag(args, "-name")
		if name == "" && len(args) > 1 {
			name = args[1]
		}
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: -name flag required")
			os.Exit(1)
		}
		if projectID == "" {
			projectID = slugifyIdentifier(name)
		}
		desc := extractFlag(args, "-description")
		if desc == "" {
			desc = name
		}
		body := map[string]string{"id": projectID, "name": name}
		body["description"] = desc
		data, err := client.Request("POST", "/api/v1/projects", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project created:")
		printResponseData(data, config)
	case "update":
		projectID := extractFlag(args, "-project_id")
		if projectID == "" && len(args) > 1 {
			projectID = args[1]
		}
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
		printResponseData(data, config)
	case "delete", "rm":
		projectID := extractFlag(args, "-project_id")
		if projectID == "" && len(args) > 1 {
			projectID = args[1]
		}
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
		projectID := extractFlag(args, "-project_id")
		projectName := extractFlag(args, "-name")

		// Positional fallback remains supported for compatibility.
		if projectID == "" && projectName == "" && len(args) > 1 {
			projectID = args[1]
		}

		if projectID == "" && projectName == "" {
			fmt.Fprintln(os.Stderr, "Error: -project_id or -name required")
			os.Exit(1)
		}

		if projectID == "" && projectName != "" {
			// Resolve project name to ID.
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
			for _, p := range projects {
				name, okName := p["name"].(string)
				id, okID := p["id"].(string)
				if okName && okID && name == projectName {
					projectID = id
					break
				}
			}
			if projectID == "" {
				fmt.Fprintf(os.Stderr, "Error: project '%s' not found\n", projectName)
				os.Exit(1)
			}
		}

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
			printResponseData(data, config)
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

func resolveProjectIDForChildResource(args []string, config *cli.Config) string {
	projectID := extractFlag(args, "-project_id")
	if projectID != "" {
		return projectID
	}
	if strings.TrimSpace(config.ProjectID) != "" {
		return config.ProjectID
	}
	return "default"
}

func handleProjectFileCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		printProjectFileHelp()
		printCurrentDefaultProject(config)
		return
	}

	subcommand := args[0]
	projectID := resolveProjectIDForChildResource(args, config)
	basePath := "/api/v1/projects/" + projectID + "/files"

	switch subcommand {
	case "list", "ls":
		data, err := client.Request("GET", basePath, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
			return
		}
		var files []map[string]interface{}
		if err := parseJSON(data, &files); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d files in project %s:\n\n", len(files), projectID)
		for _, f := range files {
			fmt.Printf("ID:   %v\n", f["id"])
			fmt.Printf("Name: %v\n", f["name"])
			fmt.Println()
		}
	case "get":
		fileID := extractFlag(args, "-file_id")
		if fileID == "" && len(args) > 1 {
			fileID = args[1]
		}
		if fileID == "" {
			fmt.Fprintln(os.Stderr, "Error: file ID required")
			os.Exit(1)
		}
		data, err := client.Request("GET", basePath+"/"+fileID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
	case "create":
		name := extractFlag(args, "-name")
		if name == "" && len(args) > 1 {
			name = args[1]
		}
		if name == "" {
			fmt.Fprintln(os.Stderr, "Error: file name required")
			os.Exit(1)
		}
		body := map[string]interface{}{
			"name":    name,
			"content": extractFlag(args, "-content"),
		}
		data, err := client.Request("POST", basePath, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project file created:")
		printResponseData(data, config)
	case "update":
		fileID := extractFlag(args, "-file_id")
		if fileID == "" && len(args) > 1 {
			fileID = args[1]
		}
		if fileID == "" {
			fmt.Fprintln(os.Stderr, "Error: file ID required")
			os.Exit(1)
		}
		body := map[string]interface{}{}
		if name := extractFlag(args, "-name"); name != "" {
			body["name"] = name
		}
		if content := extractFlag(args, "-content"); content != "" {
			body["content"] = content
		}
		if len(body) == 0 {
			fmt.Fprintln(os.Stderr, "Error: at least one field to update required")
			os.Exit(1)
		}
		data, err := client.Request("PUT", basePath+"/"+fileID, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project file updated:")
		printResponseData(data, config)
	case "delete", "rm":
		fileID := extractFlag(args, "-file_id")
		if fileID == "" && len(args) > 1 {
			fileID = args[1]
		}
		if fileID == "" {
			fmt.Fprintln(os.Stderr, "Error: file ID required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", basePath+"/"+fileID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Project file '%s' deleted\n", fileID)
	default:
		fmt.Fprintf(os.Stderr, "Unknown project file subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleProjectNoteCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		printProjectNoteHelp()
		printCurrentDefaultProject(config)
		return
	}

	subcommand := args[0]
	projectID := resolveProjectIDForChildResource(args, config)
	basePath := "/api/v1/projects/" + projectID + "/notes"

	switch subcommand {
	case "list", "ls":
		data, err := client.Request("GET", basePath, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
			return
		}
		var notes []map[string]interface{}
		if err := parseJSON(data, &notes); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d notes in project %s:\n\n", len(notes), projectID)
		for _, n := range notes {
			fmt.Printf("ID:    %v\n", n["id"])
			fmt.Printf("Title: %v\n", n["title"])
			fmt.Println()
		}
	case "get":
		noteID := extractFlag(args, "-note_id")
		if noteID == "" && len(args) > 1 {
			noteID = args[1]
		}
		if noteID == "" {
			fmt.Fprintln(os.Stderr, "Error: note ID required")
			os.Exit(1)
		}
		data, err := client.Request("GET", basePath+"/"+noteID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
	case "create":
		title := extractFlag(args, "-title")
		if title == "" && len(args) > 1 {
			title = args[1]
		}
		if title == "" {
			fmt.Fprintln(os.Stderr, "Error: note title required")
			os.Exit(1)
		}
		body := map[string]interface{}{
			"title":   title,
			"content": extractFlag(args, "-content"),
		}
		data, err := client.Request("POST", basePath, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project note created:")
		printResponseData(data, config)
	case "update":
		noteID := extractFlag(args, "-note_id")
		if noteID == "" && len(args) > 1 {
			noteID = args[1]
		}
		if noteID == "" {
			fmt.Fprintln(os.Stderr, "Error: note ID required")
			os.Exit(1)
		}
		body := map[string]interface{}{}
		if title := extractFlag(args, "-title"); title != "" {
			body["title"] = title
		}
		if content := extractFlag(args, "-content"); content != "" {
			body["content"] = content
		}
		if len(body) == 0 {
			fmt.Fprintln(os.Stderr, "Error: at least one field to update required")
			os.Exit(1)
		}
		data, err := client.Request("PUT", basePath+"/"+noteID, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project note updated:")
		printResponseData(data, config)
	case "delete", "rm":
		noteID := extractFlag(args, "-note_id")
		if noteID == "" && len(args) > 1 {
			noteID = args[1]
		}
		if noteID == "" {
			fmt.Fprintln(os.Stderr, "Error: note ID required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", basePath+"/"+noteID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Project note '%s' deleted\n", noteID)
	default:
		fmt.Fprintf(os.Stderr, "Unknown project note subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleTaskCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		printTaskHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
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
			printResponseData(data, config)
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
		printResponseData(data, config)
	case "create":
		title := extractFlag(args, "-title")
		if title == "" && len(args) > 1 {
			title = args[1]
		}
		if title == "" {
			fmt.Fprintln(os.Stderr, "Error: -title flag required (or provide positional title)")
			os.Exit(1)
		}
		body := map[string]interface{}{"title": title}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		} else {
			body["description"] = ""
		}
		if projectID := extractFlag(args, "-project_id"); projectID != "" {
			body["project_id"] = projectID
		} else if config.ProjectID != "" {
			body["project_id"] = config.ProjectID
		} else {
			body["project_id"] = "default"
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
		printResponseData(data, config)
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
		printResponseData(data, config)
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
		printResponseData(data, config)
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
		printResponseData(data, config)
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
			printResponseData(data, config)
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
		printResponseData(data, config)
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
		printResponseData(data, config)
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
			printResponseData(data, config)
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

func resolveCommentEntity(args []string, config *cli.Config) (string, string) {
	if projectID := extractFlag(args, "-project_id"); projectID != "" {
		return "project", projectID
	}
	if taskID := extractFlag(args, "-task_id"); taskID != "" {
		return "task", taskID
	}
	entityType := extractFlag(args, "-entity_type")
	entityID := extractFlag(args, "-entity_id")
	if entityType == "project" && entityID == "" {
		if strings.TrimSpace(config.ProjectID) != "" {
			entityID = config.ProjectID
		} else {
			entityID = "default"
		}
	}
	return entityType, entityID
}

func handleCommentCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		printCommentHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		entityType, entityID := resolveCommentEntity(args, config)
		if entityType == "" || entityID == "" {
			fmt.Fprintln(os.Stderr, "Error: target required (-project_id or -task_id)")
			os.Exit(1)
		}
		path := "/api/v1/comments?entity_type=" + entityType + "&entity_id=" + entityID
		if hasFlag(args, "-include_deleted") {
			path += "&include_deleted=true"
		}
		data, err := client.Request("GET", path, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
			return
		}
		var comments []map[string]interface{}
		if err := parseJSON(data, &comments); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d comments on %s %s:\n\n", len(comments), entityType, entityID)
		for _, c := range comments {
			fmt.Printf("ID:      %v\n", c["id"])
			fmt.Printf("Owner:   %v\n", c["owner_username"])
			fmt.Printf("Deleted: %v\n", c["is_deleted"])
			fmt.Printf("Text:    %v\n", c["text"])
			fmt.Println()
		}
	case "get":
		commentID := extractFlag(args, "-comment_id")
		if commentID == "" && len(args) > 1 {
			commentID = args[1]
		}
		if commentID == "" {
			fmt.Fprintln(os.Stderr, "Error: comment ID required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/comments/"+commentID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
	case "create":
		entityType, entityID := resolveCommentEntity(args, config)
		if entityType == "" || entityID == "" {
			fmt.Fprintln(os.Stderr, "Error: target required (-project_id or -task_id)")
			os.Exit(1)
		}
		text := extractFlag(args, "-text")
		if text == "" {
			fmt.Fprintln(os.Stderr, "Error: -text flag required")
			os.Exit(1)
		}
		body := map[string]string{
			"entity_type": entityType,
			"entity_id":   entityID,
			"text":        text,
		}
		data, err := client.Request("POST", "/api/v1/comments", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Comment created:")
		printResponseData(data, config)
	case "update":
		commentID := extractFlag(args, "-comment_id")
		if commentID == "" && len(args) > 1 {
			commentID = args[1]
		}
		if commentID == "" {
			fmt.Fprintln(os.Stderr, "Error: comment ID required")
			os.Exit(1)
		}
		text := extractFlag(args, "-text")
		if text == "" {
			fmt.Fprintln(os.Stderr, "Error: -text flag required")
			os.Exit(1)
		}
		body := map[string]string{"text": text}
		data, err := client.Request("PUT", "/api/v1/comments/"+commentID, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Comment updated:")
		printResponseData(data, config)
	case "delete", "rm":
		commentID := extractFlag(args, "-comment_id")
		if commentID == "" && len(args) > 1 {
			commentID = args[1]
		}
		if commentID == "" {
			fmt.Fprintln(os.Stderr, "Error: comment ID required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", "/api/v1/comments/"+commentID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Comment soft-deleted")
	case "history":
		commentID := extractFlag(args, "-comment_id")
		if commentID == "" && len(args) > 1 {
			commentID = args[1]
		}
		if commentID == "" {
			fmt.Fprintln(os.Stderr, "Error: comment ID required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/comments/"+commentID+"/history", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
	default:
		fmt.Fprintf(os.Stderr, "Unknown comment subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleUserCommand(client *cli.Client, config *cli.Config, args []string) {
	if len(args) == 0 {
		printUserHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		data, err := client.Request("GET", "/api/v1/users", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
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
	case "get":
		userID := extractFlag(args, "-user_id")
		if userID == "" {
			userID = extractFlag(args, "-username")
		}
		if userID == "" && len(args) > 1 {
			userID = args[1]
		}
		if userID == "" {
			fmt.Fprintln(os.Stderr, "Error: -user_id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/users/"+userID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
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
		printResponseData(data, config)
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
	case "delete", "rm":
		// There is no hard-delete user endpoint; treat delete/rm as soft-delete (disable).
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
		fmt.Printf("User '%s' soft-deleted (disabled)\n", username)
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
		body := map[string]string{"new_password": newPassword}
		_, err := client.Request("POST", "/api/v1/users/"+username+"/reset-password", body)
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
		printWorkerHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		data, err := client.Request("GET", "/api/v1/workers", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
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
			printResponseData(data, config)
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
		printRoleHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		data, err := client.Request("GET", "/api/v1/roles", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
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
		id := extractFlag(args, "-role_id")
		if id == "" {
			fmt.Fprintln(os.Stderr, "Error: -role_id flag required")
			os.Exit(1)
		}
		data, err := client.Request("GET", "/api/v1/roles/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printResponseData(data, config)
	case "create":
		title := extractFlag(args, "-title")
		if title == "" {
			title = extractFlag(args, "-name")
		}
		if title == "" {
			fmt.Fprintln(os.Stderr, "Error: -title flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{
			"name":  title,
			"scope": "global",
		}
		if desc := extractFlag(args, "-description"); desc != "" {
			body["description"] = desc
		} else {
			body["description"] = title
		}
		if goals := extractFlag(args, "-goals"); goals != "" {
			body["goals"] = goals
		} else {
			body["goals"] = "{}"
		}
		if scope := extractFlag(args, "-scope"); scope != "" {
			body["scope"] = scope
		}
		if projectID := extractFlag(args, "-project_id"); projectID != "" {
			body["project_id"] = projectID
		}
		data, err := client.Request("POST", "/api/v1/roles", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Role created:")
		printResponseData(data, config)
	case "update":
		roleID := extractFlag(args, "-role_id")
		if roleID == "" {
			fmt.Fprintln(os.Stderr, "Error: -role_id flag required")
			os.Exit(1)
		}
		body := map[string]interface{}{}
		if title := extractFlag(args, "-title"); title != "" {
			body["name"] = title
		}
		if name := extractFlag(args, "-name"); name != "" {
			body["name"] = name
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
		printResponseData(data, config)
	case "delete", "rm":
		roleID := extractFlag(args, "-role_id")
		if roleID == "" && len(args) > 1 {
			roleID = args[1]
		}
		if roleID == "" {
			fmt.Fprintln(os.Stderr, "Error: -role_id flag required")
			os.Exit(1)
		}
		_, err := client.Request("DELETE", "/api/v1/roles/"+roleID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Role deleted")
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
			printResponseData(data, config)
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
		printConfigHelp()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		data, err := client.Request("GET", "/api/v1/config", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if config.JSON {
			printResponseData(data, config)
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
		printResponseData(data, config)
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
		printResponseData(data, config)
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

func slugifyIdentifier(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	nonAlphaNum := regexp.MustCompile(`[^a-z0-9]+`)
	slug := nonAlphaNum.ReplaceAllString(lower, "_")
	slug = strings.Trim(slug, "_")
	if slug == "" {
		return "project"
	}
	return slug
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func printResponseData(data []byte, config *cli.Config) {
	if config.JSON {
		printPrettyJSON(data)
		return
	}
	printHumanJSON(data)
}

func printPrettyJSON(data []byte) {
	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		fmt.Println(string(data))
		return
	}
	pretty, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(string(pretty))
}

func printHumanJSON(data []byte) {
	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		fmt.Println(string(data))
		return
	}
	printHumanValue(decoded, 0)
}

func printHumanValue(v interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch typed := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(typed))
		for k := range typed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			val := typed[key]
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				fmt.Printf("%s%s:\n", prefix, key)
				printHumanValue(val, indent+1)
			default:
				fmt.Printf("%s%s: %v\n", prefix, key, val)
			}
		}
	case []interface{}:
		for _, item := range typed {
			switch item.(type) {
			case map[string]interface{}, []interface{}:
				fmt.Printf("%s-\n", prefix)
				printHumanValue(item, indent+1)
			default:
				fmt.Printf("%s- %v\n", prefix, item)
			}
		}
	default:
		fmt.Printf("%s%v\n", prefix, typed)
	}
}

func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		// Use crypto/rand for secure random selection
		randomByte := make([]byte, 1)
		_, err := cryptorand.Read(randomByte)
		if err != nil {
			// Fallback to less secure but still usable method
			password[i] = charset[i%len(charset)]
		} else {
			password[i] = charset[int(randomByte[0])%len(charset)]
		}
	}
	return string(password)
}

// promptMatrixPassword prompts for a password with matrix-style changing characters
func promptMatrixPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Read password without echo using terminal raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	var password []byte
	matrixChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	displayChars := make([]rune, 0)
	animationCounters := make([]int, 0) // Track animation cycles for each character

	// Animation ticker for matrix effect (100ms = slower by 50%)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Channel for keyboard input
	inputChan := make(chan byte, 1)
	done := make(chan bool)

	// Goroutine to read keyboard input
	go func() {
		buf := make([]byte, 1)
		for {
			select {
			case <-done:
				return
			default:
				n, err := os.Stdin.Read(buf)
				if err != nil || n == 0 {
					continue
				}
				inputChan <- buf[0]
			}
		}
	}()

	// Main loop
	for {
		select {
		case char := <-inputChan:
			if char == 13 || char == 10 { // Enter key
				close(done)
				// Settle all remaining characters to asterisks before exiting
				if len(displayChars) > 0 {
					for i := range displayChars {
						displayChars[i] = '*'
					}
					fmt.Print("\r" + prompt)
					for i := 0; i < len(displayChars); i++ {
						fmt.Print("*")
					}
				}
				fmt.Println() // Move to next line
				return string(password), nil
			} else if char == 127 || char == 8 { // Backspace
				if len(password) > 0 {
					password = password[:len(password)-1]
					displayChars = displayChars[:len(displayChars)-1]
					animationCounters = animationCounters[:len(animationCounters)-1]
					// Clear line and reprint
					fmt.Print("\r" + prompt)
					for i := 0; i < len(displayChars); i++ {
						if displayChars[i] == '*' {
							fmt.Print("*")
						} else {
							fmt.Print(string(displayChars[i]))
						}
					}
					fmt.Print(" \b") // Clear last char
				}
			} else if char >= 32 && char < 127 { // Printable characters
				password = append(password, char)
				displayChars = append(displayChars, rune(matrixChars[rand.Intn(len(matrixChars))]))
				animationCounters = append(animationCounters, 0)
				fmt.Print(string(displayChars[len(displayChars)-1]))
			}

		case <-ticker.C:
			// Animate characters that haven't settled yet
			if len(displayChars) > 0 {
				needsRedraw := false
				for i := range displayChars {
					if displayChars[i] != '*' {
						animationCounters[i]++
						// After 8 animation cycles (0.8 seconds), settle to *
						if animationCounters[i] >= 8 {
							displayChars[i] = '*'
							needsRedraw = true
						} else {
							// Still animating - change to random character
							displayChars[i] = rune(matrixChars[rand.Intn(len(matrixChars))])
							needsRedraw = true
						}
					}
				}
				// Redraw if any character changed
				if needsRedraw {
					fmt.Print("\r" + prompt)
					for i := 0; i < len(displayChars); i++ {
						if displayChars[i] == '*' {
							fmt.Print("*")
						} else {
							fmt.Print(string(displayChars[i]))
						}
					}
				}
			}
		}
	}
}

func handleStatusCommand(client *cli.Client, config *cli.Config) {
	data, err := client.Request("GET", "/api/v1/status", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError fetching status: %v\n", err)
		os.Exit(1)
	}

	// Pretty print the JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintf(os.Stderr, "\nError parsing status response: %v\n", err)
		os.Exit(1)
	}

	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError formatting status: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(prettyJSON))
	printCurrentDefaultProject(config)
}

func handleLoginCommand(config *cli.Config) {
	// Prompt for credentials if not provided
	if config.Username == "" {
		fmt.Print("username: ")
		fmt.Scanln(&config.Username)
	}
	if config.Password == "" {
		var err error
		config.Password, err = promptMatrixPassword("password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
			os.Exit(1)
		}
	}

	if config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "\nError: Username and password are required")
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
		fmt.Fprintf(os.Stderr, "\nLogin failed: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "\nFailed to parse login response: %v\n", err)
		os.Exit(1)
	}

	// Save session token
	if err := cli.SaveSessionToken(loginResp.Token, loginResp.RefreshToken, loginResp.ExpiresAt, config.ServerURL, config.Username); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to save session token: %v\n", err)
	}

	// Also save credentials as fallback
	if err := cli.SaveCredentials(config.ServerURL, config.Username, config.Password); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to save credentials: %v\n", err)
	}

	if config.JSON {
		printResponseData(data, config)
	} else {
		fmt.Printf("\nSession token saved to ~/.config/sf/session.json\n")
		fmt.Printf("Token expires at: %s\n", loginResp.ExpiresAt)
	}
}

func handleRegisterCommand(config *cli.Config) {
	// Prompt for credentials if not provided
	if config.Username == "" {
		fmt.Print("username: ")
		fmt.Scanln(&config.Username)
	}
	if config.Password == "" {
		var err error
		config.Password, err = promptMatrixPassword("password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
			os.Exit(1)
		}
	}

	if config.Username == "" || config.Password == "" {
		fmt.Fprintln(os.Stderr, "\nError: Username and password are required")
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
		fmt.Fprintf(os.Stderr, "\nRegistration failed: %v\n", err)
		os.Exit(1)
	}

	// Save credentials
	if err := cli.SaveCredentials(config.ServerURL, config.Username, config.Password); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to save credentials: %v\n", err)
	}

	if config.JSON {
		printResponseData(data, config)
	} else {
		fmt.Println("\nCredentials saved to ~/.config/sf/credentials.json")
		fmt.Println("You can now use 'sf' commands without providing credentials.")
	}
}

func handleLogoutCommand() {
	// Clear session token
	sessionErr := cli.ClearSessionToken()
	sessionDeleted := sessionErr == nil
	sessionNotExist := os.IsNotExist(sessionErr)

	// Clear saved credentials
	credErr := cli.ClearCredentials()
	credDeleted := credErr == nil
	credNotExist := os.IsNotExist(credErr)

	// Clear project context
	contextErr := cli.ClearProjectContext()
	contextDeleted := contextErr == nil
	contextNotExist := os.IsNotExist(contextErr)

	anythingDeleted := sessionDeleted || credDeleted || contextDeleted
	anythingExisted := !sessionNotExist || !credNotExist || !contextNotExist

	if anythingDeleted {
		fmt.Println("Logged out successfully")
		if sessionDeleted {
			fmt.Println("  - Session token cleared")
		}
		if credDeleted {
			fmt.Println("  - Credentials cleared")
		}
		if contextDeleted {
			fmt.Println("  - Project context cleared")
		}
	} else if !anythingExisted {
		fmt.Println("No active session found")
	} else {
		// Files exist but couldn't be deleted
		fmt.Println("Error clearing session data")
		if sessionErr != nil && !sessionNotExist {
			fmt.Printf("  - Session token: %v\n", sessionErr)
		}
		if credErr != nil && !credNotExist {
			fmt.Printf("  - Credentials: %v\n", credErr)
		}
		if contextErr != nil && !contextNotExist {
			fmt.Printf("  - Project context: %v\n", contextErr)
		}
	}
}
