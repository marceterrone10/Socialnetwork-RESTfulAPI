package mailer

import "embed"

const (
	FromName            = "Social Network"
	maxRetries          = 3
	userWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "template"
var FS embed.FS

type Client interface {
	Send(templateField, username string, email string, data any, isSandbox bool) error
}
