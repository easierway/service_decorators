package service_decorators

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/easierway/g_met"
)

func originalFunction(a int, b int) (int, error) {
	return a + b, nil
}

func Example() {
	// Encapsulate the original function as service_decorators.serviceFunc method signature
	type encapsulatedReq struct {
		a int
		b int
	}

	encapsulatedFn := func(req Request) (Response, error) {
		innerReq, ok := req.(encapsulatedReq)
		if !ok {
			return nil, errors.New("invalid parameters")
		}
		return originalFunction(innerReq.a, innerReq.b)
	}

	// Add the logics with decorators
	// 1. Create the decorators
	var (
		retryDec        *RetryDecorator
		circuitBreakDec *CircuitBreakDecorator
		metricDec       *MetricDecorator
		err             error
	)

	if retryDec, err = CreateRetryDecorator(3 /*max retry times*/, time.Second*1,
		time.Second*1, retriableChecker); err != nil {
		panic(err)
	}

	if circuitBreakDec, err = CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 100).
		WithMaxCurrentRequests(1000).
		Build(); err != nil {
		panic(err)
	}

	gmet := g_met.CreateGMetInstanceByDefault("g_met_config/gmet_config.xml")
	if metricDec, err = CreateMetricDecorator(gmet).
		NeedsRecordingTimeSpent().Build(); err != nil {
		panic(err)
	}

	// 2. decorate the encapsulted function with decorators
	// be careful of the order of the decorators
	decFn := circuitBreakDec.Decorate(metricDec.Decorate(retryDec.Decorate(encapsulatedFn)))
	ret, err := decFn(encapsulatedReq{1, 2})
	fmt.Println(ret, err)
	//Output: 3 <nil>
}

func TestExample(t *testing.T) {
	Example()
}
