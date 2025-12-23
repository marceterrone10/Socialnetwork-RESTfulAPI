package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGridMailer(fromEmail, apiKey string) *SendGridMailer { // constructor del struct
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username string, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(username, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "template/"+templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := 0; i < maxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email (request error): %v", err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		if response.StatusCode < 200 || response.StatusCode >= 300 {
			log.Printf("SendGrid API error: status=%d body=%s", response.StatusCode, response.Body)
			return fmt.Errorf("sendgrid API returned status %d: %s", response.StatusCode, response.Body)
		}

		log.Printf("Email sent successfully: %v", response.StatusCode)
		return nil
	}
	return errors.New("failed to send email")
}
