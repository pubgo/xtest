package xtest

import (
	"fmt"
	fuzz "github.com/google/gofuzz"
	"github.com/smartystreets/assertions/should"
	"math/rand"
	"reflect"
	"time"
)

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
		return ErrDurZero
	}

	if fn == nil {
		return ErrParamIsNil
	}

	var ch = make(chan error, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				switch err := err.(type) {
				case error:
					ch <- err
				default:
					ch <- fmt.Errorf("%s", err)
				}
			}
		}()
		fn()
		ch <- nil
	}()

	select {
	case err := <-ch:
		return err
	case <-time.After(dur):
		return ErrFuncTimeout
	}
}

func CostWith(fn func()) (dur time.Duration) {
	if fn == nil {
		panic(ErrParamIsNil)
	}

	defer func(start time.Time) {
		dur = time.Since(start)
	}(time.Now())

	fn()
	return
}

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

// RangeBytes ...
func RangeBytes(min, max int) []byte {
	var dt = make([]byte, Range(min, max))
	rand.Read(dt)
	return dt
}

// RangeString ...
func RangeString(min, max int) string {
	return string(RangeBytes(min, max))
}

// RangeDur ...
func RangeDur(min, max time.Duration) time.Duration {
	return time.Duration(Range(int(min), int(max)))
}

// Range ...
func Range(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
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
