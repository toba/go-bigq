---
# gobig-lr24
title: 'Phase 1: Bootstrap — Submodule + Bazel Static Library'
status: completed
type: epic
priority: normal
created_at: 2026-01-31T01:17:35Z
updated_at: 2026-01-31T02:50:48Z
parent: gobig-iut0
sync:
    clickup:
        synced_at: "2026-02-21T00:27:58Z"
        task_id: 868hk0170
---

Set up project structure, git submodule for googlesql 2025.12.1, Bazel build for static lib, custom C bridge (not goccy copy), Go bindings.

## Checklist
- [x] Initialize go.mod
- [x] Add googlesql git submodule at 2025.12.1
- [x] Create project directory structure
- [x] Write Makefile for Bazel build
- [x] Write custom C bridge (zetasql_bridge.h/.cc) wrapping parse/analyze/catalog
- [x] Write Go CGO bridge (internal/bridge/bridge.go)
- [x] Write public Go API (zetasql/zetasql.go)
- [x] Write schema loading (internal/schema/)
- [x] Write catalog builder (internal/catalog/)
- [x] Write linter core (internal/lint/)
- [x] Write CLI (cmd/bigq/)
- [x] Build static library via Bazel (cc_static_library)
- [x] Copy headers for CGO compilation
- [x] Link against static lib + ICU
- [x] Verify go build succeeds
- [x] Verify ParseStatement smoke test passes
- [x] Verify AnalyzeStatement with catalog passes
- [ ] Update collect_headers.sh for fully automated builds
