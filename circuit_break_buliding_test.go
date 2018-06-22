package service_decorators

import (
	"testing"
	"time"
)

func checkUnexpectedError(err error, t *testing.T) {
	if err != nil {
		t.Error("unexpected error occurred")
	}
}

func checkTimeoutSetting(expected time.Duration, actual time.Duration, t *testing.T) {
	if actual != expected {
		t.Errorf("The timeout should be set as %d, but it is %d", expected, actual)
	}
}

func checkMaxCurReq(expected int, actual int, t *testing.T) {
	if actual != expected {
		t.Errorf("maxCurRequest should be %d, but it is %d",
			expected, actual)
	}
}

func TestBuildCircuitBreakDecoratorWithSettings(t *testing.T) {
	settingTimeout := time.Second * 10
	settingMaxCurReq := 10
	cbDecorator, err := CreateCircuitBreakDecorator().
		WithTimeout(settingTimeout).
		Build()
	checkUnexpectedError(err, t)
	maxCurReq := cbDecorator.Config.maxCurrentRequests
	checkMaxCurReq(0, maxCurReq, t)
	timeOut := cbDecorator.Config.timeout
	checkTimeoutSetting(settingTimeout, timeOut, t)

	cbDecorator, err = CreateCircuitBreakDecorator().
		WithTimeout(settingTimeout).
		WithMaxCurrentRequests(settingMaxCurReq).
		Build()
	checkUnexpectedError(err, t)
	maxCurReq = cbDecorator.Config.maxCurrentRequests
	checkMaxCurReq(settingMaxCurReq, maxCurReq, t)
	timeOut = cbDecorator.Config.timeout
	checkTimeoutSetting(settingTimeout, timeOut, t)

}
