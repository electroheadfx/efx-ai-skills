---
phase: 4
slug: browser-integration
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-05
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (Go convention) |
| **Quick run command** | `go test ./internal/tui/ -run Browser -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/tui/ -run Browser -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | BRWS-04 | unit | `go test ./internal/tui/ -run TestOpenInBrowser -count=1` | Wave 0 | pending |
| 04-01-02 | 01 | 1 | BRWS-01 | unit | `go test ./internal/tui/ -run TestURLForAPISkill -count=1` | Wave 0 | pending |
| 04-01-03 | 01 | 1 | BRWS-05 | unit | `go test ./internal/tui/ -run TestURLForAPISkill/playbooks -count=1` | Wave 0 | pending |
| 04-01-04 | 01 | 1 | BRWS-03 | unit | `go test ./internal/tui/ -run TestRegistryBaseURL -count=1` | Wave 0 | pending |
| 04-01-05 | 01 | 2 | BRWS-01 | unit | `go test ./internal/tui/ -run TestSearchOpenKey -count=1` | Wave 0 | pending |
| 04-01-06 | 01 | 2 | BRWS-02 | unit | `go test ./internal/tui/ -run TestURLForManagedSkill -count=1` | Wave 0 | pending |
| 04-01-07 | 01 | 2 | BRWS-03 | unit | `go test ./internal/tui/ -run TestConfigOpenKey -count=1` | Wave 0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/tui/browser_test.go` — stubs for BRWS-01 through BRWS-05 (URL construction + browser dispatch)
- [ ] `internal/tui/browser.go` — openInBrowser() + URL resolution helpers

*Wave 0 creates the utility module and test file that all subsequent tasks depend on.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Browser actually opens URL | BRWS-04 | Requires desktop environment + browser | Press [o] in search view, verify browser opens correct URL |
| Playbooks URL loads correct page | BRWS-05 | Requires live network + browser | Press [o] on a playbooks.com skill, verify correct page loads |
| TUI does not freeze on [o] | BRWS-04 | Requires interactive TUI session | Press [o], verify TUI remains responsive |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
