package codersdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EnhancedExternalAuthProvider is a constant that represents enhanced
// support for a type of external authentication. All of the Git providers
// are examples of enhanced, because they support intercepting "git clone".
type EnhancedExternalAuthProvider string

func (e EnhancedExternalAuthProvider) String() string {
	return string(e)
}

// Git returns whether the provider is a Git provider.
func (e EnhancedExternalAuthProvider) Git() bool {
	switch e {
	case EnhancedExternalAuthProviderGitHub,
		EnhancedExternalAuthProviderGitLab,
		EnhancedExternalAuthProviderBitBucketCloud,
		EnhancedExternalAuthProviderBitBucketServer,
		EnhancedExternalAuthProviderAzureDevops,
		EnhancedExternalAuthProviderAzureDevopsEntra,
		EnhancedExternalAuthProviderGitea:
		return true
	default:
		return false
	}
}

const (
	EnhancedExternalAuthProviderAzureDevops EnhancedExternalAuthProvider = "azure-devops"
	// Authenticate to ADO using an app registration in Entra ID
	EnhancedExternalAuthProviderAzureDevopsEntra EnhancedExternalAuthProvider = "azure-devops-entra"
	EnhancedExternalAuthProviderGitHub           EnhancedExternalAuthProvider = "github"
	EnhancedExternalAuthProviderGitLab           EnhancedExternalAuthProvider = "gitlab"
	// EnhancedExternalAuthProviderBitBucketCloud is the Bitbucket Cloud provider.
	// Not to be confused with the self-hosted 'EnhancedExternalAuthProviderBitBucketServer'
	EnhancedExternalAuthProviderBitBucketCloud  EnhancedExternalAuthProvider = "bitbucket-cloud"
	EnhancedExternalAuthProviderBitBucketServer EnhancedExternalAuthProvider = "bitbucket-server"
	EnhancedExternalAuthProviderSlack           EnhancedExternalAuthProvider = "slack"
	EnhancedExternalAuthProviderJFrog           EnhancedExternalAuthProvider = "jfrog"
	EnhancedExternalAuthProviderGitea           EnhancedExternalAuthProvider = "gitea"
)

type ExternalAuth struct {
	Authenticated      bool   `json:"authenticated"`
	Device             bool   `json:"device"`
	DisplayName        string `json:"display_name"`
	SupportsRevocation bool   `json:"supports_revocation"`

	// User is the user that authenticated with the provider.
	User *ExternalAuthUser `json:"user"`
	// AppInstallable is true if the request for app installs was successful.
	AppInstallable bool `json:"app_installable"`
	// AppInstallations are the installations that the user has access to.
	AppInstallations []ExternalAuthAppInstallation `json:"installations"`
	// AppInstallURL is the URL to install the app.
	AppInstallURL string `json:"app_install_url"`
}

type ListUserExternalAuthResponse struct {
	Providers []ExternalAuthLinkProvider `json:"providers"`
	// Links are all the authenticated links for the user.
	// If a link has a provider ID that does not exist, then that provider
	// is no longer configured, rendering it unusable. It is still valuable
	// to include these links so that the user can unlink them.
	Links []ExternalAuthLink `json:"links"`
}

type DeleteExternalAuthByIDResponse struct {
	// TokenRevoked set to true if token revocation was attempted and was successful
	TokenRevoked         bool   `json:"token_revoked"`
	TokenRevocationError string `json:"token_revocation_error,omitempty"`
}

// ExternalAuthLink is a link between a user and an external auth provider.
// It excludes information that requires a token to access, so can be statically
// built from the database and configs.
type ExternalAuthLink struct {
	ProviderID      string    `json:"provider_id"`
	CreatedAt       time.Time `json:"created_at" format:"date-time"`
	UpdatedAt       time.Time `json:"updated_at" format:"date-time"`
	HasRefreshToken bool      `json:"has_refresh_token"`
	Expires         time.Time `json:"expires" format:"date-time"`
	Authenticated   bool      `json:"authenticated"`
	ValidateError   string    `json:"validate_error"`
}

