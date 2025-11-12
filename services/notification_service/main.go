package main

import (
	"github.com/uloamaka/notification-service/email_worker/consumer"
	"github.com/uloamaka/notification-service/email_worker/processor"
    emailservice "github.com/uloamaka/notification-service/email_service/sender"
)

func main() {
	smtp := emailservice.NewSMTPProvider()
    p := processor.NewEmailProcessor(smtp)
    c := consumer.NewKafkaEmailConsumer(p)

    c.Start("email.jobs")
}