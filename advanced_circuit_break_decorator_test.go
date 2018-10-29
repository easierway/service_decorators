package service_decorators

import (
	"errors"
	"testing"
	"time"
)

func TestInvokingFallbackWhenBeyondThreshold(t *testing.T) {
	cntFallback := 0
	fallbackFn := func(req Request, lastErr error) (Response, error) {
		cntFallback++
		return nil, nil
	}
	serviceFn := func(req Request) (Response, error) {
		return nil, errors.New("error")
	}
	errDistinguisherFn := func(err error) bool {
		return true
	}
	dec := CreateAdvancedCircuitBreakDecorator(3, time.Second*1, time.Second*2,
		errDistinguisherFn, fallbackFn)

	decFn := dec.Decorate(serviceFn)
	for i := 0; i < 6; i++ {
		decFn("input")
		t.Logf("errCnt:%d fallbackCnt:%d\n", dec.ErrorCounter, cntFallback)
	}
	expectedFallbackTimes := 3
	if cntFallback != expectedFallbackTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedFallbackTimes, cntFallback)
	}
}

func TestErrorCntReset(t *testing.T) {
	cntFallback := 0
	cntBackend := 0
	fallbackFn := func(req Request, lastErr error) (Response, error) {
		cntFallback++
		return nil, nil
	}
	serviceFn := func(req Request) (Response, error) {
		cntBackend++
		if cntBackend > 3 {
			return nil, nil
		}
		return nil, errors.New("error")
	}
	errDistinguisherFn := func(err error) bool {
		return true
	}
	dec := CreateAdvancedCircuitBreakDecorator(3, time.Second*2, time.Millisecond*100,
		errDistinguisherFn, fallbackFn)

	decFn := dec.Decorate(serviceFn)
	for i := 0; i < 6; i++ {
		decFn("input")
		if i == 3 {
			time.Sleep(time.Millisecond * 200)
		}
		t.Logf("errCnt:%d fallbackCnt:%d\n", dec.ErrorCounter, cntFallback)
	}
	expectedFallbackTimes := 1
	expectedBackendTimes := 5
	if cntFallback != expectedFallbackTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedFallbackTimes, cntFallback)
	}
	if cntBackend != expectedBackendTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedBackendTimes, cntBackend)
	}
}

func TestRetryBackendFnAndNotRecovered(t *testing.T) {
	cntFallback := 0
	cntBackend := 0
	fallbackFn := func(req Request, lastErr error) (Response, error) {
		cntFallback++
		return nil, nil
	}
	serviceFn := func(req Request) (Response, error) {
		cntBackend++
		return nil, errors.New("error")
	}
	errDistinguisherFn := func(err error) bool {
		return true
	}
	dec := CreateAdvancedCircuitBreakDecorator(3, time.Second*2, time.Millisecond*100,
		errDistinguisherFn, fallbackFn)

	decFn := dec.Decorate(serviceFn)
	for i := 0; i < 6; i++ {
		decFn("input")
		if i == 3 {
			time.Sleep(time.Millisecond * 200)
		}
		t.Logf("errCnt:%d fallbackCnt:%d\n", dec.ErrorCounter, cntFallback)
	}
	expectedFallbackTimes := 2
	expectedBackendTimes := 4
	if cntFallback != expectedFallbackTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedFallbackTimes, cntFallback)
	}
	if cntBackend != expectedBackendTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedBackendTimes, cntBackend)
	}
}

func TestRetryBackendFnAndRecovered(t *testing.T) {
	cntFallback := 0
	cntBackend := 0
	fallbackFn := func(req Request, lastErr error) (Response, error) {
		cntFallback++
		return nil, nil
	}
	serviceFn := func(req Request) (Response, error) {
		cntBackend++
		if cntBackend > 3 {
			return nil, nil
		}
		return nil, errors.New("error")
	}
	errDistinguisherFn := func(err error) bool {
		return true
	}
	dec := CreateAdvancedCircuitBreakDecorator(3, time.Second*2, time.Millisecond*100,
		errDistinguisherFn, fallbackFn)

	decFn := dec.Decorate(serviceFn)
	for i := 0; i < 6; i++ {
		decFn("input")
		if i == 3 {
			time.Sleep(time.Millisecond * 200)
		}
		t.Logf("errCnt:%d fallbackCnt:%d\n", dec.ErrorCounter, cntFallback)
	}
	expectedFallbackTimes := 1
	expectedBackendTimes := 5
	if cntFallback != expectedFallbackTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedFallbackTimes, cntFallback)
	}
	if cntBackend != expectedBackendTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedBackendTimes, cntBackend)
	}
}

func TestInConcurrentEnv(t *testing.T) {
	cntFallback := 0
	cntBackend := 0
	fallbackFn := func(req Request, lastErr error) (Response, error) {
		cntFallback++
		return nil, nil
	}
	serviceFn := func(req Request) (Response, error) {
		cntBackend++
		if cntBackend > 3 {
			return nil, nil
		}
		return nil, errors.New("error")
	}
	errDistinguisherFn := func(err error) bool {
		return true
	}
	dec := CreateAdvancedCircuitBreakDecorator(3, time.Second*2, time.Millisecond*100,
		errDistinguisherFn, fallbackFn)

	decFn := dec.Decorate(serviceFn)
	for i := 0; i < 8; i++ {
		go func(i int) {
			if i > 4 {
				time.Sleep(time.Millisecond * 200)
			}
			decFn("input")
			t.Logf("errCnt:%d fallbackCnt:%d\n", dec.ErrorCounter, cntFallback)
		}(i)
	}
	time.Sleep(time.Millisecond * 300)
	expectedFallbackTimes := 2
	expectedBackendTimes := 6
	if cntFallback != expectedFallbackTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedFallbackTimes, cntFallback)
	}
	if cntBackend != expectedBackendTimes {
		t.Errorf("expected times is %d, but the actual times is %d\n",
			expectedBackendTimes, cntBackend)
	}
}
