# Codebase Concerns

**Analysis Date:** 2026-03-05

## Tech Debt

**Dead Code: `internal/config` package is never imported:**
- Issue: The entire `internal/config/config.go` package (156 lines) defines `Config`, `Registry`, `ProviderConfig` types and `Load`/`Save`/`DefaultConfig` functions, but no other package imports it. The TUI reimplements its own config loading in `internal/tui/config.go` (`loadConfigFromFile`) and `internal/tui/status.go` (`detectProviders`).
- Files: `internal/config/config.go`
- Impact: Confusion about which config path is authoritative. Two separate config schemas exist -- the `internal/config` package uses `map[string]ProviderConfig` while `internal/tui/config.go` uses `[]string` for `enabled_providers`. Adding features requires understanding which system is actually used.
- Fix approach: Either delete `internal/config/config.go` entirely, or refactor all config loading in `internal/tui/` to use the centralized `config.Load()`/`config.Save()` API. The latter is preferred since `internal/config` has a cleaner design.

**Dead Code: `internal/provider` package is never imported:**
- Issue: The entire `internal/provider/provider.go` package (139 lines) defines a `Provider` struct and detection logic, but is never imported. The TUI redefines its own `Provider` struct in `internal/tui/status.go` with a slightly different shape (adds `Path` and `Synced` fields, drops `SkillsPath`). Provider detection is reimplemented in `internal/tui/status.go:detectProviders()`.
- Files: `internal/provider/provider.go`, `internal/tui/status.go`
- Impact: Two competing `Provider` type definitions. Changes to provider logic must be traced carefully to the actually-used implementation.
- Fix approach: Delete `internal/provider/provider.go` or refactor to use it as the single source of truth, extending its struct to include the fields the TUI needs.

**Duplicate type definitions:**
- Issue: `Registry` struct is defined in both `internal/config/config.go:17` and `internal/tui/config.go:15`. `Provider` struct is defined in both `internal/provider/provider.go:9` and `internal/tui/status.go:14`. These are independent types that don't share code.
- Files: `internal/config/config.go`, `internal/tui/config.go`, `internal/provider/provider.go`, `internal/tui/status.go`
- Impact: Any change to the domain model requires updating multiple locations. Risk of divergence between the "real" and "dead" types.
- Fix approach: Consolidate types into the packages that are actually imported. Remove the dead package definitions.

**Duplicate GitHub content fetching logic:**
- Issue: SKILL.md fetching from GitHub is implemented three times with slightly different URL patterns:
  1. `internal/api/client.go:FetchSkillContent` (3 URL patterns, never called)
  2. `internal/tui/preview.go:fetchSkillContent` (8 URL patterns, used by preview)
  3. `internal/skill/store.go:installDirect` (2 URL patterns, used by install)
- Files: `internal/api/client.go:94-122`, `internal/tui/preview.go:196-250`, `internal/skill/store.go:57-106`
- Impact: Inconsistent URL fallback behavior. The preview tries 8 paths while install tries 2. A fix to URL resolution must be applied in three places.
- Fix approach: Create a single `FetchSkillContent` function in `internal/api/client.go` that all callers use, with the most comprehensive set of URL patterns.

**Hardcoded version strings scattered across views:**
- Issue: Version is displayed as a hardcoded string in each TUI view rather than using the `version` variable from `cmd/efx-skills/main.go`. Additionally, the versions are inconsistent: `status.go` shows "v0.1.4" while `search.go` and `preview.go` show "v0.1.3". The Makefile has `VERSION=0.1.2` and `main.go` has `version = "0.1.3"`.
- Files: `internal/tui/status.go:206`, `internal/tui/search.go:246`, `internal/tui/preview.go:172`, `cmd/efx-skills/main.go:11`, `Makefile:4`
- Impact: Users see different version numbers depending on which view they are on. Version bumps require editing 5 files.
- Fix approach: Pass the version string from `main.go` through to the TUI model, or use a shared `version` package-level variable. Update Makefile VERSION to match.

