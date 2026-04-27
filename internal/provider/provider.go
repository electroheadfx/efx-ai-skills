package provider

import (
	"os"
	"path/filepath"
)

// Provider represents an AI coding agent provider
type Provider struct {
	Name       string
	SkillsPath string
	Configured bool
	SkillCount int
}

// Definition describes a known provider and how to find its skills directory.
type Definition struct {
	Name           string
	DefaultEnabled bool
	Path           func(home string) string
}

var definitions = []Definition{
	{Name: "claude", DefaultEnabled: true, Path: func(h string) string { return filepath.Join(h, ".claude", "skills") }},
	{Name: "cursor", DefaultEnabled: true, Path: func(h string) string { return filepath.Join(h, ".cursor", "skills") }},
	{Name: "qoder", DefaultEnabled: true, Path: func(h string) string { return filepath.Join(h, ".qoder", "skills") }},
	{Name: "windsurf", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".windsurf", "skills") }},
	{Name: "copilot", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".copilot", "skills") }},
	{Name: "opencode", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".config", "opencode", "skills") }},
	{Name: "codex", DefaultEnabled: false, Path: func(h string) string { return filepath.Join(h, ".codex", "skills") }},
}

// Definitions returns the known provider catalog.
func Definitions() []Definition {
	defs := make([]Definition, len(definitions))
	copy(defs, definitions)
	return defs
}

// DetectAll detects all providers on the system
func DetectAll() []Provider {
	home := os.Getenv("HOME")
	var providers []Provider

	for _, def := range definitions {
		p := Provider{
			Name:       def.Name,
			SkillsPath: def.Path(home),
		}

		if info, err := os.Stat(p.SkillsPath); err == nil && info.IsDir() {
			p.Configured = true

			if entries, err := os.ReadDir(p.SkillsPath); err == nil {
				for _, e := range entries {
					if e.Name() != ".DS_Store" {
						p.SkillCount++
					}
				}
			}
		}

		providers = append(providers, p)
	}

	return providers
}

// Get returns a specific provider by name
func Get(name string) *Provider {
	home := os.Getenv("HOME")

	for _, def := range definitions {
		if def.Name == name {
			p := &Provider{
				Name:       def.Name,
				SkillsPath: def.Path(home),
			}

			if info, err := os.Stat(p.SkillsPath); err == nil && info.IsDir() {
				p.Configured = true

				if entries, err := os.ReadDir(p.SkillsPath); err == nil {
					for _, e := range entries {
						if e.Name() != ".DS_Store" {
							p.SkillCount++
						}
					}
				}
			}

			return p
		}
	}

	return nil
}

// GetConfigured returns only configured providers
func GetConfigured() []Provider {
	all := DetectAll()
	var configured []Provider

	for _, p := range all {
		if p.Configured {
			configured = append(configured, p)
		}
	}

	return configured
}

// ListSkills returns skills installed for a provider
func (p *Provider) ListSkills() ([]string, error) {
	if !p.Configured {
		return nil, nil
	}

	entries, err := os.ReadDir(p.SkillsPath)
	if err != nil {
		return nil, err
	}

	var skills []string
	for _, e := range entries {
		if e.Name() != ".DS_Store" {
			skills = append(skills, e.Name())
		}
	}

	return skills, nil
}

// HasSkill checks if provider has a specific skill
func (p *Provider) HasSkill(skillName string) bool {
	skillPath := filepath.Join(p.SkillsPath, skillName)
	_, err := os.Stat(skillPath)
	return err == nil
}

// Configure creates the provider directory if it doesn't exist
func (p *Provider) Configure() error {
	if err := os.MkdirAll(p.SkillsPath, 0755); err != nil {
		return err
	}
	p.Configured = true
	return nil
}
