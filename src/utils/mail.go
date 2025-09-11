package utils

import (
	"fmt"
	"go-auth-system/src/config"

	"gopkg.in/gomail.v2"
)

type MailService struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

func NewMailService() *MailService {
	return &MailService{
		SMTPHost:     config.GetSMTPHost(),
		SMTPPort:     config.GetSMTPPort(),
		SMTPUsername: config.GetSMTPUsername(),
		SMTPPassword: config.GetSMTPPassword(),
		FromEmail:    config.GetSMTPUsername(),
	}
}

func (ms *MailService) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", ms.FromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(ms.SMTPHost, ms.SMTPPort, ms.SMTPUsername, ms.SMTPPassword)

	return d.DialAndSend(m)
}

func (ms *MailService) SendVerificationEmail(to, token string) error {
	verificationURL := fmt.Sprintf("http://localhost:8080/auth/verify?token=%s", token)

	tmpl := `
		<html>
		<body>
			<h2>Email Verification</h2>
			<p>Please click the link below to verify your email address:</p>
			<a href="%s">Verify Email</a>
			<p>This link will expire in 24 hours.</p>
		</body>
		</html>
	`

	body := fmt.Sprintf(tmpl, verificationURL)
	return ms.SendEmail(to, "Email Verification", body)
}

func (ms *MailService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	tmpl := `
		<html>
		<body>
			<h2>Password Reset</h2>
			<p>You requested a password reset. Click the link below to reset your password:</p>
			<a href="%s">Reset Password</a>
			<p>This link will expire in 1 hour.</p>
			<p>If you did not request this, please ignore this email.</p>
		</body>
		</html>
	`

	body := fmt.Sprintf(tmpl, resetURL)
	return ms.SendEmail(to, "Password Reset", body)
}
