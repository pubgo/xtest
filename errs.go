package xtest

import (
	"github.com/pubgo/xerror"
)

var (
	ErrXTest                     = xerror.New("grpcTest error")
	ErrParamIsNil                = ErrXTest.New("the parameter is nil")
	ErrFuncTimeout               = ErrXTest.New("the func is timeout")
	ErrParamTypeNotFunc          = ErrXTest.New("the type of the parameters is not func")
	ErrDurZero                   = ErrXTest.New("the duration time must more than zero")
	ErrInputParamsNotMatch       = ErrXTest.New("the input params of func is not match")
	ErrInputOutputParamsNotMatch = ErrXTest.New("the input num and output num of the callback func is not match")
	ErrFuncOutputTypeNotMatch    = ErrXTest.New("the  output type of the callback func is not match")
	ErrForeachParameterNil       = ErrXTest.New("the parameter of [Foreach] must not be nil")
)
