// package circuit_breaker

// import (
//     "github.com/sony/gobreaker"
//     "time"
// )

// func NewEmailCircuitBreaker() *gobreaker.CircuitBreaker {
//     st := gobreaker.Settings{
//         Name:        "smtp-breaker",
//         MaxRequests: 3,                       // when half-open
//         Interval:    30 * time.Second,        // reset rolling counts
//         Timeout:     15 * time.Second,        // how long before trying again
//     }
//     return gobreaker.NewCircuitBreaker(st)
// }

package circuit_breaker

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

func NewEmailCircuitBreaker() *gobreaker.CircuitBreaker[any] {
	settings := gobreaker.Settings{
		Name:          "email_smtp_breaker",
		MaxRequests:   5,
		Interval:      30 * time.Second,
		Timeout:       60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip when failures exceed 60% of total
			failRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests > 10 && failRatio > 0.6
		},
	}
	return gobreaker.NewCircuitBreaker[any](settings)
}
