---
# gobig-s36o
title: Support IF/END IF blocks (BigQuery scripting)
status: completed
type: bug
priority: normal
created_at: 2026-01-31T04:14:54Z
updated_at: 2026-01-31T04:19:26Z
---

go-bigq lint fails on IF/END IF control flow blocks:
```sql
IF scores_count = 0 THEN
  SELECT ERROR(FORMAT(...));
END IF;
```
These are BQ scripting control flow. ZetaSQL rejects both the IF and END keywords at top level.

Related: go-bigq-yqn9 (DECLARE)