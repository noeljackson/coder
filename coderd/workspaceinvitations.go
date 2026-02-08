package coderd

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"cdr.dev/slog/v3"

	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbtime"
	"github.com/coder/coder/v2/coderd/email"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/coderd/httpmw"
	"github.com/coder/coder/v2/codersdk"
)

const (
	// invitationTokenLength is the length of the invitation token.
	invitationTokenLength = 32
	// invitationExpiryDays is how long an invitation is valid.
	invitationExpiryDays = 7
)

// generateInvitationToken generates a secure random token for workspace invitations.
func generateInvitationToken() (string, error) {
	bytes := make([]byte, invitationTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// @Summary Create workspace invitation
// @ID create-workspace-invitation
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Workspaces
// @Param workspace path string true "Workspace ID" format(uuid)
// @Param request body codersdk.CreateWorkspaceInvitationRequest true "Invitation request"
// @Success 201 {object} codersdk.WorkspaceInvitation
// @Router /workspaces/{workspace}/invitations [post]
func (api *API) createWorkspaceInvitation(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspace := httpmw.WorkspaceParam(r)
	apiKey := httpmw.APIKey(r)

	var req codersdk.CreateWorkspaceInvitationRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	// Generate secure token
	token, err := generateInvitationToken()
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to generate invitation token.",
			Detail:  err.Error(),
		})
		return
	}

	expiresAt := dbtime.Now().Add(invitationExpiryDays * 24 * time.Hour)

	invitation, err := api.Database.CreateWorkspaceInvitation(ctx, database.CreateWorkspaceInvitationParams{
		WorkspaceID: workspace.ID,
		InviterID:   apiKey.UserID,
		Email:       req.Email,
		AccessLevel: string(req.AccessLevel),
		Token:       token,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to create invitation.",
			Detail:  err.Error(),
		})
		return
	}

	// Send invitation email if Resend is configured
	if resendAPIKey := os.Getenv("CODER_RESEND_API_KEY"); resendAPIKey != "" {
		inviter, _ := api.Database.GetUserByID(ctx, apiKey.UserID)
		inviterName := "A team member"
		if inviter.Name != "" {
			inviterName = inviter.Name
		} else if inviter.Username != "" {
			inviterName = inviter.Username
		}

		accessURL := os.Getenv("CODER_ACCESS_URL")
		if accessURL == "" {
			accessURL = "https://coder.example.com"
		}
		acceptURL := fmt.Sprintf("%s/invitation/%s", accessURL, token)

		emailClient := email.NewResendClient(email.ResendConfig{
			APIKey:    resendAPIKey,
			FromEmail: os.Getenv("CODER_EMAIL_FROM"),
			FromName:  "Coder",
		})

		err := emailClient.SendWorkspaceInvitation(ctx, req.Email, email.WorkspaceInvitationData{
			InviterName:   inviterName,
			WorkspaceName: workspace.Name,
			AccessLevel:   string(req.AccessLevel),
			AcceptURL:     acceptURL,
			ExpiresAt:     expiresAt,
		})
		if err != nil {
			// Log the error but don't fail the request
			api.Logger.Warn(ctx, "failed to send invitation email",
				slog.Error(err),
				slog.F("email", req.Email),
				slog.F("workspace", workspace.Name),
			)
		}
	}

	httpapi.Write(ctx, rw, http.StatusCreated, convertWorkspaceInvitation(invitation, token))
}

