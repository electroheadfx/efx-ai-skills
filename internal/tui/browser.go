package tui

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/lmarques/efx-skills/internal/api"
)

// openInBrowser opens the given URL in the user's default browser.
// It uses cmd.Start() so the TUI does not block waiting for the browser.
func openInBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// urlForAPISkill returns a browser-friendly URL for an api.Skill based on its registry.
//
// For playbooks.com: returns https://playbooks.com/skills/{Source}/{Name} when both
// Source and Name are non-empty, otherwise falls back to https://playbooks.com.
//
// For other registries (skills.sh, github, etc.): returns https://github.com/{Source}
// when Source is non-empty, otherwise returns "".
func urlForAPISkill(s api.Skill) string {
	switch s.Registry {
	case "playbooks.com":
		if s.Source != "" && s.Name != "" {
			return fmt.Sprintf("https://playbooks.com/skills/%s/%s", s.Source, s.Name)
		}
		return "https://playbooks.com"
	default:
		if s.Source != "" {
			return fmt.Sprintf("https://github.com/%s", s.Source)
		}
		return ""
	}
}

// registryBaseURL maps a registry API name to its browser-friendly base URL.
func registryBaseURL(name string) string {
	switch name {
	case "skills.sh":
		return "https://skills.sh"
	case "playbooks.com":
		return "https://playbooks.com"
	default:
		return ""
	}
}

// urlForManagedSkill looks up a skill's URL from the config skills array by name.
// Returns "" if no match is found or skills is nil.
// This function is pure (no disk I/O) -- the caller passes the skills slice.
func urlForManagedSkill(skillName string, skills []SkillMeta) string {
	for _, s := range skills {
		if s.Name == skillName {
			return s.URL
		}
	}
	return ""
}
