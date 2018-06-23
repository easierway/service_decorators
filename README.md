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

![image](https://github.com/easierway/service_decorators/doc_pics/other_functions.png)

Most of them are much complicated than RPC, in some cases they are even complicated than your core business logic.

For example, the code section of rate limit

```Go
...
bucket := make(chan struct{}, numOfReqs)
//fill the bucket firstly
for j := 0; j < numOfReqs; j++ {
  bucket <- struct{}{}
}

go func() {
  for _ = range time.Tick(interval) {
    for i := 0; i < numOfReqs; i++ {
      bucket <- struct{}{}
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

![image](https://github.com/easierway/service_decorators/doc_pics/to_do.png)

The goal of the project is to reduce your work on what you have to do and let you only focus on what you want to do.

All these common functions (e.g. rate limit, circuit break, metric) are encapsulated in the components, and these common logics can be injected in your service implementation transparently by leveraging decorator pattern.

![image](https://github.com/easierway/service_decorators/doc_pics/decorator_pattern.png)

![image](https://github.com/easierway/service_decorators/doc_pics/decorators.png)

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
