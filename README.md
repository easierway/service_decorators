# Simplify microservice development
## What’s the most complicated part of implementing a microservice except the core business logic?

You might think it must be RPC end point part, which makes your business logic as a real service can be accessed from the network.

But, this is not true. By leveraging the opensource RPC packages, such as, apache thrift, gRPC, this part could be extremely easy except defining your service interface with some IDL.

The following codes are about starting a service server based on thrift.
```Go
transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
serverTransport, err := thrift.NewTServerSocket(NetworkAddr)
if err != nil {
	os.Exit(1)
}
processor := calculator_thrift.NewCalculatorProcessor(&calculatorHandler{})
server := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
server.Serve()
```
As we all known, to build a robust and maintainable service is not an easy task. There are a lot of tricky work as the following:

![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/other_functions.jpg)

Most of them are much complicated than RPC, in some cases they are even complicated than your core business logic.

For example, the code section of rate limit

```Go
...
bucket := make(chan struct{}, tokenBucketSize)
//fill the bucket firstly
for j := 0; j < tokenBucketSize; j++ {
	bucket <- struct{}{}
}
go func() {
	for _ = range time.Tick(interval) {
		for i := 0; i < numOfReqs; i++ {
			bucket <- struct{}{}
			sleepTime := interval / time.Duration(numOfReqs)
			time.Sleep(time.Nanosecond * sleepTime)
		}
	}
}()
...

select {
	case <-dec.tokenBucket:
		return executeBusinessLog(req),nil
	default:
		return errors.New("beyond the rate limit")
}

...

```

Just as the following diagram shows normally what you have to do is much more than what you want to do when implementing a microservice.

![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/to_do.jpg)

The goal of the project is to reduce your work on what you have to do and let you only focus on what you want to do.

All these common functions (e.g. rate limit, circuit break, metric) are encapsulated in the components, and these common logics can be injected in your service implementation transparently by leveraging decorator pattern.

![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/decorator_pattern.jpg)

![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/decorators.jpg)

The following is an example of adding rate limit, circuit break and metrics functions with the prebuilt decorators.

```Go
if rateLimitDec, err = service_decorators.CreateRateLimitDecorator(time.Millisecond*1, 100); err != nil {
		return nil, err
	}

	if circuitBreakDec, err = service_decorators.CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 100).
		WithMaxCurrentRequests(1000).
		WithTimeoutFallbackFunction(addFallback).
		WithBeyondMaxConcurrencyFallbackFunction(addFallback).
		Build(); err != nil {
		return nil, err
	}

	gmet := g_met.CreateGMetInstanceByDefault("g_met_config/gmet_config.xml")
	if metricDec, err = service_decorators.CreateMetricDecorator(gmet).
		NeedsRecordingTimeSpent().Build(); err != nil {
		return nil, err
	}
	decFn := rateLimitDec.Decorate(circuitBreakDec.Decorate(metricDec.Decorate(innerFn)))
  ...
```
Refer to the example: https://github.com/easierway/service_decorators_example/blob/master/example_service_test.go

## Not only for building service
You definitely can use this project in the cases besides building a service. For example, when invoking a remote service, you have to consider on fault-over and metrics. In this case, you can leverage the decorators to simplify your code as following:
```Go
func originalFunction(a int, b int) (int, error) {
	return a + b, nil
}

func Example() {
	// Encapsulate the original function as service_decorators.serviceFunc method signature
	type encapsulatedReq struct {
		a int
		b int
	}

	encapsulatedFn := func(req Request) (Response, error) {
		innerReq, ok := req.(encapsulatedReq)
		if !ok {
			return nil, errors.New("invalid parameters")
		}
		return originalFunction(innerReq.a, innerReq.b)
	}

	// Add the logics with decorators
	// 1. Create the decorators
	var (
		retryDec        *RetryDecorator
		circuitBreakDec *CircuitBreakDecorator
		metricDec       *MetricDecorator
		err             error
	)

	if retryDec, err = CreateRetryDecorator(3 /*max retry times*/, time.Second*1,
		time.Second*1, retriableChecker); err != nil {
		panic(err)
	}

	if circuitBreakDec, err = CreateCircuitBreakDecorator().
		WithTimeout(time.Millisecond * 100).
		WithMaxCurrentRequests(1000).
		Build(); err != nil {
		panic(err)
	}

	gmet := g_met.CreateGMetInstanceByDefault("g_met_config/gmet_config.xml")
	if metricDec, err = CreateMetricDecorator(gmet).
		NeedsRecordingTimeSpent().Build(); err != nil {
		panic(err)
	}

	// 2. decorate the encapsulted function with decorators
	// be careful of the order of the decorators
	decFn := circuitBreakDec.Decorate(metricDec.Decorate(retryDec.Decorator(encapsulatedFn)))
	ret, err := decFn(encapsulatedReq{1, 2})
	fmt.Println(ret, err)
	//Output: 3 <nil>
}
```
## Decorators
### Decorators List
1. Rate Limit Decorator
2. Circuit Break Decorator
3. Advanced Circuit Break Decorator
4. Metric Decorator
5. Retry Decorator
6. Chaos Engineering Decorator

