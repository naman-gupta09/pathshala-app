package utils

import (
	"fmt"
	"html"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendResetEmailGrid(toEmail string, resetLink string) error {
	from := mail.NewEmail("Pathashala", os.Getenv("EMAIL_FROM")) // e.g., noreply@yourdomain.com
	subject := "Reset Your Password"
	to := mail.NewEmail("User", toEmail)

	safeLink := html.EscapeString(resetLink) // This is important

	// Plain text fallback for email clients that don't support HTML
	plainTextContent := fmt.Sprintf(`
		Hi there,

		We received a request to reset your Pathashala account password.

		Click the link below to reset your password:
		%s

		If you didn’t request this, you can ignore this email.

		Thanks,
		The Pathashala Team
	`, resetLink)

	// Branded, styled HTML content
	htmlContent := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #ddd; border-radius: 10px;">
		<h2 style="color: #333;">Reset Your Password</h2>
		<p>Hi there,</p>
		<p>We received a request to reset your <strong>Pathashala</strong> account password. Click the button below to reset it:</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background-color: #007BFF; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px;">Reset Password</a>
		</p>
		<p>If you didn’t request this, you can safely ignore this email.</p>
		<p style="margin-top: 40px;">Thanks,<br><strong>The Pathashala Team</strong></p>
		</div>
`, safeLink)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}

	return nil
}
