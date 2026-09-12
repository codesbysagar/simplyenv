package core

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

// VarRestore captures the previous state of a variable before simplyenv set it.
type VarRestore struct {
	HadOriginal bool   `json:"h"`
	OrigValue   string `json:"v,omitempty"`
}

// State represents the snapshot of environment variables loaded by simplyenv.
type State struct {
	ConfigPath string                `json:"p"`
	Restores   map[string]VarRestore `json:"r"`
}

// VarKeys returns a sorted list of variable keys managed in this state.
func (s *State) VarKeys() []string {
	if s == nil || len(s.Restores) == 0 {
		return nil
	}
	keys := make([]string, 0, len(s.Restores))
	for k := range s.Restores {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// BuildNewState builds the state for newly exported variables.
// If a variable was already tracked in prevState, its original value is preserved.
// If it was not tracked in prevState, os.LookupEnv checks if the user had an ambient value.
func BuildNewState(configPath string, newVars map[string]string, prevState State) State {
	restores := make(map[string]VarRestore, len(newVars))

	for key := range newVars {
		if !IsValidKey(key) {
			continue
		}

		if prevRestore, exists := prevState.Restores[key]; exists {
			// Variable was already being tracked; preserve its initial shell value
			restores[key] = prevRestore
		} else {
			// Variable is newly introduced by simplyenv; capture shell ambient state
			val, hadOriginal := os.LookupEnv(key)
			restores[key] = VarRestore{
				HadOriginal: hadOriginal,
				OrigValue:   val,
			}
		}
	}

	return State{
		ConfigPath: configPath,
		Restores:   restores,
	}
}

// EncodeState serializes the State to a URL-safe base64 JSON string.
func EncodeState(state State) string {
	if len(state.Restores) == 0 {
		return ""
	}
	data, err := json.Marshal(state)
	if err != nil {
		return ""
	}
	return "v2:" + base64.RawURLEncoding.EncodeToString(data)
}

// DecodeState parses a state string (supporting both legacy "pathHash:k1,k2" and "v2:<base64>").
func DecodeState(stateStr string) State {
	if stateStr == "" {
		return State{Restores: make(map[string]VarRestore)}
	}

	// v2 Base64 format
	if strings.HasPrefix(stateStr, "v2:") {
		encoded := strings.TrimPrefix(stateStr, "v2:")
		data, err := base64.RawURLEncoding.DecodeString(encoded)
		if err == nil {
			var st State
			if err := json.Unmarshal(data, &st); err == nil {
				if st.Restores == nil {
					st.Restores = make(map[string]VarRestore)
				}
				return st
			}
		}
	}

	// Fallback / legacy format: pathHash:key1,key2,key3
	parts := strings.SplitN(stateStr, ":", 2)
	restores := make(map[string]VarRestore)
	if len(parts) == 2 {
		keys := strings.Split(parts[1], ",")
		for _, k := range keys {
			k = strings.TrimSpace(k)
			if k != "" {
				restores[k] = VarRestore{HadOriginal: false}
			}
		}
	}

	return State{
		ConfigPath: "",
		Restores:   restores,
	}
}
