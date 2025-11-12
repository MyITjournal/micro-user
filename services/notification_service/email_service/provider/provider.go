package provider

import "github.com/uloamaka/notification-service/email_service/models"


type EmailProvider interface {
    Send(req models.EmailRequest) error
}