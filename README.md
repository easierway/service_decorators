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
