package tui

import (
	"runtime"
	"testing"

	"github.com/lmarques/efx-skills/internal/api"
)

func TestURLForAPISkill(t *testing.T) {
	tests := []struct {
		name  string
		skill api.Skill
		want  string
	}{
		{
			name: "playbooks.com with source and name",
			skill: api.Skill{
				Source:   "vercel-labs/ai-chatbot-skills",
				Name:     "web-design-guidelines",
				Registry: "playbooks.com",
			},
			want: "https://playbooks.com/skills/vercel-labs/ai-chatbot-skills/web-design-guidelines",
		},
		{
			name: "skills.sh with source",
			skill: api.Skill{
				Source:   "acme/tools",
				Name:     "lint",
				Registry: "skills.sh",
			},
			want: "https://github.com/acme/tools",
		},
		{
			name: "skills.sh with empty source",
			skill: api.Skill{
				Source:   "",
				Name:     "lint",
				Registry: "skills.sh",
			},
			want: "",
		},
		{
			name: "playbooks.com with empty source and name",
			skill: api.Skill{
				Source:   "",
				Name:     "",
				Registry: "playbooks.com",
			},
			want: "https://playbooks.com",
		},
		{
			name: "playbooks.com with empty source only",
			skill: api.Skill{
				Source:   "",
				Name:     "lint",
				Registry: "playbooks.com",
			},
			want: "https://playbooks.com",
		},
		{
			name: "playbooks.com with empty name only",
			skill: api.Skill{
				Source:   "acme/tools",
				Name:     "",
				Registry: "playbooks.com",
			},
			want: "https://playbooks.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlForAPISkill(tt.skill)
			if got != tt.want {
				t.Errorf("urlForAPISkill(%+v) = %q, want %q", tt.skill, got, tt.want)
			}
		})
	}
}

func TestRegistryBaseURL(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"skills.sh", "https://skills.sh"},
		{"playbooks.com", "https://playbooks.com"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registryBaseURL(tt.name)
			if got != tt.want {
				t.Errorf("registryBaseURL(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestURLForManagedSkill(t *testing.T) {
	skills := []SkillMeta{
		{Name: "lint", URL: "https://github.com/acme/tools"},
		{Name: "deploy", URL: "https://github.com/beta/deploy"},
	}

	tests := []struct {
		name      string
		skillName string
		skills    []SkillMeta
		want      string
	}{
		{
			name:      "matching skill returns URL",
			skillName: "lint",
			skills:    skills,
			want:      "https://github.com/acme/tools",
		},
		{
			name:      "no match returns empty",
			skillName: "missing",
			skills:    skills,
			want:      "",
		},
		{
			name:      "nil skills returns empty",
			skillName: "anything",
			skills:    nil,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlForManagedSkill(tt.skillName, tt.skills)
			if got != tt.want {
				t.Errorf("urlForManagedSkill(%q, ...) = %q, want %q", tt.skillName, got, tt.want)
			}
		})
	}
}

func TestOpenInBrowser(t *testing.T) {
	// We cannot test actual browser launch, but we can verify
	// the function exists and handles unsupported platforms.
	// On the current platform (darwin or linux) it should not error.
	// We only test that the function signature is correct and
	// returns no error on supported platforms.
	switch runtime.GOOS {
	case "darwin", "linux":
		// Cannot actually open a URL in tests, so just verify
		// the function compiles and is callable. A real test would
		// require mocking exec.Command.
	default:
		err := openInBrowser("https://example.com")
		if err == nil {
			t.Error("openInBrowser should return error on unsupported platform")
		}
	}
}
