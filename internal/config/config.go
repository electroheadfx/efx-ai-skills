package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Registries []Registry                `json:"registries"`
	Repos      []string                  `json:"repos"`
	Providers  map[string]ProviderConfig `json:"providers"`
}

// Registry represents a skill registry
type Registry struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home := os.Getenv("HOME")

	return &Config{
		Registries: []Registry{
			{Name: "skills.sh", URL: "https://skills.sh/api/search", Enabled: true},
			{Name: "playbooks.com", URL: "https://playbooks.com/api/skills", Enabled: true},
		},
		Repos: []string{
			"yoanbernabeu/grepai-skills",
			"better-auth/skills",
			"awni/mlx-skills",
		},
		Providers: map[string]ProviderConfig{
			"claude":   {Enabled: true, Path: filepath.Join(home, ".claude", "skills")},
			"cursor":   {Enabled: true, Path: filepath.Join(home, ".cursor", "skills")},
			"qoder":    {Enabled: true, Path: filepath.Join(home, ".qoder", "skills")},
			"windsurf": {Enabled: false, Path: filepath.Join(home, ".windsurf", "skills")},
			"copilot":  {Enabled: false, Path: filepath.Join(home, ".copilot", "skills")},
			"opencode": {Enabled: false, Path: filepath.Join(home, ".opencode", "skills")},
		},
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "efx-skills", "config.json")
}

// Load loads configuration from file
func Load() (*Config, error) {
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	path := ConfigPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddRepo adds a custom repository
func (c *Config) AddRepo(repo string) {
	// Check if already exists
	for _, r := range c.Repos {
		if r == repo {
			return
		}
	}
	c.Repos = append(c.Repos, repo)
}

// RemoveRepo removes a custom repository
func (c *Config) RemoveRepo(repo string) {
	for i, r := range c.Repos {
		if r == repo {
			c.Repos = append(c.Repos[:i], c.Repos[i+1:]...)
			return
		}
	}
}

// EnableRegistry enables or disables a registry
func (c *Config) EnableRegistry(name string, enabled bool) {
	for i, r := range c.Registries {
		if r.Name == name {
			c.Registries[i].Enabled = enabled
			return
		}
	}
}

// EnableProvider enables or disables a provider
func (c *Config) EnableProvider(name string, enabled bool) {
	if p, ok := c.Providers[name]; ok {
		p.Enabled = enabled
		c.Providers[name] = p
	}
}

// GetEnabledRegistries returns only enabled registries
func (c *Config) GetEnabledRegistries() []Registry {
	var enabled []Registry
	for _, r := range c.Registries {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled
}

// GetEnabledProviders returns only enabled providers
func (c *Config) GetEnabledProviders() map[string]ProviderConfig {
	enabled := make(map[string]ProviderConfig)
	for name, p := range c.Providers {
		if p.Enabled {
			enabled[name] = p
		}
	}
	return enabled
}
