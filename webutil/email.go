package webutil

import (
	"fmt"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailMessage struct {
	FromEmail   string
	FromName    string
	ToEmail     string
	ToName      string
	Subject     string
	PlainText   string
	HTML        string
	Attachments []mail.Attachment
}

// EmailSender defines the interface for sending emails.
type EmailSender interface {
	SendEmail(msg ...EmailMessage) error
}

// SendGrid implements EmailSender for SendGrid.
type SendGrid struct {
	apiKey string
}

// NewSendGrid creates a new SendGrid.
func NewSendGrid(apiKey string) *SendGrid {
	return &SendGrid{apiKey: apiKey}
}

// SendEmail sends an email using the SendGrid API.
func (s *SendGrid) SendEmail(msgs ...EmailMessage) error {
	for _, msg := range msgs {
		from := mail.NewEmail(msg.FromName, msg.FromEmail)
		to := mail.NewEmail(msg.ToName, msg.ToEmail)
		message := mail.NewSingleEmail(from, msg.Subject, to, msg.PlainText, msg.HTML)

		client := sendgrid.NewSendClient(s.apiKey)
		response, err := client.Send(message)
		if err != nil {
			return fmt.Errorf("error sending email via SendGrid: %w", err)
		}

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			log.Printf("Email sent successfully! Status: %d", response.StatusCode)
			return nil
		}

		return fmt.Errorf("SendGrid API error: status %d, body %s", response.StatusCode, response.Body)
	}

	return nil
}
