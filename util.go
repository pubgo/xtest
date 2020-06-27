package xtest

import (
	"fmt"
	"github.com/pubgo/xerror"
	"reflect"
	"runtime"
	"sync"
	"time"
)

func Wrap(fn interface{}) func(...interface{}) func(...interface{}) (err error) {
	if fn == nil {
		xerror.Panic(ErrParamIsNil)
	}

	_tr := tryWrap(reflect.ValueOf(fn))
	return func(args ...interface{}) func(...interface{}) (err error) {
		var _args = valueGet()
		defer valuePut(_args)

		for _, k := range args {
			_args = append(_args, reflect.ValueOf(k))
		}
		_tr1 := _tr(_args...)
		return func(cfn ...interface{}) (err error) {
			var _cfn = valueGet()
			defer valuePut(_cfn)

			for _, k := range cfn {
				_cfn = append(_cfn, reflect.ValueOf(k))
			}
			return _tr1(_cfn...)
		}
	}
}

func tryWrap(fn reflect.Value) func(...reflect.Value) func(...reflect.Value) (err error) {
	if fn.Type().Kind() != reflect.Func {
		xerror.Panic(ErrParamTypeNotFunc)
	}

	numIn := fn.Type().NumIn()
	var variadicType reflect.Value
	var isVariadic = fn.Type().IsVariadic()
	if isVariadic {
		variadicType = reflect.New(fn.Type().In(numIn - 1).Elem()).Elem()
	}

	numOut := fn.Type().NumOut()
	return func(args ...reflect.Value) func(...reflect.Value) (err error) {
		if !isVariadic && numIn != len(args) || isVariadic && len(args) < numIn-1 {
			xerror.PanicF(ErrInputParamsNotMatch, "func: %s, func(%d,%d)", fn.Type(), numIn, len(args))
		}

		for i, k := range args {
			if !k.IsValid() {
				args[i] = reflect.New(fn.Type().In(i)).Elem()
				continue
			}

			switch k.Kind() {
			case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
				if k.IsNil() {
					args[i] = reflect.New(fn.Type().In(i)).Elem()
					continue
				}
			}

			if isVariadic {
				args[i] = variadicType
			}

			args[i] = k
		}

		return func(cfn ...reflect.Value) (err error) {
			defer func() {
				if err1 := recover(); err1 != nil {
					switch err1 := err1.(type) {
					case error:
						err = err1
					default:
						err = fmt.Errorf("%s -> [func: %s] [input: %s]", err1, fn.String(), args)
					}
				}
			}()
			defer valuePut(args)

			_c := fn.Call(args)
			if len(cfn) > 0 && cfn[0].IsValid() && !cfn[0].IsZero() {
				if cfn[0].Type().NumIn() != numOut {
					xerror.PanicF(ErrInputOutputParamsNotMatch, "[%d]<->[%d]", cfn[0].Type().NumIn(), fn.Type().NumOut())
				}

				if cfn[0].Type().NumIn() != 0 && cfn[0].Type().In(0) != fn.Type().Out(0) {
					xerror.PanicF(ErrFuncOutputTypeNotMatch, "[%s]<->[%s]", cfn[0].Type().In(0), fn.Type().Out(0))
				}
				cfn[0].Call(_c)
			}
			return
		}
	}
}

var _valuePool = sync.Pool{
	New: func() interface{} {
		return []reflect.Value{}
	},
}

func valueGet() []reflect.Value {
	return _valuePool.Get().([]reflect.Value)
}

func valuePut(v []reflect.Value) {
	_valuePool.Put(v[:0])
}

func FuncSprint(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	fn := args[0]
	if fn == nil || !reflect.ValueOf(fn).IsValid() || reflect.ValueOf(fn).IsNil() {
		return "nil"
	}

	return fmt.Sprintf("[%d][%s]", reflect.ValueOf(fn).Pointer(), reflect.TypeOf(fn).String())
}

func SliceOf(args ...interface{}) []interface{} {
	return args
}

// Tick 简单定时器
// Example: Tick(100, time.Second)
func Tick(args ...interface{}) <-chan time.Time {
	var n int
	var dur time.Duration

	for _, arg := range args {
		switch ag := arg.(type) {
		case int:
			n = ag
		case time.Duration:
			dur = ag
		}
	}

	if n <= 0 {
		n = 1
	}
	if dur <= 0 {
		dur = time.Second
	}

	var c = make(chan time.Time)
	go func() {
		tk := time.NewTicker(dur)
		var i int
		for t := range tk.C {
			if i == n {
				tk.Stop()
				break
			}
			i++
			c <- t
		}
		close(c)
	}()
	return c
}

// Count ...
func Count(n int) <-chan int {
	var c = make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			c <- i
		}
		close(c)
	}()
	return c
}

func Try(fn func()) (e error) {
	if fn == nil {
		return ErrParamIsNil
	}

	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				e = err
			default:
				e = fmt.Errorf("%s", err)
			}
		}
	}()
	fn()
	return
}

func TimeoutWith(dur time.Duration, fn func()) error {
	if dur < 0 {
		return xerror.WrapF(ErrDurZero, "dur: %s", dur)
	}

	if fn == nil {
		return ErrParamIsNil
	}

	var ch = make(chan error, 1)
	go func() {
		defer xerror.RespChanErr(ch)
		fn()
		ch <- nil
	}()

	select {
	case err := <-ch:
		return err
	case <-time.After(dur):
		return xerror.Wrap(ErrFuncTimeout)
	}
}

func CostWith(fn func()) (dur time.Duration) {
	if fn == nil {
		xerror.Panic(ErrParamIsNil)
		return
	}

	defer func(start time.Time) {
		dur = time.Since(start)
	}(time.Now())

	fn()
	return
}

func PrintMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("HeapAlloc = %v HeapIdel= %v HeapSys = %v  HeapReleased = %v\n", m.HeapAlloc/1024, m.HeapIdle/1024, m.HeapSys/1024, m.HeapReleased/1024)
}
