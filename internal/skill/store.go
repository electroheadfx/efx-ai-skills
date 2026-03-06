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

// NewStore creates a new skill store. If skillsPath is empty, it defaults
// to ~/.agents/skills. The lock file is placed alongside the skills directory.
func NewStore(skillsPath string) *Store {
	if skillsPath == "" {
		home := os.Getenv("HOME")
		skillsPath = filepath.Join(home, ".agents", "skills")
	}
	return &Store{
		BaseDir:  skillsPath,
		LockFile: filepath.Join(filepath.Dir(skillsPath), ".skill-lock.json"),
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
	// Capture output silently to avoid breaking the TUI
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
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

// gitHubAPIBaseURL is the base URL for GitHub API calls.
// Tests override this to point to httptest.NewServer.
var gitHubAPIBaseURL = "https://api.github.com"

// LockEntry represents an entry in the lock file
type LockEntry struct {
	Source          string `json:"source"`
	SourceType      string `json:"sourceType"`
	SourceURL       string `json:"sourceUrl"`
	SkillPath       string `json:"skillPath,omitempty"`
	SkillFolderHash string `json:"skillFolderHash"`
	CommitHash      string `json:"commitHash"`
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

// AddToLock adds a skill to the lock file with an optional commit hash.
func (s *Store) AddToLock(skillName, source, commitHash string) error {
	lock, err := s.ReadLockFile()
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	lock.Skills[skillName] = LockEntry{
		Source:      source,
		SourceType:  "github",
		SourceURL:   fmt.Sprintf("https://github.com/%s.git", source),
		CommitHash:  commitHash,
		InstalledAt: now,
		UpdatedAt:   now,
	}

	return s.WriteLockFile(lock)
}

// RemoveFromLock removes a skill entry from the lock file.
// If the skill is not present, this is a no-op.
func (s *Store) RemoveFromLock(skillName string) error {
	lock, err := s.ReadLockFile()
	if err != nil {
		return err
	}
	delete(lock.Skills, skillName)
	return s.WriteLockFile(lock)
}

// FetchLatestCommitHash fetches the HEAD commit SHA for a GitHub owner/repo.
func FetchLatestCommitHash(owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?per_page=1", gitHubAPIBaseURL, owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetching latest commit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var commits []struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("decoding commit response: %w", err)
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found for %s/%s", owner, repo)
	}

	return commits[0].SHA, nil
}

// CheckForUpdate checks whether a skill has an upstream update available.
func (s *Store) CheckForUpdate(skillName string) (hasUpdate bool, currentHash string, latestHash string, err error) {
	lock, err := s.ReadLockFile()
	if err != nil {
		return false, "", "", err
	}

	entry, ok := lock.Skills[skillName]
	if !ok {
		return false, "", "", fmt.Errorf("skill %q not found in lock file", skillName)
	}

	// Parse owner/repo from source
	parts := strings.Split(entry.Source, "/")
	if len(parts) < 2 {
		return false, "", "", fmt.Errorf("invalid source format: %s", entry.Source)
	}

	latestHash, err = FetchLatestCommitHash(parts[0], parts[1])
	if err != nil {
		return false, entry.CommitHash, "", err
	}

	currentHash = entry.CommitHash

	// Empty stored hash means legacy install -- always treat as updateable
	if currentHash == "" {
		return true, currentHash, latestHash, nil
	}

	return currentHash != latestHash, currentHash, latestHash, nil
}

// UpdateSkill re-downloads a skill and updates the lock entry with the new commit hash.
func (s *Store) UpdateSkill(skillName string) error {
	lock, err := s.ReadLockFile()
	if err != nil {
		return err
	}

	entry, ok := lock.Skills[skillName]
	if !ok {
		return fmt.Errorf("skill %q not found in lock file", skillName)
	}

	// Re-install from source
	if err := s.Install(entry.Source, skillName); err != nil {
		return fmt.Errorf("reinstalling %s: %w", skillName, err)
	}

	// Fetch latest commit hash
	parts := strings.Split(entry.Source, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid source format: %s", entry.Source)
	}
	latestHash, err := FetchLatestCommitHash(parts[0], parts[1])
	if err != nil {
		return fmt.Errorf("fetching commit hash for %s: %w", skillName, err)
	}

	// Update lock entry
	entry.CommitHash = latestHash
	entry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	lock.Skills[skillName] = entry

	return s.WriteLockFile(lock)
}

// UpdateAllSkills iterates all locked skills and updates each that has a newer upstream commit.
// Returns the list of skill names that were updated. Individual failures are collected but do not
// stop processing of remaining skills.
func (s *Store) UpdateAllSkills() (updated []string, err error) {
	lock, err := s.ReadLockFile()
	if err != nil {
		return nil, err
	}

	var errs []string
	for name := range lock.Skills {
		hasUpdate, _, _, checkErr := s.CheckForUpdate(name)
		if checkErr != nil {
			errs = append(errs, fmt.Sprintf("%s: check failed: %v", name, checkErr))
			continue
		}
		if !hasUpdate {
			continue
		}

		if updateErr := s.UpdateSkill(name); updateErr != nil {
			errs = append(errs, fmt.Sprintf("%s: update failed: %v", name, updateErr))
			continue
		}
		updated = append(updated, name)
	}

	if len(errs) > 0 {
		return updated, fmt.Errorf("some skills failed to update: %s", strings.Join(errs, "; "))
	}
	return updated, nil
}
