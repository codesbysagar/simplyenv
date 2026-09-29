package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecurityMiddlewareHostValidation(t *testing.T) {
	handler := NewHandler()

	// Allowed hosts
	allowedHosts := []string{"127.0.0.1:8085", "localhost:8085", "127.0.0.1", "localhost", "[::1]:8085"}
	for _, host := range allowedHosts {
		req := httptest.NewRequest(http.MethodGet, "/api/system/cwd", nil)
		req.Host = host
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 for host %q, got %d", host, rec.Code)
		}
	}

	// Disallowed hosts (DNS rebinding attacks)
	disallowedHosts := []string{"evil.com", "attacker.net:8085", "192.168.1.5:8085", "example.com"}
	for _, host := range disallowedHosts {
		req := httptest.NewRequest(http.MethodGet, "/api/system/cwd", nil)
		req.Host = host
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for host %q, got %d", host, rec.Code)
		}
	}
}

func TestSecurityMiddlewareOriginValidation(t *testing.T) {
	handler := NewHandler()

	// Allowed origins
	allowedOrigins := []string{
		"http://127.0.0.1:8085",
		"http://localhost:8085",
		"http://127.0.0.1",
		"http://localhost",
	}
	for _, origin := range allowedOrigins {
		req := httptest.NewRequest(http.MethodGet, "/api/system/cwd", nil)
		req.Host = "127.0.0.1:8085"
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 for origin %q, got %d", origin, rec.Code)
		}
	}

	// Disallowed origins (Cross-origin tampering)
	disallowedOrigins := []string{
		"https://evil.com",
		"http://attacker.org:8085",
		"https://google.com",
	}
	for _, origin := range disallowedOrigins {
		req := httptest.NewRequest(http.MethodGet, "/api/system/cwd", nil)
		req.Host = "127.0.0.1:8085"
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for origin %q, got %d", origin, rec.Code)
		}
	}
}

func TestSecurityMiddlewareCrossSiteStateModification(t *testing.T) {
	handler := NewHandler()

	// State-modifying POST request with Sec-Fetch-Site: cross-site
	body := strings.NewReader(`{"shell":"bash"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/system/install-hook", body)
	req.Host = "127.0.0.1:8085"
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for Sec-Fetch-Site: cross-site, got %d", rec.Code)
	}

	// State-modifying POST request with foreign Referer
	body2 := strings.NewReader(`{"shell":"bash"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/system/install-hook", body2)
	req2.Host = "127.0.0.1:8085"
	req2.Header.Set("Referer", "https://malicious-site.com/attack")
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for malicious Referer, got %d", rec2.Code)
	}
}

func TestProjectsAndModulesAPI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "simplyenv-api-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	handler := NewHandler()

	projDir := filepath.Join(tempDir, "sample-project")
	_ = os.MkdirAll(projDir, 0755)

	// 1. Register project via POST /api/projects
	body := strings.NewReader(fmt.Sprintf(`{"path":%q}`, projDir))
	req := httptest.NewRequest(http.MethodPost, "/api/projects", body)
	req.Host = "localhost:8085"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 from POST /api/projects, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Add module via POST /api/modules
	modBody := strings.NewReader(fmt.Sprintf(`{"project_path":%q,"name":"microservice-A","rel_path":"microservice-A"}`, projDir))
	modReq := httptest.NewRequest(http.MethodPost, "/api/modules", modBody)
	modReq.Host = "localhost:8085"
	modRec := httptest.NewRecorder()
	handler.ServeHTTP(modRec, modReq)

	if modRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 from POST /api/modules, got %d: %s", modRec.Code, modRec.Body.String())
	}

	// 3. Get projects via GET /api/projects and verify module is present
	listReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	listReq.Host = "localhost:8085"
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 from GET /api/projects, got %d", listRec.Code)
	}
	if !strings.Contains(listRec.Body.String(), "microservice-A") {
		t.Errorf("Expected response to contain microservice-A, got %s", listRec.Body.String())
	}

	// 4. Delete module via DELETE /api/modules
	delBody := strings.NewReader(fmt.Sprintf(`{"project_path":%q,"module_path":%q}`, projDir, filepath.Join(projDir, "microservice-A")))
	delReq := httptest.NewRequest(http.MethodDelete, "/api/modules", delBody)
	delReq.Host = "localhost:8085"
	delRec := httptest.NewRecorder()
	handler.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 from DELETE /api/modules, got %d: %s", delRec.Code, delRec.Body.String())
	}
}

