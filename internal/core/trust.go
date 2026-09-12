package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var trustMu sync.RWMutex

// TrustRegistry stores the path-to-hash mapping of allowed environment files.
type TrustRegistry struct {
	Allowed map[string]string `json:"allowed"`
}

func getTrustFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "simplyenv")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "allowed.json"), nil
}

// ComputeConfigHash computes the SHA-256 checksum of the given configuration file.
func ComputeConfigHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// LoadTrustRegistry loads the trusted configs registry from disk.
func LoadTrustRegistry() (*TrustRegistry, error) {
	trustMu.RLock()
	defer trustMu.RUnlock()

	path, err := getTrustFilePath()
	if err != nil {
		return &TrustRegistry{Allowed: make(map[string]string)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TrustRegistry{Allowed: make(map[string]string)}, nil
		}
		return nil, err
	}

	var reg TrustRegistry
	if err := json.Unmarshal(data, &reg); err != nil || reg.Allowed == nil {
		return &TrustRegistry{Allowed: make(map[string]string)}, nil
	}

	return &reg, nil
}

// Save writes the trust registry back to disk.
func (r *TrustRegistry) Save() error {
	trustMu.Lock()
	defer trustMu.Unlock()

	path, err := getTrustFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsConfigAllowed checks if the configuration file is allowed and its content hash matches.
// If SIMPLYENV_AUTO_ALLOW is set to 1 or true, it bypasses the check.
func IsConfigAllowed(configPath string) (bool, error) {
	if configPath == "" {
		return true, nil
	}

	// Bypass if explicitly configured via environment variable
	autoAllow := strings.ToLower(os.Getenv("SIMPLYENV_AUTO_ALLOW"))
	if autoAllow == "1" || autoAllow == "true" {
		return true, nil
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return false, err
	}

	currentHash, err := ComputeConfigHash(absPath)
	if err != nil {
		return false, err
	}

	reg, err := LoadTrustRegistry()
	if err != nil {
		return false, err
	}

	allowedHash, exists := reg.Allowed[absPath]
	if !exists {
		return false, nil
	}

	return allowedHash == currentHash, nil
}

// AllowConfig marks the given configuration file as trusted with its current content hash.
func AllowConfig(configPath string) (string, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", err
	}

	hash, err := ComputeConfigHash(absPath)
	if err != nil {
		return "", err
	}

	reg, err := LoadTrustRegistry()
	if err != nil {
		return "", err
	}

	reg.Allowed[absPath] = hash
	if err := reg.Save(); err != nil {
		return "", err
	}

	return hash, nil
}

// DenyConfig revokes trust for the specified configuration file.
func DenyConfig(configPath string) error {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}

	reg, err := LoadTrustRegistry()
	if err != nil {
		return err
	}

	delete(reg.Allowed, absPath)
	return reg.Save()
}
