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

func NewEmail(sender, subject, body string, to []string) Email {
	return Email{
		Sender:  sender,
		To:      to,
		Subject: subject,
		Body:    body,
	}
}

type Emailer struct {
	Auth    smtp.Auth
	Address string
	From    string
}

func NewEmailer(auth smtp.Auth, addr string, from string) *Emailer {
	return &Emailer{
		Auth:    auth,
		Address: addr,
		From:    from,
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

func (em *Emailer) SendEmail(mail Email) error {
	err := smtp.SendMail(em.Address, em.Auth, em.From, mail.To, buildEmail(mail))
	return err
}
