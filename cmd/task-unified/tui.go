package main

import (
	"fmt"
	"os"

	"github.com/simonski/task/internal/cli"
	"github.com/simonski/task/internal/tui"
)

func tuiMain() {
	// Parse command line args for config
	config, err := cli.NewConfig(os.Args[1:])
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
