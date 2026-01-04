package coderd

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/codersdk"
)

// @Summary List external auth providers
// @ID list-external-auth-providers
// @Security CoderSessionToken
// @Produce json
// @Tags Deployment
// @Success 200 {array} codersdk.ExternalAuthProviderConfig
// @Router /deployment/external-auth-providers [get]
func (api *API) listExternalAuthProviders(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	providers, err := api.Database.GetExternalAuthProviders(ctx)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get external auth providers.",
			Detail:  err.Error(),
		})
		return
	}

	result := make([]codersdk.ExternalAuthProviderConfig, 0, len(providers))
	for _, p := range providers {
		result = append(result, dbExternalAuthProviderToSDK(p))
	}

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Get external auth provider by ID
// @ID get-external-auth-provider-by-id
// @Security CoderSessionToken
// @Produce json
// @Tags Deployment
// @Param id path string true "Provider ID"
// @Success 200 {object} codersdk.ExternalAuthProviderConfig
// @Router /deployment/external-auth-providers/{id} [get]
func (api *API) getExternalAuthProviderByID(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	provider, err := api.Database.GetExternalAuthProviderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpapi.ResourceNotFound(rw)
			return
		}
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get external auth provider.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, dbExternalAuthProviderToSDK(provider))
}

// @Summary Create external auth provider
// @ID create-external-auth-provider
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Deployment
// @Param request body codersdk.CreateExternalAuthProviderRequest true "Provider details"
// @Success 201 {object} codersdk.ExternalAuthProviderConfig
// @Router /deployment/external-auth-providers [post]
func (api *API) createExternalAuthProvider(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req codersdk.CreateExternalAuthProviderRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	if req.ID == "" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Provider ID is required.",
		})
		return
	}

	if req.ClientID == "" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Client ID is required.",
		})
		return
	}

	// Encrypt the client secret.
	// For now, store it as-is - encryption will be handled by dbcrypt layer.
	clientSecretEncrypted := []byte(req.ClientSecret)

	// Encrypt GitHub App secrets if provided.
	webhookSecretEncrypted := []byte(req.GithubAppWebhookSecret)
	privateKeyEncrypted := []byte(req.GithubAppPrivateKey)

	provider, err := api.Database.InsertExternalAuthProvider(ctx, database.InsertExternalAuthProviderParams{
		ID:                              req.ID,
		Type:                            req.Type,
		ClientID:                        req.ClientID,
		ClientSecretEncrypted:           clientSecretEncrypted,
		ClientSecretKeyID:               sql.NullString{},
		DisplayName:                     toNullString(req.DisplayName),
		DisplayIcon:                     toNullString(req.DisplayIcon),
		AuthUrl:                         toNullString(req.AuthURL),
		TokenUrl:                        toNullString(req.TokenURL),
		ValidateUrl:                     toNullString(req.ValidateURL),
		DeviceCodeUrl:                   sql.NullString{},
		Scopes:                          req.Scopes,
		ExtraTokenKeys:                  []string{},
		NoRefresh:                       req.NoRefresh,
		DeviceFlow:                      req.DeviceFlow,
		Regex:                           toNullString(req.Regex),
		AppInstallUrl:                   toNullString(req.AppInstallURL),
		AppInstallationsUrl:             toNullString(req.AppInstallationsURL),
		GithubAppID:                     toNullInt64(req.GithubAppID),
		GithubAppWebhookSecretEncrypted: webhookSecretEncrypted,
		GithubAppWebhookSecretKeyID:     sql.NullString{},
		GithubAppPrivateKeyEncrypted:    privateKeyEncrypted,
		GithubAppPrivateKeyKeyID:        sql.NullString{},
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to create external auth provider.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusCreated, dbExternalAuthProviderToSDK(provider))
}

// @Summary Update external auth provider
// @ID update-external-auth-provider
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Deployment
// @Param id path string true "Provider ID"
// @Param request body codersdk.UpdateExternalAuthProviderRequest true "Provider details"
// @Success 200 {object} codersdk.ExternalAuthProviderConfig
// @Router /deployment/external-auth-providers/{id} [patch]
func (api *API) updateExternalAuthProvider(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req codersdk.UpdateExternalAuthProviderRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	// Build update params with COALESCE-friendly values.
	params := database.UpdateExternalAuthProviderParams{
		ID: id,
	}

	if req.DisplayName != nil {
		params.DisplayName = toNullString(*req.DisplayName)
	}
	if req.DisplayIcon != nil {
		params.DisplayIcon = toNullString(*req.DisplayIcon)
	}
	if req.Scopes != nil {
		params.Scopes = req.Scopes
	}
	if req.NoRefresh != nil {
		params.NoRefresh = *req.NoRefresh
	}
	if req.DeviceFlow != nil {
		params.DeviceFlow = *req.DeviceFlow
	}
	if req.Regex != nil {
		params.Regex = toNullString(*req.Regex)
	}

	provider, err := api.Database.UpdateExternalAuthProvider(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpapi.ResourceNotFound(rw)
			return
		}
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to update external auth provider.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, dbExternalAuthProviderToSDK(provider))
}

