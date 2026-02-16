---
description: Stage all changes and commit with a descriptive message
---

## Sync Beans to ClickUp

Before staging, sync beans in the background (non-blocking):

```bash
beanup --config .beans.clickup.yml sync &
```

## Step 1: Run Critical Review (Parallel)

**IMPORTANT**: Before committing, run build and tests **in parallel** (single message, multiple Bash calls).

Execute these commands concurrently:
1. `go build ./...` - verify compilation
2. `go test ./...` - run test suite
3. `go vet ./...` - check for issues

Then review the diff for security/quality issues.

If any command fails or review finds blocking issues, report them and STOP. Do not proceed to commit.

## Step 2: Stage and Commit

1. Run `git status --short` to see changes
2. Run `git diff HEAD` to review all changes
3. Stage all relevant changes
4. Commit with a concise, descriptive message:
   - Lowercase, imperative mood (e.g., "add feature" not "Added feature")
   - Focus on "why" not just "what"
   - Include affected bean IDs if applicable
5. Run `git status` to confirm the commit succeeded

## Step 3: Push, Version, and Release (if requested)

If $ARGUMENTS contains "push" or user requested push:

1. Get the latest version tag: `git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1`
2. Get the previous tag's commit to see what changed: `git log <latest-tag>..HEAD --oneline`
3. Determine version increment based on changes since last tag:
   - **patch** (x.y.Z): Bug fixes, minor improvements, documentation
   - **minor** (x.Y.0): New features, new tools, new flags
   - **major** (X.0.0): Breaking changes, API changes, removed functionality
4. Ask user to confirm the version increment (show current version and proposed new version)
5. After confirmation, push and tag:
   ```bash
   git push
   git tag v<new-version>
   git push origin v<new-version>
   ```

### Step 4: Build and Release Locally

Build the release binary locally (no CI — the ZetaSQL static library is pre-built):

```bash
CGO_ENABLED=1 go build -ldflags "-X main.version=v<new-version>" -o go-bigq ./cmd/bigq/
./go-bigq version
tar -czvf go-bigq-v<new-version>-arm64.tar.gz go-bigq
shasum -a 256 go-bigq-v<new-version>-arm64.tar.gz > go-bigq-v<new-version>-arm64.tar.gz.sha256
```

Generate release notes and create the GitHub release:

```bash
NOTES=$(git log --pretty=format:"- %s" <prev-tag>..v<new-version>)
gh release create v<new-version> --repo toba/go-bigq --title "v<new-version>" --notes "$NOTES" \
  go-bigq-v<new-version>-arm64.tar.gz \
  go-bigq-v<new-version>-arm64.tar.gz.sha256
```

Clean up build artifacts:

```bash
rm -f go-bigq go-bigq-v<new-version>-arm64.tar.gz go-bigq-v<new-version>-arm64.tar.gz.sha256
```

### Version Examples

- Current: v1.2.3
  - Bug fix → v1.2.4 (patch)
  - New ParseScript support → v1.3.0 (minor)
  - Changed CLI argument names → v2.0.0 (major)
