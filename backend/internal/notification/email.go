package notification

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/resend/resend-go/v3"
)

type EmailSender interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

type emailSender struct {
	client *resend.Client
	from   string
}

func NewEmailSender(apiKey string, from string) (EmailSender, error) {
	return &emailSender{
		client: resend.NewClient(apiKey),
		from:   from,
	}, nil
}

func (e *emailSender) Send(ctx context.Context, recipient string, subject string, body string) error {
	address, err := mail.ParseAddress(recipient)
	if err != nil || address.Address != recipient {
		return fmt.Errorf("invalid recipient address %q", recipient)
	}
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("email subject contains a newline")
	}
	_, err = e.client.Emails.Send(&resend.SendEmailRequest{
		From:    e.from,
		To:      []string{recipient},
		Subject: subject,
		Html:    body,
	})
	if err != nil {
		return err
	}
	return nil
}
