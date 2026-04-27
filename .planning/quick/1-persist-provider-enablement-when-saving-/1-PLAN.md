# Quick Task 1: Persist provider enablement when saving skills for newly configured providers

**Date:** 2026-04-27
**Mode:** quick
**Description:** Persist provider enablement when saving skills for newly configured providers

## Goal

Fix the Codex provider management flow so saving skill toggles for an unconfigured provider creates links and persists the provider in `enabled_providers`, preventing status reloads from showing Codex as not configured.

## Tasks

### Task 1: Add regression coverage for newly configured provider saves

**Files:**
- `internal/tui/config_test.go`

**Action:**
- Add a failing test showing `applySkillChanges` for an unconfigured Codex provider creates the selected skill link and persists `codex` into config `Providers`.
- Add a second regression test showing first-time config creation preserves already detected providers while adding `codex`.

**Verify:**
- `go test ./internal/tui -run 'TestApplySkillChanges.*' -count=1` fails before implementation and passes after implementation.

**Done:**
- Tests cover the reported toggle-loss/status-loss regression and provider preservation edge case.

### Task 2: Persist provider configuration and surface save errors

**Files:**
- `internal/tui/manage.go`

**Action:**
- Change `applySkillChanges` to return an error.
- When saving an unconfigured provider, create the provider directory, load or initialize config, preserve detected providers if the config has no provider list, append the provider name, save config, and mark the local provider copy configured before applying symlink changes.
- Return filesystem, relative-path, symlink, remove, and config-save errors.
- Update the save key handler to return `errMsg` when saving fails.

**Verify:**
- `go test ./internal/tui -run 'TestApplySkillChanges.*' -count=1`
- `go test ./...`
- `go build -o ./bin/efx-skills ./cmd/efx-skills/`

**Done:**
- Codex remains configured after saving and skill toggle links are visible after reload.
