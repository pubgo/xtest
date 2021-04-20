package xtest

import (
	"fmt"
	"testing"
	"time"

	"github.com/pubgo/xerror"
	"github.com/smartystreets/goconvey/convey"
	"github.com/smartystreets/gunit"
)

func TestXTest(t *testing.T) {
	defer Leak(t)
	gunit.Run(new(xtestFixture), t, gunit.Options.AllSequential())
}

type xtestFixture struct {
	*gunit.Fixture
}

func TestRangeString(t *testing.T) {
	convey.Convey("RangeString", t, func() {
		fn := TestFunc(func(min, max int) {
			convey.Convey(fmt.Sprint("min=", min, "  max=", max), func() {
				defer xerror.Resp(func(err xerror.XErr) {
					switch err.Error() {
					case "invalid argument to Intn", "runtime error: makeslice: len out of range":
						convey.So(err, convey.ShouldNotEqual, "")
					default:
						xerror.Exit(err)
					}
				})

				dt := MockString(min, max)
				convey.So(len(dt) < max && len(dt) >= min, convey.ShouldBeTrue)
			})
		})
		fn.In(-10, 0, 10)
		fn.In(-10, 0, 10, 20)
		fn.Do()
	})
}

func TestBen(t *testing.T) {
	MemStatsPrint()
	bm := Benchmark(1000).
		Do(func(b B) {
			time.Sleep(time.Millisecond)
		})
	MemStatsPrint()
	fmt.Println(bm)
	Debugln()
}
