package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverModules(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "simplyenv-modules-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create root config
	rootCfg := filepath.Join(tempDir, ".simplyenv")
	if err := WriteEnvFile(rootCfg, map[string]string{"ROOT_KEY": "root_val"}); err != nil {
		t.Fatalf("Failed to write root config: %v", err)
	}

	// Create microservice-A
	msADir := filepath.Join(tempDir, "microservice-A")
	if err := os.MkdirAll(msADir, 0755); err != nil {
		t.Fatalf("Failed to create msA dir: %v", err)
	}
	msACfg := filepath.Join(msADir, ".simplyenv")
	if err := WriteEnvFile(msACfg, map[string]string{"PORT": "3001", "SERVICE": "A"}); err != nil {
		t.Fatalf("Failed to write msA config: %v", err)
	}

	// Create microservice-B
	msBDir := filepath.Join(tempDir, "microservice-B")
	if err := os.MkdirAll(msBDir, 0755); err != nil {
		t.Fatalf("Failed to create msB dir: %v", err)
	}
	msBCfg := filepath.Join(msBDir, ".simplyenv")
	if err := WriteEnvFile(msBCfg, map[string]string{"PORT": "3002", "SERVICE": "B", "DEBUG": "true"}); err != nil {
		t.Fatalf("Failed to write msB config: %v", err)
	}

	// Also create an ignored dir like node_modules with a dummy .simplyenv to make sure it's ignored
	nmDir := filepath.Join(tempDir, "node_modules", "somepkg")
	_ = os.MkdirAll(nmDir, 0755)
	_ = WriteEnvFile(filepath.Join(nmDir, ".simplyenv"), map[string]string{"IGNORE": "true"})

	// Run discovery
	modules := DiscoverModules(tempDir)

	if len(modules) != 3 {
		t.Fatalf("Expected 3 modules (root, microservice-A, microservice-B), got %d: %+v", len(modules), modules)
	}

	modMap := make(map[string]Module)
	for _, m := range modules {
		modMap[m.Name] = m
	}

	// Check root
	rootMod, hasRoot := modMap["root"]
	if !hasRoot || rootMod.VarCount != 1 {
		t.Errorf("Expected root module with 1 var, got %+v", rootMod)
	}

	// Check msA
	msAMod, hasA := modMap["microservice-A"]
	if !hasA || msAMod.VarCount != 2 {
		t.Errorf("Expected microservice-A with 2 vars, got %+v", msAMod)
	}

	// Check msB
	msBMod, hasB := modMap["microservice-B"]
	if !hasB || msBMod.VarCount != 3 {
		t.Errorf("Expected microservice-B with 3 vars, got %+v", msBMod)
	}
}

func TestAddAndRemoveModule(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "simplyenv-addmod-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock home directory for isolated registry test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	projDir := filepath.Join(tempDir, "sample-project")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("Failed to create projDir: %v", err)
	}

	// Register project
	proj, err := RegisterProject(projDir)
	if err != nil {
		t.Fatalf("Failed to register project: %v", err)
	}

	// Add module-A
	modA, err := AddModule(projDir, "service-auth", "services/auth")
	if err != nil {
		t.Fatalf("Failed to add module: %v", err)
	}
	if modA.Name != "service-auth" || modA.RelPath != "services/auth" {
		t.Errorf("Unexpected module A: %+v", modA)
	}
	if !modA.HasConfig {
		t.Errorf("Expected module A to have config file created")
	}

	// Verify file was created on disk
	expectedCfg := filepath.Join(projDir, "services", "auth", ".simplyenv")
	if _, err := os.Stat(expectedCfg); err != nil {
		t.Errorf("Config file was not created on disk: %v", err)
	}

	// Verify in registry
	reg, err := LoadRegistry()
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}
	var loadedProj *Project
	for i := range reg.Projects {
		if reg.Projects[i].Path == proj.Path {
			loadedProj = &reg.Projects[i]
			break
		}
	}
	if loadedProj == nil || len(loadedProj.Modules) != 1 {
		t.Fatalf("Expected 1 module in registry, got %+v", loadedProj)
	}

	// Remove module
	if err := RemoveModule(projDir, modA.Path); err != nil {
		t.Fatalf("Failed to remove module: %v", err)
	}

	// Verify removed in registry
	reg2, _ := LoadRegistry()
	for i := range reg2.Projects {
		if reg2.Projects[i].Path == proj.Path {
			if len(reg2.Projects[i].Modules) != 0 {
				t.Errorf("Expected 0 modules after removal, got %d", len(reg2.Projects[i].Modules))
			}
		}
	}
}
