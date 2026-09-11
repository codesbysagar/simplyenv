package core

import (
	"strings"
	"testing"
)

func TestParseRawImportEnvFormat(t *testing.T) {
	raw := `
# Configuration
PORT=4000
SECRET="super_secret_value"
HOST='127.0.0.1'
export REDIS_URL=redis://localhost:6379
`
	res, err := ParseRawImport(raw)
	if err != nil {
		t.Fatalf("ParseRawImport failed: %v", err)
	}

	if res["PORT"] != "4000" {
		t.Errorf("Expected PORT=4000, got %s", res["PORT"])
	}
	if res["SECRET"] != "super_secret_value" {
		t.Errorf("Expected SECRET=super_secret_value, got %s", res["SECRET"])
	}
	if res["HOST"] != "127.0.0.1" {
		t.Errorf("Expected HOST=127.0.0.1, got %s", res["HOST"])
	}
	if res["REDIS_URL"] != "redis://localhost:6379" {
		t.Errorf("Expected REDIS_URL=redis://localhost:6379, got %s", res["REDIS_URL"])
	}
}

func TestParseRawImportJSONFormat(t *testing.T) {
	jsonRaw := `{
		"PORT": "5000",
		"APP_NAME": "simplyenv-test"
	}`
	res, err := ParseRawImport(jsonRaw)
	if err != nil {
		t.Fatalf("ParseRawImport JSON failed: %v", err)
	}

	if res["PORT"] != "5000" {
		t.Errorf("Expected PORT=5000, got %s", res["PORT"])
	}
	if res["APP_NAME"] != "simplyenv-test" {
		t.Errorf("Expected APP_NAME=simplyenv-test, got %s", res["APP_NAME"])
	}
}

func TestFormatExport(t *testing.T) {
	vars := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	// Test shell
	shellOut, err := FormatExport(vars, FormatShell)
	if err != nil {
		t.Fatalf("FormatExport shell failed: %v", err)
	}
	if !strings.Contains(shellOut, "export FOO=\"bar\"") {
		t.Errorf("Missing export statement in shell output: %s", shellOut)
	}

	// Test json
	jsonOut, err := FormatExport(vars, FormatJSON)
	if err != nil {
		t.Fatalf("FormatExport json failed: %v", err)
	}
	if !strings.Contains(jsonOut, `"FOO": "bar"`) {
		t.Errorf("Missing JSON key in json output: %s", jsonOut)
	}
}
