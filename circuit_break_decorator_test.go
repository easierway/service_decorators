package service_decorators

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func MockServiceFn(req Request) (Response, error) {
	if op, ok := req.(int); ok {
		return Response(op + 1), nil
	}
	return nil, errors.New("Not support input param")
}

func MockServiceLongRunFn(req Request) (Response, error) {
	if op, ok := req.(int); ok {
		time.Sleep(time.Second * 1)
		return Response(op + 1), nil
	}

	return nil, errors.New("Not support input param")
}

func MockFallbackFn(req Request, err error) (Response, error) {
	return Response(-2), nil
}

func TestCircuitBreakHappyCase(t *testing.T) {
	cbDec, err := CreateCircuitBreakDecorator().
		WithTimeout(time.Second * 5).
		WithMaxCurrentRequests(10).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceFn)
	ret, err := decoratedFn(10)
	if err != nil {
		t.Errorf("Unexpected error happened %v", err)
		return
	}
	if ret != 11 {
		t.Errorf("Expected return value is %d, but actual is %d", 11, ret)
		return
	}
	fmt.Printf("Return value is %d\n", ret)
}

func TestCircuitBreakTimeoutCase(t *testing.T) {
	cbDec, err := CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 5).
		WithMaxCurrentRequests(10).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceLongRunFn)
	_, err = decoratedFn(10)
	if err != ErrorCircuitBreakTimeout {
		t.Errorf("Unexpected error happened! %v", err)
	}
}

func TestCircuitBreakTimeoutCaseWithFallback(t *testing.T) {
	cbDec, err := CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 5).
		WithTimeoutFallbackFunction(MockFallbackFn).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceLongRunFn)
	ret, err := decoratedFn(10)
	if err != nil {
		t.Errorf("Unexpected error happened! %v", err)
	}
	if ret != -2 {
		t.Error("Timeout fallback didn't work well!")
	}

}

type fnResponse struct {
	ret Response
	err error
}

func callFnConcurrently(fn ServiceFunc, fnReq int,
	numOfGoroutines int, respChan chan fnResponse,
	launchingInterval time.Duration) {
	for i := 0; i < numOfGoroutines; i++ {
		go func() {
			ret, err := fn(fnReq)
			resp := fnResponse{ret, err}
			respChan <- resp
		}()
		time.Sleep(launchingInterval)
	}
}

func TestCircuitBreakBeyondMaxConcurrentReqCase(t *testing.T) {
	maxCurReqSetting := 10
	cbDec, err := CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 1500).
		WithMaxCurrentRequests(maxCurReqSetting).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceLongRunFn)
	numOfGoroutines := maxCurReqSetting + 1
	respChan := make(chan fnResponse, numOfGoroutines)
	callFnConcurrently(decoratedFn, 10, numOfGoroutines, respChan, 0)

	hasMaxCurError := false
	for fnRet := range respChan {
		if fnRet.err == nil {
			if fnRet.ret.(int) != 11 {
				t.Errorf("The expected return value is %d, but actual is %d", 11, fnRet)
				break
			}
		} else {
			if fnRet.err == ErrorCircuitBreakTooManyConcurrentRequests {
				hasMaxCurError = true
				break
			}
		}
	}
	if !hasMaxCurError {
		t.Error("MaxConcurrentRequests setting doesn't work!")
	}
}

func TestCircuitBreakBeyondMaxConcurrentReqCaseWithFallback(t *testing.T) {
	maxCurReqSetting := 10
	cbDec, err := CreateCircuitBreakDecorator().
		WithMaxCurrentRequests(maxCurReqSetting).
		WithTimeout(time.Millisecond * 1500).
		WithBeyondMaxConcurrencyFallbackFunction(MockFallbackFn).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceLongRunFn)
	numOfGoroutines := maxCurReqSetting + 1
	respChan := make(chan fnResponse, numOfGoroutines)
	callFnConcurrently(decoratedFn, 10, numOfGoroutines, respChan, 0)

	hasFallbackInvoked := false
	var fnRet fnResponse
	for j := 0; j < numOfGoroutines; j++ {
		fnRet = <-respChan
		if fnRet.err == nil {
			if fnRet.ret.(int) == -2 {
				hasFallbackInvoked = true
			}
		} else {
			t.Errorf("Unexpected error happened. %v", fnRet.err)
		}
	}
	if !hasFallbackInvoked {
		t.Error("MaxConcurrentFallback setting didn't work well!")
	}
}

func TestCircuitBreakInMaxConcurrentReqCase(t *testing.T) {
	maxCurReqSetting := 10
	cbDec, err := CreateCircuitBreakDecorator().
		WithTimeout(time.Second * 3).
		WithMaxCurrentRequests(maxCurReqSetting).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceLongRunFn)
	numOfGoroutines := maxCurReqSetting
	respChan := make(chan fnResponse, numOfGoroutines)
	callFnConcurrently(decoratedFn, 10, numOfGoroutines, respChan, 0)
	var fnRet fnResponse
	for j := 0; j < numOfGoroutines; j++ {
		fnRet = <-respChan
		if fnRet.err == nil {
			if fnRet.ret.(int) != 11 {
				t.Errorf("The expected return value is %d, but actual is %d", 11, fnRet)
				break
			}
		} else {
			t.Errorf("Unexpected error happened {%v}", fnRet.err)
		}
	}

}

func TestCircuitBreakInMaxConcurrentReqCaseForProperFequency(t *testing.T) {
	maxCurReqSetting := 3
	cbDec, err := CreateCircuitBreakDecorator().
		WithTimeout(time.Second * 2).
		WithMaxCurrentRequests(maxCurReqSetting).
		Build()
	checkUnexpectedError(err, t)
	decoratedFn := cbDec.Decorate(MockServiceLongRunFn)
	numOfGoroutines := maxCurReqSetting + 1
	respChan := make(chan fnResponse, numOfGoroutines)
	callFnConcurrently(decoratedFn, 10, numOfGoroutines, respChan, time.Millisecond*1500)
	var fnRet fnResponse
	for j := 0; j < numOfGoroutines; j++ {
		fnRet = <-respChan
		if fnRet.err == nil {
			if fnRet.ret.(int) != 11 {
				t.Errorf("The expected return value is %d, but actual is %d", 11, fnRet)
				break
			}
		} else {
			t.Errorf("Unexpected error happened {%v}", fnRet.err)
		}
	}
}
