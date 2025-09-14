package main

import (
	"fmt"
	"log"
	"os"
	"simplyenv/internal/core" // Assumes module name is 'simplyenv'
)

func main() {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Search for the config file
	configPath, err := core.FindConfig(wd)
	if err != nil {
		fmt.Println("No .simplyenv file found in this or any parent directory.")
		return
	}

	fmt.Printf("Found config file at: %s\n", configPath)
}
