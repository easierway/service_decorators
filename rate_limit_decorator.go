package service_decorators

import (
	"errors"
	"time"

	"golang.org/x/time/rate"
)

// ErrorRateLimitDecoratorConfig occurred when the configurations are invalid
var ErrorRateLimitDecoratorConfig = errors.New("rate limit configuration is wrong")

// ErrorBeyondRateLimit occurred when current request rate is beyond the limit
var ErrorBeyondRateLimit = errors.New("current request rate is beyond the limit")

// RateLimitDecorator provides the rate limit control
// RateLimitDecoratorConfig is the rate limit Configurations
// Rate = NumOfRequests / Interval
type RateLimitDecorator struct {
	interval      time.Duration
	numOfRequests int
	limiter       *rate.Limiter
}

// CreateRateLimitDecorator is to create a RateLimitDecorator
func CreateRateLimitDecorator(interval time.Duration, numOfReqs int, tokenBucketSize int) (*RateLimitDecorator, error) {
	if interval == 0 || numOfReqs <= 0 {
		return nil, ErrorRateLimitDecoratorConfig

	}
	qps := 1 / (interval / time.Duration(numOfReqs)).Seconds()
	l := rate.NewLimiter(rate.Limit(qps), tokenBucketSize)

	return &RateLimitDecorator{
		interval:      interval,
		numOfRequests: numOfReqs,
		limiter:       l,
	}, nil
}

func (dec *RateLimitDecorator) tryToGetToken() bool {
	return dec.limiter.Allow()
}

// Decorate function is to add request rate limit logic to the function
func (dec *RateLimitDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		if dec.numOfRequests > 0 {
			if !dec.tryToGetToken() {
				return nil, ErrorBeyondRateLimit
			}

		}
		return innerFn(req)
	}
}
