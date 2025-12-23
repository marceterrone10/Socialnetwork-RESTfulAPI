package mailer

import "embed"

const (
	FromName            = "Social Network"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "template"
var FS embed.FS

type Client interface {
	Send(templateFile, username string, email string, data any, isSandbox bool) error
}
