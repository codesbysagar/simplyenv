package shell

import (
	"fmt"
	"strings"
)

// ... FormatForShell function remains the same ...
func FormatForShell(envVars map[string]string) string {
	var builder strings.Builder
	for key, value := range envVars {
		builder.WriteString(fmt.Sprintf("export %s=\"%s\";\n", key, value))
	}
	return builder.String()
}

// FormatUnset creates unset commands for the given keys.
func FormatUnset(keys map[string]bool) string {
	var builder strings.Builder
	for key := range keys {
		builder.WriteString(fmt.Sprintf("unset %s;\n", key))
	}
	return builder.String()
}

// FormatState creates the command to update the SIMPLYENV_STATE variable.
func FormatState(stateStr string) string {
	if stateStr == "" {
		return "unset SIMPLYENV_STATE;\n"
	}
	return fmt.Sprintf("export SIMPLYENV_STATE=\"%s\";\n", stateStr)
}
