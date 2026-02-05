package main

import (
	"fmt"
	"os"
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
	case "project", "task", "user", "role", "config":
		runCLI(os.Args[1:])
	case "version", "-v", "--version":
		fmt.Printf("task version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("tsf")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage: task <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  server        Start the HTTP server")
	fmt.Println("  orchestrator  Start the orchestrator daemon")
	fmt.Println("  worker        Start a worker daemon")
	fmt.Println("  initdb        Initialize the database")
	fmt.Println("  tui           Start the Terminal User Interface")
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
	fmt.Println("Use 'task <command> -h' for more information about a command.")
}

func runServer(args []string) {
	os.Args = append([]string{"task"}, args...)
	serverMain()
}

func runOrchestrator(args []string) {
	os.Args = append([]string{"task"}, args...)
	orchestratorMain()
}

func runWorker(args []string) {
	os.Args = append([]string{"task"}, args...)
	workerMain()
}

func runInitDB(args []string) {
	os.Args = append([]string{"task"}, args...)
	initDBMain()
}

func runCLI(args []string) {
	os.Args = append([]string{"task"}, args...)
	cliMain()
}

func runTUI(args []string) {
	os.Args = append([]string{"task"}, args...)
	tuiMain()
}
