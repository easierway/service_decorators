package service_decorators

import (
	"testing"
	"time"
)

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Error("Unexpected error happened.", err)
	}
}

func TestRateLimitDecoratorHappyCase(t *testing.T) {
	dec, err := CreateRateLimitDecorator(time.Second*1, 2, 3)
	checkErr(err, t)
	decFn := dec.Decorate(MockServiceLongRunFn)
	ret, err := decFn(10)
	checkErr(err, t)
	checkCnt(ret.(int), 11, t)
}

type RateLimitSetting struct {
	numOfReqs  int
	invterval  time.Duration
	bucketSize int
}

func checkRateLimitDecorator(ratelimit RateLimitSetting, invokingInterval time.Duration,
	t *testing.T) bool {
	numOfReqs := 5
	dec, err := CreateRateLimitDecorator(ratelimit.invterval, ratelimit.numOfReqs, ratelimit.bucketSize)
	checkErr(err, t)
	decFn := dec.Decorate(MockServiceLongRunFn)
	respChan := make(chan fnResponse, numOfReqs)
	callFnConcurrently(decFn, 10, numOfReqs, respChan, invokingInterval)
	didRatelimitErrHappened := false
	for j := 0; j < numOfReqs; j++ {
		resp := <-respChan
		if resp.err == ErrorBeyondRateLimit {
			didRatelimitErrHappened = true
		}
	}
	return didRatelimitErrHappened
}

func TestRateLimitDecoratorBeyondRateLimitCase(t *testing.T) {
	if !checkRateLimitDecorator(RateLimitSetting{2, time.Second * 1, 2},
		time.Millisecond*10, t) {
		t.Error("Rate limit didn't work well!")
	}
}

func TestRateLimitDecoratorInRateLimitCase(t *testing.T) {
	if checkRateLimitDecorator(RateLimitSetting{20, time.Second * 1, 20},
		time.Millisecond*60, t) {
		t.Error("Rate limit didn't work well!")
	}
}
