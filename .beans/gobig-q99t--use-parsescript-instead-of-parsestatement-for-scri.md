---
# gobig-q99t
title: Use ParseScript instead of ParseStatement for scripting support
status: completed
type: feature
priority: normal
created_at: 2026-01-31T04:22:28Z
updated_at: 2026-01-31T04:26:21Z
---

Replace the skip-scripting-keywords approach with ZetaSQL's ParseScript API, which parses full multi-statement scripts including DECLARE, SET, ASSERT, IF/END IF. ParseScript is a superset of ParseStatement — regular DML/DDL/DQL statements are still fully syntax-validated.

## Checklist

- [ ] Find ParseScript in the googlesql headers and understand its signature
- [ ] Add ParseScript C bridge function in bridge.go / zetasql_bridge.cc
- [ ] Expose ParseScript in the bigq package
- [ ] Update linter to use ParseScript on the full SQL string instead of split+skip
- [ ] Remove isScriptingStatement and splitStatements (or keep splitStatements for line tracking)
- [ ] Update tests
- [ ] Remove scripting-skip mention from README (it's now real parsing)