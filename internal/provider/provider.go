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

// Known providers and their paths
var KnownProviders = []struct {
	Name     string
	PathFunc func(home string) string
}{
	{"claude", func(h string) string { return filepath.Join(h, ".claude", "skills") }},
	{"cursor", func(h string) string { return filepath.Join(h, ".cursor", "skills") }},
	{"qoder", func(h string) string { return filepath.Join(h, ".qoder", "skills") }},
	{"windsurf", func(h string) string { return filepath.Join(h, ".windsurf", "skills") }},
	{"copilot", func(h string) string { return filepath.Join(h, ".copilot", "skills") }},
	{"opencode", func(h string) string { return filepath.Join(h, ".opencode", "skills") }},
}

// DetectAll detects all providers on the system
func DetectAll() []Provider {
	home := os.Getenv("HOME")
	var providers []Provider

	for _, kp := range KnownProviders {
		p := Provider{
			Name:       kp.Name,
			SkillsPath: kp.PathFunc(home),
		}

		// Check if directory exists
		if info, err := os.Stat(p.SkillsPath); err == nil && info.IsDir() {
			p.Configured = true

			// Count skills
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

	for _, kp := range KnownProviders {
		if kp.Name == name {
			p := &Provider{
				Name:       kp.Name,
				SkillsPath: kp.PathFunc(home),
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
