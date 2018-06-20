package service_decorators

import (
	"testing"
)

func checkCnt(cnt int, expectedValue int, t *testing.T) {
	if cnt != expectedValue {
		t.Errorf(
			"expected value is %d, but actual value is %d",
			expectedValue, cnt)
	}
}

func TestDecoratedActionsHaveBeenInject(t *testing.T) {
	dec := createDemoDecorator()
	decoratedSum := dec.Decorate(sum)
	if ret, err := decoratedSum(sumRequest([]int{1, 2, 3, 4, 5})); err != nil {
		t.Error(err)
	} else {
		checkCnt(dec.InvokedCnt, 1, t)
		checkCnt(dec.FailedCnt, 0, t)
		checkCnt(dec.SucceedCnt, 1, t)
		t.Logf("Sum is %d", ret)
	}
}