**Unimplemented CLI commands (stubs):**
- Issue: `RunInstall` and `RunSync` in `internal/tui/app.go` are stubs with TODO comments. They print placeholder messages and return nil.
- Files: `internal/tui/app.go:193-197` (`RunInstall`), `internal/tui/app.go:233-237` (`RunSync`)
- Impact: Users running `efx-skills install <skill>` or `efx-skills sync` see a misleading success message but nothing happens. The TUI-based install (from search view) works, but the CLI command does not.
- Fix approach: Implement the CLI install/sync commands or remove them from the command registration in `main.go` and document them as TUI-only features.

**Redundant `max` function:**
- Issue: A custom `max(a, b int) int` function is defined in `internal/tui/preview.go:252-257`. Go 1.21+ (this project uses Go 1.22) has a built-in `max` function.
- Files: `internal/tui/preview.go:252-257`
- Impact: Minor. The custom function shadows the builtin and is never actually called.
- Fix approach: Delete the function. If it were called, the builtin would work identically.

## Known Bugs

**`defer resp.Body.Close()` inside loops leaks file descriptors:**
- Symptoms: When fetching SKILL.md from GitHub, the code tries multiple URLs in a loop. Each `defer resp.Body.Close()` only executes when the enclosing function returns, not when the loop iteration ends. Non-200 responses accumulate open connections until the function exits.
- Files: `internal/api/client.go:110`, `internal/tui/preview.go:237`, `internal/skill/store.go:89`
- Trigger: Searching for a skill where the first several URL patterns return 404 before finding the content. Each 404 response body stays open.
- Workaround: In practice, the loop iterates at most 8 times, so the leak is bounded. But it is a correctness issue.
- Fix: Replace `defer resp.Body.Close()` inside loops with immediate `resp.Body.Close()` calls, or extract the loop body into a separate function.

**`installDirect` leaks response body on retry path:**
- Symptoms: In `internal/skill/store.go:76-94`, the first `http.Get` creates `resp` with a deferred close. If it gets a non-200, a second `http.Get` reassigns `resp`, but the first response body's defer still references the old variable binding.
- Files: `internal/skill/store.go:76-94`
- Trigger: First URL returns non-200, second URL is tried.
- Workaround: None needed in practice since the old response body will be closed when the function returns, but the first body stays open longer than necessary.

**`os.RemoveAll` used instead of `os.Remove` for symlink removal:**
- Symptoms: When unlinking a skill from a provider in the manage view, `os.RemoveAll(linkPath)` is used. If the symlink target is a real directory (not a symlink) due to a manual copy instead of a symlink, `RemoveAll` would recursively delete all contents.
- Files: `internal/tui/manage.go:403`
- Trigger: A skill was manually copied (not symlinked) to a provider directory, then user deselects it in the manage view.
- Workaround: The normal flow creates symlinks, so `RemoveAll` on a symlink just removes the symlink. But this is a safety concern.
- Fix: Use `os.Remove` instead, which only removes a single file/symlink. If removal of a real directory is needed, require explicit confirmation.

**Silently ignored errors in `applySkillChanges`:**
- Symptoms: Both `os.Symlink` and `os.RemoveAll` in `applySkillChanges` ignore their error return values. The `filepath.Rel` error is also silently discarded with `relPath, _ := filepath.Rel(...)`.
- Files: `internal/tui/manage.go:387-406`
- Trigger: Any filesystem error during skill linking/unlinking (permission denied, disk full, etc.).
- Workaround: None. User gets no feedback that the operation failed.
- Fix: Collect errors and report them back via the TUI message system.

**Lock file error silently ignored during install:**
- Symptoms: `_ = store.AddToLock(s.Name, s.Source)` in `internal/tui/search.go:117` discards the error. If the lock file cannot be written, the skill appears installed but is not tracked.
- Files: `internal/tui/search.go:117`
- Trigger: Permission issues or disk full when writing `~/.agents/.skill-lock.json`.
- Fix: Propagate the error and show a warning to the user.

