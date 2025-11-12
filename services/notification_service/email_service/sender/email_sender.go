package sender

import (
	"github.com/uloamaka/notification-service/email_service/models"
	"github.com/uloamaka/notification-service/email_service/provider"
)

type EmailSender interface {
    SendEmail(job interface{}) error
}

type smtpSender struct {
    provider provider.EmailProvider
}

func NewSMTPProvider() EmailSender {
    return &smtpSender{provider: provider.NewSMTPProvider()}
}

func (s *smtpSender) SendEmail(job interface{}) error {
    j := job.(models.EmailRequest)
    return s.provider.Send(j)
}
