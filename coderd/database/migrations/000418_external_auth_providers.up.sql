-- Add external_auth to crypto_key_feature enum for encrypting client secrets
ALTER TYPE crypto_key_feature ADD VALUE IF NOT EXISTS 'external_auth';

-- Table for dynamically created external auth providers (e.g., GitHub Apps via manifest)
CREATE TABLE IF NOT EXISTS external_auth_providers (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL DEFAULT 'github',
    client_id TEXT NOT NULL,
    client_secret_encrypted BYTEA,
    client_secret_key_id TEXT,
    display_name TEXT,
    display_icon TEXT,
    auth_url TEXT,
    token_url TEXT,
    validate_url TEXT,
    device_code_url TEXT,
    scopes TEXT[] DEFAULT '{}',
    extra_token_keys TEXT[] DEFAULT '{}',
    no_refresh BOOLEAN NOT NULL DEFAULT FALSE,
    device_flow BOOLEAN NOT NULL DEFAULT FALSE,
    regex TEXT,
    app_install_url TEXT,
    app_installations_url TEXT,
    -- GitHub App specific fields
    github_app_id BIGINT,
    github_app_webhook_secret_encrypted BYTEA,
    github_app_webhook_secret_key_id TEXT,
    github_app_private_key_encrypted BYTEA,
    github_app_private_key_key_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE external_auth_providers IS 'Stores dynamically created external auth providers like GitHub Apps created via manifest flow';
COMMENT ON COLUMN external_auth_providers.client_secret_encrypted IS 'Encrypted OAuth client secret using dbcrypt';
COMMENT ON COLUMN external_auth_providers.github_app_id IS 'GitHub App ID for GitHub Apps created via manifest';

-- Table for tracking GitHub App manifest flow state
CREATE TABLE IF NOT EXISTS external_auth_manifest_states (
    state TEXT PRIMARY KEY,
    redirect_uri TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '10 minutes'
);

COMMENT ON TABLE external_auth_manifest_states IS 'Temporary state storage for GitHub App manifest creation flow';

-- Index for cleaning up expired states
CREATE INDEX IF NOT EXISTS idx_external_auth_manifest_states_expires_at ON external_auth_manifest_states(expires_at);
