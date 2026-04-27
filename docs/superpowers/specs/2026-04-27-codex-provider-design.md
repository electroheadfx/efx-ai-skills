# Codex Provider Design

## Goal

Add OpenAI Codex as a supported skills provider and bump the CLI version to `0.2.1`. Provider definitions should be maintained in one place so future provider additions do not require editing multiple hard-coded lists.

## Architecture

`internal/provider` will become the single source of truth for known provider definitions. It will expose the provider name and skills directory path calculation for each provider. Codex will be added as `codex` with the skills directory `~/.codex/skills`.

Existing TUI provider detection and config defaults will consume the shared provider catalog instead of defining their own provider lists. This keeps provider discovery, status display, config toggles, and lower-level provider lookup aligned.

## Components

- `internal/provider`: owns the known provider catalog and path resolution.
- `internal/tui/status.go`: detects configured providers by iterating over the shared catalog.
- `internal/config/config.go`: builds default provider configuration from the shared catalog.
- `cmd/efx-skills/main.go`: bumps the visible CLI version to `0.2.1`.
- `README.md`: documents Codex in supported providers and config examples.

## Data flow

At runtime, status/config code loads enabled provider names from `~/.config/efx-skills/config.json` when present. It then iterates over the shared provider definitions, resolves each path from `$HOME`, checks whether the skills directory exists, and counts linked skills for enabled providers.

When no config file exists, provider detection keeps the existing behavior of treating an existing provider directory as configured.

## Error handling

The change does not introduce new external calls or new error states. Existing behavior remains: unreadable or missing provider directories simply appear as unconfigured or as having no counted skills.

## Testing

Tests should verify that Codex is included in the centralized provider catalog and that config/status defaults consume the shared catalog. Existing config serialization tests should continue to pass. After implementation, run the relevant Go tests and the required build command:

```bash
go build -o ./bin/efx-skills ./cmd/efx-skills/
```
