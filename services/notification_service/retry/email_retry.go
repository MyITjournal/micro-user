package retry

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

func NewExponentialBackoff() *backoff.ExponentialBackOff{
    b := backoff.NewExponentialBackOff()
    b.InitialInterval = 1 * time.Second
    b.MaxInterval = 1 * time.Minute
    b.MaxElapsedTime = 5 * time.Minute
    return b
}