// ExternalAuthLinkProvider are the static details of a provider.
type ExternalAuthLinkProvider struct {
	ID                            string   `json:"id"`
	Type                          string   `json:"type"`
	Device                        bool     `json:"device"`
	DisplayName                   string   `json:"display_name"`
	DisplayIcon                   string   `json:"display_icon"`
	AllowRefresh                  bool     `json:"allow_refresh"`
	AllowValidate                 bool     `json:"allow_validate"`
	SupportsRevocation            bool     `json:"supports_revocation"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

type ExternalAuthAppInstallation struct {
	ID           int              `json:"id"`
	Account      ExternalAuthUser `json:"account"`
	ConfigureURL string           `json:"configure_url"`
}

type ExternalAuthUser struct {
	ID         int64  `json:"id"`
	Login      string `json:"login"`
	AvatarURL  string `json:"avatar_url"`
	ProfileURL string `json:"profile_url"`
	Name       string `json:"name"`
}

// ExternalAuthDevice is the response from the device authorization endpoint.
// See: https://tools.ietf.org/html/rfc8628#section-3.2
type ExternalAuthDevice struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type ExternalAuthDeviceExchange struct {
	DeviceCode string `json:"device_code"`
}

func (c *Client) ExternalAuthDeviceByID(ctx context.Context, provider string) (ExternalAuthDevice, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/external-auth/%s/device", provider), nil)
	if err != nil {
		return ExternalAuthDevice{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ExternalAuthDevice{}, ReadBodyAsError(res)
	}
	var extAuth ExternalAuthDevice
	return extAuth, json.NewDecoder(res.Body).Decode(&extAuth)
}

// ExchangeGitAuth exchanges a device code for an external auth token.
func (c *Client) ExternalAuthDeviceExchange(ctx context.Context, provider string, req ExternalAuthDeviceExchange) error {
	res, err := c.Request(ctx, http.MethodPost, fmt.Sprintf("/api/v2/external-auth/%s/device", provider), req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return ReadBodyAsError(res)
	}
	return nil
}

// ExternalAuthByID returns the external auth for the given provider by ID.
func (c *Client) ExternalAuthByID(ctx context.Context, provider string) (ExternalAuth, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/external-auth/%s", provider), nil)
	if err != nil {
		return ExternalAuth{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ExternalAuth{}, ReadBodyAsError(res)
	}
	var extAuth ExternalAuth
	return extAuth, json.NewDecoder(res.Body).Decode(&extAuth)
}

// UnlinkExternalAuthByID deletes the external auth for the given provider by ID
// for the user. This does not revoke the token from the IDP.
func (c *Client) UnlinkExternalAuthByID(ctx context.Context, provider string) (DeleteExternalAuthByIDResponse, error) {
	noRevoke := DeleteExternalAuthByIDResponse{TokenRevoked: false}
	res, err := c.Request(ctx, http.MethodDelete, fmt.Sprintf("/api/v2/external-auth/%s", provider), nil)
	if err != nil {
		return noRevoke, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return noRevoke, ReadBodyAsError(res)
	}
	var resp DeleteExternalAuthByIDResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return noRevoke, err
	}
	return resp, nil
}

// ListExternalAuths returns the available external auth providers and the user's
// authenticated links if they exist.
func (c *Client) ListExternalAuths(ctx context.Context) (ListUserExternalAuthResponse, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/external-auth", nil)
	if err != nil {
		return ListUserExternalAuthResponse{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ListUserExternalAuthResponse{}, ReadBodyAsError(res)
	}
	var extAuth ListUserExternalAuthResponse
	return extAuth, json.NewDecoder(res.Body).Decode(&extAuth)
}

// ExternalAuthProviderConfig is a database-stored external auth provider configuration.
// This is separate from the file-based configuration and supports runtime management.
type ExternalAuthProviderConfig struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	ClientID    string    `json:"client_id"`
	DisplayName string    `json:"display_name,omitempty"`
	DisplayIcon string    `json:"display_icon,omitempty"`
	AuthURL     string    `json:"auth_url,omitempty"`
	TokenURL    string    `json:"token_url,omitempty"`
	ValidateURL string    `json:"validate_url,omitempty"`
	Scopes      []string  `json:"scopes,omitempty"`
	NoRefresh   bool      `json:"no_refresh"`
	DeviceFlow  bool      `json:"device_flow"`
	Regex       string    `json:"regex,omitempty"`
	CreatedAt   time.Time `json:"created_at" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at" format:"date-time"`

	// GitHub App specific fields
	AppInstallURL      string `json:"app_install_url,omitempty"`
	AppInstallationsURL string `json:"app_installations_url,omitempty"`
	GithubAppID        int64  `json:"github_app_id,omitempty"`
}

// CreateExternalAuthProviderRequest is used to create a new external auth provider.
type CreateExternalAuthProviderRequest struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	DisplayName  string   `json:"display_name,omitempty"`
	DisplayIcon  string   `json:"display_icon,omitempty"`
	AuthURL      string   `json:"auth_url,omitempty"`
	TokenURL     string   `json:"token_url,omitempty"`
	ValidateURL  string   `json:"validate_url,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	NoRefresh    bool     `json:"no_refresh"`
	DeviceFlow   bool     `json:"device_flow"`
	Regex        string   `json:"regex,omitempty"`

	// GitHub App specific fields
	AppInstallURL         string `json:"app_install_url,omitempty"`
	AppInstallationsURL   string `json:"app_installations_url,omitempty"`
	GithubAppID           int64  `json:"github_app_id,omitempty"`
	GithubAppWebhookSecret string `json:"github_app_webhook_secret,omitempty"`
	GithubAppPrivateKey   string `json:"github_app_private_key,omitempty"`
}

// UpdateExternalAuthProviderRequest is used to update an external auth provider.
type UpdateExternalAuthProviderRequest struct {
	DisplayName *string  `json:"display_name,omitempty"`
	DisplayIcon *string  `json:"display_icon,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	NoRefresh   *bool    `json:"no_refresh,omitempty"`
	DeviceFlow  *bool    `json:"device_flow,omitempty"`
	Regex       *string  `json:"regex,omitempty"`
}

