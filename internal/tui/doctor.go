package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/lmarques/efx-skills/internal/skill"
)

// DoctorIssue represents a single diagnostic finding.
type DoctorIssue struct {
	Severity  string // "error", "warning", "info"
	Category  string // "missing", "untracked", "backfill"
	SkillName string
	Message   string
	Fix       string
}

// DoctorReport holds the complete diagnostic result.
type DoctorReport struct {
	Issues             []DoctorIssue
	MissingFromFS      []string
	UntrackedOnFS      []string
	BackfillCandidates []string
	EnrichCandidates   []string
}

// knownCorrespondences maps skill directory names to their known metadata for
// pre-v0.2.0 skills that were installed before origin tracking existed. This is
// the fallback when the lock file has no entry for a skill.
var knownCorrespondences = map[string]SkillMeta{
	// anthropics/skills
	"skill-creator": {
		Owner:    "anthropics/skills",
		Name:     "skill-creator",
		Registry: "github",
		URL:      "https://github.com/anthropics/skills",
	},
	// davis7dotsh/better-context
	"btca-cli": {
		Owner:    "davis7dotsh/better-context",
		Name:     "btca-cli",
		Registry: "github",
		URL:      "https://github.com/davis7dotsh/better-context/tree/main/skills/btca-cli",
	},
	// manaflow-ai/cmux
	"cmux": {
		Owner:    "manaflow-ai/cmux",
		Name:     "cmux",
		Registry: "github",
		URL:      "https://github.com/manaflow-ai/cmux/tree/main/skills/cmux",
	},
	"cmux-browser": {
		Owner:    "manaflow-ai/cmux",
		Name:     "cmux-browser",
		Registry: "github",
		URL:      "https://github.com/manaflow-ai/cmux/tree/main/skills/cmux-browser",
	},
	// psycho-baller/ai-agents-config
	"cmux-and-worktrees": {
		Owner:    "psycho-baller/ai-agents-config",
		Name:     "cmux-and-worktrees",
		Registry: "github",
		URL:      "https://github.com/psycho-baller/ai-agents-config/tree/main/skills",
	},
	// github/awesome-copilot
	"excalidraw-diagram-generator": {
		Owner:    "github/awesome-copilot",
		Name:     "excalidraw-diagram-generator",
		Registry: "github",
		URL:      "https://github.com/github/awesome-copilot/tree/main/skills/excalidraw-diagram-generator",
	},
	// vercel-labs/skills
	"find-skills": {
		Owner:    "vercel-labs/skills",
		Name:     "find-skills",
		Registry: "vercel",
		URL:      "https://github.com/vercel-labs/skills/blob/main/skills/find-skills",
	},
	// vercel-labs/agent-browser
	"agent-browser": {
		Owner:    "vercel-labs/agent-browser",
		Name:     "agent-browser",
		Registry: "github",
		URL:      "https://github.com/vercel-labs/agent-browser",
	},
	// yoanbernabeu/grepai-skills
	"grepai-chunking": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-chunking",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/indexing/grepai-chunking",
	},
	"grepai-config-reference": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-config-reference",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/configuration/grepai-config-reference",
	},
	"grepai-embeddings-lmstudio": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-embeddings-lmstudio",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/embeddings/grepai-embeddings-lmstudio",
	},
	"grepai-embeddings-ollama": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-embeddings-ollama",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/embeddings/grepai-embeddings-ollama",
	},
	"grepai-embeddings-openai": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-embeddings-openai",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/embeddings/grepai-embeddings-openai",
	},
	"grepai-ignore-patterns": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-ignore-patterns",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/configuration/grepai-ignore-patterns",
	},
	"grepai-init": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-init",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/configuration/grepai-init",
	},
	"grepai-installation": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-installation",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/getting-started/grepai-installation",
	},
	"grepai-languages": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-languages",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/advanced/grepai-languages",
	},
	"grepai-mcp-claude": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-mcp-claude",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/integration/grepai-mcp-claude",
	},
	"grepai-mcp-cursor": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-mcp-cursor",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/integration/grepai-mcp-cursor",
	},
	"grepai-mcp-tools": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-mcp-tools",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/integration/grepai-mcp-tools",
	},
	"grepai-ollama-setup": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-ollama-setup",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/getting-started/grepai-ollama-setup",
	},
	"grepai-quickstart": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-quickstart",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/getting-started/grepai-quickstart",
	},
	"grepai-search-advanced": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-search-advanced",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/search/grepai-search-advanced",
	},
	"grepai-search-basics": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-search-basics",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/search/grepai-search-basics",
	},
	"grepai-search-boosting": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-search-boosting",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/search/grepai-search-boosting",
	},
	"grepai-search-tips": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-search-tips",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/search/grepai-search-tips",
	},
	"grepai-storage-gob": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-storage-gob",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/storage/grepai-storage-gob",
	},
	"grepai-storage-qdrant": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-storage-qdrant",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/storage/grepai-storage-qdrant",
	},
	"grepai-storage-postgres": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-storage-postgres",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/storage/grepai-storage-postgres",
	},
	"grepai-trace-callers": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-trace-callers",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/trace/grepai-trace-callers",
	},
	"grepai-trace-callees": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-trace-callees",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/trace/grepai-trace-callees",
	},
	"grepai-trace-graph": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-trace-graph",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/trace/grepai-trace-graph",
	},
	"grepai-troubleshooting": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-troubleshooting",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/advanced/grepai-troubleshooting",
	},
	"grepai-watch-daemon": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-watch-daemon",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/indexing/grepai-watch-daemon",
	},
	"grepai-workspaces": {
		Owner:    "yoanbernabeu/grepai-skills",
		Name:     "grepai-workspaces",
		Registry: "github",
		URL:      "https://github.com/yoanbernabeu/grepai-skills/tree/main/skills/advanced/grepai-workspaces",
	},
}

