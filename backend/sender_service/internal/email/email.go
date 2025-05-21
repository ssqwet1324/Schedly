package email

import (
	"gopkg.in/gomail.v2"
)

type Email struct {
	From     string
	Host     string
	Port     int
	Username string
	Password string
}

func (e *Email) SendEmail(subject, to, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer(e.Host, e.Port, e.Username, e.Password)

	return d.DialAndSend(m)
}
