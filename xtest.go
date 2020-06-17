package xtest

import (
	fuzz "github.com/google/gofuzz"
	"github.com/pubgo/xerror"
	"reflect"
)

var fns []interface{}

// MockRegister ...
func MockRegister(fns ...interface{}) {
	for _, fn := range fns {
		if fn == nil {
			xerror.Panic(ErrParamIsNil)
		}

		if reflect.TypeOf(fn).Kind() != reflect.Func {
			xerror.Panic(ErrParamTypeNotFunc)
		}

		fns = append(fns, fn)
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
			xerror.Panic(ErrForeachParameterNil)
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

type xtest struct {
	fn     interface{}
	params [][]interface{}
}

func (t *xtest) In(args ...interface{}) {
	var params [][]interface{}
	if len(t.params) == 0 {
		for _, arg := range args {
			params = append(params, []interface{}{arg})
		}
	} else {
		for _, p := range t.params {
			for _, arg := range args {
				params = append(params, append(p, arg))
			}
		}
	}
	t.params = params
}

func (t *xtest) Do() {
	wfn := Wrap(t.fn)
	for _, param := range t.params {
		_ = wfn(param...)()
	}
	return
}

func TestFuncWith(fn interface{}) *xtest {
	if fn == nil {
		xerror.Panic(ErrParamIsNil)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		xerror.Panic(ErrParamTypeNotFunc)
	}

	return &xtest{fn: fn}
}