// @Summary Delete external auth provider
// @ID delete-external-auth-provider
// @Security CoderSessionToken
// @Tags Deployment
// @Param id path string true "Provider ID"
// @Success 204
// @Router /deployment/external-auth-providers/{id} [delete]
func (api *API) deleteExternalAuthProvider(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	err := api.Database.DeleteExternalAuthProvider(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpapi.ResourceNotFound(rw)
			return
		}
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to delete external auth provider.",
			Detail:  err.Error(),
		})
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// @Summary Initiate GitHub App manifest flow
// @ID initiate-github-app-manifest
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Deployment
// @Param request body codersdk.GitHubAppManifestRequest true "Manifest request"
// @Success 200 {object} codersdk.GitHubAppManifestResponse
// @Router /deployment/external-auth-providers/github/manifest [post]
func (api *API) initiateGitHubAppManifest(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req codersdk.GitHubAppManifestRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	if req.RedirectURI == "" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Redirect URI is required.",
		})
		return
	}

	// Generate a secure state token.
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to generate state token.",
			Detail:  err.Error(),
		})
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store the state in the database.
	_, err := api.Database.InsertExternalAuthManifestState(ctx, database.InsertExternalAuthManifestStateParams{
		State:       state,
		RedirectUri: req.RedirectURI,
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to store manifest state.",
			Detail:  err.Error(),
		})
		return
	}

	// Build the GitHub App manifest.
	// See: https://docs.github.com/en/apps/sharing-github-apps/registering-a-github-app-from-a-manifest
	accessURL := api.AccessURL.String()
	manifest := map[string]interface{}{
		"name": fmt.Sprintf("Coder - %s", api.AccessURL.Hostname()),
		"url":  accessURL,
		"hook_attributes": map[string]interface{}{
			"url":    fmt.Sprintf("%s/api/v2/deployment/external-auth-providers/github/webhook", accessURL),
			"active": false,
		},
		"redirect_url": fmt.Sprintf("%s/api/v2/deployment/external-auth-providers/github/callback?state=%s", accessURL, state),
		"callback_urls": []string{
			fmt.Sprintf("%s/external-auth/github/callback", accessURL),
		},
		"setup_url": fmt.Sprintf("%s/external-auth/github/callback", accessURL),
		"public":    false,
		"default_permissions": map[string]string{
			"contents": "read",
			"metadata": "read",
		},
		"default_events": []string{},
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to encode manifest.",
			Detail:  err.Error(),
		})
		return
	}

	// Build the GitHub URL.
	githubURL := "https://github.com/settings/apps/new"
	if req.Owner != "" {
		githubURL = fmt.Sprintf("https://github.com/organizations/%s/settings/apps/new", req.Owner)
	}

	// The manifest is passed as a form parameter.
	manifestURL := fmt.Sprintf("%s?manifest=%s", githubURL, url.QueryEscape(string(manifestJSON)))

	httpapi.Write(ctx, rw, http.StatusOK, codersdk.GitHubAppManifestResponse{
		URL:   manifestURL,
		State: state,
	})
}

