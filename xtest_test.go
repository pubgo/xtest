package xtest

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

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

func TestExampleFixture(t *testing.T) {
	Run(t, new(ExampleFixture))
}

type ExampleFixture struct {
	*Fixture
	i int
}

func (t *ExampleFixture) Setup() {
	t.InitHandlerParam("Hello", func() interface{} {
		return &Hello{
			Name:   RandS("hello", "world", "world1"),
			HName1: RandS("hello", "world", "world1"),
		}
	})

	t.i++
	log.Println("SetupStuff", t.i)
}

func (t *ExampleFixture) Teardown() {
}

type Hello struct {
	Name   string `json:"name"`
	HName1 string `json:"hName1"`
}

// This is an actual test case:
func (t *ExampleFixture) Hello(req *Hello) (*Hello, error) {
	return req, errors.New("ss")
}
