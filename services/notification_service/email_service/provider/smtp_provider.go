package provider

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/uloamaka/notification-service/email_service/models"
	"gopkg.in/gomail.v2"
	// "notification_service/services/email_service/models"
)

type SMTPProvider struct {
    host string
    port int
    user string
    pass string
    from string
}

func NewSMTPProvider() *SMTPProvider {
    _ = godotenv.Load()

    port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

    return &SMTPProvider{
        host: os.Getenv("SMTP_HOST"),
        port: port,
        user: os.Getenv("SMTP_USER"),
        pass: os.Getenv("SMTP_PASS"),
        from: os.Getenv("EMAIL_FROM"),
    }
}

func (s *SMTPProvider) Send(req models.EmailRequest) error {
    m := gomail.NewMessage()
    m.SetHeader("From", s.from)
    m.SetHeader("To", req.To)
    m.SetHeader("Subject", req.Subject)
    m.SetBody("text/html", req.HtmlBody)

    d := gomail.NewDialer(s.host, s.port, s.user, s.pass)
    if err := d.DialAndSend(m); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }
    return nil
}
