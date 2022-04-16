package xtest

import (
	fuzz "github.com/google/gofuzz"
	"github.com/pubgo/xerror"

	"math/rand"
	"reflect"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// MockBytes ...
func MockBytes(min, max int) []byte {
	var dt = make([]byte, MockInt(min, max))
	rand.Read(dt)
	return dt
}

// MockString ...
func MockString(min, max int) string {
	return string(MockBytes(min, max))
}

// MockDur ...
func MockDur(min, max time.Duration) time.Duration {
	return time.Duration(MockInt(int(min), int(max)))
}

// MockInt ...
func MockInt(min, max int) int {
	if min >= max {
		return max
	}
	return min + rand.Intn(max-min)
}

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

// MockCheck ...
func MockCheck(args ...bool) {
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

// InErrs ...
func InErrs(d error, ds ...error) (b bool) {
	for _, d1 := range ds {
		if d == d1 {
			return true
		}
	}
	return false
}
