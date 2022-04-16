package main_test

import (
	"errors"
	"testing"

	"github.com/pubgo/xtest"
)

func TestExample(t *testing.T) {
	xtest.Run(t, new(Example))
}

type Example struct {
	*xtest.Fixture
	i int
}

func (t *Example) Setup() {
	t.InitHandlerParam("Hello", func() interface{} {
		return &Hello{
			Name:   xtest.RandS("hello", "world", "world1"),
			HName1: xtest.RandS("hello", "world", "world1"),
		}
	})

	t.i++
}

func (t *Example) Teardown() {
}

type Hello struct {
	Name   string `json:"name"`
	HName1 string `json:"hName1"`
}

// This is an actual test case:
func (t *Example) Hello(req *Hello) (*Hello, error) {
	return req, xtest.Fail(errors.New("ss"))
}
