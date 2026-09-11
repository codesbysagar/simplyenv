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
