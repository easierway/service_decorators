package service_decorators

import (
	"testing"
	"time"
)

type MockConfigStorage struct {
	ConfigStr string
}

func serviceFn(req Request) (Response, error) {
	return "", nil
}

func (storage *MockConfigStorage) Get(name string) ([]byte, error) {
	return []byte(storage.ConfigStr), nil
}

func TestSlowResponse(t *testing.T) {
	testConfigStr := `{
	   "IsToInjectChaos" : true,
	   "AdditionalResponseTime" : 500,
	   "ChaosRate" : 100
	 }`
	storage := &MockConfigStorage{ConfigStr: testConfigStr}
	chaosDec, err := CreateChaosEngineeringDecorator(storage, "chaos_config", nil, 10*time.Millisecond)
	if err != nil {
		t.Error(err)
	}
	decFn := chaosDec.Decorate(serviceFn)
	start_t := time.Now()
	decFn("")
	timeEscaped := time.Since(start_t).Seconds()
	t.Logf("Time escaped:%v seconds.", timeEscaped)
	if timeEscaped < 0.5 {
		t.Error("Failed to simulate slow response.")
	}
	storage.ConfigStr = `{
	   "IsToInjectChaos" : false,
	   "AdditionalResponseTime" : 500,
	   "ChaosRate" : 100
	 }`
	time.Sleep(20 * time.Millisecond)
	start_t = time.Now()
	decFn("")
	timeEscaped = time.Since(start_t).Seconds()
	t.Logf("Time escaped:%v seconds.", timeEscaped)
	if timeEscaped >= 0.5 {
		t.Error("Failed to switch off chaos.")
	}
}

func TestChaosFnWithRateControl(t *testing.T) {
	testConfigStr := `{
	   "IsToInjectChaos" : true,
	   "AdditionalResponseTime" : 0,
	   "ChaosRate" : 50
	 }`
	cnt := 0
	chaosFn := func(req Request) (Response, error) {
		cnt++
		return nil, nil
	}
	storage := &MockConfigStorage{ConfigStr: testConfigStr}
	chaosDec, err := CreateChaosEngineeringDecorator(storage, "chaos_config", chaosFn, 10*time.Millisecond)
	if err != nil {
		t.Error(err)
	}
	decFn := chaosDec.Decorate(serviceFn)
	for i := 0; i < 100000; i++ {
		decFn("")
	}
	t.Logf("The chaos function has been invoked %d times.", cnt)
	invokedRate := float64(cnt) / 100000.0
	if invokedRate > 0.55 || invokedRate < 0.45 {
		t.Error("Failed to control chaos injection rate.")
	}
}

func TestInvalidConfig(t *testing.T) {
	testConfigStr := `[
		 "IsToInChaos" : true,
		 "AdditionalResponseTime" : 0,
		 "ChaosRate" : 50
	 }`
	storage := &MockConfigStorage{ConfigStr: testConfigStr}
	chaosDec, err := CreateChaosEngineeringDecorator(storage, "chaos_config", nil, 10*time.Millisecond)
	if err == nil {
		t.Error("The error is expected here.")
	}
	initConfig := ChaosEngineeringConfig{
		IsToInjectChaos:        false,
		AdditionalResponseTime: 0,
		ChaosRate:              0,
	}
	if config, ok := chaosDec.config.Load().(*ChaosEngineeringConfig); ok {
		if *config != initConfig {
			t.Errorf("Unexpected configuration: %v", config)
		}
	} else {
		t.Errorf("Unexpected configuration: %v", config)
	}
}