## Security Considerations

**Command injection via `npx` with user-supplied source:**
- Risk: `internal/skill/store.go:42-54` passes `source` directly as an argument to `exec.Command("npx", ...)`. While Go's `exec.Command` is not vulnerable to shell injection (it doesn't use a shell), the `source` string comes from API responses and could contain unexpected values that `npx skills add` might interpret maliciously.
- Files: `internal/skill/store.go:42-54`
- Current mitigation: Go's `exec.Command` passes arguments directly to the process, avoiding shell expansion. The `npx skills` tool itself would need to have vulnerabilities for this to be exploitable.
- Recommendations: Validate `source` format (must match `owner/repo` pattern) before passing to `npx`. Add allowlist validation for characters.

**No input sanitization on GitHub URLs:**
- Risk: User-provided skill names and sources are interpolated directly into GitHub URLs via `fmt.Sprintf`. While unlikely to cause SSRF (the URLs are constructed to point to `raw.githubusercontent.com`), a malicious registry could return source strings containing path traversal characters.
- Files: `internal/api/client.go:98-100`, `internal/tui/preview.go:220-228`, `internal/skill/store.go:75-84`
- Current mitigation: URLs are only used for GET requests to `raw.githubusercontent.com`.
- Recommendations: Validate that owner/repo/path components contain only alphanumeric characters, hyphens, and underscores.

**Config file written with world-readable permissions:**
- Risk: `os.WriteFile(configFile, jsonData, 0644)` makes the config readable by all users on the system. While the config does not currently contain secrets, it could in the future if API keys are added.
- Files: `internal/tui/config.go:252`, `internal/config/config.go:94`
- Current mitigation: No secrets are stored in config currently.
- Recommendations: Use `0600` permissions for the config file as a defensive measure.

## Performance Bottlenecks

**Sequential HTTP requests for skill content resolution:**
- Problem: `fetchSkillContent` in `internal/tui/preview.go` tries up to 8 different GitHub URLs sequentially, each with a 10-second timeout. In the worst case, this can take up to 80 seconds before returning an error.
- Files: `internal/tui/preview.go:196-250`
- Cause: Sequential fallback through URL patterns with no parallelism or early termination.
- Improvement path: Use concurrent HTTP requests with `context.Context` and cancel remaining requests once one succeeds. Alternatively, reduce the timeout for fallback attempts (e.g., 3 seconds each) or cache successful URL patterns.

**`detectProviders` called repeatedly without caching:**
- Problem: `detectProviders()` in `internal/tui/status.go` reads the config file from disk and scans 6 provider directories on every call. It is called at least 4 times: status init, config init, search install flow, and `RunList`.
- Files: `internal/tui/status.go:80-148`
- Cause: No caching of provider state. Each call re-reads the config JSON and re-scans directories.
- Improvement path: Cache the result within a session and invalidate only when the user explicitly changes config or after an install/uninstall operation.

**No HTTP response caching for API calls:**
- Problem: Search and trending API calls to `skills.sh` and `playbooks.com` are made fresh on every search, even for repeated queries within the same session.
- Files: `internal/api/client.go`, `internal/api/skillssh.go`, `internal/api/playbooks.go`
- Cause: No in-memory cache for API responses.
- Improvement path: Add a simple TTL-based in-memory cache for search results (e.g., 5-minute TTL).

## Fragile Areas

**TUI package is a monolithic dependency hub:**
- Files: `internal/tui/app.go`, `internal/tui/status.go`, `internal/tui/search.go`, `internal/tui/manage.go`, `internal/tui/config.go`, `internal/tui/preview.go`
- Why fragile: The `tui` package contains business logic (provider detection, skill installation, config persistence), presentation logic (views), and data types (Provider, Registry) all in one package. The `search.go` directly calls `skill.NewStore()` and `detectProviders()`, tightly coupling search UI to filesystem operations.
- Safe modification: When modifying one view file, test all views since they share types and helper functions. The `detectProviders` function in `status.go` is used by `config.go`, `search.go`, and `app.go`.
- Test coverage: Zero -- no test files exist anywhere in the project.

