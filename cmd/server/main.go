package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/simonski/task/internal/db"
	"github.com/simonski/task/internal/server"
	"github.com/simonski/task/internal/web"
)

func main() {
	// Parse command line flags
	dbPath := flag.String("f", "", "Path to database file (default: ~/.config/task/task.db)")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	// Determine database path
	finalPath := *dbPath
	if finalPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		finalPath = filepath.Join(home, ".config", "task", "task.db")
	}

	// Check if database exists
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		log.Fatalf("Database does not exist at %s. Please run 'task initdb' first.", finalPath)
	}

	// Open database
	database, err := db.Open(finalPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create server
	srv := server.New(database)
	srv.SetWebFS(web.FS)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Database: %s", finalPath)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
