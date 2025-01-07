package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Email struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

type MailSender struct {
	Auth smtp.Auth
	From string
}

func NewMailSender(auth smtp.Auth, from string) MailSender {
	return MailSender{
		Auth: auth,
		From: from,
	}
}

func buildEmail(mail Email) []byte {
	message := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	message += fmt.Sprintf("From: %s\r\n", mail.Sender)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	message += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	message += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return []byte(message)
}

func (ms *MailSender) SendEmail(addr string, mail Email) error {
	err := smtp.SendMail(addr, ms.Auth, ms.From, mail.To, buildEmail(mail))
	return err
}
