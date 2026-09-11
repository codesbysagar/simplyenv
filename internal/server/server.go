package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"simplyenv/internal/core"
	"simplyenv/internal/shell"
	"simplyenv/internal/ui"
	"strings"
	"time"
)

// StartServer starts the simplyenv HTTP server and launches the UI in the default browser.
func StartServer(port int, autoOpen bool) error {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/projects", handleProjects)
	mux.HandleFunc("/api/env", handleEnv)
	mux.HandleFunc("/api/import", handleImport)
	mux.HandleFunc("/api/export", handleExport)
	mux.HandleFunc("/api/system/cwd", handleCwd)
	mux.HandleFunc("/api/system/shell", handleShellInfo)
	mux.HandleFunc("/api/system/install-hook", handleInstallHook)
	mux.HandleFunc("/api/system/uninstall-hook", handleUninstallHook)

	// Static UI Assets (Embedded)
	fileServer := http.FileServer(ui.GetFileSystem())
	mux.Handle("/", fileServer)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// If requested port is busy, fallback to any free port
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to bind port: %w", err)
		}
	}
	defer listener.Close()

	actualPort := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", actualPort)

	fmt.Println("==================================================")
	fmt.Printf(" simplyenv UI running at: %s\n", url)
	fmt.Println(" Press Ctrl+C to stop the server")
	fmt.Println("==================================================")

	if autoOpen {
		go func() {
			time.Sleep(150 * time.Millisecond)
			_ = OpenBrowser(url)
		}()
	}

	return http.Serve(listener, mux)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		reg, err := core.LoadRegistry()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{"projects": reg.Projects})

	case http.MethodPost:
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Path) == "" {
			errorResponse(w, http.StatusBadRequest, "Invalid directory path")
			return
		}

		// Ensure directory exists
		if fi, err := os.Stat(req.Path); err != nil || !fi.IsDir() {
			errorResponse(w, http.StatusBadRequest, "Path is not an existing directory")
			return
		}

		proj, err := core.RegisterProject(req.Path)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// If no config file exists yet, create an initial empty .simplyenv
		if !proj.HasConfig {
			targetFile := filepath.Join(req.Path, ".simplyenv")
			_ = core.WriteEnvFile(targetFile, map[string]string{})
			proj.HasConfig = true
			proj.ConfigFile = ".simplyenv"
		}

		jsonResponse(w, http.StatusOK, proj)

	case http.MethodDelete:
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "Invalid request")
			return
		}
		if err := core.UnregisterProject(req.Path); err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]bool{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func resolveConfigFile(dirPath string) string {
	configPath, err := core.FindConfig(dirPath)
	if err == nil {
		return configPath
	}
	return filepath.Join(dirPath, ".simplyenv")
}

func handleEnv(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dirPath := r.URL.Query().Get("path")
		if dirPath == "" {
			errorResponse(w, http.StatusBadRequest, "Missing path parameter")
			return
		}

		configPath, err := core.FindConfig(dirPath)
		if err != nil {
			// No config file exists yet
			jsonResponse(w, http.StatusOK, map[string]interface{}{
				"vars":        map[string]string{},
				"config_file": ".simplyenv",
			})
			return
		}

		envVars, err := core.ParseEnvFile(configPath)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"vars":        envVars,
			"config_file": filepath.Base(configPath),
		})

	case http.MethodPost:
		var req struct {
			Path  string `json:"path"`
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" || req.Key == "" {
			errorResponse(w, http.StatusBadRequest, "Invalid input")
			return
		}

		targetFile := resolveConfigFile(req.Path)
		if err := core.SetEnvVar(targetFile, req.Key, req.Value); err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Ensure registered in catalog
		_, _ = core.RegisterProject(req.Path)

		jsonResponse(w, http.StatusOK, map[string]bool{"success": true})

	case http.MethodDelete:
		var req struct {
			Path string `json:"path"`
			Key  string `json:"key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" || req.Key == "" {
			errorResponse(w, http.StatusBadRequest, "Invalid input")
			return
		}

		targetFile := resolveConfigFile(req.Path)
		if err := core.DeleteEnvVar(targetFile, req.Key); err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]bool{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path      string `json:"path"`
		Content   string `json:"content"`
		Overwrite bool   `json:"overwrite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		errorResponse(w, http.StatusBadRequest, "Invalid import payload")
		return
	}

	importedVars, err := core.ParseRawImport(req.Content)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Failed to parse import content: "+err.Error())
		return
	}

	targetFile := resolveConfigFile(req.Path)
	existingVars, _ := core.ParseEnvFile(targetFile)
	if existingVars == nil {
		existingVars = make(map[string]string)
	}

	for k, v := range importedVars {
		if req.Overwrite || existingVars[k] == "" {
			existingVars[k] = v
		}
	}

	if err := core.WriteEnvFile(targetFile, existingVars); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to write env file: "+err.Error())
		return
	}

	_, _ = core.RegisterProject(req.Path)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(importedVars),
		"total":   len(existingVars),
	})
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dirPath := r.URL.Query().Get("path")
	format := core.ExportFormat(r.URL.Query().Get("format"))
	if dirPath == "" {
		errorResponse(w, http.StatusBadRequest, "Missing path parameter")
		return
	}

	targetFile := resolveConfigFile(dirPath)
	envVars, _ := core.ParseEnvFile(targetFile)
	if envVars == nil {
		envVars = make(map[string]string)
	}

	content, err := core.FormatExport(envVars, format)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"content": content})
}

func handleCwd(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"path": wd})
}

func handleShellInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shellType := r.URL.Query().Get("shell")
	if shellType == "" {
		shellType = shell.DetectUserShell()
	}

	installed, profilePath, err := shell.IsHookInstalled(shellType)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"shell":        shellType,
		"profile_path": profilePath,
		"installed":    installed,
		"hook_cmd":     shell.GetHookCommand(shellType),
	})
}

func handleInstallHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Shell string `json:"shell"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Shell == "" {
		req.Shell = shell.DetectUserShell()
	}

	profilePath, err := shell.InstallHook(req.Shell)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"profile_path": profilePath,
		"message":      fmt.Sprintf("Hook installed successfully into %s", profilePath),
	})
}

func handleUninstallHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Shell string `json:"shell"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Shell == "" {
		req.Shell = shell.DetectUserShell()
	}

	profilePath, err := shell.UninstallHook(req.Shell)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"profile_path": profilePath,
		"message":      fmt.Sprintf("Hook removed from %s", profilePath),
	})
}

// OpenBrowser attempts to open the specified URL in the default web browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

