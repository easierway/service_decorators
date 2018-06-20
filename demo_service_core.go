package service_decorators

import (
	"errors"
)

type sumRequest []int

type sumResponse int

func sum(req Request) (Response, error) {
	if request, ok := req.(sumRequest); ok {
		var sum int
		for _, i := range request {
			sum += i
		}
		return sum, nil
	}
	return nil, errors.New("The request []int is required. ")

}
