---
# core-9exd
title: Evaluate BigQuery SQL linter for warehouse models
status: draft
type: task
priority: normal
tags:
    - bigquery
created_at: 2026-01-30T17:21:39Z
updated_at: 2026-01-31T00:56:13Z
sync:
    clickup:
        synced_at: "2026-02-21T00:27:57Z"
        task_id: 868hk016y
---

## Context

Our warehouse model SQL is embedded in Go string literals with concatenated table references (e.g. `+ outputTable +`). We currently have:
- A narrow unit test (`TestScriptModels_DECLAREBeforeDDL`) that catches DECLARE ordering issues
- An integration test (`TestAllModelsGenerateValidSQL`) that does BigQuery DryRun but requires GCP credentials

A general-purpose BigQuery linter would catch more issues without needing credentials.

## Options

### 1. SQLFluff (Python)
- GitHub: https://github.com/sqlfluff/sqlfluff
- Most popular open-source SQL linter, multi-dialect (BigQuery, Snowflake, etc.)
- 70+ built-in rules, auto-fix, dbt support
- `sqlfluff lint query.sql --dialect bigquery`
- Downside: Python dependency in a Go project

### 2. bqvalid (Go)
- GitHub: https://github.com/hirosassa/bqvalid
- BigQuery-specific, Go-native
- Lighter weight, fewer rules, less active development

### 3. ZetaSQL / GoogleSQL (Google's own parser)
- https://github.com/google/googlesql — the actual SQL engine BigQuery uses internally
- Latest release: 2025.12.1 (actively maintained)
- Provides `execute_query` prebuilt binary for Linux and macOS
- Can parse and analyze SQL with full semantic validation — type checking, function validation, schema-aware column references
- Not a library you link against — it's a C++ codebase with a CLI tool

### 4. go-zetasql (Go bindings via cgo)
- https://github.com/goccy/go-zetasql — Go bindings wrapping ZetaSQL via cgo
- Last release: v0.5.5 (Dec 2023) — stale, may not match latest BigQuery features
- Exposes `zetasql.ParseStatement()` for syntax and `zetasql.AnalyzeStatement()` for full semantic analysis with a catalog
- Requires `CGO_ENABLED=1` and `CXX=clang++`, slow first build (compiles all of ZetaSQL)
- Could define a SimpleCatalog with your table schemas and validate SQL offline in a Go test

### 5. Shell out to ZetaSQL execute_query (parse-only)
- Write a Go test that shells out to the `execute_query` binary (like we do with sqruff) in parse-only mode
- No schema catalog needed for syntax validation — catches DECLARE ordering, invalid function names, malformed expressions

## Recommendation

The pragmatic path: shell out to the googlesql `execute_query` binary from a Go test, same pattern as sqruff. Parse-only mode catches syntax errors without needing a schema catalog. For full semantic validation (wrong column names, type mismatches), the existing DryRun integration test already covers that.

go-zetasql is the richer option (full analysis in-process) but it's stale and the cgo build complexity conflicts with our "few dependencies, simple builds" values.

## Integration approach

Since SQL is in Go strings, any linter needs a harness:
1. Generate SQL from each model via `m.SQL(execCtx)` (already done in tests)
2. Write to temp file
3. Run linter on temp file
4. Report errors back in test output

Could be implemented as:
- A Go test that shells out to the linter (like we do with `golangci-lint`)
- A standalone lint script (like `scripts/lint-sql.sh`)

## Checklist

- [ ] Evaluate SQLFluff BigQuery dialect coverage (does it catch DECLARE ordering, syntax errors, etc.)
- [ ] Evaluate bqvalid rule coverage
- [ ] Evaluate ZetaSQL execute_query parse-only mode coverage
- [ ] Pick one and add as dev dependency (mise or pip)
- [ ] Wire up as test or lint script against all MaterializeScript models
- [ ] Add to CI / pre-push hooks
