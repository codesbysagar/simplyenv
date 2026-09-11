package shell

import (
	"fmt"
	"strings"
)

// FormatForShell generates shell export statements.
func FormatForShell(envVars map[string]string) string {
	var builder strings.Builder
	for key, value := range envVars {
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "$", "\\$")
		escaped = strings.ReplaceAll(escaped, "`", "\\`")
		builder.WriteString(fmt.Sprintf("export %s=\"%s\";\n", key, escaped))
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

// GenerateHook returns the shell hook snippet for the specified shell.
func GenerateHook(shellType string) string {
	switch strings.ToLower(shellType) {
	case "zsh":
		return `# simplyenv zsh hook
_simplyenv_hook() {
  trap -- '' SIGINT;
  eval "$(simplyenv eval 2>/dev/null)";
  trap - SIGINT;
}
autoload -Uz add-zsh-hook
add-zsh-hook chpwd _simplyenv_hook
add-zsh-hook precmd _simplyenv_hook
`
	case "fish":
		return `# simplyenv fish hook
function __simplyenv_export_eval --on-event fish_prompt
  simplyenv eval 2>/dev/null | source
end
`
	case "pwsh", "powershell":
		return `# simplyenv powershell hook
function Invoke-Simplyenv {
  $out = simplyenv eval 2>$null
  if ($out) {
    Invoke-Expression $out
  }
}
$oldPrompt = $function:prompt
function prompt {
  Invoke-Simplyenv
  & $oldPrompt
}
`
	case "bash":
		fallthrough
	default:
		return `# simplyenv bash hook
_simplyenv_hook() {
  local previous_exit_status=$?;
  trap -- '' SIGINT;
  eval "$(simplyenv eval 2>/dev/null)";
  trap - SIGINT;
  return $previous_exit_status;
};
if ! [[ "$PROMPT_COMMAND" =~ _simplyenv_hook ]]; then
  PROMPT_COMMAND="_simplyenv_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
`
	}
}
