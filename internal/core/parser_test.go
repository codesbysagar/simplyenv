package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFileAndWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, ".simplyenv")

	initialContent := `# Project environment
API_KEY="secret_val_123"
DATABASE_URL=postgres://localhost:5432/test
export DEBUG="true"
# Comment line
EMPTY_LINE=
`
	if err := os.WriteFile(filePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	vars, err := ParseEnvFile(filePath)
	if err != nil {
		t.Fatalf("ParseEnvFile failed: %v", err)
	}

	if vars["API_KEY"] != "secret_val_123" {
		t.Errorf("Expected API_KEY to be 'secret_val_123', got '%s'", vars["API_KEY"])
	}
	if vars["DATABASE_URL"] != "postgres://localhost:5432/test" {
		t.Errorf("Expected DATABASE_URL to be 'postgres://localhost:5432/test', got '%s'", vars["DATABASE_URL"])
	}
	if vars["DEBUG"] != "true" {
		t.Errorf("Expected DEBUG to be 'true', got '%s'", vars["DEBUG"])
	}

	// Test SetEnvVar
	if err := SetEnvVar(filePath, "NEW_VAR", "hello_world"); err != nil {
		t.Fatalf("SetEnvVar failed: %v", err)
	}

	updated, err := ParseEnvFile(filePath)
	if err != nil {
		t.Fatalf("ParseEnvFile after set failed: %v", err)
	}
	if updated["NEW_VAR"] != "hello_world" {
		t.Errorf("Expected NEW_VAR to be 'hello_world', got '%s'", updated["NEW_VAR"])
	}

	// Test DeleteEnvVar
	if err := DeleteEnvVar(filePath, "API_KEY"); err != nil {
		t.Fatalf("DeleteEnvVar failed: %v", err)
	}

	afterDelete, err := ParseEnvFile(filePath)
	if err != nil {
		t.Fatalf("ParseEnvFile after delete failed: %v", err)
	}
	if _, exists := afterDelete["API_KEY"]; exists {
		t.Errorf("Expected API_KEY to be deleted")
	}
}

func TestIsValidKey(t *testing.T) {
	valid := []string{"FOO", "foo", "VAR_1", "_SECRET", "A1_B2_C3", "a"}
	for _, k := range valid {
		if !IsValidKey(k) {
			t.Errorf("Expected %q to be valid key", k)
		}
	}

	invalid := []string{
		"FOO; rm -rf /",
		"KEY$(whoami)",
		"KEY`id`",
		"FOO BAR",
		"FOO-BAR",
		"1START_WITH_NUMBER",
		"VAR.NAME",
		"KEY:VAL",
		"",
	}
	for _, k := range invalid {
		if IsValidKey(k) {
			t.Errorf("Expected %q to be invalid key", k)
		}
	}
}

func TestRejectMaliciousKeys(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, ".simplyenv")

	maliciousContent := `
VALID_KEY="safe_value"
MALICIOUS; rm -rf /="bad"
KEY$(id)="bad"
INVALID-NAME="bad"
`
	if err := os.WriteFile(filePath, []byte(maliciousContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	vars, err := ParseEnvFile(filePath)
	if err != nil {
		t.Fatalf("ParseEnvFile failed: %v", err)
	}

	if len(vars) != 1 {
		t.Errorf("Expected exactly 1 valid variable, got %d: %v", len(vars), vars)
	}
	if vars["VALID_KEY"] != "safe_value" {
		t.Errorf("Expected VALID_KEY to be safe_value, got %q", vars["VALID_KEY"])
	}

	// Setting invalid key should return an error
	if err := SetEnvVar(filePath, "BAD; KEY", "val"); err == nil {
		t.Errorf("Expected error setting invalid key, got nil")
	}
}