**Config persistence has two incompatible schemas:**
- Files: `internal/config/config.go` (unused schema), `internal/tui/config.go` (active schema)
- Why fragile: The active config schema (`ConfigData` in `tui/config.go`) stores providers as `[]string` (enabled provider names), while the dead schema (`Config` in `config/config.go`) stores them as `map[string]ProviderConfig`. If someone references the wrong package, config loading silently produces wrong data.
- Safe modification: Only modify `internal/tui/config.go` for config changes. Ignore `internal/config/config.go` unless consolidating.
- Test coverage: None.

**Provider path definitions duplicated in 3 locations:**
- Files: `internal/provider/provider.go:17-27`, `internal/tui/status.go:98-108`, `internal/config/config.go:43-50`
- Why fragile: Adding a new AI provider requires updating provider paths in all three locations (though only `internal/tui/status.go` is actively used). Missing an update would cause inconsistent behavior.
- Safe modification: Only update `internal/tui/status.go:providerDefs` for the active codebase. But ideally consolidate into a single definition.

## Scaling Limits

**Linear skill scanning on every operation:**
- Current capacity: Works fine for dozens of skills.
- Limit: With hundreds or thousands of installed skills, `os.ReadDir` scanning in `loadSkillsForProvider`, `detectProviders`, and `ListInstalled` will become noticeably slow.
- Scaling path: Use the lock file (`~/.agents/.skill-lock.json`) as a skill index instead of scanning directories. The lock file already exists but is only used for install tracking, not as a read cache.

**Single-threaded search across registries:**
- Current capacity: Two registries (skills.sh, playbooks.com).
- Limit: Adding more registries makes `SearchAll` linearly slower since each is queried sequentially.
- Scaling path: Query registries concurrently with `sync.WaitGroup` or `errgroup`.

## Dependencies at Risk

**None critical.** All dependencies are well-maintained Charmbracelet ecosystem packages (bubbletea, bubbles, lipgloss, glamour) and spf13/cobra. These are among the most popular Go TUI and CLI libraries.

## Missing Critical Features

**No update/upgrade mechanism for installed skills:**
- Problem: Once a skill is installed, there is no way to check for updates or upgrade it. The `sync` command is stubbed out.
- Blocks: Users must manually re-install skills to get updates.

**No uninstall from central storage:**
- Problem: The manage view can unlink skills from providers but cannot remove them from `~/.agents/skills/`. There is no `uninstall` command.
- Blocks: Disk space reclamation and clean removal of unwanted skills.

**Custom repos are not searchable:**
- Problem: Custom GitHub repos added via config are stored but never actually queried during search. `searchSkills()` in `internal/tui/search.go:352-354` only calls `api.SearchAll(query, 50)` which searches `skills.sh` and `playbooks.com`. The repos list is decorative.
- Files: `internal/tui/search.go:352-354`, `internal/tui/config.go:74-80`
- Blocks: The entire custom repos feature is non-functional beyond storing repo names.

## Test Coverage Gaps

**No tests exist in the entire codebase:**
- What's not tested: Everything. No `*_test.go` files exist anywhere in the project.
- Files: All 14 `.go` files have zero test coverage.
- Risk: Any refactoring (especially consolidating the duplicate types and dead code) could introduce regressions with no safety net. The business logic in `applySkillChanges`, `searchSkills`, `detectProviders`, and `installDirect` is particularly risky to modify without tests.
- Priority: High. At minimum, unit tests are needed for:
  - `internal/api/` - API response parsing (mock HTTP responses)
  - `internal/skill/store.go` - Lock file management, install logic
  - `internal/tui/status.go:detectProviders` - Provider detection logic
  - `internal/tui/manage.go:extractGroup` - Group name extraction
  - `internal/tui/manage.go:applySkillChanges` - Symlink management

---

*Concerns audit: 2026-03-05*
