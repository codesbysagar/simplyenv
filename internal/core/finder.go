package core

import (
	"errors"
	"os"
	"path/filepath"
)

// A prioritized list of config filenames to search for.
var configFileNames = []string{".simplyenv", ".envrc"}

// FindConfig searches for a config file, starting from startDir and moving upwards.
// It returns the full path to the file if found, or an error if not.
func FindConfig(startDir string) (string, error) {
	dir := startDir
	for {
		// Check for each config file name in our prioritized list
		for _, fileName := range configFileNames {
			configPath := filepath.Join(dir, fileName)
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil // Found the highest priority file
			}
		}

		// Move up to the parent directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// We've hit the root directory without finding any config file
			return "", errors.New("config file not found")
		}
		dir = parentDir
	}
}
