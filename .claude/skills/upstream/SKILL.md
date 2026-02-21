---
name: upstream
description: |
  Check upstream repos for new changes that may be worth incorporating. Use when:
  (1) User says /upstream
  (2) User asks to "check upstream" or "what changed upstream"
  (3) User wants to know if upstream repos have new commits
  (4) User asks about syncing with or pulling from upstream sources
---

# Upstream Change Tracker

Check upstream repos for new commits, classify changes by relevance, and present a summary.

## Upstream Repos

| Repo | Default Branch | Relationship | Used By |
|------|---------------|-------------|---------|
| `google/googlesql` | `master` | Git submodule (pinned at tag) | `third_party/googlesql/` → compiled to `lib/libzetasql.a` via Bazel, linked via CGO in `internal/bridge/` |

### google/googlesql

Core SQL parser and analyzer engine. This is the project's only major dependency — a C++ library compiled as a static archive and linked via CGO. The submodule is pinned to a tagged release (currently `2025.12.1`). Updates require rebuilding the static library via `make lib`.

**What to watch for:**
- New tagged releases (version format: `YYYY.MM.N`)
- Bug fixes to SQL parsing or analysis
- New BigQuery SQL features (new statement types, functions, type support)
- Changes to the C++ API surface used by `internal/bridge/`
- Breaking changes to Bazel build targets

## Workflow

### Step 1: Read Marker File

Read `.claude/skills/upstream/references/last-checked.json`.

- **If the file does not exist** → this is a first run. Set `FIRST_RUN=true`.
- **If the file exists** → parse the JSON to get `last_checked_sha` and `last_checked_date` per repo.

### Step 2: Fetch Changes

#### First Run (no marker file)

Fetch the last 30 commits and recent releases:

```bash
gh api "repos/google/googlesql/commits?per_page=30&sha=master" --jq '[.[] | {sha: .sha, date: .commit.committer.date, message: (.commit.message | split("\n") | .[0]), author: .commit.author.name}]'
```

```bash
gh api "repos/google/googlesql/releases?per_page=10" --jq '[.[] | {tag: .tag_name, date: .published_at, name: .name}]'
```

#### Subsequent Runs (marker file exists)

Use the compare API:

```bash
gh api "repos/google/googlesql/compare/{LAST_SHA}...master" --jq '{total_commits: .total_commits, commits: [.commits[] | {sha: .sha, date: .commit.committer.date, message: (.commit.message | split("\n") | .[0]), author: .commit.author.name}], files: [.files[].filename]}'
```

Also check for new releases since last check:

```bash
gh api "repos/google/googlesql/releases?per_page=10" --jq '[.[] | {tag: .tag_name, date: .published_at, name: .name}]'
```

**Fallback:** If the compare API returns 404 (e.g. force-push rewrote history), fall back to date-based query:

```bash
gh api "repos/google/googlesql/commits?since={LAST_DATE}&sha=master&per_page=100" --jq '[.[] | {sha: .sha, date: .commit.committer.date, message: (.commit.message | split("\n") | .[0]), author: .commit.author.name}]'
```

### Step 3: Classify Changed Files by Relevance

| Relevance | Path Patterns |
|-----------|--------------|
| **HIGH** | `zetasql/parser/**`, `zetasql/analyzer/**`, `zetasql/public/**` (API surface, parser, analyzer — directly affects our CGO bridge) |
| **MEDIUM** | `zetasql/resolved_ast/**`, `zetasql/scripting/**`, `zetasql/common/**`, `BUILD`, `*.bzl`, `MODULE.bazel` (AST types, scripting support, build changes) |
| **LOW** | `docs/**`, `README.md`, `LICENSE`, `.github/**`, `zetasql/tools/**`, `zetasql/testing/**`, `java/**`, `javatests/**` (docs, tooling, Java bindings we don't use) |

Files not matching any pattern → **MEDIUM** (unknown = worth reviewing).

### Step 4: Present Summary

Format the output as follows:

```
# Upstream Changes

## google/googlesql (N new commits since YYYY-MM-DD)

**Current submodule pin:** 2025.12.1 (sha)
**Latest release:** YYYY.MM.N

### Commits
- `abc1234` Fix parser handling of QUALIFY clause — @author (2025-05-01)
- `def5678` Add support for PIPE syntax — @author (2025-04-28)

### Changed Files

**HIGH relevance** (parser/analyzer/public API):
- zetasql/parser/parser.cc
- zetasql/public/simple_catalog.h

**MEDIUM relevance** (AST/scripting/build):
- zetasql/scripting/script_executor.cc

**LOW relevance** (docs/tooling):
- README.md

### New Releases Since Last Check
- 2025.06.1 (2025-06-15)

**Assessment:** Summary of whether an update is warranted and what it would involve.

---

## Overall Recommendation
(Summarize: are there new releases to upgrade to, any high-relevance changes, suggested action)
```

If the repo has **no new commits**, show:

```
## google/googlesql — No new commits since last check (YYYY-MM-DD)
```

### Step 5: Update Marker File

Build the new marker JSON with the HEAD SHA and current date.

- **First run:** Write the marker file automatically (tell the user it was created).
- **Subsequent runs:** Ask the user "Update the last-checked markers to current HEAD?" before writing.

Write to `.claude/skills/upstream/references/last-checked.json`:

```json
{
  "google/googlesql": {
    "last_checked_sha": "<HEAD_SHA>",
    "last_checked_date": "<ISO_DATE>"
  }
}
```
