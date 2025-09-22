package main

import (
	"fmt"
	"os"
	"simplyenv/internal/core"
	"simplyenv/internal/shell"
)

func main() {
	// 1. Get previous state from environment variable
	prevState := core.DecodeState(os.Getenv("SIMPLYENV_STATE"))

	// 2. Find the config file for the current directory
	wd, _ := os.Getwd()
	configPath, err := core.FindConfig(wd)

	// --- Calculate Differences ---

	// Variables to unset are all the keys from the previous state by default
	varsToUnset := make(map[string]bool)
	for _, key := range prevState.VarKeys {
		if key != "" {
			varsToUnset[key] = true
		}
	}

	var newEnvVars map[string]string
	var newStateStr string

	if err == nil { // A config file was found
		newEnvVars, _ = core.ParseEnvFile(configPath)
		newStateStr = core.EncodeState(configPath, newEnvVars)

		// Don't unset variables that are present in the new environment
		for key := range newEnvVars {
			delete(varsToUnset, key)
		}
	}

	// 3. Generate the shell commands
	output := shell.FormatUnset(varsToUnset) // We'll create this function next
	output += shell.FormatForShell(newEnvVars)
	output += shell.FormatState(newStateStr) // And this one too

	fmt.Print(output)
}
