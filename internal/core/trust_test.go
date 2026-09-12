package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrustFlow(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	testDir := t.TempDir()
	envPath := filepath.Join(testDir, ".simplyenv")

	// 1. Initially create a file externally (e.g. git clone)
	if err := os.WriteFile(envPath, []byte("API_KEY=initial_value\n"), 0644); err != nil {
		t.Fatalf("Failed to write initial env file: %v", err)
	}

	// Should not be allowed yet
	allowed, err := IsConfigAllowed(envPath)
	if err != nil {
		t.Fatalf("IsConfigAllowed failed: %v", err)
	}
	if allowed {
		t.Errorf("Expected newly cloned config to be untrusted by default")
	}

	// 2. Allow config
	hash, err := AllowConfig(envPath)
	if err != nil {
		t.Fatalf("AllowConfig failed: %v", err)
	}
	if hash == "" {
		t.Errorf("Expected non-empty hash from AllowConfig")
	}

	// Now it should be allowed
	allowed, err = IsConfigAllowed(envPath)
	if err != nil {
		t.Fatalf("IsConfigAllowed failed: %v", err)
	}
	if !allowed {
		t.Errorf("Expected config to be allowed after AllowConfig")
	}

	// 3. Simulate file modification (e.g. untrusted git pull with altered secrets/code)
	if err := os.WriteFile(envPath, []byte("API_KEY=modified_untrusted_value\n"), 0644); err != nil {
		t.Fatalf("Failed to write modified env file: %v", err)
	}

	// Should now be blocked due to hash mismatch
	allowed, err = IsConfigAllowed(envPath)
	if err != nil {
		t.Fatalf("IsConfigAllowed failed: %v", err)
	}
	if allowed {
		t.Errorf("Expected modified config to be blocked due to content hash mismatch")
	}

	// 4. Re-allow after user reviews changes
	_, err = AllowConfig(envPath)
	if err != nil {
		t.Fatalf("AllowConfig failed: %v", err)
	}
	allowed, _ = IsConfigAllowed(envPath)
	if !allowed {
		t.Errorf("Expected config to be allowed after re-allowing")
	}

	// 5. Deny config
	if err := DenyConfig(envPath); err != nil {
		t.Fatalf("DenyConfig failed: %v", err)
	}
	allowed, _ = IsConfigAllowed(envPath)
	if allowed {
		t.Errorf("Expected config to be blocked after DenyConfig")
	}
}

func TestWriteEnvFileAutoAllows(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	testDir := t.TempDir()
	envPath := filepath.Join(testDir, ".simplyenv")

	// Local user edits via WriteEnvFile
	err := WriteEnvFile(envPath, map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/test",
	})
	if err != nil {
		t.Fatalf("WriteEnvFile failed: %v", err)
	}

	// Should be auto-allowed because it was an intentional local user action
	allowed, err := IsConfigAllowed(envPath)
	if err != nil {
		t.Fatalf("IsConfigAllowed failed: %v", err)
	}
	if !allowed {
		t.Errorf("Expected WriteEnvFile to automatically allow locally authored configs")
	}
}

func TestAutoAllowEnvBypass(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	testDir := t.TempDir()
	envPath := filepath.Join(testDir, ".simplyenv")
	_ = os.WriteFile(envPath, []byte("KEY=val\n"), 0644)

	t.Setenv("SIMPLYENV_AUTO_ALLOW", "1")
	allowed, err := IsConfigAllowed(envPath)
	if err != nil || !allowed {
		t.Errorf("Expected SIMPLYENV_AUTO_ALLOW=1 to bypass trust check")
	}
}
