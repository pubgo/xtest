package xtest

import (
	"fmt"
	"log"
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
	dd convey.FailureMode
	fv convey.C
	*gunit.Fixture
}

func TestRangeString(t *testing.T) {
	convey.Convey("RangeString", t, func() {
		fn := TestFunc("RangeString", func(min, max int) {
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

var i = 1

func TestExampleFixture(t *testing.T) {
	Run(t, new(ExampleFixture))
}

type ExampleFixture struct {
	*gunit.Fixture
	i int
}

func (t *ExampleFixture) Teardown() {
}

func (t *ExampleFixture) GenReq() map[string]Request {
	return map[string]Request{
		"Hello": {
			Gen: func(name string, name1 string) *Hello {
				return &Hello{Name: name, HName1: name1}
			},
			Data: [][]interface{}{
				{"hello", "world", "world1"},
				{"hello", "world", "world1"},
			},
		},
	}
}

func (t *ExampleFixture) Setup() {
	t.i++
	i++
	log.Println("SetupStuff", t.i, i)
}

type Hello struct {
	Name   string
	HName1 string
}

// This is an actual test case:
func (t *ExampleFixture) Hello(req *Hello) (*Hello, error) {
	return req, nil
}
