# Coder Fork - Custom Features

This fork of [coder/coder](https://github.com/coder/coder) adds features for workspace collaboration and simplified deployment.

**Upstream**: https://github.com/coder/coder
**Fork**: https://github.com/noeljackson/coder

## Custom Features

### Workspace Invitations & Collaborators

Share workspaces with other users via email invitations.

**Access Levels:**
- `readonly` - View workspace only
- `use` - Connect to and use workspace
- `admin` - Full control including managing collaborators

**API Endpoints:**

| Endpoint                                     | Method | Description             |
|----------------------------------------------|--------|-------------------------|
| `/api/v2/workspaces/{id}/invitations`        | POST   | Create invitation       |
| `/api/v2/workspaces/{id}/invitations`        | GET    | List invitations        |
| `/api/v2/workspaces/{id}/invitations/{id}`   | DELETE | Cancel invitation       |
| `/api/v2/invitations/{token}`                | GET    | Get invitation by token |
| `/api/v2/invitations/{token}/accept`         | POST   | Accept invitation       |
| `/api/v2/invitations/{token}/decline`        | POST   | Decline invitation      |
| `/api/v2/workspaces/{id}/collaborators`      | GET    | List collaborators      |
| `/api/v2/workspaces/{id}/collaborators/{id}` | PATCH  | Update access level     |
| `/api/v2/workspaces/{id}/collaborators/{id}` | DELETE | Remove collaborator     |
| `/api/v2/users/me/workspace-collaborations`  | GET    | My collaborations       |
| `/api/v2/users/me/workspace-invitations`     | GET    | My pending invitations  |

**Files:**
- `codersdk/workspaceinvitations.go` - SDK types and client methods
- `coderd/workspaceinvitations.go` - API handlers
- `coderd/database/queries/workspaceinvitations.sql` - Database queries
- `site/src/pages/InvitationPage/` - Invitation acceptance UI
- `site/src/modules/workspaces/WorkspaceCollaborators/` - Collaborator management UI
- `site/src/api/queries/workspaceInvitations.ts` - Frontend API hooks

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
- Auto-detects latest upstream stable version
- Publishes to `ghcr.io/noeljackson/coder`
- Includes embedded frontend (full server binary)
- Alpine-based image with Terraform pre-installed

**Trigger:**
- Push to `main` branch (including automated sync merges)
- Manual dispatch with optional version override

**Files:**
- `.github/workflows/build-release.yaml`

### Full Pipeline

Upstream change → daily sync auto-merges → push to `main` → Docker build → `ghcr.io/noeljackson/coder:latest`

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

**Common Merge Conflicts:**
- `coderd/database/dbmetrics/querymetrics.go` - Keep `ExpireWorkspaceInvitations`
- `site/src/api/typesGenerated.ts` - Keep `UpdateWorkspaceCollaboratorRequest`

## Database Schema

This fork adds two tables:

### workspace_invitations

| Column       | Type      | Description                                |
|--------------|-----------|--------------------------------------------|
| id           | UUID      | Primary key                                |
| workspace_id | UUID      | Target workspace                           |
| inviter_id   | UUID      | User who sent invitation                   |
| email        | TEXT      | Invitee email                              |
| access_level | TEXT      | readonly/use/admin                         |
| token        | TEXT      | Unique acceptance token                    |
| status       | TEXT      | pending/accepted/declined/expired/canceled |
| expires_at   | TIMESTAMP | Expiration time                            |
| created_at   | TIMESTAMP | Creation time                              |
| responded_at | TIMESTAMP | Response time (nullable)                   |

### workspace_collaborators

| Column       | Type      | Description               |
|--------------|-----------|---------------------------|
| id           | UUID      | Primary key               |
| workspace_id | UUID      | Target workspace          |
| user_id      | UUID      | Collaborator user         |
| access_level | TEXT      | readonly/use/admin        |
| invited_by   | UUID      | Who added them (nullable) |
| created_at   | TIMESTAMP | Creation time             |

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
