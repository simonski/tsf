package main

import (
"fmt"
"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Task CLI - use 'task help' for usage")
		os.Exit(0)
	}
	fmt.Printf("Command: %s (implementation in progress)\n", os.Args[1])
}
