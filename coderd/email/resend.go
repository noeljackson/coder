package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ResendClient handles sending emails via Resend API.
type ResendClient struct {
	apiKey     string
	fromEmail  string
	fromName   string
	httpClient *http.Client
}

// ResendConfig holds configuration for Resend email service.
type ResendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

// NewResendClient creates a new Resend email client.
func NewResendClient(cfg ResendConfig) *ResendClient {
	return &ResendClient{
		apiKey:    cfg.APIKey,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendEmail sends an email via Resend API.
func (c *ResendClient) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	payload := map[string]interface{}{
		"from":    fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail),
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
		"text":    textBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			StatusCode int    `json:"statusCode"`
			Message    string `json:"message"`
			Name       string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("resend API error (status %d)", resp.StatusCode)
		}
		return fmt.Errorf("resend API error: %s - %s", errResp.Name, errResp.Message)
	}

	return nil
}

// WorkspaceInvitationData holds data for workspace invitation email.
type WorkspaceInvitationData struct {
	InviterName   string
	WorkspaceName string
	AccessLevel   string
	AcceptURL     string
	ExpiresAt     time.Time
}

// SendWorkspaceInvitation sends a workspace invitation email.
func (c *ResendClient) SendWorkspaceInvitation(ctx context.Context, toEmail string, data WorkspaceInvitationData) error {
	subject := fmt.Sprintf("%s invited you to collaborate on %s", data.InviterName, data.WorkspaceName)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .button { display: inline-block; background: #6366f1; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500; }
        .button:hover { background: #4f46e5; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 14px; color: #666; }
        .access-level { display: inline-block; background: #f3f4f6; padding: 4px 8px; border-radius: 4px; font-weight: 500; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Workspace Invitation</h1>
        </div>
        <p>Hi,</p>
        <p><strong>%s</strong> has invited you to collaborate on the workspace <strong>%s</strong>.</p>
        <p>Your access level: <span class="access-level">%s</span></p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" class="button">Accept Invitation</a>
        </p>
        <div class="footer">
            <p>This invitation expires on %s.</p>
            <p>If you didn't expect this invitation, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`, data.InviterName, data.WorkspaceName, data.AccessLevel, data.AcceptURL, data.ExpiresAt.Format("January 2, 2006"))

	text := fmt.Sprintf(`
Workspace Invitation

Hi,

%s has invited you to collaborate on the workspace "%s".

Your access level: %s

Accept the invitation by visiting:
%s

This invitation expires on %s.

If you didn't expect this invitation, you can safely ignore this email.
`, data.InviterName, data.WorkspaceName, data.AccessLevel, data.AcceptURL, data.ExpiresAt.Format("January 2, 2006"))

	return c.SendEmail(ctx, toEmail, subject, html, text)
}
