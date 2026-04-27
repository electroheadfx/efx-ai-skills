# Quick Task 1 Summary: Persist provider enablement when saving skills for newly configured providers

**Date:** 2026-04-27
**Status:** Complete

## What changed

- Added regression coverage in `internal/tui/config_test.go` for saving skill toggles on an unconfigured Codex provider.
- Updated `applySkillChanges` in `internal/tui/manage.go` to persist the provider name into `enabled_providers` when a provider is configured through the manage view.
- Preserved already detected providers when creating a config file from the manage save flow.
- Surfaced save/link errors through the TUI instead of ignoring them.
- Rebuilt `bin/efx-skills` with the required project build command.

## Verification

- `go test ./internal/tui -run 'TestApplySkillChanges.*' -count=1`
- `go test ./...`
- `go build -o ./bin/efx-skills ./cmd/efx-skills/`
- `./bin/efx-skills --version` returned `efx-skills version 0.2.1`
- User confirmed the Codex provider save flow now works.

## Result

Saving selected skills for Codex now creates the links and keeps Codex configured after reload, so toggles no longer disappear because of missing provider persistence.
