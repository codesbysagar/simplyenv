package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetHookCommand(t *testing.T) {
	bashCmd := GetHookCommand("bash")
	if !strings.Contains(bashCmd, "simplyenv hook bash") {
		t.Errorf("Expected bash hook command, got: %s", bashCmd)
	}

	zshCmd := GetHookCommand("zsh")
	if !strings.Contains(zshCmd, "simplyenv hook zsh") {
		t.Errorf("Expected zsh hook command, got: %s", zshCmd)
	}

	fishCmd := GetHookCommand("fish")
	if !strings.Contains(fishCmd, "simplyenv hook fish") {
		t.Errorf("Expected fish hook command, got: %s", fishCmd)
	}
}

func TestInstallAndUninstallHook(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	profilePath := filepath.Join(tempHome, ".bashrc")
	initialContent := "# Existing bashrc\nexport PATH=/usr/bin:$PATH\n"
	if err := os.WriteFile(profilePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial profile: %v", err)
	}

	// 1. Check initially not installed
	installed, path, err := IsHookInstalled("bash")
	if err != nil {
		t.Fatalf("IsHookInstalled failed: %v", err)
	}
	if installed {
		t.Errorf("Expected hook to not be installed initially")
	}
	if path != profilePath {
		t.Errorf("Expected path %s, got %s", profilePath, path)
	}

	// 2. Install hook
	installedPath, err := InstallHook("bash")
	if err != nil {
		t.Fatalf("InstallHook failed: %v", err)
	}
	if installedPath != profilePath {
		t.Errorf("Expected installedPath %s, got %s", profilePath, installedPath)
	}

	// Verify installed
	installed, _, _ = IsHookInstalled("bash")
	if !installed {
		t.Errorf("Expected hook to be installed now")
	}

	// Verify idempotency (second install should return error/skip)
	_, err = InstallHook("bash")
	if err == nil {
		t.Errorf("Expected error when installing already installed hook")
	}

	// 3. Uninstall hook
	_, err = UninstallHook("bash")
	if err != nil {
		t.Fatalf("UninstallHook failed: %v", err)
	}

	installed, _, _ = IsHookInstalled("bash")
	if installed {
		t.Errorf("Expected hook to be uninstalled")
	}

	// Verify original content was preserved
	content, _ := os.ReadFile(profilePath)
	if !strings.Contains(string(content), "export PATH=/usr/bin:$PATH") {
		t.Errorf("Original bashrc content was corrupted: %s", string(content))
	}
}

func TestFormatUnsetProtectedVars(t *testing.T) {
	keysToUnset := map[string]bool{
		"PATH":          true,
		"HOME":          true,
		"USER":          true,
		"SHELL":         true,
		"CUSTOM_VAR":    true,
		"ANOTHER_VAR":   true,
		"MALICIOUS; rm": true,
	}

	result := FormatUnset(keysToUnset)

	// Protected vars must NEVER be unset
	if strings.Contains(result, "unset PATH") {
		t.Errorf("FormatUnset emitted 'unset PATH', which corrupts shell!")
	}
	if strings.Contains(result, "unset HOME") {
		t.Errorf("FormatUnset emitted 'unset HOME', which corrupts shell!")
	}
	if strings.Contains(result, "unset USER") {
		t.Errorf("FormatUnset emitted 'unset USER', which corrupts shell!")
	}
	if strings.Contains(result, "unset SHELL") {
		t.Errorf("FormatUnset emitted 'unset SHELL', which corrupts shell!")
	}
	if strings.Contains(result, "rm") {
		t.Errorf("FormatUnset emitted invalid key containing 'rm'!")
	}

	// Legitimate custom keys should be unset
	if !strings.Contains(result, "unset CUSTOM_VAR;") {
		t.Errorf("FormatUnset failed to unset CUSTOM_VAR: %s", result)
	}
	if !strings.Contains(result, "unset ANOTHER_VAR;") {
		t.Errorf("FormatUnset failed to unset ANOTHER_VAR: %s", result)
	}
}

func TestFormatForShellKeyValidation(t *testing.T) {
	envVars := map[string]string{
		"VALID_KEY":         "123",
		"INJECT; echo PWN":  "attack",
		"KEY$(whoami)":      "attack",
	}

	result := FormatForShell(envVars)

	if !strings.Contains(result, "export VALID_KEY=\"123\";") {
		t.Errorf("FormatForShell missed VALID_KEY: %s", result)
	}
	if strings.Contains(result, "PWN") || strings.Contains(result, "whoami") {
		t.Errorf("FormatForShell exported malicious keys: %s", result)
	}
}
