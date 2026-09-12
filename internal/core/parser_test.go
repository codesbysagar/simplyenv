package core

import (
	"os"
	"path/filepath"
	"strings"
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

func TestStateEncodingAndRestoration(t *testing.T) {
	// Set ambient environment variable
	t.Setenv("PORT", "3000")

	// 1. Enter dir1: defines PORT=8080 and NEW_VAR=hello
	dir1Vars := map[string]string{
		"PORT":    "8080",
		"NEW_VAR": "hello",
	}
	state1 := BuildNewState("/dir1", dir1Vars, State{})

	if !state1.Restores["PORT"].HadOriginal || state1.Restores["PORT"].OrigValue != "3000" {
		t.Errorf("Expected PORT to have HadOriginal=true and OrigValue='3000', got %+v", state1.Restores["PORT"])
	}
	if state1.Restores["NEW_VAR"].HadOriginal {
		t.Errorf("Expected NEW_VAR to have HadOriginal=false, got %+v", state1.Restores["NEW_VAR"])
	}

	// 2. Encode and Decode
	encoded := EncodeState(state1)
	if !strings.HasPrefix(encoded, "v2:") {
		t.Errorf("Expected encoded state to start with v2:, got %s", encoded)
	}
	decoded := DecodeState(encoded)
	if decoded.ConfigPath != "/dir1" {
		t.Errorf("Expected decoded path /dir1, got %s", decoded.ConfigPath)
	}
	if decoded.Restores["PORT"].OrigValue != "3000" {
		t.Errorf("Expected decoded PORT restore '3000', got %s", decoded.Restores["PORT"].OrigValue)
	}

	// 3. Transition to dir2: defines PORT=9000 (preserves original PORT from state1)
	dir2Vars := map[string]string{
		"PORT": "9000",
	}
	state2 := BuildNewState("/dir2", dir2Vars, decoded)
	if state2.Restores["PORT"].OrigValue != "3000" {
		t.Errorf("Expected dir2 to preserve original PORT=3000, got %s", state2.Restores["PORT"].OrigValue)
	}

	// 4. Test legacy decode fallback
	legacyState := DecodeState("fakehash:KEY1,KEY2")
	if len(legacyState.Restores) != 2 {
		t.Errorf("Expected 2 restores from legacy state, got %d", len(legacyState.Restores))
	}
	if legacyState.Restores["KEY1"].HadOriginal {
		t.Errorf("Expected legacy key to default to HadOriginal=false")
	}
}
