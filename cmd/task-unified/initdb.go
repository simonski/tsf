package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/simonski/task/internal/db"
)

func initDBMain() {
	dbPath := flag.String("f", "", "Path to database file (default: ~/.config/task/task.db)")
	force := flag.Bool("force", false, "Force rebuild database (removes existing database)")
	flag.Parse()

	finalPath := *dbPath
	if finalPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		finalPath = filepath.Join(home, ".config", "task", "task.db")
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

	adminPassword, orchestratorPassword, err := database.InitializeDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
		os.Remove(finalPath)
		os.Exit(1)
	}

	fmt.Println("Database initialized successfully!")
	fmt.Printf("Location: %s\n\n", finalPath)
	fmt.Println("Admin credentials:")
	fmt.Println("  Username: admin")
	fmt.Printf("  Password: %s\n\n", adminPassword)
	fmt.Println("Orchestrator credentials:")
	fmt.Println("  Username: orchestrator")
	fmt.Printf("  Password: %s\n\n", orchestratorPassword)
	fmt.Println("IMPORTANT: Save these credentials securely. They cannot be recovered.")
}
