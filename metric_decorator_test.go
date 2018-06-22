package service_decorators

import (
	"errors"
	"testing"

	"github.com/easierway/g_met"
)

type memoryMet struct {
	metrics []g_met.MetricItem
}

func (met *memoryMet) Send(metrics ...g_met.MetricItem) error {
	met.metrics = metrics
	return nil
}

func (met *memoryMet) Close() error {
	return nil
}

func (met *memoryMet) Flush() {
}

func MockServiceFnWithErr(req Request) (Response, error) {
	return nil, errors.New("Unexpected Error")
}

func MockErrorClassifier(err error) (string, bool) {
	if err != nil {
		return err.Error(), true
	}
	return "N/A", false
}

func checkInnerFunc(ret Response, err error, t *testing.T) {
	if err != nil {
		t.Error("Unexpected error occurred.", err)
	}
	if ret != 11 {
		t.Errorf("Expected value is %d, but actual is %d", 11, ret)
		return
	}
}

func TestMetricsWithTimeSpentRecords(t *testing.T) {
	met := memoryMet{}
	dec := CreateMetricDecorator(&met).NeedsRecordingTimeSpent().Build()
	decFn := dec.Decorate(MockServiceLongRunFn)
	ret, err := decFn(10)
	checkInnerFunc(ret, err, t)
	if len(met.metrics) != 1 {
		t.Errorf("the metrics is not expected %v", met.metrics)
		return
	}
	if met.metrics[0].Key != TimeSpent {
		t.Errorf("the metrics is not expected %v", met.metrics)
		return
	}
	t.Log(met.metrics[0])
}

func TestMetricsWithErrorRecords(t *testing.T) {
	met := memoryMet{}
	dec := CreateMetricDecorator(&met).WithErrorClassifier(MockErrorClassifier).Build()
	decFn := dec.Decorate(MockServiceFnWithErr)
	_, err := decFn(10)
	if err == nil {
		t.Error("An error is expected")
		return
	}
	if len(met.metrics) != 1 {
		t.Errorf("the metrics is not expected %v", met.metrics)
		return
	}
	if met.metrics[0].Key != OccurredError {
		t.Errorf("the metrics is not expected %v", met.metrics)
		return
	}
	t.Log(met.metrics[0])
}

func TestMetricsWithoutAnyRecords(t *testing.T) {
	met := memoryMet{}
	dec := CreateMetricDecorator(&met).Build()
	decFn := dec.Decorate(MockServiceFnWithErr)
	_, err := decFn(10)
	if err == nil {
		t.Error("An error is expected")
		return
	}
	if len(met.metrics) != 0 {
		t.Errorf("the metrics is not expected %v", met.metrics)
		return
	}
	t.Log(met.metrics)
}
