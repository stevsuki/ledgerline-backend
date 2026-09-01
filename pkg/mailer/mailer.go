// Package mailer: domain.Mailer implementation over SMTP.
package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/wneessen/go-mail"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

//go:embed templates/*.html
var templateFS embed.FS

type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromName    string
	FromAddress string
	AppName     string
	Timeout     time.Duration
	TLS         string // mandatory | opportunistic | none
}

type SMTPMailer struct {
	client *mail.Client
	cfg    Config
	tpl    *template.Template
}

// Compile-time check against the domain contract.
var _ domain.Mailer = (*SMTPMailer)(nil)

func NewSMTP(cfg Config) (*SMTPMailer, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.FromAddress == "" {
		return nil, fmt.Errorf("sender address is required")
	}

	tpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse template email: %w", err)
	}

	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithTimeout(cfg.Timeout),
		mail.WithTLSPolicy(tlsPolicy(cfg.TLS)),
	}

	// Dev servers like Mailpit use no authentication.
	if cfg.Username == "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthNoAuth))
	} else {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("create smtp client: %w", err)
	}
	return &SMTPMailer{client: client, cfg: cfg, tpl: tpl}, nil
}

func (m *SMTPMailer) SendPasswordResetOTP(ctx context.Context, data domain.PasswordResetOTPMail) error {
	view := map[string]any{
		"AppName":   m.cfg.AppName,
		"FullName":  data.FullName,
		"OTP":       data.OTP,
		"ExpiresIn": humanizeDuration(data.ExpiresIn),
	}

	html, err := m.render("password_reset_otp.html", view)
	if err != nil {
		return err
	}

	plain := fmt.Sprintf(
		"Hi %s,\n\nYour %s password reset code is: %s\n\n"+
			"This code is valid for %s and can only be used once.\n"+
			"Ignore this email if you did not request a password reset.\n",
		data.FullName, m.cfg.AppName, data.OTP, view["ExpiresIn"],
	)

	subject := fmt.Sprintf("%s - Your password reset code", m.cfg.AppName)

	return m.send(ctx, data.Email, subject, plain, html)
}

func (m *SMTPMailer) render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := m.tpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("render template %s: %w", name, err)
	}
	return buf.String(), nil
}

func (m *SMTPMailer) send(ctx context.Context, to, subject, plain, html string) error {
	msg := mail.NewMsg()
	if err := msg.FromFormat(m.cfg.FromName, m.cfg.FromAddress); err != nil {
		return fmt.Errorf("set pengirim: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set penerima: %w", err)
	}

	msg.Subject(subject)
	// The last part is the one email clients prefer.
	msg.SetBodyString(mail.TypeTextPlain, plain)
	msg.AddAlternativeString(mail.TypeTextHTML, html)

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func tlsPolicy(s string) mail.TLSPolicy {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return mail.NoTLS
	case "opportunistic":
		return mail.TLSOpportunistic
	default:
		return mail.TLSMandatory
	}
}

func humanizeDuration(d time.Duration) string {
	if d <= 0 {
		return "a few minutes"
	}
	if m := int(d.Minutes()); m < 60 {
		return fmt.Sprintf("%d minutes", m)
	}
	return fmt.Sprintf("%d hours", int(d.Hours()))
}
