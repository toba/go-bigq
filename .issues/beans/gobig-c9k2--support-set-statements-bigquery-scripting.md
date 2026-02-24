---
# gobig-c9k2
title: Support SET statements (BigQuery scripting)
status: completed
type: bug
priority: normal
created_at: 2026-01-31T04:14:54Z
updated_at: 2026-01-31T04:47:49Z
sync:
    clickup:
        synced_at: "2026-02-21T00:27:56Z"
        task_id: 868hk016u
---

go-bigq lint fails on SET statements. These are BQ scripting syntax for variable assignment:
```sql
SET inserted_rows = (SELECT COUNT(*) FROM ...);
SET scores_count = (SELECT COUNT(*) FROM ...);
```
SET can span multiple lines (subquery on following lines). ZetaSQL doesn't handle these as top-level statements.

Related: go-bigq-yqn9 (DECLARE)
