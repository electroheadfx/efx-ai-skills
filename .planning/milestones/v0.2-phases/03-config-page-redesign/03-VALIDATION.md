---
phase: 3
slug: config-page-redesign
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-05
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (uses `go test`) |
| **Quick run command** | `go test ./internal/tui/ -run TestConfig -v` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/tui/ -run TestConfig -v`
- **After every plan wave:** Run `go test ./... -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | CONF-01 | unit | `go test ./internal/tui/ -run TestRegistryDisplayName -v` | No - W0 | pending |
| 03-01-02 | 01 | 1 | CONF-01 | unit | `go test ./internal/tui/ -run TestConfigViewRegistryFriendlyNames -v` | No - W0 | pending |
| 03-01-03 | 01 | 1 | CONF-02 | unit | `go test ./internal/tui/ -run TestConfigViewRepoTwoColumn -v` | No - W0 | pending |
| 03-01-04 | 01 | 1 | CONF-03 | unit | `go test ./internal/tui/ -run TestConfigViewProvidersLabel -v` | No - W0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `TestRegistryDisplayName` in `config_test.go` — stubs for CONF-01 mapping logic
- [ ] `TestConfigViewRegistryFriendlyNames` in `config_test.go` — covers CONF-01 View output
- [ ] `TestConfigViewRepoTwoColumn` in `config_test.go` — covers CONF-02 View output
- [ ] `TestConfigViewProvidersLabel` in `config_test.go` — covers CONF-03 View output

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Bold URL rendering visible in terminal | CONF-01 | ANSI bold codes present in string but visual verification needed | Run app, navigate to config, verify URLs appear bold |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
