package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectUserShell determines the active shell from $SHELL or OS default.
func DetectUserShell() string {
	shellEnv := os.Getenv("SHELL")
	if shellEnv != "" {
		base := strings.ToLower(filepath.Base(shellEnv))
		switch {
		case strings.Contains(base, "zsh"):
			return "zsh"
		case strings.Contains(base, "fish"):
			return "fish"
		case strings.Contains(base, "bash"):
			return "bash"
		}
	}

	if runtime.GOOS == "windows" {
		return "pwsh"
	}

	// Default fallback on macOS is zsh, Linux is bash
	if runtime.GOOS == "darwin" {
		return "zsh"
	}
	return "bash"
}

// GetProfilePath returns the target shell config file path for the given shell.
func GetProfilePath(shellType string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch strings.ToLower(shellType) {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	case "pwsh", "powershell":
		if runtime.GOOS == "windows" {
			docs := filepath.Join(home, "Documents", "PowerShell")
			return filepath.Join(docs, "Microsoft.PowerShell_profile.ps1"), nil
		}
		return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"), nil
	case "bash":
		fallthrough
	default:
		// On macOS, bash often uses .bash_profile; on Linux .bashrc
		if runtime.GOOS == "darwin" {
			bashProfile := filepath.Join(home, ".bash_profile")
			if _, err := os.Stat(bashProfile); err == nil {
				return bashProfile, nil
			}
		}
		return filepath.Join(home, ".bashrc"), nil
	}
}

// GetHookCommand returns the single-line eval hook for the profile.
func GetHookCommand(shellType string) string {
	switch strings.ToLower(shellType) {
	case "zsh":
		return `eval "$(simplyenv hook zsh)"`
	case "fish":
		return `simplyenv hook fish | source`
	case "pwsh", "powershell":
		return `Invoke-Expression (simplyenv hook pwsh)`
	case "bash":
		fallthrough
	default:
		return `eval "$(simplyenv hook bash)"`
	}
}

// IsHookInstalled checks if the hook line already exists in the profile.
func IsHookInstalled(shellType string) (bool, string, error) {
	profilePath, err := GetProfilePath(shellType)
	if err != nil {
		return false, "", err
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, profilePath, nil
		}
		return false, profilePath, err
	}

	content := string(data)
	return strings.Contains(content, "simplyenv hook"), profilePath, nil
}

// InstallHook safely appends the hook into the user's shell configuration file.
func InstallHook(shellType string) (string, error) {
	installed, profilePath, err := IsHookInstalled(shellType)
	if err != nil {
		return "", err
	}
	if installed {
		return profilePath, fmt.Errorf("hook is already installed in %s", profilePath)
	}

	// Ensure parent directory exists (e.g. for fish ~/.config/fish)
	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return "", err
	}

	hookCmd := GetHookCommand(shellType)
	entry := fmt.Sprintf("\n# simplyenv shell integration\n%s\n", hookCmd)

	file, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		return "", err
	}

	return profilePath, nil
}

// UninstallHook safely removes the simplyenv hook lines from the shell profile.
func UninstallHook(shellType string) (string, error) {
	profilePath, err := GetProfilePath(shellType)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		return profilePath, err
	}

	lines := strings.Split(string(data), "\n")
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "simplyenv hook") || strings.Contains(line, "# simplyenv shell integration") {
			continue
		}
		newLines = append(newLines, line)
	}

	cleaned := strings.Join(newLines, "\n")
	if err := os.WriteFile(profilePath, []byte(cleaned), 0644); err != nil {
		return profilePath, err
	}

	return profilePath, nil
}
