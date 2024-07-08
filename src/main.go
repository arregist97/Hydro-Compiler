package main

import (
	"fmt"
	"os"
)

func main() {
	// Check if a command line argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command line argument.")
		return
	}

	// Get the command line argument
	arg := os.Args[1]

	// Print the argument
	fmt.Println("Command line argument:", arg)
}
