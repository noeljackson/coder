-- name: CreateWorkspaceInvitation :one
INSERT INTO workspace_invitations (
    workspace_id,
    inviter_id,
    email,
    access_level,
    token,
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetWorkspaceInvitationByID :one
SELECT * FROM workspace_invitations
WHERE id = $1
LIMIT 1;

-- name: GetWorkspaceInvitationByToken :one
SELECT * FROM workspace_invitations
WHERE token = $1
LIMIT 1;

-- name: GetWorkspaceInvitationsByWorkspaceID :many
SELECT * FROM workspace_invitations
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: GetPendingWorkspaceInvitationsByEmail :many
SELECT * FROM workspace_invitations
WHERE email = $1 AND status = 'pending' AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: UpdateWorkspaceInvitationStatus :one
UPDATE workspace_invitations
SET status = $2, responded_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspaceInvitation :exec
DELETE FROM workspace_invitations
WHERE id = $1;

-- name: ExpireWorkspaceInvitations :exec
UPDATE workspace_invitations
SET status = 'expired'
WHERE status = 'pending' AND expires_at <= NOW();

-- name: CreateWorkspaceCollaborator :one
INSERT INTO workspace_collaborators (
    workspace_id,
    user_id,
    access_level,
    invited_by
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkspaceCollaboratorByID :one
SELECT * FROM workspace_collaborators
WHERE id = $1
LIMIT 1;

-- name: GetWorkspaceCollaboratorsByWorkspaceID :many
SELECT * FROM workspace_collaborators
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: GetWorkspaceCollaboratorByUserAndWorkspace :one
SELECT * FROM workspace_collaborators
WHERE workspace_id = $1 AND user_id = $2
LIMIT 1;

-- name: GetWorkspaceCollaborationsByUserID :many
SELECT * FROM workspace_collaborators
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateWorkspaceCollaboratorAccessLevel :one
UPDATE workspace_collaborators
SET access_level = $2
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspaceCollaborator :exec
DELETE FROM workspace_collaborators
WHERE id = $1;

-- name: DeleteWorkspaceCollaboratorByUserAndWorkspace :exec
DELETE FROM workspace_collaborators
WHERE workspace_id = $1 AND user_id = $2;
