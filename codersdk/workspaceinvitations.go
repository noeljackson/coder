package codersdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// WorkspaceAccessLevel defines the level of access for workspace collaborators.
type WorkspaceAccessLevel string

const (
	WorkspaceAccessLevelReadonly WorkspaceAccessLevel = "readonly"
	WorkspaceAccessLevelUse      WorkspaceAccessLevel = "use"
	WorkspaceAccessLevelAdmin    WorkspaceAccessLevel = "admin"
)

// WorkspaceInvitationStatus defines the status of a workspace invitation.
type WorkspaceInvitationStatus string

const (
	WorkspaceInvitationStatusPending  WorkspaceInvitationStatus = "pending"
	WorkspaceInvitationStatusAccepted WorkspaceInvitationStatus = "accepted"
	WorkspaceInvitationStatusDeclined WorkspaceInvitationStatus = "declined"
	WorkspaceInvitationStatusExpired  WorkspaceInvitationStatus = "expired"
	WorkspaceInvitationStatusCanceled WorkspaceInvitationStatus = "canceled"
)

// WorkspaceInvitation represents an invitation to collaborate on a workspace.
type WorkspaceInvitation struct {
	ID          uuid.UUID                 `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID                 `json:"workspace_id" format:"uuid"`
	InviterID   uuid.UUID                 `json:"inviter_id" format:"uuid"`
	Email       string                    `json:"email"`
	AccessLevel WorkspaceAccessLevel      `json:"access_level"`
	Token       string                    `json:"token,omitempty"` // Only shown on creation
	Status      WorkspaceInvitationStatus `json:"status"`
	ExpiresAt   time.Time                 `json:"expires_at" format:"date-time"`
	CreatedAt   time.Time                 `json:"created_at" format:"date-time"`
	RespondedAt *time.Time                `json:"responded_at,omitempty" format:"date-time"`

	// Populated fields
	InviterUsername string `json:"inviter_username,omitempty"`
	WorkspaceName   string `json:"workspace_name,omitempty"`
}

// WorkspaceCollaborator represents a user with access to a workspace.
type WorkspaceCollaborator struct {
	ID          uuid.UUID            `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID            `json:"workspace_id" format:"uuid"`
	UserID      uuid.UUID            `json:"user_id" format:"uuid"`
	AccessLevel WorkspaceAccessLevel `json:"access_level"`
	InvitedBy   *uuid.UUID           `json:"invited_by,omitempty" format:"uuid"`
	CreatedAt   time.Time            `json:"created_at" format:"date-time"`

	// Populated fields
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// CreateWorkspaceInvitationRequest is the request body for creating a workspace invitation.
type CreateWorkspaceInvitationRequest struct {
	Email       string               `json:"email" validate:"required,email"`
	AccessLevel WorkspaceAccessLevel `json:"access_level" validate:"required,oneof=readonly use admin"`
}

// UpdateWorkspaceCollaboratorRequest is the request body for updating a collaborator's access level.
type UpdateWorkspaceCollaboratorRequest struct {
	AccessLevel WorkspaceAccessLevel `json:"access_level" validate:"required,oneof=readonly use admin"`
}

// CreateWorkspaceInvitation creates a new invitation for a workspace.
func (c *Client) CreateWorkspaceInvitation(ctx context.Context, workspaceID uuid.UUID, req CreateWorkspaceInvitationRequest) (WorkspaceInvitation, error) {
	res, err := c.Request(ctx, http.MethodPost, fmt.Sprintf("/api/v2/workspaces/%s/invitations", workspaceID), req)
	if err != nil {
		return WorkspaceInvitation{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return WorkspaceInvitation{}, ReadBodyAsError(res)
	}
	var invitation WorkspaceInvitation
	return invitation, json.NewDecoder(res.Body).Decode(&invitation)
}

// GetWorkspaceInvitations returns all invitations for a workspace.
func (c *Client) GetWorkspaceInvitations(ctx context.Context, workspaceID uuid.UUID) ([]WorkspaceInvitation, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/workspaces/%s/invitations", workspaceID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ReadBodyAsError(res)
	}
	var invitations []WorkspaceInvitation
	return invitations, json.NewDecoder(res.Body).Decode(&invitations)
}

// DeleteWorkspaceInvitation cancels a pending invitation.
func (c *Client) DeleteWorkspaceInvitation(ctx context.Context, workspaceID, invitationID uuid.UUID) error {
	res, err := c.Request(ctx, http.MethodDelete, fmt.Sprintf("/api/v2/workspaces/%s/invitations/%s", workspaceID, invitationID), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return ReadBodyAsError(res)
	}
	return nil
}

// AcceptWorkspaceInvitation accepts an invitation using the invitation token.
func (c *Client) AcceptWorkspaceInvitation(ctx context.Context, token string) (WorkspaceCollaborator, error) {
	res, err := c.Request(ctx, http.MethodPost, fmt.Sprintf("/api/v2/invitations/%s/accept", token), nil)
	if err != nil {
		return WorkspaceCollaborator{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return WorkspaceCollaborator{}, ReadBodyAsError(res)
	}
	var collaborator WorkspaceCollaborator
	return collaborator, json.NewDecoder(res.Body).Decode(&collaborator)
}

// DeclineWorkspaceInvitation declines an invitation using the invitation token.
func (c *Client) DeclineWorkspaceInvitation(ctx context.Context, token string) error {
	res, err := c.Request(ctx, http.MethodPost, fmt.Sprintf("/api/v2/invitations/%s/decline", token), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return ReadBodyAsError(res)
	}
	return nil
}

// GetWorkspaceInvitationByToken retrieves invitation details using the token.
func (c *Client) GetWorkspaceInvitationByToken(ctx context.Context, token string) (WorkspaceInvitation, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/invitations/%s", token), nil)
	if err != nil {
		return WorkspaceInvitation{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return WorkspaceInvitation{}, ReadBodyAsError(res)
	}
	var invitation WorkspaceInvitation
	return invitation, json.NewDecoder(res.Body).Decode(&invitation)
}

// GetWorkspaceCollaborators returns all collaborators for a workspace.
func (c *Client) GetWorkspaceCollaborators(ctx context.Context, workspaceID uuid.UUID) ([]WorkspaceCollaborator, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/workspaces/%s/collaborators", workspaceID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ReadBodyAsError(res)
	}
	var collaborators []WorkspaceCollaborator
	return collaborators, json.NewDecoder(res.Body).Decode(&collaborators)
}

// UpdateWorkspaceCollaborator updates a collaborator's access level.
func (c *Client) UpdateWorkspaceCollaborator(ctx context.Context, workspaceID, collaboratorID uuid.UUID, req UpdateWorkspaceCollaboratorRequest) (WorkspaceCollaborator, error) {
	res, err := c.Request(ctx, http.MethodPatch, fmt.Sprintf("/api/v2/workspaces/%s/collaborators/%s", workspaceID, collaboratorID), req)
	if err != nil {
		return WorkspaceCollaborator{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return WorkspaceCollaborator{}, ReadBodyAsError(res)
	}
	var collaborator WorkspaceCollaborator
	return collaborator, json.NewDecoder(res.Body).Decode(&collaborator)
}

// DeleteWorkspaceCollaborator removes a collaborator from a workspace.
func (c *Client) DeleteWorkspaceCollaborator(ctx context.Context, workspaceID, collaboratorID uuid.UUID) error {
	res, err := c.Request(ctx, http.MethodDelete, fmt.Sprintf("/api/v2/workspaces/%s/collaborators/%s", workspaceID, collaboratorID), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return ReadBodyAsError(res)
	}
	return nil
}

// GetMyWorkspaceCollaborations returns workspaces the current user is a collaborator on.
func (c *Client) GetMyWorkspaceCollaborations(ctx context.Context) ([]WorkspaceCollaborator, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/users/me/workspace-collaborations", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ReadBodyAsError(res)
	}
	var collaborations []WorkspaceCollaborator
	return collaborations, json.NewDecoder(res.Body).Decode(&collaborations)
}

// GetMyPendingInvitations returns pending workspace invitations for the current user.
func (c *Client) GetMyPendingInvitations(ctx context.Context) ([]WorkspaceInvitation, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/users/me/workspace-invitations", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ReadBodyAsError(res)
	}
	var invitations []WorkspaceInvitation
	return invitations, json.NewDecoder(res.Body).Decode(&invitations)
}
