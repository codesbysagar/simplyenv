package server

import (
	"net/http"
	"net/http/httptest"
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
