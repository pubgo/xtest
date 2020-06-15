package xtest

import (
	fuzz "github.com/google/gofuzz"
	"github.com/smartystreets/assertions/should"
	"reflect"
)

var fns []interface{}

// MockRegister ...
func MockRegister(fns ...interface{}) {
	for _, fn := range fns {
		if fn == nil {
			panic(ErrParamIsNil)
		}
		if reflect.TypeOf(fn).Kind() != reflect.Func {
			panic(ErrParamTypeNotFunc)
		}
		fns = append(fns, fn)
	}
}

// InErrs ...
func InErrs(d error, ds ...error) (b bool) {
	return should.Contain(ds, d) == ""
}

// AssertErrs ...
func AssertErrs(d error, ds ...error) {
	if !InErrs(d, ds...) {
		logger.Fatalf("%s, %s", d, ds)
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
			logger.Fatalln("the parameter of [Foreach] must not be nil")
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

func (t *xtest) In(args ...interface{}) *xtest {
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
	return t
}

func (t *xtest) Do() (err error) {
	wfn := Wrap(t.fn)
	for _, param := range t.params {
		if err1 := wfn(param...)(); err1 != nil {
			err = err1
		}
	}
	return
}

func TestFuncWith(fn interface{}) *xtest {
	if fn == nil {
		logger.Fatalln(ErrParamIsNil)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		logger.Fatalln(ErrParamTypeNotFunc)
	}

	return &xtest{fn: fn}
}
