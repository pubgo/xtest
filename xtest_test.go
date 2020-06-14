package xtest_test

import (
	"errors"
	"github.com/pubgo/xtest"
	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
	"testing"
	"time"
)

func TestXTest(t *testing.T) {
	gunit.Run(new(xtestFixture), t, gunit.Options.AllSequential())
}

type xtestFixture struct {
	*gunit.Fixture
}

func (t *xtestFixture) TestTick() {
	fn := xtest.TestFuncWith(func(args ...interface{}) {
		i := 0
		for range xtest.Tick(args...) {
			i++
		}
		t.So(xtest.SliceOf(1, 10), should.Contain, i)
	})
	fn.In(10, -1)
	fn.In(time.Millisecond * 10)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestCount() {
	fn := xtest.TestFuncWith(func(n int) {
		i := 0
		for range xtest.Count(n) {
			i++
		}
		t.So(xtest.SliceOf(0, 10), should.Contain, i)
	})
	fn.In(10, -1)
	t.So(fn.Do(), should.Equal, nil)

}

func (t *xtestFixture) TestRangeString() {
	fn := xtest.TestFuncWith(func(min, max int) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			switch err := err.(type) {
			case error:
				t.Println(err.Error())
				t.So(xtest.SliceOf("invalid argument to Intn", "runtime error: makeslice: len out of range"), should.Contain, err.Error())
			case string:
				t.So("invalid argument to Intn", should.Equal, err)
			default:
				panic(err)
			}
		}()

		dt := xtest.RangeString(min, max)
		t.Assert(len(dt) < max && len(dt) >= min)
	})
	fn.In(-10, 0, 10)
	fn.In(-10, 0, 10, 20)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestMockRegister() {
	fn := xtest.TestFuncWith(func(fns ...interface{}) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			switch err := err.(type) {
			case error:
				t.So(xtest.SliceOf(xtest.ErrParamIsNil, xtest.ErrParamTypeNotFunc), should.Contain, err)
			default:
				panic(err)
			}
		}()

		xtest.MockRegister(fns...)
	})
	fn.In(
		nil,
		"hello",
		func() {},
	)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestFuncCost() {
	fn := xtest.TestFuncWith(func(fn func()) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			t.So(err, should.HaveSameTypeAs, errors.New(""))
			err = err.(error)
			t.So(err, should.Equal, xtest.ErrParamIsNil)
		}()

		t.So(xtest.SliceOf(time.Duration(1), time.Duration(0)), should.Contain, xtest.CostWith(fn)/time.Millisecond)
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
	fn := xtest.TestFuncWith(func(fn func()) {
		err := xtest.Try(fn)
		xtest.AssertErrs(err, nil, xtest.ErrParamIsNil, e)
		switch err {
		case xtest.ErrParamIsNil:
			t.So(fn, should.Equal, nil)
		}
	})
	fn.In(
		nil,
		func() {},
		func() { panic("error") },
	)
	t.So(fn.Do(), should.Equal, nil)
}

func (t *xtestFixture) TestTimeoutWith() {
	var err1 = errors.New("hello")
	fn := xtest.TestFuncWith(func(dur time.Duration, fn func()) error {
		err := xtest.TimeoutWith(dur, fn)
		xtest.AssertErrs(err, nil, xtest.ErrParamIsNil, xtest.ErrFuncTimeout, xtest.ErrDurZero, err1)

		switch err {
		case xtest.ErrParamIsNil:
			t.So(fn, should.Equal, nil)
		case xtest.ErrFuncTimeout:
			t.So(xtest.CostWith(fn), should.BeGreaterThan, dur)
		case xtest.ErrDurZero:
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
	t.So(fn.Do(), should.Equal, nil)
}
