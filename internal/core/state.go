package core

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// State represents the previously loaded environment.
type State struct {
	PathHash string
	VarKeys  []string
}

// EncodeState creates a string representation of the current state.
// Format: pathHash:key1,key2,key3
func EncodeState(configPath string, envVars map[string]string) string {
	hasher := sha256.New()
	hasher.Write([]byte(configPath))
	pathHash := hex.EncodeToString(hasher.Sum(nil))

	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort for a consistent order

	return pathHash + ":" + strings.Join(keys, ",")
}

// DecodeState parses the state string from the environment variable.
func DecodeState(stateStr string) State {
	if stateStr == "" {
		return State{}
	}
	parts := strings.SplitN(stateStr, ":", 2)
	if len(parts) != 2 {
		return State{}
	}
	return State{
		PathHash: parts[0],
		VarKeys:  strings.Split(parts[1], ","),
	}
}
