package core

import (
	"errors"
	"os"
	"path/filepath"
)

const ConfigFileName = ".simplyenv"

// FindConfig searches for the .simplyenv file, starting from startDir and moving upwards.
// It returns the full path to the file if found, or an error if not.
func FindConfig(startDir string) (string, error) {
	dir := startDir
	for {
		// Construct the full path for the config file in the current directory
		configPath := filepath.Join(dir, ConfigFileName)

		// Check if the file exists
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil // Found it!
		}

		// Move up to the parent directory
		parentDir := filepath.Dir(dir)

		// If the parent directory is the same as the current one, we've hit the root
		if parentDir == dir {
			return "", errors.New("config file not found")
		}
		dir = parentDir
	}
}
