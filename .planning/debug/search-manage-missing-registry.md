---
status: diagnosed
trigger: "Search view doesn't show registry source column; manage view doesn't show origin/registry on skill folders"
created: 2026-03-06T00:00:00Z
updated: 2026-03-06T00:00:00Z
---

## Current Focus

hypothesis: Both views lack Registry column/display -- search.go renders Name+Source+Popularity but no Registry; manage.go renders checkbox+name+status but no origin/registry
test: Read rendering code in View() methods for both files
expecting: Confirm absence of Registry field usage in rendering
next_action: N/A -- diagnosis complete

## Symptoms

expected: Search results display a visible column showing registry source (e.g., playbooks, vercel, github). Manage view shows origin/registry on skill folders.
actual: Search view shows only Name, Source (owner/repo), and popularity count. Manage view shows only checkbox, skill name, and status (add/remove).
errors: None -- functional gap, not runtime error
reproduction: Open TUI, search for any skill, observe columns. Navigate to provider manage view, observe skill items.
started: Always been this way -- columns were never added

## Eliminated

(none -- single hypothesis confirmed)

## Evidence

- timestamp: 2026-03-06T00:00:00Z
  checked: api.Skill struct in internal/api/client.go (lines 55-63)
  found: Struct has Registry field (string, json:"registry"). Populated as "skills.sh" or "playbooks.com" by the respective search functions.
  implication: The data IS available on every search result -- it's just not rendered.

- timestamp: 2026-03-06T00:00:00Z
  checked: search.go View() method (lines 251-370), specifically result rendering loop (lines 306-335)
  found: Three columns rendered -- Name (35% width), Source (55% width), Popularity (6 chars). Line 324 formats: `nameFmt+" "+sourceFmt+" %6s"` using skill.Name, skill.Source, and popularity. skill.Registry is never referenced in View().
  implication: Registry column is simply absent from the search results table layout.

- timestamp: 2026-03-06T00:00:00Z
  checked: manage.go View() method (lines 564-717), specifically skill item rendering (lines 646-681)
  found: Skill items render: indent + checkbox + displayName + status. Line 668: `fmt.Sprintf("    %s %s%s", checkbox, displayName, status)`. No origin/registry info. The SkillEntry struct (lines 23-28) only has Name, Group, Linked, Selected -- no Registry or Origin field.
  implication: manage.go's data model (SkillEntry) doesn't even carry registry info from disk. loadSkillsForProvider() reads directory entries and checks symlinks but never looks up config metadata (SkillMeta which has Registry).

- timestamp: 2026-03-06T00:00:00Z
  checked: SkillMeta in config.go (lines 36-41) and loadSkillsForProvider in manage.go (lines 104-164)
  found: SkillMeta has Owner, Name, Registry, URL fields and IS stored in config.json. But loadSkillsForProvider() never reads config -- it only reads filesystem directories. SkillEntry struct has no Registry/Owner field.
  implication: Even though registry metadata is persisted in config.json, the manage view never loads or uses it.

- timestamp: 2026-03-06T00:00:00Z
  checked: registryDisplayName() in config.go (lines 70-78)
  found: Maps "skills.sh" -> "Vercel", "playbooks.com" -> "Playbooks". This helper exists but is only used in the config view for registry labels, never in search or manage views.
  implication: A display-name mapper exists and could be reused for both search and manage views.

## Resolution

root_cause: |
  TWO separate gaps:

  1. SEARCH VIEW (search.go): The api.Skill struct has a `Registry` field that is correctly populated
     ("skills.sh" or "playbooks.com") by both SearchSkillsSh() and SearchPlaybooks(). However, the
     View() method at lines 320-327 only renders three columns (Name, Source, Popularity) and never
     references `skill.Registry`. The column width calculation at lines 262-264 only allocates space
     for name (35%), source (55%), and count (8 chars fixed), with no allocation for a registry column.

  2. MANAGE VIEW (manage.go): The SkillEntry struct (lines 23-28) has no Registry or Origin field at
     all. The loadSkillsForProvider() function (lines 104-164) builds SkillEntry values from filesystem
     directory listings only -- it never consults config.json where SkillMeta (with Registry field)
     is stored. The View() rendering at line 668 formats only checkbox + displayName + status.
     Group headers (line 637) show arrow + checkbox + groupName + counts but no origin info either.

fix: N/A (diagnosis only)
verification: N/A
files_changed: []
