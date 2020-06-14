package xtest

import (
	"errors"
	"fmt"
)

var ErrXTest = func(err ...string) error {
	if len(err) == 0 {
		return errors.New("xtest error")
	}
	return fmt.Errorf("xtest error: %s", err[0])
}
var ErrParamIsNil = ErrXTest("the parameter is nil")
var ErrFuncTimeout = ErrXTest("the func is timeout")
var ErrParamTypeNotFunc = ErrXTest("the type of the parameters is not func")
var ErrDurZero = ErrXTest("duration time must more than zero")
