package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Module represents a microservice or sub-environment within a project.
type Module struct {
	Name       string `json:"name"`        // e.g. "microservice-A" or "root"
	RelPath    string `json:"rel_path"`    // e.g. "microservice-A" or "."
	Path       string `json:"path"`        // absolute path: /project/path/microservice-A
	ConfigFile string `json:"config_file"` // e.g. ".simplyenv"
	HasConfig  bool   `json:"has_config"`  // true if config file exists
	IsAllowed  bool   `json:"is_allowed"`  // true if trusted
	VarCount   int    `json:"var_count"`   // number of variables in this module
	LastActive int64  `json:"last_active"` // timestamp
}

// Project represents a registered directory managed by simplyenv.
type Project struct {
	Path             string   `json:"path"`
	Name             string   `json:"name"`
	ConfigFile       string   `json:"config_file"` // e.g. ".simplyenv"
	HasConfig        bool     `json:"has_config"`
	IsAllowed        bool     `json:"is_allowed"`
	LastActive       int64    `json:"last_active"`
	Modules          []Module `json:"modules"`
	UntrackedModules []string `json:"untracked_modules,omitempty"`
}

// Registry stores the list of known projects.
type Registry struct {
	Projects []Project `json:"projects"`
}

func getRegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "simplyenv")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects.json"), nil
}

// findDirectConfigFile checks if dir directly contains .simplyenv or .envrc (without searching parent directories).
func findDirectConfigFile(dir string) (string, bool) {
	for _, fn := range configFileNames {
		cfgPath := filepath.Join(dir, fn)
		if fi, err := os.Stat(cfgPath); err == nil && !fi.IsDir() {
			return cfgPath, true
		}
	}
	return "", false
}

// DiscoverModules scans a project directory for modules / microservices containing .simplyenv or .envrc.
func DiscoverModules(projectPath string) []Module {
	var modules []Module
	now := time.Now().Unix()

	// 1. Check if the project root itself has a direct config
	if rootCfg, hasRoot := findDirectConfigFile(projectPath); hasRoot {
		allowed, _ := IsConfigAllowed(rootCfg)
		vars, _ := ParseEnvFile(rootCfg)
		modules = append(modules, Module{
			Name:       "root",
			RelPath:    ".",
			Path:       projectPath,
			ConfigFile: filepath.Base(rootCfg),
			HasConfig:  true,
			IsAllowed:  allowed,
			VarCount:   len(vars),
			LastActive: now,
		})
	}

	// 2. Scan subdirectories up to 2 levels deep
	ignoredFolders := map[string]bool{
		".git": true, ".firebase": true, "node_modules": true, "vendor": true,
		"target": true, "dist": true, "build": true, ".next": true, ".nuxt": true,
		"__pycache__": true, ".vscode": true, ".idea": true, ".github": true, ".gemini": true,
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return modules
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subName := entry.Name()
		if strings.HasPrefix(subName, ".") || ignoredFolders[subName] {
			continue
		}

		subPath := filepath.Join(projectPath, subName)
		if cfg, hasCfg := findDirectConfigFile(subPath); hasCfg {
			allowed, _ := IsConfigAllowed(cfg)
			vars, _ := ParseEnvFile(cfg)
			modules = append(modules, Module{
				Name:       subName,
				RelPath:    subName,
				Path:       subPath,
				ConfigFile: filepath.Base(cfg),
				HasConfig:  true,
				IsAllowed:  allowed,
				VarCount:   len(vars),
				LastActive: now,
			})
			continue
		}

		// Check 1 level deeper (e.g., services/microservice-A or packages/app)
		nestedEntries, err := os.ReadDir(subPath)
		if err != nil {
			continue
		}
		for _, nEntry := range nestedEntries {
			if !nEntry.IsDir() {
				continue
			}
			nName := nEntry.Name()
			if strings.HasPrefix(nName, ".") || ignoredFolders[nName] {
				continue
			}
			nPath := filepath.Join(subPath, nName)
			if nCfg, nHasCfg := findDirectConfigFile(nPath); nHasCfg {
				allowed, _ := IsConfigAllowed(nCfg)
				vars, _ := ParseEnvFile(nCfg)
				rel, _ := filepath.Rel(projectPath, nPath)
				modules = append(modules, Module{
					Name:       filepath.ToSlash(rel),
					RelPath:    filepath.ToSlash(rel),
					Path:       nPath,
					ConfigFile: filepath.Base(nCfg),
					HasConfig:  true,
					IsAllowed:  allowed,
					VarCount:   len(vars),
					LastActive: now,
				})
			}
		}
	}

	return modules
}