// @Summary List workspace invitations
// @ID list-workspace-invitations
// @Security CoderSessionToken
// @Produce json
// @Tags Workspaces
// @Param workspace path string true "Workspace ID" format(uuid)
// @Success 200 {array} codersdk.WorkspaceInvitation
// @Router /workspaces/{workspace}/invitations [get]
func (api *API) listWorkspaceInvitations(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspace := httpmw.WorkspaceParam(r)

	invitations, err := api.Database.GetWorkspaceInvitationsByWorkspaceID(ctx, workspace.ID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to list invitations.",
			Detail:  err.Error(),
		})
		return
	}

	result := make([]codersdk.WorkspaceInvitation, 0, len(invitations))
	for _, inv := range invitations {
		// Don't include the token in list responses
		result = append(result, convertWorkspaceInvitation(inv, ""))
	}

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Delete workspace invitation
// @ID delete-workspace-invitation
// @Security CoderSessionToken
// @Tags Workspaces
// @Param workspace path string true "Workspace ID" format(uuid)
// @Param invitation path string true "Invitation ID" format(uuid)
// @Success 204
// @Router /workspaces/{workspace}/invitations/{invitation} [delete]
func (api *API) deleteWorkspaceInvitation(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	invitationIDStr := chi.URLParam(r, "invitation")

	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid invitation ID.",
		})
		return
	}

	invitation, err := api.Database.GetWorkspaceInvitationByID(ctx, invitationID)
	if errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Invitation not found.",
		})
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get invitation.",
			Detail:  err.Error(),
		})
		return
	}

	// Update status to canceled
	_, err = api.Database.UpdateWorkspaceInvitationStatus(ctx, database.UpdateWorkspaceInvitationStatusParams{
		ID:     invitation.ID,
		Status: "canceled",
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to cancel invitation.",
			Detail:  err.Error(),
		})
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// @Summary Get invitation by token
// @ID get-workspace-invitation-by-token
// @Produce json
// @Tags Workspaces
// @Param token path string true "Invitation token"
// @Success 200 {object} codersdk.WorkspaceInvitation
// @Router /invitations/{token} [get]
func (api *API) getWorkspaceInvitationByToken(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := chi.URLParam(r, "token")

	invitation, err := api.Database.GetWorkspaceInvitationByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Invitation not found.",
		})
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get invitation.",
			Detail:  err.Error(),
		})
		return
	}

	// Get workspace and inviter details for the response
	workspace, err := api.Database.GetWorkspaceByID(ctx, invitation.WorkspaceID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get workspace details.",
			Detail:  err.Error(),
		})
		return
	}

	inviter, err := api.Database.GetUserByID(ctx, invitation.InviterID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get inviter details.",
			Detail:  err.Error(),
		})
		return
	}

	result := convertWorkspaceInvitation(invitation, "")
	result.WorkspaceName = workspace.Name
	result.InviterUsername = inviter.Username

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Accept workspace invitation
// @ID accept-workspace-invitation
// @Security CoderSessionToken
// @Produce json
// @Tags Workspaces
// @Param token path string true "Invitation token"
// @Success 200 {object} codersdk.WorkspaceCollaborator
// @Router /invitations/{token}/accept [post]
func (api *API) acceptWorkspaceInvitation(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := chi.URLParam(r, "token")
	apiKey := httpmw.APIKey(r)

	invitation, err := api.Database.GetWorkspaceInvitationByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Invitation not found.",
		})
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get invitation.",
			Detail:  err.Error(),
		})
		return
	}

	// Check if invitation is still valid
	if invitation.Status != "pending" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: fmt.Sprintf("Invitation is %s.", invitation.Status),
		})
		return
	}
	if invitation.ExpiresAt.Before(dbtime.Now()) {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invitation has expired.",
		})
		return
	}

	// Get current user's email to verify it matches the invitation
	user, err := api.Database.GetUserByID(ctx, apiKey.UserID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get user.",
			Detail:  err.Error(),
		})
		return
	}

	// Email check (optional - could allow any authenticated user)
	if user.Email != invitation.Email {
		httpapi.Write(ctx, rw, http.StatusForbidden, codersdk.Response{
			Message: "This invitation was sent to a different email address.",
		})
		return
	}

	// Check if already a collaborator
	_, err = api.Database.GetWorkspaceCollaboratorByUserAndWorkspace(ctx, database.GetWorkspaceCollaboratorByUserAndWorkspaceParams{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      apiKey.UserID,
	})
	if err == nil {
		httpapi.Write(ctx, rw, http.StatusConflict, codersdk.Response{
			Message: "You are already a collaborator on this workspace.",
		})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to check existing collaboration.",
			Detail:  err.Error(),
		})
		return
	}

	// Create collaborator
	collaborator, err := api.Database.CreateWorkspaceCollaborator(ctx, database.CreateWorkspaceCollaboratorParams{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      apiKey.UserID,
		AccessLevel: invitation.AccessLevel,
		InvitedBy:   uuid.NullUUID{UUID: invitation.InviterID, Valid: true},
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to create collaboration.",
			Detail:  err.Error(),
		})
		return
	}

	// Update invitation status
	_, err = api.Database.UpdateWorkspaceInvitationStatus(ctx, database.UpdateWorkspaceInvitationStatusParams{
		ID:     invitation.ID,
		Status: "accepted",
	})
	if err != nil {
		// Log but don't fail - collaborator was created
		api.Logger.Warn(ctx, "failed to update invitation status", slog.Error(err))
	}

	result := convertWorkspaceCollaborator(collaborator)
	result.Username = user.Username
	result.Email = user.Email
	result.AvatarURL = user.AvatarURL

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Decline workspace invitation
// @ID decline-workspace-invitation
// @Security CoderSessionToken
// @Tags Workspaces
// @Param token path string true "Invitation token"
// @Success 204
// @Router /invitations/{token}/decline [post]
func (api *API) declineWorkspaceInvitation(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := chi.URLParam(r, "token")

	invitation, err := api.Database.GetWorkspaceInvitationByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Invitation not found.",
		})
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get invitation.",
			Detail:  err.Error(),
		})
		return
	}

	if invitation.Status != "pending" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: fmt.Sprintf("Invitation is already %s.", invitation.Status),
		})
		return
	}

	_, err = api.Database.UpdateWorkspaceInvitationStatus(ctx, database.UpdateWorkspaceInvitationStatusParams{
		ID:     invitation.ID,
		Status: "declined",
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to decline invitation.",
			Detail:  err.Error(),
		})
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// @Summary List workspace collaborators
// @ID list-workspace-collaborators
// @Security CoderSessionToken
// @Produce json
// @Tags Workspaces
// @Param workspace path string true "Workspace ID" format(uuid)
// @Success 200 {array} codersdk.WorkspaceCollaborator
// @Router /workspaces/{workspace}/collaborators [get]
func (api *API) listWorkspaceCollaborators(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspace := httpmw.WorkspaceParam(r)

	collaborators, err := api.Database.GetWorkspaceCollaboratorsByWorkspaceID(ctx, workspace.ID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to list collaborators.",
			Detail:  err.Error(),
		})
		return
	}

	// Get user details for each collaborator
	result := make([]codersdk.WorkspaceCollaborator, 0, len(collaborators))
	for _, collab := range collaborators {
		user, err := api.Database.GetUserByID(ctx, collab.UserID)
		if err != nil {
			continue // Skip if user not found
		}
		c := convertWorkspaceCollaborator(collab)
		c.Username = user.Username
		c.Email = user.Email
		c.AvatarURL = user.AvatarURL
		result = append(result, c)
	}

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Update workspace collaborator access level
// @ID update-workspace-collaborator
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Workspaces
// @Param workspace path string true "Workspace ID" format(uuid)
// @Param collaborator path string true "Collaborator ID" format(uuid)
// @Param request body codersdk.UpdateWorkspaceCollaboratorRequest true "Update request"
// @Success 200 {object} codersdk.WorkspaceCollaborator
// @Router /workspaces/{workspace}/collaborators/{collaborator} [patch]
func (api *API) updateWorkspaceCollaborator(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collaboratorIDStr := chi.URLParam(r, "collaborator")

	collaboratorID, err := uuid.Parse(collaboratorIDStr)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid collaborator ID.",
		})
		return
	}

	var req codersdk.UpdateWorkspaceCollaboratorRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	collaborator, err := api.Database.UpdateWorkspaceCollaboratorAccessLevel(ctx, database.UpdateWorkspaceCollaboratorAccessLevelParams{
		ID:          collaboratorID,
		AccessLevel: string(req.AccessLevel),
	})
	if errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Collaborator not found.",
		})
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to update collaborator.",
			Detail:  err.Error(),
		})
		return
	}

	// Get user details
	user, err := api.Database.GetUserByID(ctx, collaborator.UserID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get user details.",
			Detail:  err.Error(),
		})
		return
	}

	result := convertWorkspaceCollaborator(collaborator)
	result.Username = user.Username
	result.Email = user.Email
	result.AvatarURL = user.AvatarURL

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Remove workspace collaborator
// @ID delete-workspace-collaborator
// @Security CoderSessionToken
// @Tags Workspaces
// @Param workspace path string true "Workspace ID" format(uuid)
// @Param collaborator path string true "Collaborator ID" format(uuid)
// @Success 204
// @Router /workspaces/{workspace}/collaborators/{collaborator} [delete]
func (api *API) deleteWorkspaceCollaborator(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collaboratorIDStr := chi.URLParam(r, "collaborator")

	collaboratorID, err := uuid.Parse(collaboratorIDStr)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid collaborator ID.",
		})
		return
	}

	err = api.Database.DeleteWorkspaceCollaborator(ctx, collaboratorID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to remove collaborator.",
			Detail:  err.Error(),
		})
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// @Summary Get current user's workspace collaborations
// @ID get-my-workspace-collaborations
// @Security CoderSessionToken
// @Produce json
// @Tags Users
// @Success 200 {array} codersdk.WorkspaceCollaborator
// @Router /users/me/workspace-collaborations [get]
func (api *API) getMyWorkspaceCollaborations(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey := httpmw.APIKey(r)

	collaborations, err := api.Database.GetWorkspaceCollaborationsByUserID(ctx, apiKey.UserID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get collaborations.",
			Detail:  err.Error(),
		})
		return
	}

	result := make([]codersdk.WorkspaceCollaborator, 0, len(collaborations))
	for _, collab := range collaborations {
		result = append(result, convertWorkspaceCollaborator(collab))
	}

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// @Summary Get current user's pending workspace invitations
// @ID get-my-pending-workspace-invitations
// @Security CoderSessionToken
// @Produce json
// @Tags Users
// @Success 200 {array} codersdk.WorkspaceInvitation
// @Router /users/me/workspace-invitations [get]
func (api *API) getMyPendingInvitations(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey := httpmw.APIKey(r)

	// Get user email
	user, err := api.Database.GetUserByID(ctx, apiKey.UserID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get user.",
			Detail:  err.Error(),
		})
		return
	}

	invitations, err := api.Database.GetPendingWorkspaceInvitationsByEmail(ctx, user.Email)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get invitations.",
			Detail:  err.Error(),
		})
		return
	}

	result := make([]codersdk.WorkspaceInvitation, 0, len(invitations))
	for _, inv := range invitations {
		// Get workspace and inviter details
		workspace, err := api.Database.GetWorkspaceByID(ctx, inv.WorkspaceID)
		if err != nil {
			continue
		}
		inviter, err := api.Database.GetUserByID(ctx, inv.InviterID)
		if err != nil {
			continue
		}

		converted := convertWorkspaceInvitation(inv, "")
		converted.WorkspaceName = workspace.Name
		converted.InviterUsername = inviter.Username
		result = append(result, converted)
	}

	httpapi.Write(ctx, rw, http.StatusOK, result)
}

