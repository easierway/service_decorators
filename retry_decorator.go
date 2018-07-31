package service_decorators

import (
	"errors"
	"time"
)

type retryDecoratorConfig struct {
	maxRetryTimes     int
	retryInterval     time.Duration
	intervalIncrement time.Duration
	retriableChecker  func(err error) bool
}

// RetryDecorator is to add the retry logic to the decorated method.
type RetryDecorator struct {
	config *retryDecoratorConfig
}

// CreateRetryDecorator is to create RetryDecorator according to the settings
// maxRetryTimes : max retry times
// retryInterval, intervalIncrement : the sleep time before next retrying is  retryInterval + (retry times - 1) * intervalIncrement
// retriableChecker : the function to check wether the error is retriable
func CreateRetryDecorator(maxRetryTimes int, retryInterval time.Duration,
	intervalIncrement time.Duration,
	retriableChecker func(err error) bool) (*RetryDecorator, error) {
	if maxRetryTimes <= 0 || retryInterval <= 0 ||
		intervalIncrement < 0 || retriableChecker == nil {
		return nil, errors.New("invalid configurations")
	}
	config := retryDecoratorConfig{
		maxRetryTimes, retryInterval, intervalIncrement, retriableChecker,
	}
	return &RetryDecorator{&config}, nil
}

// Decorator function is to add the retry logic to the decorated method
func (dec *RetryDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		var (
			res Response
			err error
		)
		var interval = dec.config.retryInterval
		for i := 0; i <= dec.config.maxRetryTimes; i++ {
			res, err = innerFn(req)
			if err == nil {
				return res, err
			}
			if !dec.config.retriableChecker(err) {
				return res, err
			}
			time.Sleep(interval)
			interval = interval + dec.config.intervalIncrement
		}
		return res, err
	}
}
