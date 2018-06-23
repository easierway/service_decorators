package service_decorators

import (
	"errors"
	"time"
)

//ErrorRateLimitDecoratorConfig occurred when the configurations are invalid
var ErrorRateLimitDecoratorConfig = errors.New("rate limit configuration is wrong")

//ErrorBeyondRateLimit occurred when current request rate is beyond the limit
var ErrorBeyondRateLimit = errors.New("current request rate is beyond the limit")

//RateLimitDecorator provides the rate limit control
//RateLimitDecoratorConfig is the rate limit Configurations
//Rate = NumOfRequests / Interval
type RateLimitDecorator struct {
	interval      time.Duration
	numOfRequests int
	tokenBucket   chan struct{}
}

//CreateRateLimitDecorator is to create a RateLimitDecorator
func CreateRateLimitDecorator(interval time.Duration, numOfReqs int, tokenBucketSize int) (*RateLimitDecorator, error) {
	if interval == 0 || numOfReqs <= 0 {
		return nil, ErrorRateLimitDecoratorConfig

	}
	bucket := make(chan struct{}, tokenBucketSize)
	//fill the bucket firstly
	for j := 0; j < tokenBucketSize; j++ {
		bucket <- struct{}{}
	}
	go func() {
		for _ = range time.Tick(interval) {
			for i := 0; i < numOfReqs; i++ {
				bucket <- struct{}{}
				sleepTime := interval / time.Duration(numOfReqs)
				time.Sleep(time.Nanosecond * sleepTime)
			}
		}
	}()
	return &RateLimitDecorator{
		interval:      interval,
		numOfRequests: numOfReqs,
		tokenBucket:   bucket,
	}, nil
}

func (dec *RateLimitDecorator) tryToGetToken() bool {
	select {
	case <-dec.tokenBucket:
		return true
	default:
		return false
	}
}

//Decorate function is to add request rate limit logic to the function
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
