package xtest

import (
	"errors"
	"fmt"
)

var (
	ErrXTest = func(err ...string) error {
		if len(err) == 0 {
			return errors.New("xtest error")
		}
		return fmt.Errorf("xtest error: %s", err[0])
	}
	ErrParamIsNil       = ErrXTest("the parameter is nil")
	ErrFuncTimeout      = ErrXTest("the func is timeout")
	ErrParamTypeNotFunc = ErrXTest("the type of the parameters is not func")
	ErrDurZero          = ErrXTest("duration time must more than zero")
)
