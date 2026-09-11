package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ParseRawImport parses text in either JSON or .env/key-value format.
func ParseRawImport(content string) (map[string]string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return make(map[string]string), nil
	}

	// 1. Try parsing as JSON first
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &jsonMap); err == nil {
			result := make(map[string]string)
			for k, v := range jsonMap {
				result[k] = fmt.Sprintf("%v", v)
			}
			return result, nil
		}
	}

	// 2. Parse as standard .env / key-value lines
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Support "export KEY=VALUE"
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])

		// Strip surrounding quotes if matched
		if (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) ||
			(strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'")) {
			if len(v) >= 2 {
				v = v[1 : len(v)-1]
			}
		}

		if k != "" {
			result[k] = v
		}
	}

	return result, scanner.Err()
}

// ExportFormat represents supported export formats.
type ExportFormat string

const (
	FormatEnv    ExportFormat = "env"
	FormatJSON   ExportFormat = "json"
	FormatShell  ExportFormat = "shell"
	FormatDocker ExportFormat = "docker"
)

// FormatExport converts a map of environment variables into the requested format string.
func FormatExport(envVars map[string]string, format ExportFormat) (string, error) {
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	switch format {
	case FormatJSON:
		data, err := json.MarshalIndent(envVars, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case FormatShell:
		var b strings.Builder
		for _, k := range keys {
			val := strings.ReplaceAll(envVars[k], "\"", "\\\"")
			b.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, val))
		}
		return b.String(), nil

	case FormatDocker:
		var b strings.Builder
		for _, k := range keys {
			val := strings.ReplaceAll(envVars[k], "\"", "\\\"")
			b.WriteString(fmt.Sprintf("ENV %s=\"%s\"\n", k, val))
		}
		return b.String(), nil

	case FormatEnv:
		fallthrough
	default:
		var b strings.Builder
		for _, k := range keys {
			val := strings.ReplaceAll(envVars[k], "\"", "\\\"")
			b.WriteString(fmt.Sprintf("%s=\"%s\"\n", k, val))
		}
		return b.String(), nil
	}
}
