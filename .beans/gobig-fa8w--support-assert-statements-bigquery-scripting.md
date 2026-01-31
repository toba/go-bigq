---
# gobig-fa8w
title: Support ASSERT statements (BigQuery scripting)
status: completed
type: bug
priority: normal
created_at: 2026-01-31T04:14:54Z
updated_at: 2026-01-31T04:19:26Z
---

go-bigq lint fails on ASSERT statements. BQ scripting guardrails:
```sql
ASSERT run_date IS NOT NULL
  AS 'Guardrail failed: ...';
ASSERT inserted_rows > 0
  AS 'Guardrail failed: ...';
```
ASSERT can span multiple lines (AS clause on next line). ZetaSQL doesn't parse these.

Related: go-bigq-yqn9 (DECLARE)