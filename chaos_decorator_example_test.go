package service_decorators

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

// Example_ChaosEngineeringDecorator_InjectError is the example for error injection.
// To run the example, put the following configuration into Consul KV storage with the key "ChaosExample"
// {
//   "IsToInjectChaos" : true,
//   "AdditionalResponseTime" : 0,
//   "ChaosRate" : 10
// }
func Example_ChaosEngineeringDecorator_InjectError() {
	serviceFn := func(req Request) (Response, error) {
		return "Service is invoked", nil
	}
	ErrorInjectionFn := func(req Request) (Response, error) {
		return "Error Injection", errors.New("Failed to process.")
	}
	storage, err := CreateConsulConfigStorage(&api.Config{})
	if err != nil {
		fmt.Println("You might need to start Cousul server.", err)
		return
	}
	chaosDec, err := CreateChaosEngineeringDecorator(storage, "ChaosExample",
		ErrorInjectionFn, 10*time.Millisecond)
	if err != nil {
		fmt.Println("You might need to start Cousul server.", err)
		return
	}
	decFn := chaosDec.Decorate(serviceFn)
	for i := 0; i < 10; i++ {
		ret, _ := decFn("")
		fmt.Printf("Output is %s\n", ret)
	}
	//You have 10% probability to get "Error Injection"
}

// Example_ChaosEngineeringDecorator_InjectError is the example for slow response injection.
// To run the example, put the following configuration into Consul KV storage with the key "ChaosExample"
// {
//   "IsToInjectChaos" : true,
//   "AdditionalResponseTime" : 100, // response time will be increased 100ms
//   "ChaosRate" : 10
// }
func Example_ChaosEngineeringDecorator_InjectSlowResponse() {
	serviceFn := func(req Request) (Response, error) {
		return "Service is invoked", nil
	}
	ErrorInjectionFn := func(req Request) (Response, error) {
		return "Error Injection", errors.New("Failed to process.")
	}
	storage, err := CreateConsulConfigStorage(&api.Config{})
	if err != nil {
		fmt.Println("You might need to start Cousul server.", err)
		return
	}
	chaosDec, err := CreateChaosEngineeringDecorator(storage, "ChaosExample",
		ErrorInjectionFn, 10*time.Millisecond)
	if err != nil {
		fmt.Println("You might need to start Cousul server.", err)
		return
	}
	decFn := chaosDec.Decorate(serviceFn)
	for i := 0; i < 10; i++ {
		tStart := time.Now()
		ret, _ := decFn("")
		fmt.Printf("Output is %s. Time escaped: %f ms\n", ret,
			time.Since(tStart).Seconds()*1000)
	}

	//You have 10% probability to get slow response.

}

// func TestExample_ChaosEngineeringDecorator_InjectError(t *testing.T) {
// 	Example_ChaosEngineeringDecorator_InjectError()
// }

func TestExample_ChaosEngineeringDecorator_InjectSlowResponse(t *testing.T) {
	Example_ChaosEngineeringDecorator_InjectSlowResponse()
}