### CircuitBreakDecorator
Circuit breaker is the key part fault tolerance and recovery oriented solution. Circuit breaker is to stop cascading failure and enable resilience in complex distributed systems where failure is inevitable.

Stop cascading failures. Fallbacks and graceful degradation. Fail fast and rapid recovery.
CircuitBreakDecorator would interrupt client invoking when its time spent being longer than the expectation. A timeout exception or the result coming from the degradation method will be return to client.
The similar logic is also working for the concurrency limit.
https://github.com/easierway/service_decorators_example/blob/master/example_service_test.go#L46

### AdvancedCircuitBreakDecorator
AdvancedCircuitBreakDecorator is a stateful circuit breaker. Not like CircuitBreakDecorator, which each client call will invoke the service function wrapped by the decorators finally, AdvancedCircuitBreakDecorator is rarely invoked the service function when it's in "OPEN" state. Refer to the following state flow.
![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/circuit_breaker_states_transtion.png)

To use AdvancedCircuitBreakDecorator to handle the timeout and max concurrency limit, the service will be decorated by both CircuitBreakDecorator and AdvancedCircuitBreakDecorator.

Be careful:
1 AdvancedCircuitBreakDecorator should be put out of CircuitBreakDecorator to get timeout or beyond max concurrency errors.
2 To let AdvancedCircuitBreakDecorator catch the errors and process the faults, not setting the fallback methods for CircuitBreakDecorator.
![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/AdvancedCircuitBreakDecorator.png)

####



### ChaosEngineeringDecorator
#### What is Chaos Engineering?
According to the principles of chaos engineering, chaos engineering is “the discipline of experimenting on a distributed system in order to build confidence in the system’s capability to withstand turbulent conditions in production.”

#### Why Chaos Engineering?
You learn how to fix the things that often break.  
You don’t learn how to fix the things that rarely break.  
If something hurts, do it more often!
For today’s large scale distributed system, normally it is impossible to simulate all the cases (including the failure modes) in a nonproductive environment. Chaos engineering is about experimenting with continuous low level of breakage to make sure the system can handle the big things.

#### What can ChaosEngineeringDecorator help in your chaos engineering practice?
By ChaosEngineeringDecorator you can inject the failures (such as, slow response, error response) into the distributed system under control. This would help you to test and improve the resilience of your system.
![image](https://github.com/easierway/service_decorators/blob/master/doc_pics/chaos_engineering_dec.png)

#### How to ChaosEngineeringDecorator?
##### Configurations
The following is the configuration about ChaosEngineeringDecorator.
```Javascript
{
	 "IsToInjectChaos" : true, // Is it to start chaos injection, if it is false, all chaos injects (chaos function, slow response) will be stopped
	 "AdditionalResponseTime" : 500, // Inject additional time spent (milseconds) to simulate slow response.
	 "ChaosRate" : 40 // The proportion of the chaos response, the range is 0-100
 }
 ```
 The configuration is stored in the storage that can be accessed with ["ConfigStorage"](https://github.com/easierway/service_decorators/blob/master/config_storage.go#L6) interface.

##### Inject Slow Response
```Go
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
```
##### Inject Error Response
```Go
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
```
