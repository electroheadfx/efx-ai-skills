package skill

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Store handles local skill storage
type Store struct {
	BaseDir  string // ~/.agents/skills
	LockFile string // ~/.agents/.skill-lock.json
}

// NewStore creates a new skill store
func NewStore() *Store {
	home := os.Getenv("HOME")
	return &Store{
		BaseDir:  filepath.Join(home, ".agents", "skills"),
		LockFile: filepath.Join(home, ".agents", ".skill-lock.json"),
	}
}

// Install installs a skill from a source
func (s *Store) Install(source, skillName string) error {
	// Use npx skills if available
	if _, err := exec.LookPath("npx"); err == nil {
		return s.installViaSkills(source, skillName)
	}

	// Fallback to direct download
	return s.installDirect(source, skillName)
}

// installViaSkills uses npx skills add command
func (s *Store) installViaSkills(source, skillName string) error {
	args := []string{"skills", "add", source, "-g", "-y"}
	if skillName != "" {
		args = append(args, "--skill", skillName)
	}

	cmd := exec.Command("npx", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// installDirect downloads skill directly from GitHub
func (s *Store) installDirect(source, skillName string) error {
	// Parse source (e.g., "owner/repo")
	parts := strings.Split(source, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid source format: %s (expected owner/repo)", source)
	}

	owner := parts[0]
	repo := parts[1]

	// Create skill directory
	skillDir := filepath.Join(s.BaseDir, skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	// Download SKILL.md
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/skills/%s/SKILL.md", owner, repo, skillName)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try alternative path
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s/SKILL.md", owner, repo, skillName)
		resp, err = http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download skill: %s", resp.Status)
		}
	}

	// Write SKILL.md
	skillFile := filepath.Join(skillDir, "SKILL.md")
	out, err := os.Create(skillFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// LinkToProvider creates a symlink from provider skills dir to central storage
func (s *Store) LinkToProvider(skillName, providerPath string) error {
	// Ensure provider directory exists
	if err := os.MkdirAll(providerPath, 0755); err != nil {
		return err
	}

	// Source path (central storage)
	sourcePath := filepath.Join(s.BaseDir, skillName)

	// Target path (provider)
	targetPath := filepath.Join(providerPath, skillName)

	// Remove existing link if present
	os.Remove(targetPath)

	// Create relative symlink
	relPath, err := filepath.Rel(providerPath, sourcePath)
	if err != nil {
		return err
	}

	return os.Symlink(relPath, targetPath)
}

// UnlinkFromProvider removes a symlink from provider skills dir
func (s *Store) UnlinkFromProvider(skillName, providerPath string) error {
	targetPath := filepath.Join(providerPath, skillName)
	return os.Remove(targetPath)
}

// ListInstalled returns all installed skills
func (s *Store) ListInstalled() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		return nil, err
	}

	var skills []string
	for _, e := range entries {
		if e.IsDir() {
			skills = append(skills, e.Name())
		}
	}

	return skills, nil
}

// IsInstalled checks if a skill is installed
func (s *Store) IsInstalled(skillName string) bool {
	skillPath := filepath.Join(s.BaseDir, skillName, "SKILL.md")
	_, err := os.Stat(skillPath)
	return err == nil
}

// LockEntry represents an entry in the lock file
type LockEntry struct {
	Source          string `json:"source"`
	SourceType      string `json:"sourceType"`
	SourceURL       string `json:"sourceUrl"`
	SkillPath       string `json:"skillPath,omitempty"`
	SkillFolderHash string `json:"skillFolderHash"`
	InstalledAt     string `json:"installedAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// LockFile represents the skill lock file
type LockFile struct {
	Version int                  `json:"version"`
	Skills  map[string]LockEntry `json:"skills"`
}

// ReadLockFile reads the skill lock file
func (s *Store) ReadLockFile() (*LockFile, error) {
	data, err := os.ReadFile(s.LockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{Version: 3, Skills: make(map[string]LockEntry)}, nil
		}
		return nil, err
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// WriteLockFile writes the skill lock file
func (s *Store) WriteLockFile(lock *LockFile) error {
	dir := filepath.Dir(s.LockFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.LockFile, data, 0644)
}

// AddToLock adds a skill to the lock file
func (s *Store) AddToLock(skillName, source string) error {
	lock, err := s.ReadLockFile()
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	lock.Skills[skillName] = LockEntry{
		Source:      source,
		SourceType:  "github",
		SourceURL:   fmt.Sprintf("https://github.com/%s.git", source),
		InstalledAt: now,
		UpdatedAt:   now,
	}

	return s.WriteLockFile(lock)
}
