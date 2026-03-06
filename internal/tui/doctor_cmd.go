package tui

import (
	"fmt"
	"strings"
)

// RunDoctor orchestrates the full doctor diagnostic flow.
// It loads config, runs diagnostics, prints the report, and optionally
// backfills legacy skills when the fix flag is set.
func RunDoctor(fix bool) error {
	// Load config (nil is fine -- use empty defaults)
	cfg := loadConfigFromFile()
	var skills []SkillMeta
	if cfg != nil {
		skills = cfg.Skills
	}

	// Determine skills path
	skillsPath := getSkillsPath()

	// Run diagnostics
	report, err := RunDiagnostics(skillsPath, skills)
	if err != nil {
		return fmt.Errorf("diagnostics failed: %w", err)
	}

	// Print formatted report
	fmt.Print(FormatReport(report))
	fmt.Println()

	// If --fix and there are backfill candidates, run backfill
	if fix && len(report.BackfillCandidates) > 0 {
		backfilled, err := BackfillLegacySkills(report.BackfillCandidates, skillsPath)
		if err != nil {
			return fmt.Errorf("backfill failed: %w", err)
		}
		if len(backfilled) > 0 {
			fmt.Printf("Backfilled %d skills: %s\n", len(backfilled), strings.Join(backfilled, ", "))
		}
	}

	// If --fix and there are enrich candidates, enrich existing entries
	if fix && len(report.EnrichCandidates) > 0 {
		enriched, err := EnrichExistingSkills(skillsPath)
		if err != nil {
			return fmt.Errorf("enrich failed: %w", err)
		}
		if len(enriched) > 0 {
			fmt.Printf("Enriched %d skills: %s\n", len(enriched), strings.Join(enriched, ", "))
		}
	}

	// Count error-severity issues
	errorCount := 0
	for _, issue := range report.Issues {
		if issue.Severity == "error" {
			errorCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("doctor found %d issue(s)", errorCount)
	}

	return nil
}
