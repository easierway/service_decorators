package service_decorators

import (
	"errors"
	"testing"
	"time"
)

var ErrorConnection = errors.New("connection exception")
var retriableChecker = func(err error) bool {
	if err == ErrorConnection {
		return true
	}
	return false
}

func TestRetryWhenRetriableErrorOccurred(t *testing.T) {
	maxRetryTimes := 3
	cntExecution := 0
	connectionErrFn := func(req Request) (Response, error) {
		cntExecution++
		return cntExecution, ErrorConnection
	}
	retryDec, err := CreateRetryDecorator(maxRetryTimes, time.Second*1, time.Second*1, retriableChecker)
	checkErr(err, t)
	decFn := retryDec.Decorate(connectionErrFn)
	res, decErr := decFn(1)
	if decErr != ErrorConnection {
		t.Error("The connection exception is expected.")
	}
	if res.(int) != maxRetryTimes+1 {
		t.Errorf("The expected execution times is %v, the actual is %v", maxRetryTimes+1, res)
	}
}

func TestRetryWhenRetriableErrorOccurredAndRecovered(t *testing.T) {
	maxRetryTimes := 3
	cntExecution := 0
	connectionErrFn := func(req Request) (Response, error) {
		cntExecution++
		if cntExecution == 2 {
			return cntExecution, nil
		}
		return cntExecution, ErrorConnection
	}
	retryDec, err := CreateRetryDecorator(maxRetryTimes, time.Second*1, time.Second*1, retriableChecker)
	checkErr(err, t)
	decFn := retryDec.Decorate(connectionErrFn)
	res, decErr := decFn(1)
	checkErr(decErr, t)
	if res.(int) != 2 {
		t.Errorf("The expected execution times is %v, the actual is %v", maxRetryTimes+1, res)
	}
}

func TestRetryWithoutError(t *testing.T) {
	maxRetryTimes := 3
	cntExecution := 0
	connectionErrFn := func(req Request) (Response, error) {
		cntExecution++
		return cntExecution, nil
	}
	retryDec, err := CreateRetryDecorator(maxRetryTimes, time.Second*1, time.Second*1, retriableChecker)
	checkErr(err, t)
	decFn := retryDec.Decorate(connectionErrFn)
	res, decErr := decFn(1)
	checkErr(decErr, t)
	if res.(int) != 1 {
		t.Errorf("The expected execution times is %v, the actual is %v", 1, res)
	}
}

func TestRetryWhenUnretriableErrorOccurred(t *testing.T) {
	maxRetryTimes := 3
	cntExecution := 0
	connectionErrFn := func(req Request) (Response, error) {
		cntExecution++
		return cntExecution, errors.New("other exception")
	}
	retryDec, err := CreateRetryDecorator(maxRetryTimes, time.Second*1, time.Second*1, retriableChecker)
	checkErr(err, t)
	decFn := retryDec.Decorate(connectionErrFn)
	res, decErr := decFn(1)
	if decErr == nil || decErr == ErrorConnection {
		t.Error("The error is not expected.")
	}
	if res.(int) != 1 {
		t.Errorf("The expected execution times is %v, the actual is %v", 1, res)
	}
}

func TestCreateDecorateWithInvalidSettings(t *testing.T) {
	var err1, err2 error
	_, err1 = CreateRetryDecorator(0, time.Second*1, time.Second*1, retriableChecker)
	if err1 == nil {
		t.Error("Setting error is expected")
	}
	_, err2 = CreateRetryDecorator(3, 0, time.Second*1, retriableChecker)
	if err2 == nil {
		t.Error("Setting error is expected")
	}

}
