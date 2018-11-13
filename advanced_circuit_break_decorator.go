package service_decorators

import (
	"sync/atomic"
	"time"
)

// ErrorDistinguisherFn is to decide if the error should be counted.
type ErrorDistinguisherFn func(err error) bool

// AdvancedCircuitBreakDecorator is the advanced CircuitBeakDecorator,
// which can support complex circuit break strategy.
// For circuitBreak states transition, please refer to
// https://github.com/easierway/service_decorators/blob/master/doc_pics/circuit_breaker_states_transtion.png
//
// 1. Failure frequency will cause circuit breaker state to open state
// -- Failure frequency involves ErrorCount and ResetIntervalOfErrorCount
//    They are used to count the errors occurred in the time window.
// -- ErrorDistinguisher is to decide what kind of errors would be counted with ErrorCounter.
// -- ErrorCounter is to count the continuous occurred errors in the time window
// -- ResetIntervalOfErrorCount is the interval of resetting error counter to 0
// 2. When circuit breaker is in open state, the requests will be processed
//    by fallback function (FallbackFn)
// 3. BackendRetryInterval is to check if the backend service is health/recovered.
//    The time interval since the last time of backend service invoked is beyond BackendRetryInterval
//    even the circuit breaker is in close state, current request will be passed to backend service.
//    If the request be processed successfully or no counted errors happen (decided by ErrorDistinguisher)
//    the circuit break will switch to close state
type AdvancedCircuitBreakDecorator struct {
	ErrorCounter       int64
	ErrorDistinguisher ErrorDistinguisherFn

	ErrorFrequencyThreshold     int64
	LastError                   error
	ResetIntervalOfErrorCounter time.Duration
	BackendRetryInterval        time.Duration
	lastErrorOccuredTime        time.Time
	lastBackendInvokingTime     time.Time
	FallbackFn                  ServiceFallbackFunc
}

func CreateAdvancedCircuitBreakDecorator(
	errorFrequencyThreshold int64,
	resetIntervalOfErrorCounter time.Duration,
	backendRetryInterval time.Duration,
	errorDistinguisher ErrorDistinguisherFn,
	fallbackFn ServiceFallbackFunc) *AdvancedCircuitBreakDecorator {
	return &AdvancedCircuitBreakDecorator{
		ErrorFrequencyThreshold:     errorFrequencyThreshold,
		ResetIntervalOfErrorCounter: resetIntervalOfErrorCounter,
		BackendRetryInterval:        backendRetryInterval,
		ErrorDistinguisher:          errorDistinguisher,
		FallbackFn:                  fallbackFn,
		lastBackendInvokingTime:     time.Now(),
	}
}

func (dec *AdvancedCircuitBreakDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		now := time.Now()
		durErr := now.Sub(dec.lastErrorOccuredTime)
		durRetry := now.Sub(dec.lastBackendInvokingTime)
		if durErr > dec.ResetIntervalOfErrorCounter {
			atomic.StoreInt64(&dec.ErrorCounter, 0)
		}
		if durRetry < dec.BackendRetryInterval &&
			atomic.LoadInt64(&dec.ErrorCounter) >= dec.ErrorFrequencyThreshold {
			return dec.FallbackFn(req, dec.LastError)
		}
		ret, err := innerFn(req)
		dec.lastBackendInvokingTime = now
		if err == nil {
			atomic.StoreInt64(&dec.ErrorCounter, 0)
		} else {
			if dec.ErrorDistinguisher(err) {
				atomic.AddInt64(&dec.ErrorCounter, 1)
				dec.lastErrorOccuredTime = now
			}
		}
		return ret, err
	}
}
