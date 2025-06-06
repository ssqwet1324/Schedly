package email

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"sender_service/internal/entity"
)

type EmailService struct {
	From     string
	Host     string
	Port     int
	Username string
	Password string
}

func NewEmailService(from string, host string, port int, username string, password string) *EmailService {
	return &EmailService{
		From:     from,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

// SendEmail - отправка уведомления на почту
func (e *EmailService) SendEmail(message entity.EmailMessage) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", message.To)
	m.SetHeader("Subject", message.Subject)
	m.SetBody("text/plain", message.Content)

	fmt.Printf("Sending email to: %s, subject: %s\n", message.To, message.Subject)

	d := gomail.NewDialer(e.Host, e.Port, e.Username, e.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("sending email failed with gomail: %w", err)
	}

	return nil
}