// LoadRegistry reads known projects from ~/.config/simplyenv/projects.json.
func LoadRegistry() (*Registry, error) {
	path, err := getRegistryPath()
	if err != nil {
		return &Registry{Projects: []Project{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Projects: []Project{}}, nil
		}
		return nil, err
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return &Registry{Projects: []Project{}}, nil
	}

	// Refresh status and modules for each project
	for i := range reg.Projects {
		p := &reg.Projects[i]

		// 1. Direct root config
		if cfg, hasRoot := findDirectConfigFile(p.Path); hasRoot {
			p.HasConfig = true
			p.ConfigFile = filepath.Base(cfg)
			allowed, _ := IsConfigAllowed(cfg)
			p.IsAllowed = allowed
		} else {
			p.HasConfig = false
			p.ConfigFile = ""
			p.IsAllowed = false
		}

		// 2. Discover modules on disk
		discovered := DiscoverModules(p.Path)
		moduleMap := make(map[string]Module)
		for _, m := range discovered {
			moduleMap[m.Path] = m
		}

		untrackedMap := make(map[string]bool)
		for _, u := range p.UntrackedModules {
			untrackedMap[u] = true
		}

		// Merge with existing tracked modules to preserve any custom metadata
		var mergedModules []Module
		seenPaths := make(map[string]bool)

		for _, existing := range p.Modules {
			if untrackedMap[existing.Path] || untrackedMap[existing.RelPath] {
				continue
			}
			if disc, exists := moduleMap[existing.Path]; exists {
				// Refresh disk status
				existing.HasConfig = disc.HasConfig
				existing.ConfigFile = disc.ConfigFile
				existing.IsAllowed = disc.IsAllowed
				existing.VarCount = disc.VarCount
				if existing.Name == "" {
					existing.Name = disc.Name
				}
				mergedModules = append(mergedModules, existing)
				seenPaths[existing.Path] = true
			} else if fi, err := os.Stat(existing.Path); err == nil && fi.IsDir() {
				// Custom module folder exists, check for config
				if cfg, has := findDirectConfigFile(existing.Path); has {
					existing.HasConfig = true
					existing.ConfigFile = filepath.Base(cfg)
					allowed, _ := IsConfigAllowed(cfg)
					existing.IsAllowed = allowed
					vars, _ := ParseEnvFile(cfg)
					existing.VarCount = len(vars)
				} else {
					existing.HasConfig = false
					existing.ConfigFile = ""
					existing.IsAllowed = false
					existing.VarCount = 0
				}
				mergedModules = append(mergedModules, existing)
				seenPaths[existing.Path] = true
			}
		}

		// Append newly discovered modules
		for _, m := range discovered {
			if !seenPaths[m.Path] && !untrackedMap[m.Path] && !untrackedMap[m.RelPath] {
				mergedModules = append(mergedModules, m)
				seenPaths[m.Path] = true
			}
		}

		// If no modules exist at all, but root config exists, add root module
		if len(mergedModules) == 0 && p.HasConfig && !untrackedMap[p.Path] && !untrackedMap["."] {
			vars, _ := ParseEnvFile(filepath.Join(p.Path, p.ConfigFile))
			mergedModules = append(mergedModules, Module{
				Name:       "root",
				RelPath:    ".",
				Path:       p.Path,
				ConfigFile: p.ConfigFile,
				HasConfig:  true,
				IsAllowed:  p.IsAllowed,
				VarCount:   len(vars),
				LastActive: p.LastActive,
			})
		}

		// Sort modules: root first, then by name
		sort.Slice(mergedModules, func(a, b int) bool {
			if mergedModules[a].RelPath == "." {
				return true
			}
			if mergedModules[b].RelPath == "." {
				return false
			}
			return strings.ToLower(mergedModules[a].Name) < strings.ToLower(mergedModules[b].Name)
		})

		p.Modules = mergedModules
	}

	return &reg, nil
}

// SaveRegistry writes the registry back to disk.
func (r *Registry) Save() error {
	path, err := getRegistryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// RegisterProject adds or updates a directory in the registry.
func RegisterProject(dirPath string) (*Project, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, err
	}

	reg, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	cfg, hasConfig := findDirectConfigFile(absPath)
	configFile := ""
	if hasConfig {
		configFile = filepath.Base(cfg)
	}

	name := filepath.Base(absPath)
	now := time.Now().Unix()

	found := false
	var updatedProject Project
	for i := range reg.Projects {
		if reg.Projects[i].Path == absPath {
			reg.Projects[i].LastActive = now
			reg.Projects[i].HasConfig = hasConfig
			reg.Projects[i].ConfigFile = configFile
			updatedProject = reg.Projects[i]
			found = true
			break
		}
	}

	if !found {
		discovered := DiscoverModules(absPath)
		updatedProject = Project{
			Path:       absPath,
			Name:       name,
			ConfigFile: configFile,
			HasConfig:  hasConfig,
			LastActive: now,
			Modules:    discovered,
		}
		reg.Projects = append(reg.Projects, updatedProject)
	}

	// Sort projects by last active descending
	sort.Slice(reg.Projects, func(i, j int) bool {
		return reg.Projects[i].LastActive > reg.Projects[j].LastActive
	})

	if err := reg.Save(); err != nil {
		return nil, err
	}

	return &updatedProject, nil
}