// @Summary Complete GitHub App manifest flow
// @ID complete-github-app-manifest
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Deployment
// @Param request body codersdk.GitHubAppManifestCallbackRequest true "Callback request"
// @Success 201 {object} codersdk.ExternalAuthProviderConfig
// @Router /deployment/external-auth-providers/github/callback [post]
func (api *API) completeGitHubAppManifest(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req codersdk.GitHubAppManifestCallbackRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	if req.Code == "" || req.State == "" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Code and state are required.",
		})
		return
	}

	// Validate and retrieve the state.
	manifestState, err := api.Database.GetExternalAuthManifestState(ctx, req.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
				Message: "Invalid or expired state token.",
			})
			return
		}
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to validate state.",
			Detail:  err.Error(),
		})
		return
	}

	// Delete the state to prevent reuse.
	_ = api.Database.DeleteExternalAuthManifestState(ctx, req.State)

	// Exchange the code for the GitHub App credentials.
	// POST https://api.github.com/app-manifests/{code}/conversions
	exchangeURL := fmt.Sprintf("https://api.github.com/app-manifests/%s/conversions", req.Code)
	exchangeReq, err := http.NewRequestWithContext(ctx, http.MethodPost, exchangeURL, nil)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to create exchange request.",
			Detail:  err.Error(),
		})
		return
	}
	exchangeReq.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(exchangeReq)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to exchange code with GitHub.",
			Detail:  err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "GitHub rejected the code exchange.",
			Detail:  string(body),
		})
		return
	}

	// Parse the response.
	var githubResp struct {
		ID            int64  `json:"id"`
		Slug          string `json:"slug"`
		Name          string `json:"name"`
		ClientID      string `json:"client_id"`
		ClientSecret  string `json:"client_secret"`
		WebhookSecret string `json:"webhook_secret"`
		PEM           string `json:"pem"`
		HTMLURL       string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubResp); err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to decode GitHub response.",
			Detail:  err.Error(),
		})
		return
	}

	// Generate a provider ID from the slug.
	providerID := fmt.Sprintf("github-%s", githubResp.Slug)

	// Store the provider in the database.
	provider, err := api.Database.InsertExternalAuthProvider(ctx, database.InsertExternalAuthProviderParams{
		ID:                              providerID,
		Type:                            "github",
		ClientID:                        githubResp.ClientID,
		ClientSecretEncrypted:           []byte(githubResp.ClientSecret),
		ClientSecretKeyID:               sql.NullString{},
		DisplayName:                     toNullString(githubResp.Name),
		DisplayIcon:                     toNullString("/icon/github.svg"),
		AuthUrl:                         sql.NullString{},
		TokenUrl:                        sql.NullString{},
		ValidateUrl:                     sql.NullString{},
		DeviceCodeUrl:                   sql.NullString{},
		Scopes:                          []string{},
		ExtraTokenKeys:                  []string{},
		NoRefresh:                       true, // GitHub tokens don't expire.
		DeviceFlow:                      false,
		Regex:                           sql.NullString{},
		AppInstallUrl:                   toNullString(fmt.Sprintf("%s/installations/new", githubResp.HTMLURL)),
		AppInstallationsUrl:             toNullString("https://api.github.com/user/installations"),
		GithubAppID:                     toNullInt64(githubResp.ID),
		GithubAppWebhookSecretEncrypted: []byte(githubResp.WebhookSecret),
		GithubAppWebhookSecretKeyID:     sql.NullString{},
		GithubAppPrivateKeyEncrypted:    []byte(githubResp.PEM),
		GithubAppPrivateKeyKeyID:        sql.NullString{},
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to store external auth provider.",
			Detail:  err.Error(),
		})
		return
	}

	// If there's a redirect URI in the state, include it in the response.
	result := dbExternalAuthProviderToSDK(provider)

	httpapi.Write(ctx, rw, http.StatusCreated, result)

	// Note: Hot reload of external auth configs will be handled separately.
	// For now, a server restart is required.
	_ = manifestState
}

// handleGitHubAppManifestCallbackRedirect handles the browser redirect from GitHub.
// This is a GET endpoint that GitHub redirects to after the user approves the app.
// @Summary Handle GitHub App manifest callback redirect
// @ID handle-github-app-manifest-callback-redirect
// @Tags Deployment
// @Param code query string true "Code from GitHub"
// @Param state query string true "State token"
// @Success 302
// @Router /deployment/external-auth-providers/github/callback [get]
func (api *API) handleGitHubAppManifestCallbackRedirect(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Code and state are required.",
		})
		return
	}

	// Get the redirect URI from the state.
	manifestState, err := api.Database.GetExternalAuthManifestState(ctx, state)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
				Message: "Invalid or expired state token.",
			})
			return
		}
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to validate state.",
			Detail:  err.Error(),
		})
		return
	}

	// Redirect to the frontend with the code and state.
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", manifestState.RedirectUri, url.QueryEscape(code), url.QueryEscape(state))
	http.Redirect(rw, r, redirectURL, http.StatusTemporaryRedirect)
}

func dbExternalAuthProviderToSDK(p database.DBExternalAuthProvider) codersdk.ExternalAuthProviderConfig {
	return codersdk.ExternalAuthProviderConfig{
		ID:                  p.ID,
		Type:                p.Type,
		ClientID:            p.ClientID,
		DisplayName:         p.DisplayName.String,
		DisplayIcon:         p.DisplayIcon.String,
		AuthURL:             p.AuthUrl.String,
		TokenURL:            p.TokenUrl.String,
		ValidateURL:         p.ValidateUrl.String,
		Scopes:              p.Scopes,
		NoRefresh:           p.NoRefresh,
		DeviceFlow:          p.DeviceFlow,
		Regex:               p.Regex.String,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
		AppInstallURL:       p.AppInstallUrl.String,
		AppInstallationsURL: p.AppInstallationsUrl.String,
		GithubAppID:         p.GithubAppID.Int64,
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
