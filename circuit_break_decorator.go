package service_decorators

import (
	"fmt"
	"time"
)

//CircuitBreakTimeoutError happens when invoking is timeout
var CircuitBreakTimeoutError error

//CircuitBreakTooManyConcurrentRequestsError happens when the number of the concurrent requests beyonds the setting
var CircuitBreakTooManyConcurrentRequestsError error

//CircuitBreakDecoratorConfig includes the settings of CircuitBreakDecorator
type CircuitBreakDecoratorConfig struct {
	//Timeout is about function excution duration. Default timeout is 1 second
	timeout time.Duration

	//MaxCurrentRequests defines the max concurrency
	maxCurrentRequests int

	//if TimeoutFallbackFunction is defined,
	//it would be called when timeout error occurring
	timeoutFallbackFunction ServiceFunc

	//if BeyondMaxConcurrencyFallbackFunction is defined,
	//it would be called when concurrency beyonding error occurring
	beyondMaxConcurrencyFallbackFunction ServiceFunc
}

//CircuitBreakDecorator provides the circuit break,
//fallback, concurrency control
type CircuitBreakDecorator struct {
	//CircuitBreakDecoratorConfig
	Config      *CircuitBreakDecoratorConfig
	tokenBuffer chan struct{}
}

type serviceFuncResponse struct {
	resp Response
	err  error
}

//CreateCircuitBreakDecorator is the helper method of
//creating CircuitBreakDecorator.
//The settings can be defined by WithXX method chain
func CreateCircuitBreakDecorator() *CircuitBreakDecoratorConfig {
	return &CircuitBreakDecoratorConfig{
		timeout: time.Second * 1,
	}
}

//WithTimeout sets the method execution timeout
func (config *CircuitBreakDecoratorConfig) WithTimeout(timeOut time.Duration) *CircuitBreakDecoratorConfig {
	config.timeout = timeOut
	return config
}

//WithMaxCurrentRequests sets max concurrency
func (config *CircuitBreakDecoratorConfig) WithMaxCurrentRequests(maxCurReq int) *CircuitBreakDecoratorConfig {
	config.maxCurrentRequests = maxCurReq
	return config
}

//WithTimeoutFallbackFunction sets the fallback method for timeout error
func (config *CircuitBreakDecoratorConfig) WithTimeoutFallbackFunction(
	fallbackFn ServiceFunc) *CircuitBreakDecoratorConfig {
	config.timeoutFallbackFunction = fallbackFn
	return config
}

//WithBeyondMaxConcurrencyFallbackFunction sets the fallback method for beyonding max concurrency error
func (config *CircuitBreakDecoratorConfig) WithBeyondMaxConcurrencyFallbackFunction(
	fallbackFn ServiceFunc) *CircuitBreakDecoratorConfig {
	config.beyondMaxConcurrencyFallbackFunction = fallbackFn
	return config
}

//Build will create CircuitBreakDecorator with the settings defined by WithXX method chain
func (config *CircuitBreakDecoratorConfig) Build() *CircuitBreakDecorator {
	var tokenBuf chan struct{}
	if config.maxCurrentRequests != 0 {
		tokenBuf = make(chan struct{}, config.maxCurrentRequests)
		for i := 0; i < config.maxCurrentRequests; i++ {
			tokenBuf <- struct{}{}
		}
	}
	return &CircuitBreakDecorator{
		Config:      config,
		tokenBuffer: tokenBuf,
	}
}

func (dec *CircuitBreakDecorator) getToken() bool {
	select {
	case <-dec.tokenBuffer:
		return true
	default:
		return false
	}
}

func (dec *CircuitBreakDecorator) releaseToken() {
	select {
	case dec.tokenBuffer <- struct{}{}:
		return
	default:
		panic("There's a fatal bug here. Unexpected token has been returned.")
	}
}

//Decorate is to add the circuit break/concurrency control logic to the function
func (dec *CircuitBreakDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		ownToken := false
		if dec.Config.maxCurrentRequests > 0 {
			if !dec.getToken() {
				if dec.Config.beyondMaxConcurrencyFallbackFunction != nil {
					return dec.Config.beyondMaxConcurrencyFallbackFunction(req)
				}
				CircuitBreakTooManyConcurrentRequestsError = fmt.
					Errorf("The max current requests is %d",
						dec.Config.maxCurrentRequests)
				return nil, CircuitBreakTooManyConcurrentRequestsError
			}
			ownToken = true
		}
		output := make(chan serviceFuncResponse, 1)
		go func(r Request, withToken bool) {
			if withToken {
				defer dec.releaseToken()
			}
			inResp, inErr := innerFn(r)
			output <- serviceFuncResponse{
				resp: inResp,
				err:  inErr,
			}
		}(req, ownToken)
		select {
		case inServResp := <-output:
			return inServResp.resp, inServResp.err
		case <-time.After(dec.Config.timeout):
			if dec.Config.timeoutFallbackFunction != nil {
				return dec.Config.timeoutFallbackFunction(req)
			}
			CircuitBreakTimeoutError = fmt.Errorf(
				"Circuit break error-timeout! Timeout setting is %v",
				dec.Config.timeout)
			return nil, CircuitBreakTimeoutError
		}

	}

}