// RunDiagnostics compares the skills tracked in config against the skills
// installed on the filesystem, producing a structured report of inconsistencies.
// It also checks the lock file and known correspondences to identify backfill
// candidates among untracked skills.
func RunDiagnostics(skillsPath string, configSkills []SkillMeta) (*DoctorReport, error) {
	report := &DoctorReport{}

	// Build config set (by Name)
	configSet := make(map[string]bool, len(configSkills))
	for _, s := range configSkills {
		configSet[s.Name] = true
	}

	// Read filesystem
	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		return nil, fmt.Errorf("reading skills directory: %w", err)
	}

	fsSet := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			fsSet[e.Name()] = true
		}
	}

	// Skills in config but NOT on filesystem -> MissingFromFS
	for _, s := range configSkills {
		if !fsSet[s.Name] {
			report.MissingFromFS = append(report.MissingFromFS, s.Name)
			report.Issues = append(report.Issues, DoctorIssue{
				Severity:  "error",
				Category:  "missing",
				SkillName: s.Name,
				Message:   "Tracked in config but not found on filesystem",
				Fix:       fmt.Sprintf("Run 'efx install %s' to reinstall", s.Name),
			})
		}
	}

	// Read the lock file for backfill/enrich metadata
	store := skill.NewStore(skillsPath)
	lockFile, _ := store.ReadLockFile()

	// Config entries missing version/installed metadata -> EnrichCandidates
	// Only flag if the lock file has data for the specific missing field(s).
	for _, s := range configSkills {
		if (s.Version == "" || s.Installed == "") && fsSet[s.Name] {
			if lockFile != nil {
				if entry, ok := lockFile.Skills[s.Name]; ok {
					canEnrichVersion := s.Version == "" && entry.CommitHash != ""
					canEnrichInstalled := s.Installed == "" && entry.InstalledAt != ""
					if canEnrichVersion || canEnrichInstalled {
						report.EnrichCandidates = append(report.EnrichCandidates, s.Name)
						report.Issues = append(report.Issues, DoctorIssue{
							Severity:  "info",
							Category:  "enrich",
							SkillName: s.Name,
							Message:   "Config entry missing version/installed metadata",
							Fix:       "Run 'efx-skills doctor --fix' to populate from lock file",
						})
					}
				}
			}
		}
	}

	// Skills on filesystem but NOT in config -> UntrackedOnFS
	for name := range fsSet {
		if !configSet[name] {
			report.UntrackedOnFS = append(report.UntrackedOnFS, name)

			// Check if this is a backfill candidate
			isBackfillCandidate := false

			// Primary: check lock file
			if lockFile != nil {
				if _, ok := lockFile.Skills[name]; ok {
					isBackfillCandidate = true
				}
			}

			// Fallback: check known correspondences
			if !isBackfillCandidate {
				if _, ok := knownCorrespondences[name]; ok {
					isBackfillCandidate = true
				}
			}

			if isBackfillCandidate {
				report.BackfillCandidates = append(report.BackfillCandidates, name)
				report.Issues = append(report.Issues, DoctorIssue{
					Severity:  "info",
					Category:  "backfill",
					SkillName: name,
					Message:   "Can be auto-backfilled from lock file",
					Fix:       "Run doctor backfill to add metadata",
				})
			} else {
				report.Issues = append(report.Issues, DoctorIssue{
					Severity:  "warning",
					Category:  "untracked",
					SkillName: name,
					Message:   "Found on filesystem but not tracked in config",
					Fix:       "Run doctor backfill or add manually",
				})
			}
		}
	}

	// Sort slices for deterministic output
	sort.Strings(report.MissingFromFS)
	sort.Strings(report.UntrackedOnFS)
	sort.Strings(report.BackfillCandidates)
	sort.Strings(report.EnrichCandidates)

	return report, nil
}

