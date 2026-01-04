-- name: GetExternalAuthProviders :many
SELECT * FROM external_auth_providers
ORDER BY created_at ASC;

-- name: GetExternalAuthProviderByID :one
SELECT * FROM external_auth_providers
WHERE id = $1;

-- name: InsertExternalAuthProvider :one
INSERT INTO external_auth_providers (
    id,
    type,
    client_id,
    client_secret_encrypted,
    client_secret_key_id,
    display_name,
    display_icon,
    auth_url,
    token_url,
    validate_url,
    device_code_url,
    scopes,
    extra_token_keys,
    no_refresh,
    device_flow,
    regex,
    app_install_url,
    app_installations_url,
    github_app_id,
    github_app_webhook_secret_encrypted,
    github_app_webhook_secret_key_id,
    github_app_private_key_encrypted,
    github_app_private_key_key_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, NOW(), NOW()
) RETURNING *;

-- name: UpdateExternalAuthProvider :one
UPDATE external_auth_providers SET
    display_name = COALESCE($2, display_name),
    display_icon = COALESCE($3, display_icon),
    scopes = COALESCE($4, scopes),
    no_refresh = COALESCE($5, no_refresh),
    device_flow = COALESCE($6, device_flow),
    regex = COALESCE($7, regex),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteExternalAuthProvider :exec
DELETE FROM external_auth_providers
WHERE id = $1;

-- Manifest state queries
-- name: InsertExternalAuthManifestState :one
INSERT INTO external_auth_manifest_states (state, redirect_uri, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '10 minutes')
RETURNING *;

-- name: GetExternalAuthManifestState :one
SELECT * FROM external_auth_manifest_states
WHERE state = $1 AND expires_at > NOW();

-- name: DeleteExternalAuthManifestState :exec
DELETE FROM external_auth_manifest_states
WHERE state = $1;

-- name: CleanupExpiredManifestStates :exec
DELETE FROM external_auth_manifest_states
WHERE expires_at < NOW();
