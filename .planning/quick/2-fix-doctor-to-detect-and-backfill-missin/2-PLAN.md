---
phase: quick-2
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/doctor.go
  - internal/tui/doctor_test.go
autonomous: true
requirements: [QUICK-2]
must_haves:
  truths:
    - "RunDiagnostics detects config entries missing version/installed fields"
    - "BackfillLegacySkills populates Version from lock CommitHash"
    - "BackfillLegacySkills populates Installed from lock InstalledAt"
    - "EnrichExistingSkills updates existing config entries with version/installed from lock"
  artifacts:
    - path: "internal/tui/doctor.go"
      provides: "RunDiagnostics enrichment detection, BackfillLegacySkills version/installed, EnrichExistingSkills function"
    - path: "internal/tui/doctor_test.go"
      provides: "Tests for all three changes"
  key_links:
    - from: "internal/tui/doctor.go:BackfillLegacySkills"
      to: "skill.LockEntry"
      via: "CommitHash and InstalledAt fields"
      pattern: "entry\\.CommitHash|entry\\.InstalledAt"
    - from: "internal/tui/doctor.go:EnrichExistingSkills"
      to: "config.json"
      via: "saveConfigData"
      pattern: "saveConfigData"
---

<objective>
Fix doctor to detect and backfill missing version/installed fields for pre-v0.2.0 skills.

Purpose: Skills installed before v0.2.0 have SkillMeta entries in config.json that lack `version` and `installed` fields. The doctor should detect these gaps and the backfill/enrich functions should populate them from the lock file.

Output: Updated doctor.go with three fixes, all covered by tests.
</objective>

<execution_context>
@./.claude/get-shit-done/workflows/execute-plan.md
@./.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/doctor.go (RunDiagnostics, BackfillLegacySkills, DoctorReport)
@internal/tui/config.go (SkillMeta struct with Version/Installed fields, loadConfigFromFile, saveConfigData, addSkillToConfig)
@internal/tui/doctor_test.go (existing test patterns, setTestHome helper)
@internal/skill/store.go (LockEntry with CommitHash/InstalledAt/UpdatedAt fields)

<interfaces>
From internal/tui/config.go:
```go
type SkillMeta struct {
    Owner     string `json:"owner"`
    Name      string `json:"name"`
    Registry  string `json:"registry"`
    URL       string `json:"url"`
    Version   string `json:"version,omitempty"`
    Installed string `json:"installed,omitempty"`
}

func loadConfigFromFile() *ConfigData
func saveConfigData(cfg *ConfigData) error
func addSkillToConfig(meta SkillMeta) error
```

From internal/skill/store.go:
```go
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

type LockFile struct {
    Version int                  `json:"version"`
    Skills  map[string]LockEntry `json:"skills"`
}
```

From internal/tui/doctor.go:
```go
type DoctorReport struct {
    Issues             []DoctorIssue
    MissingFromFS      []string
    UntrackedOnFS      []string
    BackfillCandidates []string
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add enrichment detection to RunDiagnostics and new EnrichExistingSkills function</name>
  <files>internal/tui/doctor.go, internal/tui/doctor_test.go</files>
  <behavior>
    - Test: RunDiagnostics reports info-severity "enrich" issue when a config skill has empty Version and Installed fields, and lock file has CommitHash/InstalledAt for it
    - Test: RunDiagnostics does NOT report enrichment issue when skill already has Version and Installed populated
    - Test: RunDiagnostics adds skill names needing enrichment to a new `EnrichCandidates []string` field on DoctorReport
    - Test: EnrichExistingSkills reads config, finds entries missing version/installed, looks up lock CommitHash->Version and InstalledAt->Installed, saves config
    - Test: EnrichExistingSkills returns list of enriched skill names
    - Test: EnrichExistingSkills is a no-op (returns empty, no error) when all skills already have version/installed
  </behavior>
  <action>
    1. Add `EnrichCandidates []string` field to DoctorReport struct.

    2. In RunDiagnostics, after the existing "Skills in config but NOT on filesystem" loop and before the "Skills on filesystem but NOT in config" loop, add a new loop over configSkills that:
       - Checks if `s.Version == "" || s.Installed == ""` AND the skill exists on filesystem (`fsSet[s.Name]`)
       - If so, checks if lock file has an entry for this skill with a non-empty CommitHash or InstalledAt
       - If enrichable, appends to `report.EnrichCandidates` and adds an info-severity issue with category "enrich", message "Config entry missing version/installed metadata", fix "Run doctor enrich to populate from lock file"
    - Sort EnrichCandidates alongside the other sorted slices.

    3. Create `EnrichExistingSkills(skillsPath string) ([]string, error)` function that:
       - Calls `loadConfigFromFile()` to get current config
       - Calls `skill.NewStore(skillsPath).ReadLockFile()` to get lock data
       - Iterates config.Skills, for each entry where Version=="" or Installed=="":
         - Looks up lockFile.Skills[s.Name]
         - If found: sets Version=entry.CommitHash (if non-empty), Installed=entry.InstalledAt (if non-empty)
         - Tracks name as enriched
       - If any enriched, calls `saveConfigData(cfg)` to persist
       - Returns enriched names

    4. In BackfillLegacySkills, when building SkillMeta from lock file (the fallback path, lines 369-383), also set:
       - `meta.Version = entry.CommitHash`
       - `meta.Installed = entry.InstalledAt`
       This ensures NEW backfill entries get version/installed from the start.

    Note: knownCorrespondences entries in BackfillLegacySkills do NOT get version/installed (no lock data available for them). This is acceptable -- they will show as enrich candidates if lock data exists later.
  </action>
  <verify>
    <automated>cd /Users/lmarques/Dev/efx-skill-management && go test ./internal/tui/ -run "Doctor|Backfill|Enrich" -v</automated>
  </verify>
  <done>
    - RunDiagnostics detects config entries missing version/installed and reports them as enrich candidates
    - EnrichExistingSkills populates version/installed from lock file data and persists to config
    - BackfillLegacySkills sets Version/Installed when creating entries from lock file
    - All new behavior covered by tests, all existing tests still pass
  </done>
</task>

</tasks>

<verification>
```bash
cd /Users/lmarques/Dev/efx-skill-management && go test ./internal/tui/ -v
cd /Users/lmarques/Dev/efx-skill-management && go vet ./internal/tui/
```
</verification>

<success_criteria>
- `go test ./internal/tui/ -run "Doctor|Backfill|Enrich" -v` -- all pass
- `go vet ./internal/tui/` -- no warnings
- RunDiagnostics reports enrichment candidates for config entries missing version/installed
- BackfillLegacySkills populates Version and Installed from lock file for new entries
- EnrichExistingSkills updates existing config entries with version/installed from lock file
</success_criteria>

<output>
After completion, create `.planning/quick/2-fix-doctor-to-detect-and-backfill-missin/2-SUMMARY.md`
</output>
