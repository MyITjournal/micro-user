package processor

import (
	"context"
	"encoding/json"

	emailservice "github.com/uloamaka/notification-service/email_service/sender"
	models "github.com/uloamaka/notification-service/email_worker/model"
	"github.com/uloamaka/notification-service/retry"
	cbreaker "github.com/uloamaka/notification-service/circuit_breaker"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker/v2"
)

type EmailProcessor struct {
	sender  emailservice.EmailSender
	breaker *gobreaker.CircuitBreaker[any]
}

func NewEmailProcessor(sender emailservice.EmailSender) *EmailProcessor {
	return &EmailProcessor{
		sender:  sender,
		breaker: cbreaker.NewEmailCircuitBreaker(),
	}
}

func (p *EmailProcessor) Process(ctx context.Context, data []byte) error {
	var job models.EmailJob
	if err := json.Unmarshal(data, &job); err != nil {
		return err
	}

	operation := func() error {
		_, err := p.breaker.Execute(func() (any, error) {
			return nil, p.sender.SendEmail(job)
		})
		return err
	}

	bo := retry.NewExponentialBackoff()
	return backoff.Retry(operation, bo)
}
