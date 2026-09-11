package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Project represents a registered directory managed by simplyenv.
type Project struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	ConfigFile string `json:"config_file"` // e.g. ".simplyenv"
	HasConfig  bool   `json:"has_config"`
	LastActive int64  `json:"last_active"`
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

	// Refresh status of each project
	for i := range reg.Projects {
		p := &reg.Projects[i]
		cfg, err := FindConfig(p.Path)
		if err == nil {
			p.HasConfig = true
			p.ConfigFile = filepath.Base(cfg)
		} else {
			p.HasConfig = false
			p.ConfigFile = ""
		}
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

	cfg, cfgErr := FindConfig(absPath)
	configFile := ""
	hasConfig := false
	if cfgErr == nil {
		configFile = filepath.Base(cfg)
		hasConfig = true
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
		updatedProject = Project{
			Path:       absPath,
			Name:       name,
			ConfigFile: configFile,
			HasConfig:  hasConfig,
			LastActive: now,
		}
		reg.Projects = append(reg.Projects, updatedProject)
	}

	// Sort by last active descending
	sort.Slice(reg.Projects, func(i, j int) bool {
		return reg.Projects[i].LastActive > reg.Projects[j].LastActive
	})

	if err := reg.Save(); err != nil {
		return nil, err
	}

	return &updatedProject, nil
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
