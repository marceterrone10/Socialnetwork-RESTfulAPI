package mailer

type Client struct {
	Send(templateField string, vars any) error

}