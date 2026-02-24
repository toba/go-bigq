---
# gobig-4vwt
title: Make go-bigq Homebrew-installable
status: completed
type: feature
priority: normal
created_at: 2026-01-31T03:10:32Z
updated_at: 2026-01-31T03:12:43Z
sync:
    clickup:
        synced_at: "2026-02-21T00:27:56Z"
        task_id: 868hk016r
---

Add version ldflags, GitHub Actions release workflow, and create homebrew-bigq tap repo.

## Checklist
- [ ] Add version ldflags support to cmd/bigq/main.go
- [ ] Create .github/workflows/release.yml
- [ ] Create ../homebrew-bigq/Formula/bigq.rb
- [ ] Create ../homebrew-bigq/README.md
- [ ] Verify go build compiles