// Helper conversion functions

func convertWorkspaceInvitation(inv database.WorkspaceInvitation, token string) codersdk.WorkspaceInvitation {
	result := codersdk.WorkspaceInvitation{
		ID:          inv.ID,
		WorkspaceID: inv.WorkspaceID,
		InviterID:   inv.InviterID,
		Email:       inv.Email,
		AccessLevel: codersdk.WorkspaceAccessLevel(inv.AccessLevel),
		Status:      codersdk.WorkspaceInvitationStatus(inv.Status),
		ExpiresAt:   inv.ExpiresAt,
		CreatedAt:   inv.CreatedAt,
	}
	if token != "" {
		result.Token = token
	}
	if inv.RespondedAt.Valid {
		result.RespondedAt = &inv.RespondedAt.Time
	}
	return result
}

func convertWorkspaceCollaborator(collab database.WorkspaceCollaborator) codersdk.WorkspaceCollaborator {
	result := codersdk.WorkspaceCollaborator{
		ID:          collab.ID,
		WorkspaceID: collab.WorkspaceID,
		UserID:      collab.UserID,
		AccessLevel: codersdk.WorkspaceAccessLevel(collab.AccessLevel),
		CreatedAt:   collab.CreatedAt,
	}
	if collab.InvitedBy.Valid {
		result.InvitedBy = &collab.InvitedBy.UUID
	}
	return result
}
