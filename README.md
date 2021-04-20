# xtest
simple testing framework for go

## example

```go
package xtest

import (
	"errors"
	"fmt"
	"testing"
	"time"

    "github.com/pubgo/xerror"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/smartystreets/gunit"
)

func TestXTest(t *testing.T) {
	gunit.Run(new(xtestFixture), t, gunit.Options.AllSequential())
}

type xtestFixture struct {
	*gunit.Fixture
}

func (t *xtestFixture) TestTick() {

	fn := TestFuncWith(func(args ...interface{}) {
		defer xerror.RespExit()

		i := 0
		for range Tick(args...) {
			i++
		}
		t.So(SliceOf(1, 10), should.Contain, i)
	})
	fn.In(10, -1)
	fn.In(time.Millisecond * 10)
	fn.Do()
}

func (t *xtestFixture) TestCount() {
	fn := TestFuncWith(func(n int) {
		defer xerror.RespExit()

		i := 0
		for range Count(n) {
			i++
		}
		t.So(SliceOf(0, 10), should.Contain, i)
	})
	fn.In(10, -1)
	fn.Do()
}

func TestRangeString(t *testing.T) {
	Convey("RangeString", t, func() {
		fn := TestFuncWith(func(min, max int) {
			Convey(fmt.Sprint("min=", min, "  max=", max), func() {
				defer xerror.Resp(func(err xerror.XErr) {
					switch err.Error() {
					case "invalid argument to Intn", "runtime error: makeslice: len out of range":
						So(err, ShouldNotEqual, "")
					default:
						xerror.Exit(err)
					}
				})

				dt := RangeString(min, max)
				So(len(dt) < max && len(dt) >= min, ShouldBeTrue)
			})
		})
		fn.In(-10, 0, 10)
		fn.In(-10, 0, 10, 20)
		fn.Do()
	})
}

func (t *xtestFixture) TestFuncCost() {
	fn := TestFuncWith(func(fn func()) {
		defer xerror.Resp(func(err xerror.XErr) {
			switch err := err.Unwrap(); err {
			case ErrParamIsNil:
			default:
				xerror.Exit(err)
			}
		})

		t.So(SliceOf(time.Duration(1), time.Duration(0)), should.Contain, CostWith(fn)/time.Millisecond)
	})
	fn.In(
		nil,
		func() {},
		func() { time.Sleep(time.Millisecond) },
	)
	fn.Do()
}

func (t *xtestFixture) TestTry() {
	e := errors.New("error")
	fn := TestFuncWith(func(fn func()) {
		defer xerror.Resp(func(err xerror.XErr) {
			switch err.Unwrap() {
			case ErrParamIsNil:
				t.So(fn, ShouldBeNil)
			case e:
			default:
				xerror.Exit(err)
			}
		})
		xerror.Panic(Try(fn))
	})
	fn.In(
		nil,
		func() {},
		func() { panic(e) },
	)
	fn.Do()
}

func (t *xtestFixture) TestTimeoutWith() {
	var err1 = errors.New("hello")
	fn := TestFuncWith(func(dur time.Duration, fn func()) {
		defer xerror.Resp(func(err xerror.XErr) {
			switch err.Unwrap() {
			case ErrParamIsNil:
				t.So(fn, ShouldBeNil)
			case ErrFuncTimeout:
				t.So(CostWith(fn), should.BeGreaterThan, dur)
			case ErrDurZero:
				t.So(dur, should.BeLessThan, 0)
			case err1:
			default:
				xerror.Exit(err)
			}
		})

		xerror.Panic(TimeoutWith(dur, fn))

	})
	fn.In(time.Duration(-1), time.Millisecond*10)
	fn.In(
		nil,
		func() {},
		func() {
			time.Sleep(time.Millisecond * 20)
		},
		func() {
			panic(err1)
		},
	)
	fn.Do()
}

func TestTimeoutWith(t *testing.T) {
	var err1 = errors.New("hello")
	var err2 = "hello"
	Convey("TimeoutWith", t, func() {
		fn := TestFuncWith(func(dur time.Duration, fn func()) {
			Convey(fmt.Sprint("dur=", dur, "  fn=", FuncSprint(fn)), func() {
				defer xerror.Resp(func(err xerror.XErr) {
					switch err.Unwrap() {
					case ErrParamIsNil:
						So(fn, ShouldBeNil)
					case ErrFuncTimeout:
						So(CostWith(fn), should.BeGreaterThan, dur)
					case ErrDurZero:
						So(dur, should.BeLessThan, 0)
					case err1:
						So(nil, ShouldBeNil)
					default:
						if err.Error() == err2 {
							So(nil, ShouldBeNil)
							return
						}
						xerror.Exit(err)
					}
				})

				xerror.Panic(TimeoutWith(dur, fn))
			})
		})
		fn.In(time.Duration(-1), time.Millisecond*10)
		fn.In(
			nil,
			func() {},
			func() {
				time.Sleep(time.Millisecond * 20)
			},
			func() {
				panic(err1)
			},
			func() {
				panic(err2)
			},
		)
		fn.Do()
	})
}
```