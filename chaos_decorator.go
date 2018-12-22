package service_decorators

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sync/atomic"
	"time"
)

// ChaosEngineeringConfig is the configuration.
type ChaosEngineeringConfig struct {
	IsToInjectChaos        bool `json:"IsToInjectChaos"`        //Is it to start chaos injection, if it is false, all chaos injects (chaos function, slow response) will be stopped.
	AdditionalResponseTime int  `json:"AdditionalResponseTime"` //Inject additional time spent to simulate slow response.
	ChaosRate              int  `json:"ChaosRate"`              //The proportion of the chaos response, the range is 0-100
}

// ChaosEngineeringDecorator is to inject the failure for Chaos Engineering
type ChaosEngineeringDecorator struct {
	config          atomic.Value
	chaosResponseFn ServiceFunc
	configStorage   *ConfigStorage
}

func getChaosConfigFromStorage(configStorage ConfigStorage,
	configName string) (*ChaosEngineeringConfig, error) {
	configStr, err := configStorage.Get(configName)
	config := ChaosEngineeringConfig{
		IsToInjectChaos:        false,
		AdditionalResponseTime: 0,
		ChaosRate:              0,
	}
	if err != nil {
		return &config, err
	}
	err = json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return &ChaosEngineeringConfig{
			IsToInjectChaos:        false,
			AdditionalResponseTime: 0,
			ChaosRate:              0,
		}, err
	}
	if config.ChaosRate < 0 || config.ChaosRate > 100 {
		return &ChaosEngineeringConfig{
			IsToInjectChaos:        false,
			AdditionalResponseTime: 0,
			ChaosRate:              0,
		}, errors.New("The value of ChaosRate should be in [0,100].")
	}
	return &config, nil
}

// CreateChaosEngineeringDecorator is to create a CChaosEngineeringDecorator
// configStore: the storage is used to store the chaos configurations
// configName: the config name in the storage
// chaosResponseFn: the function is to inject the failure for chaos engineering
func CreateChaosEngineeringDecorator(configStorage ConfigStorage, configName string,
	chaosResponseFn ServiceFunc,
	refreshInterval time.Duration) (*ChaosEngineeringDecorator, error) {
	config, err := getChaosConfigFromStorage(configStorage, configName)
	dec := ChaosEngineeringDecorator{}
	dec.config.Store(config)
	dec.chaosResponseFn = chaosResponseFn
	go dec.refreshConfig(refreshInterval, configStorage, configName)
	return &dec, err
}

// Decorate function is to add chaos engineering logic to the function
func (dec *ChaosEngineeringDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		config, ok := dec.config.Load().(*ChaosEngineeringConfig)
		if !ok || config == nil {
			return innerFn(req)
		}
		if !config.IsToInjectChaos {
			return innerFn(req)
		}
		reqSeri := rand.Intn(99) + 1
		if reqSeri <= config.ChaosRate {
			if config.AdditionalResponseTime > 0 {
				time.Sleep(time.Duration(config.AdditionalResponseTime) * time.Millisecond)
			}
			if dec.chaosResponseFn != nil {
				return dec.chaosResponseFn(req)
			}
		}
		return innerFn(req)
	}
}

func (dec *ChaosEngineeringDecorator) refreshConfig(
	refreshInterval time.Duration, configStorage ConfigStorage,
	configName string) {
	if refreshInterval <= 0 {
		return
	}
	for _ = range time.Tick(refreshInterval) {
		updatedConfig, _ := getChaosConfigFromStorage(configStorage, configName)
		dec.config.Store(updatedConfig)
	}
}
