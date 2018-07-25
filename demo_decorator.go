package service_decorators

import (
	"fmt"
)

type demoDecorator struct {
	InvokedCnt int
	SucceedCnt int
	FailedCnt  int
}

func createDemoDecorator() *demoDecorator {
	return &demoDecorator{0, 0, 0}
}

func (d *demoDecorator) Decorate(innerFn ServiceFunc) ServiceFunc {
	return func(req Request) (Response, error) {
		fmt.Printf("Demo deccorator entry\n")
		d.InvokedCnt++
		resp, err := innerFn(req)
		if err != nil {
			d.FailedCnt++
		} else {
			d.SucceedCnt++
		}
		fmt.Printf("Demo deccorator exist\n")
		return resp, err

	}
}
