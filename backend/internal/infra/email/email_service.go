package email

import (
	"context"
	"fmt"

	"github.com/Faithful001/aegis/internal/domain/otp"
	"github.com/resend/resend-go/v4"
)

type IEmailService interface {
	SendOTPEmail(ctx context.Context, toEmail, name, otpCode string, purpose otp.OTPPurpose) error
	SendAlertEmail(ctx context.Context, toEmail, subject, htmlBody string) error
}

type EmailService struct {
	client   *resend.Client
	fromAddr string
}

func NewEmailService(apiKey, fromAddr string) IEmailService {
	return &EmailService{
		client:   resend.NewClient(apiKey),
		fromAddr: fromAddr,
	}
}

func (e *EmailService) SendOTPEmail(ctx context.Context, toEmail, name, otpCode string, purpose otp.OTPPurpose) error {
	var subject, body string

	switch purpose {
	case otp.OTPPurposeEmailVerification:
		subject = "Verify your Aegis email"
		body = emailVerificationHTML(name, otpCode)
	case otp.OTPPurposePasswordReset:
		subject = "Reset your Aegis password"
		body = passwordResetHTML(name, otpCode)
	default:
		subject = "Your Aegis OTP"
		body = fmt.Sprintf("<p>Your OTP is: <strong>%s</strong>. It expires in 10 minutes.</p>", otpCode)
	}

	params := &resend.SendEmailRequest{
		From:    e.fromAddr,
		To:      []string{toEmail},
		Subject: subject,
		Html:    body,
	}

	_, err := e.client.Emails.Send(params)
	return err
}

func (e *EmailService) SendAlertEmail(ctx context.Context, toEmail, subject, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    e.fromAddr,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := e.client.Emails.Send(params)
	return err
}

func emailVerificationHTML(name, otpCode string) string {
	displayName := name
	if displayName == "" {
		displayName = "there"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#09090b;font-family:'Inter',sans-serif;color:#e4e4e7;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 20px;">
    <tr><td align="center">
      <table width="480" cellpadding="0" cellspacing="0" style="background:#18181b;border-radius:16px;border:1px solid #27272a;overflow:hidden;">
        <tr>
          <td style="padding:32px 40px 24px;border-bottom:1px solid #27272a;">
            <div style="display:inline-flex;align-items:center;gap:12px;">
              <div style="width:36px;height:36px;background:#27272a;border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:20px;font-weight:bold;">Ø</div>
              <span style="font-size:18px;font-weight:700;letter-spacing:-0.5px;color:#fff;">Aegis</span>
            </div>
          </td>
        </tr>
        <tr>
          <td style="padding:32px 40px;">
            <h2 style="margin:0 0 8px;font-size:22px;font-weight:700;color:#fff;">Verify your email</h2>
            <p style="margin:0 0 28px;color:#a1a1aa;font-size:15px;">Hi %s, use the code below to verify your email address. This code expires in <strong style="color:#e4e4e7;">10 minutes</strong>.</p>
            <div style="background:#09090b;border:1px solid #27272a;border-radius:12px;padding:24px;text-align:center;margin-bottom:28px;">
              <span style="font-family:'Courier New',monospace;font-size:36px;font-weight:700;letter-spacing:12px;color:#fff;">%s</span>
            </div>
            <p style="margin:0;color:#71717a;font-size:13px;">If you didn't request this, you can safely ignore this email.</p>
          </td>
        </tr>
        <tr>
          <td style="padding:20px 40px;border-top:1px solid #27272a;">
            <p style="margin:0;color:#52525b;font-size:12px;">Aegis AI Inference Platform &bull; Sent via Resend</p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, displayName, otpCode)
}

func passwordResetHTML(name, otpCode string) string {
	displayName := name
	if displayName == "" {
		displayName = "there"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#09090b;font-family:'Inter',sans-serif;color:#e4e4e7;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 20px;">
    <tr><td align="center">
      <table width="480" cellpadding="0" cellspacing="0" style="background:#18181b;border-radius:16px;border:1px solid #27272a;overflow:hidden;">
        <tr>
          <td style="padding:32px 40px 24px;border-bottom:1px solid #27272a;">
            <div style="display:inline-flex;align-items:center;gap:12px;">
              <div style="width:36px;height:36px;background:#27272a;border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:20px;font-weight:bold;">Ø</div>
              <span style="font-size:18px;font-weight:700;letter-spacing:-0.5px;color:#fff;">Aegis</span>
            </div>
          </td>
        </tr>
        <tr>
          <td style="padding:32px 40px;">
            <h2 style="margin:0 0 8px;font-size:22px;font-weight:700;color:#fff;">Reset your password</h2>
            <p style="margin:0 0 28px;color:#a1a1aa;font-size:15px;">Hi %s, use the code below to reset your password. This code expires in <strong style="color:#e4e4e7;">10 minutes</strong>.</p>
            <div style="background:#09090b;border:1px solid #27272a;border-radius:12px;padding:24px;text-align:center;margin-bottom:28px;">
              <span style="font-family:'Courier New',monospace;font-size:36px;font-weight:700;letter-spacing:12px;color:#fff;">%s</span>
            </div>
            <p style="margin:0 0 12px;color:#71717a;font-size:13px;">This code can only be used once and will expire in 10 minutes.</p>
            <p style="margin:0;color:#71717a;font-size:13px;">If you didn't request a password reset, you can safely ignore this email.</p>
          </td>
        </tr>
        <tr>
          <td style="padding:20px 40px;border-top:1px solid #27272a;">
            <p style="margin:0;color:#52525b;font-size:12px;">Aegis AI Inference Platform &bull; Sent via Resend</p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, displayName, otpCode)
}
