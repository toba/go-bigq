---
# go-bigq-yqn9
title: Support DECLARE statements (BigQuery scripting)
status: completed
type: bug
priority: normal
created_at: 2026-01-31T04:12:03Z
updated_at: 2026-01-31T04:47:49Z
sync:
    clickup:
        synced_at: "2026-02-21T00:27:56Z"
        task_id: 868hk016w
---

go-bigq lint fails on files with DECLARE statements (BigQuery scripting syntax). ZetaSQL doesn't support DECLARE as a top-level statement.

## Context
Pacer's warehouse models (yes_scores_daily.sql, yes_features_daily.sql, yes_explain_daily.sql) use DECLARE for script variables:
```sql
DECLARE run_date DATE DEFAULT (...);
DECLARE inserted_rows INT64;
DECLARE IMPACT_HORIZON_DAYS INT64 DEFAULT 60;
```

These are valid BigQuery SQL but ZetaSQL's parser rejects them with:
```
error: parse error: 4:1: Syntax error: Unexpected keyword DECLARE
```

## Proposed Fix
Strip DECLARE lines (up to the semicolon) before passing to ZetaSQL. DECLARE is always a top-level statement so a simple regex strip should work. Could be done in the preprocessor alongside any other transformations.

## Workaround
Pacer currently strips DECLARE lines via sed in lint-bq.sh before passing to go-bigq.
