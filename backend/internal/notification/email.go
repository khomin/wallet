package notification

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
)

// EmailSender delivers transactional email. Keeping this interface small makes
// alert delivery independent from the SMTP implementation and easy to test.
type EmailSender interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

// SMTPSender sends mail through an SMTP relay. The relay is expected to handle
// authentication and delivery policy; this service only submits the message.
type SMTPSender struct {
	host string
	port int
	from string
}

func NewSMTPSender(host string, port int, from string) (*SMTPSender, error) {
	if strings.TrimSpace(host) == "" {
		return nil, fmt.Errorf("smtp host is required")
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid smtp port %d", port)
	}
	if _, err := mail.ParseAddress(from); err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}
	return &SMTPSender{host: host, port: port, from: from}, nil
}

func (s *SMTPSender) Send(ctx context.Context, recipient, subject, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	address, err := mail.ParseAddress(recipient)
	if err != nil || address.Address != recipient {
		return fmt.Errorf("invalid recipient address %q", recipient)
	}
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("email subject contains a newline")
	}

	message := strings.Join([]string{
		"From: " + s.from,
		"To: " + recipient,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	// smtp.SendMail has no context-aware API. Checking before submission still
	// prevents sends after cancellation; the relay connection remains bounded
	// by the process/network timeouts configured for the deployment.
	return smtp.SendMail(
		fmt.Sprintf("%s:%d", s.host, s.port),
		nil, // s.host when use TLS
		s.from,
		[]string{recipient},
		[]byte(message),
	)
}
