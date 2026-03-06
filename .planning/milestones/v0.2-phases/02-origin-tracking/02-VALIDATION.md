---
phase: 2
slug: origin-tracking
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-05
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib), go test |
| **Config file** | none — Go convention uses `*_test.go` files |
| **Quick run command** | `go test ./internal/tui/ -run TestConfig -v -count=1` |
| **Full suite command** | `go test ./... -v -count=1 && go build ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/tui/ -v -count=1`
- **After every plan wave:** Run `go test ./... -v -count=1 && go build ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 0 | ORIG-01, ORIG-05, ORIG-06 | unit | `go test ./internal/tui/ -run TestSearchOriginDisplay -v -count=1` | Wave 0 | pending |
| 02-01-02 | 01 | 1 | ORIG-01 | unit | `go test ./internal/tui/ -run TestSearchOriginDisplay -v -count=1` | Wave 0 | pending |
| 02-01-03 | 01 | 1 | ORIG-05 | unit | `go test ./internal/tui/ -run TestInstallWritesSkillMeta -v -count=1` | Wave 0 | pending |
| 02-01-04 | 01 | 1 | ORIG-05 | unit | `go test ./internal/tui/ -run TestInstallSkipsDuplicate -v -count=1` | Wave 0 | pending |
| 02-01-05 | 01 | 1 | ORIG-06 | unit | `go test ./internal/tui/ -run TestRemoveSkillFromConfig -v -count=1` | Wave 0 | pending |
| 02-01-06 | 01 | 1 | ORIG-06 | unit | `go test ./internal/tui/ -run TestRemoveNonexistentSkill -v -count=1` | Wave 0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/tui/config_test.go` — extend with addSkillToConfig/removeSkillFromConfig tests (ORIG-05, ORIG-06)
- [ ] Tests for ORIG-01 — test data availability rather than rendered Bubble Tea view output

*Existing test infrastructure covers base config operations.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Search results display origin visually | ORIG-01 | View rendering is string formatting in Bubble Tea | Run app, search for a skill, verify owner/repo appears next to name |

---

## Validation Sign-Off

- [ ] All tasks have automated verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
