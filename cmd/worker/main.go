package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simonski/task/internal/worker"
)

func main() {
	serverURL := flag.String("url", "http://localhost:8080", "Server URL")
	username := flag.String("username", os.Getenv("TASK_USERNAME"), "Username")
	password := flag.String("password", os.Getenv("TASK_PASSWORD"), "Password")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: TASK_USERNAME and TASK_PASSWORD required")
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