// GitHubAppManifestRequest is used to initiate the GitHub App manifest flow.
type GitHubAppManifestRequest struct {
	// Owner is the organization or user to create the GitHub App for.
	// If empty, creates a personal app.
	Owner       string `json:"owner,omitempty"`
	RedirectURI string `json:"redirect_uri"`
}

// GitHubAppManifestResponse contains the URL to redirect the user to GitHub.
type GitHubAppManifestResponse struct {
	// URL is the GitHub URL to redirect the user to for app creation.
	URL   string `json:"url"`
	State string `json:"state"`
}

// GitHubAppManifestCallbackRequest is the callback from GitHub after app creation.
type GitHubAppManifestCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// GetExternalAuthProviders returns all database-stored external auth providers.
func (c *Client) GetExternalAuthProviders(ctx context.Context) ([]ExternalAuthProviderConfig, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/deployment/external-auth-providers", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ReadBodyAsError(res)
	}
	var providers []ExternalAuthProviderConfig
	return providers, json.NewDecoder(res.Body).Decode(&providers)
}

// GetExternalAuthProvider returns a database-stored external auth provider by ID.
func (c *Client) GetExternalAuthProvider(ctx context.Context, id string) (ExternalAuthProviderConfig, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/deployment/external-auth-providers/%s", id), nil)
	if err != nil {
		return ExternalAuthProviderConfig{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ExternalAuthProviderConfig{}, ReadBodyAsError(res)
	}
	var provider ExternalAuthProviderConfig
	return provider, json.NewDecoder(res.Body).Decode(&provider)
}

// CreateExternalAuthProvider creates a new external auth provider.
func (c *Client) CreateExternalAuthProvider(ctx context.Context, req CreateExternalAuthProviderRequest) (ExternalAuthProviderConfig, error) {
	res, err := c.Request(ctx, http.MethodPost, "/api/v2/deployment/external-auth-providers", req)
	if err != nil {
		return ExternalAuthProviderConfig{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return ExternalAuthProviderConfig{}, ReadBodyAsError(res)
	}
	var provider ExternalAuthProviderConfig
	return provider, json.NewDecoder(res.Body).Decode(&provider)
}

// UpdateExternalAuthProvider updates an external auth provider.
func (c *Client) UpdateExternalAuthProvider(ctx context.Context, id string, req UpdateExternalAuthProviderRequest) (ExternalAuthProviderConfig, error) {
	res, err := c.Request(ctx, http.MethodPatch, fmt.Sprintf("/api/v2/deployment/external-auth-providers/%s", id), req)
	if err != nil {
		return ExternalAuthProviderConfig{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ExternalAuthProviderConfig{}, ReadBodyAsError(res)
	}
	var provider ExternalAuthProviderConfig
	return provider, json.NewDecoder(res.Body).Decode(&provider)
}

// DeleteExternalAuthProvider deletes an external auth provider.
func (c *Client) DeleteExternalAuthProvider(ctx context.Context, id string) error {
	res, err := c.Request(ctx, http.MethodDelete, fmt.Sprintf("/api/v2/deployment/external-auth-providers/%s", id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return ReadBodyAsError(res)
	}
	return nil
}

// InitiateGitHubAppManifest initiates the GitHub App manifest flow.
func (c *Client) InitiateGitHubAppManifest(ctx context.Context, req GitHubAppManifestRequest) (GitHubAppManifestResponse, error) {
	res, err := c.Request(ctx, http.MethodPost, "/api/v2/deployment/external-auth-providers/github/manifest", req)
	if err != nil {
		return GitHubAppManifestResponse{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return GitHubAppManifestResponse{}, ReadBodyAsError(res)
	}
	var resp GitHubAppManifestResponse
	return resp, json.NewDecoder(res.Body).Decode(&resp)
}

// CompleteGitHubAppManifest completes the GitHub App manifest flow after the callback.
func (c *Client) CompleteGitHubAppManifest(ctx context.Context, req GitHubAppManifestCallbackRequest) (ExternalAuthProviderConfig, error) {
	res, err := c.Request(ctx, http.MethodPost, "/api/v2/deployment/external-auth-providers/github/callback", req)
	if err != nil {
		return ExternalAuthProviderConfig{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return ExternalAuthProviderConfig{}, ReadBodyAsError(res)
	}
	var provider ExternalAuthProviderConfig
	return provider, json.NewDecoder(res.Body).Decode(&provider)
}
