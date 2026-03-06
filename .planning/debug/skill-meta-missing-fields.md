---
status: diagnosed
trigger: "Diagnose why installed skill metadata is missing version and timestamp"
created: 2026-03-06T00:00:00Z
updated: 2026-03-06T00:00:00Z
---

## Current Focus

hypothesis: SkillMeta struct only has 4 fields (Owner, Name, Registry, URL) - no Version or Installed fields exist
test: Read struct definition and mapping function
expecting: Missing fields in both struct and mapping
next_action: N/A - root cause confirmed

## Symptoms

expected: config.json skill entries include version and installed timestamp
actual: config.json skill entries only have owner, name, registry, url
errors: none (silent omission)
reproduction: Install any skill via search view
started: always - fields were never implemented

## Eliminated

(none needed - root cause found on first hypothesis)

## Evidence

- timestamp: 2026-03-06T00:00:00Z
  checked: SkillMeta struct in config.go (line 36-41)
  found: Only 4 fields defined - Owner, Name, Registry, URL. No Version or Installed fields.
  implication: The struct literally cannot hold version or timestamp data

- timestamp: 2026-03-06T00:00:00Z
  checked: skillMetaFromAPISkill function in config.go (line 565-572)
  found: Maps s.Source->Owner, s.Name->Name, s.Registry->Registry, constructs URL. No version or timestamp mapping.
  implication: Even if fields existed, the mapping function doesn't populate them

- timestamp: 2026-03-06T00:00:00Z
  checked: api.Skill struct in client.go (line 55-63)
  found: Fields are ID, Name, Source, Description, Installs, Stars, Registry. No version field exists on the API struct either.
  implication: Neither registry API returns version info

- timestamp: 2026-03-06T00:00:00Z
  checked: SkillsShSkill struct in skillssh.go (line 15-20)
  found: Fields are ID, Name, Installs, Source. No version field.
  implication: skills.sh API does not provide version data

- timestamp: 2026-03-06T00:00:00Z
  checked: PlaybooksSkill struct in playbooks.go (line 16-27)
  found: Fields are ID, Name, Description, ShortDescription, RepoOwner, RepoName, Path, SkillSlug, Stars, IsOfficial. No version field.
  implication: playbooks.com API does not provide version data

- timestamp: 2026-03-06T00:00:00Z
  checked: LockEntry struct in store.go (line 172-181)
  found: LockEntry HAS InstalledAt and UpdatedAt timestamps, plus CommitHash (used as version proxy). These are populated in AddToLock (line 229-237).
  implication: Version/timestamp data IS captured in lock file but NOT propagated to config.json SkillMeta

- timestamp: 2026-03-06T00:00:00Z
  checked: Install flow in search.go (line 108-143)
  found: Install calls store.AddToLock (which writes timestamp+commit to lock file) AND addSkillToConfig(skillMetaFromAPISkill(s)) (which writes only 4 fields to config). The two writes are independent - no data flows from lock entry to config entry.
  implication: The data exists at install time but the two persistence paths don't share it

## Resolution

root_cause: |
  Two independent issues:

  1. **SkillMeta struct is missing fields**: The struct (config.go:36-41) only defines Owner, Name, Registry, URL.
     It has no Version or Installed (timestamp) fields, so config.json physically cannot store them.

  2. **Version data is not available from APIs**: Neither the skills.sh API (SkillsShSkill) nor the playbooks.com
     API (PlaybooksSkill) return a version field. The closest proxy is the git commit hash, which IS fetched
     during install (via FetchLatestCommitHash) and stored in the lock file's CommitHash field.

  3. **Timestamp IS available but not propagated**: The install flow in search.go calls store.AddToLock() which
     records InstalledAt/UpdatedAt timestamps in the lock file. However, skillMetaFromAPISkill() constructs
     SkillMeta before the lock entry is written, and the two data paths never intersect.

  In summary: the lock file (.skill-lock.json) already tracks commit hash + timestamps, but the config.json
  SkillMeta struct was never designed to include those fields.

fix: (not applied - diagnose only)
verification: (not applied - diagnose only)
files_changed: []
