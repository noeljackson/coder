# Coder Fork - Custom Features

This fork of [coder/coder](https://github.com/coder/coder) adds features for simplified deployment with the Codespace platform.

**Upstream**: https://github.com/coder/coder
**Fork**: https://github.com/noeljackson/coder

## Custom Features

### Workspace Access Control

Workspace access is controlled at the Codespace platform layer via OIDC tenant
membership. Users must be members of a tenant to access its Coder instance.
The `maybeAutoConsent` function in Codespace automatically grants OAuth consent
for tenant members during OIDC login. When a user is removed from a tenant,
their consent is revoked.

The `workspace_invitations` and `workspace_collaborators` database tables exist
from a previous feature but are no longer used. The migration
(`000417_workspace_invitations.up.sql`) is preserved for safety.

### Email Integration (Resend)

Sends workspace invitation emails via [Resend](https://resend.com).

**Configuration:**

- `RESEND_API_KEY` - API key from Resend
- `RESEND_FROM_EMAIL` - Sender email address
- `RESEND_FROM_NAME` - Sender display name

**Files:**

- `coderd/email/resend.go` - Resend client implementation

## CI/CD

### Automated Upstream Sync

A daily GitHub Action keeps the fork up to date with upstream.

**How it works:**

1. Runs daily at 06:00 UTC (also supports manual trigger)
2. Fetches `upstream/main` and checks for new commits
3. If no new commits, exits early
4. Attempts `git merge upstream/main`
   - **Clean merge:** pushes to `main`, which triggers the Docker build
   - **Conflicts:** creates a PR on a `sync/upstream-YYYY-MM-DD` branch
     with the conflicting files listed for manual resolution

**Manual trigger:** Go to Actions > "Sync Upstream" > "Run workflow".

**Files:**

- `.github/workflows/sync-upstream.yaml`

### Docker Build Workflow

Custom GitHub Actions workflow for multi-arch Docker builds.

**Features:**

- Builds linux/amd64 and linux/arm64
- Auto-detects version via `scripts/version.sh` (do not manually tag)
- Publishes to `ghcr.io/noeljackson/coder`
- Includes embedded frontend (full server binary)
- Alpine-based image with Terraform pre-installed

**Trigger:**

- Push to `main` branch (including automated sync merges)
- Manual dispatch with optional version override

**Files:**

- `.github/workflows/build-release.yaml`

### Artifact Sync: Docker + Helm

The `latest` Docker tag is only applied after both the Docker image and Helm
chart are successfully pushed. This prevents a failure mode where the Codespace
platform resolves a version that has a Docker image but no corresponding Helm
chart, causing Terraform to fail with "Unable to locate chart."

**How it works:**

1. Docker image is pushed with version + SHA tags only (no `latest`)
2. Helm chart is packaged and pushed to `oci://ghcr.io/noeljackson/chart`
3. Only after the Helm chart push succeeds, the Docker image is retagged as
   `latest` via `docker buildx imagetools create`

If the Helm chart push fails, `latest` continues to point at the previous
known-good version. The Codespace version resolver also validates chart
existence server-side as a secondary safeguard.

### Full Pipeline

Upstream change → daily sync auto-merges → push to `main` → Docker build
→ Helm chart push → `latest` tag applied → `ghcr.io/noeljackson/coder:latest`

If the sync has conflicts, a PR is created instead and the build waits until
you merge that PR.

## Keeping Updated

### Automated (default)

The sync-upstream workflow handles this automatically. When conflicts occur,
you'll see a PR with instructions:

```bash
git fetch origin sync/upstream-YYYY-MM-DD
git checkout sync/upstream-YYYY-MM-DD
# Resolve conflicts, preserving fork customizations
git add .
git commit
git push origin sync/upstream-YYYY-MM-DD
# Then merge the PR on GitHub
```

### Manual (fallback)

```bash
git fetch upstream
git merge upstream/main
# Resolve conflicts (preserve fork customizations)
git push origin main
```

## Development

Use standard Coder development workflow:

```bash
# Start development server
./scripts/develop.sh

# Run tests
make test

# Generate database code after schema changes
make gen

# Lint
make lint
```
