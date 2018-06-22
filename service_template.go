/*
Package service_decorators is to simplify your work on building microservices.
The common functions for the microservices (such as, Circuit break, Rate limit,
Metric...) have be encapsulated in the reusable components(decorators).
To build a service is to decorate the core business logic with the common
decorators, so you can only focus on the core business logic.
@Auth chaocai2001@icloud.com
@Created on 2018-6
*/
package service_decorators

//Request is the interface of the service request.
type Request interface{}

//Response is the interface of the service response.
type Response interface{}

//ServiceFunc is the service function definition.
//To leverage the prebuilt decorators, the service function signature should follow it.
type ServiceFunc func(req Request) (Response, error)

//ServiceFallbackFunc is the fallback function definition
type ServiceFallbackFunc func(req Request, err error) (Response, error)

//Decorator is the interface of the decorators.
type Decorator interface {
	//Decorate function is to introdoce decorator's the functions
	Decorate(ServiceFunc) ServiceFunc
}
