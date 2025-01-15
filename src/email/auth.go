package email

import (
	"Unbewohnte/dela/conf"
	"net/smtp"
)

func Auth(conf conf.Conf) smtp.Auth {
	return smtp.PlainAuth(
		"",
		conf.Verification.Emailer.User,
		conf.Verification.Emailer.Password,
		conf.Verification.Emailer.Host,
	)
}
