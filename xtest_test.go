package xtest

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

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
		i := 0
		for range Tick(args...) {
			i++
		}
		t.So(SliceOf(1, 10), should.Contain, i)
	})
	fn.In(10, -1)
	fn.In(time.Millisecond * 10)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestCount() {
	fn := TestFuncWith(func(n int) {
		i := 0
		for range Count(n) {
			i++
		}
		t.So(SliceOf(0, 10), should.Contain, i)
	})
	fn.In(10, -1)
	t.So(fn.Do(), should.Equal, nil)

}

func (t *xtestFixture) TestRangeString() {
	fn := TestFuncWith(func(min, max int) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			switch err := err.(type) {
			case error:
				t.Println(err.Error())
				t.So(SliceOf("invalid argument to Intn", "runtime error: makeslice: len out of range"), should.Contain, err.Error())
			case string:
				t.So("invalid argument to Intn", should.Equal, err)
			default:
				panic(err)
			}
		}()

		dt := RangeString(min, max)
		t.Assert(len(dt) < max && len(dt) >= min)
	})
	fn.In(-10, 0, 10)
	fn.In(-10, 0, 10, 20)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestMockRegister() {
	fn := TestFuncWith(func(fns ...interface{}) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			switch err := err.(type) {
			case error:
				t.So(SliceOf(ErrParamIsNil, ErrParamTypeNotFunc), ShouldContain, err)
			default:
				panic(err)
			}
		}()

		MockRegister(fns...)
	})
	fn.In(
		nil,
		"hello",
		func() {},
	)
	t.So(fn.Do(), should.BeNil)
}

func (t *xtestFixture) TestFuncCost() {
	fn := TestFuncWith(func(fn func()) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			t.So(err, should.HaveSameTypeAs, errors.New(""))
			err = err.(error)
			t.So(err, should.Equal, ErrParamIsNil)
		}()

		t.So(SliceOf(time.Duration(1), time.Duration(0)), should.Contain, CostWith(fn)/time.Millisecond)
	})
	fn.In(
		nil,
		func() {},
		func() { time.Sleep(time.Millisecond) },
	)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestTry() {
	e := errors.New("error")
	fn := TestFuncWith(func(fn func()) {
		err := Try(fn)
		t.So(SliceOf(nil, ErrParamIsNil, e), ShouldContain, err)
		switch err {
		case ErrParamIsNil:
			t.So(fn, should.Equal, nil)
		}
	})
	fn.In(
		nil,
		func() {},
		func() { panic(e) },
	)
	t.So(fn.Do(), should.BeNil)
}

func (t *xtestFixture) TestTimeoutWith() {
	var err1 = errors.New("hello")
	fn := TestFuncWith(func(dur time.Duration, fn func()) error {
		err := TimeoutWith(dur, fn)
		t.So(SliceOf(nil, ErrParamIsNil, ErrFuncTimeout, ErrDurZero, err1), ShouldContain, err)

		switch err {
		case ErrParamIsNil:
			t.So(fn, should.Equal, nil)
		case ErrFuncTimeout:
			t.So(CostWith(fn), should.BeGreaterThan, dur)
		case ErrDurZero:
			t.So(dur, should.BeLessThan, 0)
		}
		return err
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
	t.So(fn.Do(), should.BeNil)
}

func TestTimeoutWith(t *testing.T) {
	var err1 = errors.New("hello")
	Convey("TimeoutWith", t, func() {
		fn := TestFuncWith(func(dur time.Duration, fn func()) {
			Convey(fmt.Sprint("dur=", dur, "  fn=", reflect.ValueOf(fn)), func() {
				err := TimeoutWith(dur, fn)
				So(SliceOf(nil, ErrParamIsNil, ErrFuncTimeout, ErrDurZero, err1), ShouldContain, err)

				switch err {
				case ErrParamIsNil:
					So(fn, should.Equal, nil)
				case ErrFuncTimeout:
					So(CostWith(fn), should.BeGreaterThan, dur)
				case ErrDurZero:
					So(dur, should.BeLessThan, 0)
				}
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
		)
		So(fn.Do(), should.BeNil)
	})
}