// BackfillLegacySkills adds metadata to config for untracked filesystem skills.
// It first checks the lock file for source information, then falls back to
// known correspondences. Skills not found in either source are skipped.
// Returns the list of successfully backfilled skill names.
func BackfillLegacySkills(candidates []string, skillsPath string) ([]string, error) {
	store := skill.NewStore(skillsPath)
	lockFile, _ := store.ReadLockFile()

	var backfilled []string

	for _, name := range candidates {
		var meta SkillMeta
		found := false

		// Primary: known correspondences (has correct skill-specific URLs)
		if km, ok := knownCorrespondences[name]; ok {
			meta = km
			found = true
		}

		// Fallback: lock file
		if !found && lockFile != nil {
			if entry, ok := lockFile.Skills[name]; ok {
				registry := entry.SourceType
				if registry == "" {
					registry = "github"
				}
				meta = SkillMeta{
					Owner:     entry.Source,
					Name:      name,
					Registry:  registry,
					URL:       entry.SourceURL,
					Version:   entry.CommitHash,
					Installed: entry.InstalledAt,
				}
				found = true
			}
		}

		if !found {
			// Skip unknown skills gracefully
			continue
		}

		if err := addSkillToConfig(meta); err != nil {
			// Continue on individual failures
			continue
		}
		backfilled = append(backfilled, name)
	}

	return backfilled, nil
}

// EnrichExistingSkills updates config entries that are missing version/installed
// fields by looking up the corresponding lock file entry. Returns the list of
// enriched skill names.
func EnrichExistingSkills(skillsPath string) ([]string, error) {
	cfg := loadConfigFromFile()
	if cfg == nil {
		return nil, nil
	}

	store := skill.NewStore(skillsPath)
	lockFile, err := store.ReadLockFile()
	if err != nil {
		return nil, fmt.Errorf("reading lock file: %w", err)
	}

	var enriched []string
	for i, s := range cfg.Skills {
		if s.Version != "" && s.Installed != "" {
			continue // already populated
		}

		entry, ok := lockFile.Skills[s.Name]
		if !ok {
			continue // no lock data for this skill
		}

		changed := false
		if s.Version == "" && entry.CommitHash != "" {
			cfg.Skills[i].Version = entry.CommitHash
			changed = true
		}
		if s.Installed == "" && entry.InstalledAt != "" {
			cfg.Skills[i].Installed = entry.InstalledAt
			changed = true
		}
		if changed {
			enriched = append(enriched, s.Name)
		}
	}

	if len(enriched) > 0 {
		if err := saveConfigData(cfg); err != nil {
			return nil, fmt.Errorf("saving config: %w", err)
		}
	}

	return enriched, nil
}

// FormatReport formats a DoctorReport as a human-readable string with severity
// icons, grouped by severity (errors first, then warnings, then info), with a
// summary line at the end.
func FormatReport(report *DoctorReport) string {
	if len(report.Issues) == 0 {
		return "No issues found - all skills healthy"
	}

	var b strings.Builder

	// Group issues by severity
	var errors, warnings, infos []DoctorIssue
	for _, issue := range report.Issues {
		switch issue.Severity {
		case "error":
			errors = append(errors, issue)
		case "warning":
			warnings = append(warnings, issue)
		case "info":
			infos = append(infos, issue)
		}
	}

	// Format each group
	formatIssues := func(issues []DoctorIssue, icon string) {
		for _, issue := range issues {
			b.WriteString(fmt.Sprintf("  %s %s: %s\n", icon, issue.SkillName, issue.Message))
			b.WriteString(fmt.Sprintf("    Fix: %s\n", issue.Fix))
		}
	}

	if len(errors) > 0 {
		formatIssues(errors, "!")
	}
	if len(warnings) > 0 {
		formatIssues(warnings, "?")
	}
	if len(infos) > 0 {
		formatIssues(infos, "i")
	}

	// Summary line
	b.WriteString(fmt.Sprintf("\n%d issues found (%d errors, %d warnings, %d info)",
		len(report.Issues), len(errors), len(warnings), len(infos)))

	return b.String()
}
