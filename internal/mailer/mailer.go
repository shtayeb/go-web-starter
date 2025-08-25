package mailer

import (
	"bytes"
	"embed"
	"go-web-starter/internal/config"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer interface {
	Send(recipient, templateFile string, data interface{}) error
}

type AppMailer struct {
	dialer *mail.Dialer
	sender string
}

func New(smtp config.SMTP) AppMailer {
	dialer := mail.NewDialer(smtp.Host, smtp.Port, smtp.Username, smtp.Password)
	dialer.Timeout = 5 * time.Second

	return AppMailer{
		dialer: dialer,
		sender: smtp.Sender,
	}

}

func (m AppMailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Mail
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Call the DialAndSend() method on the dialer, passing in the message to send. This
	// opens a connection to the SMTP server, sends the message, then closes the
	// connection. If there is a timeout, it will return a "dial tcp: i/o timeout" error.

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