// AddModule adds or creates a microservice/module under a project directory.
func AddModule(projectPath, name, relPath string) (*Module, error) {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("module name cannot be empty")
	}

	relPath = strings.TrimSpace(relPath)
	if relPath == "" {
		relPath = name
	}
	relPath = filepath.Clean(relPath)

	if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
		return nil, errors.New("invalid module relative path")
	}

	absModulePath := filepath.Join(absProjectPath, relPath)

	// Ensure module directory exists
	if err := os.MkdirAll(absModulePath, 0755); err != nil {
		return nil, err
	}

	// Ensure .simplyenv exists
	cfgPath, hasCfg := findDirectConfigFile(absModulePath)
	if !hasCfg {
		cfgPath = filepath.Join(absModulePath, ".simplyenv")
		if err := WriteEnvFile(cfgPath, map[string]string{}); err != nil {
			return nil, err
		}
		hasCfg = true
	}

	// Auto-trust the new module config
	_, _ = AllowConfig(cfgPath)
	allowed, _ := IsConfigAllowed(cfgPath)
	vars, _ := ParseEnvFile(cfgPath)
	now := time.Now().Unix()

	newModule := Module{
		Name:       name,
		RelPath:    filepath.ToSlash(relPath),
		Path:       absModulePath,
		ConfigFile: filepath.Base(cfgPath),
		HasConfig:  hasCfg,
		IsAllowed:  allowed,
		VarCount:   len(vars),
		LastActive: now,
	}

	reg, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	// Find project and add or update module
	projFound := false
	for i := range reg.Projects {
		if reg.Projects[i].Path == absProjectPath {
			projFound = true

			// Remove from UntrackedModules if it was previously untracked
			var remainingUntracked []string
			for _, u := range reg.Projects[i].UntrackedModules {
				if u != absModulePath && u != relPath && u != filepath.ToSlash(relPath) {
					remainingUntracked = append(remainingUntracked, u)
				}
			}
			reg.Projects[i].UntrackedModules = remainingUntracked

			modFound := false
			for j := range reg.Projects[i].Modules {
				if reg.Projects[i].Modules[j].Path == absModulePath {
					reg.Projects[i].Modules[j] = newModule
					modFound = true
					break
				}
			}
			if !modFound {
				reg.Projects[i].Modules = append(reg.Projects[i].Modules, newModule)
			}
			reg.Projects[i].LastActive = now
			break
		}
	}

	if !projFound {
		// Project wasn't registered yet; register it now
		proj, regErr := RegisterProject(absProjectPath)
		if regErr != nil {
			return nil, regErr
		}
		proj.Modules = append(proj.Modules, newModule)
		reg, _ = LoadRegistry()
	}

	if err := reg.Save(); err != nil {
		return nil, err
	}

	return &newModule, nil
}

// RemoveModule removes a module from project tracking.
func RemoveModule(projectPath, modulePath string) error {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	absModulePath, err := filepath.Abs(modulePath)
	if err != nil {
		return err
	}

	reg, err := LoadRegistry()
	if err != nil {
		return err
	}

	for i := range reg.Projects {
		if reg.Projects[i].Path == absProjectPath {
			newMods := make([]Module, 0, len(reg.Projects[i].Modules))
			for _, m := range reg.Projects[i].Modules {
				if m.Path != absModulePath && m.RelPath != modulePath {
					newMods = append(newMods, m)
				}
			}
			reg.Projects[i].Modules = newMods

			// Record in UntrackedModules so auto-discovery won't resurrect it
			alreadyInUntracked := false
			for _, u := range reg.Projects[i].UntrackedModules {
				if u == absModulePath {
					alreadyInUntracked = true
					break
				}
			}
			if !alreadyInUntracked {
				reg.Projects[i].UntrackedModules = append(reg.Projects[i].UntrackedModules, absModulePath)
			}
			break
		}
	}

	return reg.Save()
}

// TouchProjectAndModule updates the last active timestamp for a path (which might be a module or project).
func TouchProjectAndModule(targetPath string) {
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return
	}

	reg, err := LoadRegistry()
	if err != nil {
		return
	}

	now := time.Now().Unix()
	modified := false

	for i := range reg.Projects {
		if reg.Projects[i].Path == absPath {
			reg.Projects[i].LastActive = now
			modified = true
		}
		for j := range reg.Projects[i].Modules {
			if reg.Projects[i].Modules[j].Path == absPath {
				reg.Projects[i].Modules[j].LastActive = now
				reg.Projects[i].LastActive = now
				modified = true
			}
		}
	}

	if modified {
		_ = reg.Save()
	}
}

// UnregisterProject removes a directory from the registry.
func UnregisterProject(dirPath string) error {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return err
	}

	reg, err := LoadRegistry()
	if err != nil {
		return err
	}

	newProjects := make([]Project, 0, len(reg.Projects))
	for _, p := range reg.Projects {
		if p.Path != absPath {
			newProjects = append(newProjects, p)
		}
	}
	reg.Projects = newProjects

	return reg.Save()
}
