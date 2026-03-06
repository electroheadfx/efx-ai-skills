package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- RunDiagnostics tests ---

func TestRunDiagnostics_MissingFromFS(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory (empty -- no skills installed)
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Config says "tooling" is tracked, but it's not on the filesystem
	configSkills := []SkillMeta{
		{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools"},
	}

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	if len(report.MissingFromFS) != 1 {
		t.Fatalf("MissingFromFS = %v, want 1 entry", report.MissingFromFS)
	}
	if report.MissingFromFS[0] != "tooling" {
		t.Errorf("MissingFromFS[0] = %q, want %q", report.MissingFromFS[0], "tooling")
	}

	// Should have an error-severity issue
	found := false
	for _, issue := range report.Issues {
		if issue.SkillName == "tooling" && issue.Severity == "error" && issue.Category == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error-severity 'missing' issue for 'tooling', got issues: %+v", report.Issues)
	}
}

func TestRunDiagnostics_UntrackedOnFS(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory with "mystuff" installed
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(filepath.Join(skillsDir, "mystuff"), 0755)

	// Config has no skills
	configSkills := []SkillMeta{}

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	if len(report.UntrackedOnFS) != 1 {
		t.Fatalf("UntrackedOnFS = %v, want 1 entry", report.UntrackedOnFS)
	}
	if report.UntrackedOnFS[0] != "mystuff" {
		t.Errorf("UntrackedOnFS[0] = %q, want %q", report.UntrackedOnFS[0], "mystuff")
	}

	// Should have a warning-severity issue
	found := false
	for _, issue := range report.Issues {
		if issue.SkillName == "mystuff" && issue.Severity == "warning" && issue.Category == "untracked" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning-severity 'untracked' issue for 'mystuff', got issues: %+v", report.Issues)
	}
}

func TestRunDiagnostics_BackfillCandidate(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory with a skill on FS but not in config
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(filepath.Join(skillsDir, "grepai"), 0755)

	// Create a lock file with an entry for "grepai"
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"grepai": map[string]interface{}{
				"source":     "yoanbernabeu/grepai-skills",
				"sourceType": "github",
				"sourceUrl":  "https://github.com/yoanbernabeu/grepai-skills.git",
				"commitHash": "abc123",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	configSkills := []SkillMeta{}

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	// Should be untracked AND a backfill candidate (has lock entry)
	if len(report.UntrackedOnFS) != 1 {
		t.Fatalf("UntrackedOnFS = %v, want 1 entry", report.UntrackedOnFS)
	}
	if len(report.BackfillCandidates) != 1 {
		t.Fatalf("BackfillCandidates = %v, want 1 entry", report.BackfillCandidates)
	}
	if report.BackfillCandidates[0] != "grepai" {
		t.Errorf("BackfillCandidates[0] = %q, want %q", report.BackfillCandidates[0], "grepai")
	}

	// Should have an info-severity backfill issue
	found := false
	for _, issue := range report.Issues {
		if issue.SkillName == "grepai" && issue.Severity == "info" && issue.Category == "backfill" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected info-severity 'backfill' issue for 'grepai', got issues: %+v", report.Issues)
	}
}

func TestRunDiagnostics_BackfillCandidateFromKnownCorrespondences(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory with a known-correspondence skill but NO lock entry
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(filepath.Join(skillsDir, "agent-browser"), 0755)

	// No lock file -- should fall back to knownCorrespondences
	configSkills := []SkillMeta{}

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	if len(report.BackfillCandidates) != 1 {
		t.Fatalf("BackfillCandidates = %v, want 1 entry (from known correspondences)", report.BackfillCandidates)
	}
	if report.BackfillCandidates[0] != "agent-browser" {
		t.Errorf("BackfillCandidates[0] = %q, want %q", report.BackfillCandidates[0], "agent-browser")
	}
}

func TestRunDiagnostics_AllHealthy(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory with "tooling"
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(filepath.Join(skillsDir, "tooling"), 0755)

	// Config also has "tooling"
	configSkills := []SkillMeta{
		{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools"},
	}

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	if len(report.Issues) != 0 {
		t.Errorf("expected no issues, got: %+v", report.Issues)
	}
	if len(report.MissingFromFS) != 0 {
		t.Errorf("MissingFromFS = %v, want empty", report.MissingFromFS)
	}
	if len(report.UntrackedOnFS) != 0 {
		t.Errorf("UntrackedOnFS = %v, want empty", report.UntrackedOnFS)
	}
}

// --- BackfillLegacySkills tests ---

func TestBackfillLegacySkills_FromLockFile(t *testing.T) {
	tmp := setTestHome(t)

	// Create config directory and initial config
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Create a lock file with metadata for "grepai"
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"grepai": map[string]interface{}{
				"source":     "yoanbernabeu/grepai-skills",
				"sourceType": "github",
				"sourceUrl":  "https://github.com/yoanbernabeu/grepai-skills.git",
				"commitHash": "abc123",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	backfilled, err := BackfillLegacySkills([]string{"grepai"}, skillsDir)
	if err != nil {
		t.Fatalf("BackfillLegacySkills failed: %v", err)
	}

	if len(backfilled) != 1 {
		t.Fatalf("backfilled = %v, want 1 entry", backfilled)
	}
	if backfilled[0] != "grepai" {
		t.Errorf("backfilled[0] = %q, want %q", backfilled[0], "grepai")
	}

	// Verify skill was added to config
	cfg := loadConfigFromFile()
	if cfg == nil {
		t.Fatal("config file not created")
	}
	found := false
	for _, s := range cfg.Skills {
		if s.Name == "grepai" && s.Owner == "yoanbernabeu/grepai-skills" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("grepai not found in config skills: %+v", cfg.Skills)
	}
}

func TestBackfillLegacySkills_FromKnownCorrespondences(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory but NO lock file
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	backfilled, err := BackfillLegacySkills([]string{"agent-browser"}, skillsDir)
	if err != nil {
		t.Fatalf("BackfillLegacySkills failed: %v", err)
	}

	if len(backfilled) != 1 {
		t.Fatalf("backfilled = %v, want 1 entry", backfilled)
	}

	// Verify skill was added to config from known correspondences
	cfg := loadConfigFromFile()
	if cfg == nil {
		t.Fatal("config file not created")
	}
	found := false
	for _, s := range cfg.Skills {
		if s.Name == "agent-browser" {
			found = true
			if s.Owner != "vercel-labs/agent-browser" {
				t.Errorf("agent-browser Owner = %q, want %q", s.Owner, "vercel-labs/agent-browser")
			}
			if s.URL != "https://github.com/vercel-labs/agent-browser" {
				t.Errorf("agent-browser URL = %q, want %q", s.URL, "https://github.com/vercel-labs/agent-browser")
			}
			break
		}
	}
	if !found {
		t.Errorf("agent-browser not found in config skills: %+v", cfg.Skills)
	}
}

func TestBackfillLegacySkills_UnknownSkillSkipped(t *testing.T) {
	tmp := setTestHome(t)

	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// "unknown-skill" is not in lock or known correspondences
	backfilled, err := BackfillLegacySkills([]string{"unknown-skill"}, skillsDir)
	if err != nil {
		t.Fatalf("BackfillLegacySkills failed: %v", err)
	}

	if len(backfilled) != 0 {
		t.Errorf("backfilled = %v, want empty (unknown skill should be skipped)", backfilled)
	}
}

// --- FormatReport tests ---

func TestFormatReport_WithIssues(t *testing.T) {
	report := &DoctorReport{
		Issues: []DoctorIssue{
			{Severity: "error", Category: "missing", SkillName: "tooling", Message: "Tracked in config but not found on filesystem", Fix: "Run 'efx install tooling' to reinstall"},
			{Severity: "warning", Category: "untracked", SkillName: "mystuff", Message: "Found on filesystem but not tracked in config", Fix: "Run doctor backfill or add manually"},
			{Severity: "info", Category: "backfill", SkillName: "grepai", Message: "Can be auto-backfilled from lock file", Fix: "Run doctor backfill to add metadata"},
		},
		MissingFromFS:    []string{"tooling"},
		UntrackedOnFS:    []string{"mystuff", "grepai"},
		BackfillCandidates: []string{"grepai"},
	}

	output := FormatReport(report)

	// Should contain all skill names
	if !strings.Contains(output, "tooling") {
		t.Errorf("output missing 'tooling': %s", output)
	}
	if !strings.Contains(output, "mystuff") {
		t.Errorf("output missing 'mystuff': %s", output)
	}
	if !strings.Contains(output, "grepai") {
		t.Errorf("output missing 'grepai': %s", output)
	}

	// Should contain severity indicators
	if !strings.Contains(output, "!") {
		t.Errorf("output missing error indicator '!': %s", output)
	}
	if !strings.Contains(output, "?") {
		t.Errorf("output missing warning indicator '?': %s", output)
	}

	// Should contain summary line
	if !strings.Contains(output, "3 issues found") {
		t.Errorf("output missing summary '3 issues found': %s", output)
	}

	// Errors should appear before warnings (check order)
	errorIdx := strings.Index(output, "tooling")
	warningIdx := strings.Index(output, "mystuff")
	if errorIdx > warningIdx {
		t.Errorf("errors should appear before warnings in output")
	}
}

// --- EnrichCandidates detection tests ---

func TestRunDiagnostics_EnrichCandidate(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory with "tooling" on FS
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(filepath.Join(skillsDir, "tooling"), 0755)

	// Config has "tooling" but WITHOUT version/installed
	configSkills := []SkillMeta{
		{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools"},
	}

	// Create lock file with CommitHash and InstalledAt for "tooling"
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"tooling": map[string]interface{}{
				"source":      "acme/tools",
				"sourceType":  "github",
				"sourceUrl":   "https://github.com/acme/tools.git",
				"commitHash":  "abc123",
				"installedAt": "2025-01-15T10:00:00Z",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	// Should have enrich candidate
	if len(report.EnrichCandidates) != 1 {
		t.Fatalf("EnrichCandidates = %v, want 1 entry", report.EnrichCandidates)
	}
	if report.EnrichCandidates[0] != "tooling" {
		t.Errorf("EnrichCandidates[0] = %q, want %q", report.EnrichCandidates[0], "tooling")
	}

	// Should have info-severity "enrich" issue
	found := false
	for _, issue := range report.Issues {
		if issue.SkillName == "tooling" && issue.Severity == "info" && issue.Category == "enrich" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected info-severity 'enrich' issue for 'tooling', got issues: %+v", report.Issues)
	}
}

func TestRunDiagnostics_NoEnrichWhenAlreadyPopulated(t *testing.T) {
	tmp := setTestHome(t)

	// Create skills directory with "tooling" on FS
	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(filepath.Join(skillsDir, "tooling"), 0755)

	// Config has "tooling" WITH version and installed already set
	configSkills := []SkillMeta{
		{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools", Version: "abc123", Installed: "2025-01-15T10:00:00Z"},
	}

	// Lock file exists too
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"tooling": map[string]interface{}{
				"source":      "acme/tools",
				"sourceType":  "github",
				"sourceUrl":   "https://github.com/acme/tools.git",
				"commitHash":  "abc123",
				"installedAt": "2025-01-15T10:00:00Z",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	report, err := RunDiagnostics(skillsDir, configSkills)
	if err != nil {
		t.Fatalf("RunDiagnostics failed: %v", err)
	}

	// Should NOT have enrich candidates
	if len(report.EnrichCandidates) != 0 {
		t.Errorf("EnrichCandidates = %v, want empty (already populated)", report.EnrichCandidates)
	}

	// Should have no issues at all
	if len(report.Issues) != 0 {
		t.Errorf("expected no issues, got: %+v", report.Issues)
	}
}

// --- EnrichExistingSkills tests ---

func TestEnrichExistingSkills_PopulatesFromLock(t *testing.T) {
	tmp := setTestHome(t)

	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Create config with a skill missing version/installed
	configDir := filepath.Join(tmp, ".config", "efx-skills")
	os.MkdirAll(configDir, 0755)
	cfg := ConfigData{
		Registries: defaultRegistries(),
		Repos:      defaultRepos(),
		SkillsPath: skillsDir,
		Skills: []SkillMeta{
			{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools"},
		},
	}
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, "config.json"), cfgJSON, 0644)

	// Create lock file with data for "tooling"
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"tooling": map[string]interface{}{
				"source":      "acme/tools",
				"sourceType":  "github",
				"sourceUrl":   "https://github.com/acme/tools.git",
				"commitHash":  "def456",
				"installedAt": "2025-02-20T15:30:00Z",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	enriched, err := EnrichExistingSkills(skillsDir)
	if err != nil {
		t.Fatalf("EnrichExistingSkills failed: %v", err)
	}

	if len(enriched) != 1 {
		t.Fatalf("enriched = %v, want 1 entry", enriched)
	}
	if enriched[0] != "tooling" {
		t.Errorf("enriched[0] = %q, want %q", enriched[0], "tooling")
	}

	// Verify config was updated
	updatedCfg := loadConfigFromFile()
	if updatedCfg == nil {
		t.Fatal("config file not found after enrich")
	}
	for _, s := range updatedCfg.Skills {
		if s.Name == "tooling" {
			if s.Version != "def456" {
				t.Errorf("Version = %q, want %q", s.Version, "def456")
			}
			if s.Installed != "2025-02-20T15:30:00Z" {
				t.Errorf("Installed = %q, want %q", s.Installed, "2025-02-20T15:30:00Z")
			}
			return
		}
	}
	t.Error("tooling not found in updated config")
}

func TestEnrichExistingSkills_NoOpWhenAlreadyPopulated(t *testing.T) {
	tmp := setTestHome(t)

	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Create config with a skill that already has version/installed
	configDir := filepath.Join(tmp, ".config", "efx-skills")
	os.MkdirAll(configDir, 0755)
	cfg := ConfigData{
		Registries: defaultRegistries(),
		Repos:      defaultRepos(),
		SkillsPath: skillsDir,
		Skills: []SkillMeta{
			{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools", Version: "abc123", Installed: "2025-01-15T10:00:00Z"},
		},
	}
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, "config.json"), cfgJSON, 0644)

	enriched, err := EnrichExistingSkills(skillsDir)
	if err != nil {
		t.Fatalf("EnrichExistingSkills failed: %v", err)
	}

	if len(enriched) != 0 {
		t.Errorf("enriched = %v, want empty (all already populated)", enriched)
	}
}

func TestEnrichExistingSkills_ReturnsEnrichedNames(t *testing.T) {
	tmp := setTestHome(t)

	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Create config with two skills: one missing version/installed, one already has them
	configDir := filepath.Join(tmp, ".config", "efx-skills")
	os.MkdirAll(configDir, 0755)
	cfg := ConfigData{
		Registries: defaultRegistries(),
		Repos:      defaultRepos(),
		SkillsPath: skillsDir,
		Skills: []SkillMeta{
			{Owner: "acme/tools", Name: "tooling", Registry: "github", URL: "https://github.com/acme/tools"},
			{Owner: "org/done", Name: "done-skill", Registry: "github", URL: "https://github.com/org/done", Version: "aaa", Installed: "2025-01-01T00:00:00Z"},
		},
	}
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, "config.json"), cfgJSON, 0644)

	// Lock file has data for "tooling" only
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"tooling": map[string]interface{}{
				"source":      "acme/tools",
				"sourceType":  "github",
				"sourceUrl":   "https://github.com/acme/tools.git",
				"commitHash":  "xyz789",
				"installedAt": "2025-03-01T12:00:00Z",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	enriched, err := EnrichExistingSkills(skillsDir)
	if err != nil {
		t.Fatalf("EnrichExistingSkills failed: %v", err)
	}

	// Only "tooling" should be enriched, not "done-skill"
	if len(enriched) != 1 {
		t.Fatalf("enriched = %v, want 1 entry", enriched)
	}
	if enriched[0] != "tooling" {
		t.Errorf("enriched[0] = %q, want %q", enriched[0], "tooling")
	}
}

// --- BackfillLegacySkills version/installed tests ---

func TestBackfillLegacySkills_SetsVersionAndInstalledFromLock(t *testing.T) {
	tmp := setTestHome(t)

	skillsDir := filepath.Join(tmp, ".agents", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Create lock file with CommitHash and InstalledAt
	lockDir := filepath.Join(tmp, ".agents")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"grepai": map[string]interface{}{
				"source":      "yoanbernabeu/grepai-skills",
				"sourceType":  "github",
				"sourceUrl":   "https://github.com/yoanbernabeu/grepai-skills.git",
				"commitHash":  "commit789",
				"installedAt": "2025-06-01T08:00:00Z",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	os.WriteFile(filepath.Join(lockDir, ".skill-lock.json"), lockJSON, 0644)

	backfilled, err := BackfillLegacySkills([]string{"grepai"}, skillsDir)
	if err != nil {
		t.Fatalf("BackfillLegacySkills failed: %v", err)
	}

	if len(backfilled) != 1 {
		t.Fatalf("backfilled = %v, want 1 entry", backfilled)
	}

	// Verify config entry has version and installed from lock
	cfg := loadConfigFromFile()
	if cfg == nil {
		t.Fatal("config file not created")
	}
	for _, s := range cfg.Skills {
		if s.Name == "grepai" {
			if s.Version != "commit789" {
				t.Errorf("Version = %q, want %q", s.Version, "commit789")
			}
			if s.Installed != "2025-06-01T08:00:00Z" {
				t.Errorf("Installed = %q, want %q", s.Installed, "2025-06-01T08:00:00Z")
			}
			return
		}
	}
	t.Error("grepai not found in config skills")
}

func TestFormatReport_Healthy(t *testing.T) {
	report := &DoctorReport{
		Issues:           []DoctorIssue{},
		MissingFromFS:    []string{},
		UntrackedOnFS:    []string{},
		BackfillCandidates: []string{},
	}

	output := FormatReport(report)

	if !strings.Contains(output, "No issues found") {
		t.Errorf("expected 'No issues found' in output, got: %s", output)
	}
}
