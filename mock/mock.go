package mock

import (
	fuzz "github.com/google/gofuzz"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xtest"
	"reflect"
)

var fns []interface{}

// MockRegister ...
func MockRegister(fns ...interface{}) {
	for _, fn := range fns {
		if fn == nil {
			xerror.Panic(xtest.ErrParamIsNil)
		}

		if reflect.TypeOf(fn).Kind() != reflect.Func {
			xerror.Panic(xtest.ErrParamTypeNotFunc)
		}

		fns = append(fns, fn)
	}
}

// Check ...
func Check(args ...bool) {
	for _, arg := range args {
		if arg {
			panic(arg)
		}
	}
}

// Mock ...
func Mock(args ...interface{}) {
	var (
		n  int
		fn interface{}
		fz = fuzz.New()
	)

	for i := range args {
		if args[i] == nil {
			xerror.Panic(xtest.ErrForeachParameterNil)
		}

		switch reflect.TypeOf(args[i]).Kind() {
		case reflect.Int:
			n = args[i].(int)
		case reflect.Func:
			fn = args[i]
		}
	}

	if n <= 0 {
		n = 1
	}

	rfn := reflect.ValueOf(fn)
	numIn := reflect.TypeOf(fn).NumIn()
	var i = 0
	for i < n {
		func() {
			defer func() {
				if err := recover(); err == nil {
					i++
				}
			}()

			if numIn == 0 {
				rfn.Call([]reflect.Value{})
			} else {
				var res = reflect.New(rfn.Type().In(0))
				reflect.ValueOf(fz.Funcs(fns...).Fuzz).Call([]reflect.Value{res})
				rfn.Call([]reflect.Value{res.Elem()})
			}
		}()
	}
}

// InErrs ...
func InErrs(d error, ds ...error) (b bool) {
	for _, d1 := range ds {
		if d == d1 {
			return true
		}
	}
	return false
}
