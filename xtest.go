package xtest

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/smartystreets/assertions"
)

func Rand(data ...interface{}) interface{} {
	return data[rand.Intn(len(data))]
}

func RandS(strList ...string) string {
	return strList[rand.Intn(len(strList))]
}

type Fixture struct {
	RunNum uint
	assert *assertions.Assertion
	data   map[string]func() interface{}
}

func (t *Fixture) So(
	actual interface{},
	assert func(actual interface{}, expected ...interface{}) string,
	expected ...interface{}) bool {
	return t.assert.So(actual, assert, expected)
}

func (t *Fixture) Failed() bool {
	return t.assert.Failed()
}

func (t *Fixture) InitHandlerParam(name string, fn func() interface{}) {
	xerror.Assert(name == "" || fn == nil, "name or fn is null")

	if t.data == nil {
		t.data = make(map[string]func() interface{})
	}

	t.data[name] = fn

}

type handler struct {
	fn    reflect.Value
	stack string
}

func serviceMethod(val interface{}) map[string]*handler {
	var data = make(map[string]*handler)
	var t = reflect.TypeOf(val)
	var v = reflect.ValueOf(val)
	for i := t.NumMethod() - 1; i >= 0; i-- {
		var s = stack.Func(t.Method(i).Func.Interface())
		data[t.Method(i).Name] = &handler{
			fn:    v.Method(i),
			stack: s,
		}
	}
	return data
}

type Test interface {
	Setup()
	Teardown()
}

type grpcTest struct {
	tt      Test
	fixture *Fixture
	t       *testing.T
	srv     map[string]*handler
	fn      interface{}
}

func (t *grpcTest) Do() {
	defer xerror.RespExit()

	fmt.Println("Setup:", t.srv["Setup"].stack)
	t.tt.Setup()
	t.t.Cleanup(func() {
		fmt.Println("Teardown:", t.srv["Teardown"].stack)
		t.tt.Teardown()
	})

	var cache = make(map[string]bool)
	//var id = time.Now().Unix()

	// record uuid
	for name, h := range t.srv {
		if t.fixture.data[name] == nil {
			continue
		}

		fmt.Println(name+":", h.stack)
		t.t.Run(name, func(tt *testing.T) {
			wfn := fx.WrapRaw(h.fn)
			for i := uint(0); i < t.fixture.RunNum; i++ {
				ppp := t.fixture.data[name]()
				var nnn = fmt.Sprintf("%v", ppp)
				if cache[nnn] {
					continue
				}

				cache[nnn] = true

				tt.Run(fmt.Sprintf("%#v", ppp), func(tt *testing.T) {
					resp := wfn(ppp)
					var err = resp[1]

					// ok
					if !err.IsValid() || err.IsNil() {
						return
					}

					switch err.Interface().(type) {
					case *okErr:
						tt.Log(err.Interface())
					case *failErr:
						tt.Error(err.Interface())
					case error:
						tt.Fatal(err.Interface())
					}

					tt.Logf("resp: %#v", resp[0])
				})
			}
		})
	}
}

func Run(t *testing.T, tests ...Test) {
	assert := assertions.New(t)
	for i := range tests {
		var tt = tests[i]
		if tt == nil {
			panic(fmt.Sprintf("tests[%d] is nil", i))
		}

		optsV := reflect.ValueOf(tt).Elem().FieldByName("Fixture")
		if !optsV.IsValid() {
			panic("has not [Fixture] field")
		}

		var gt = &grpcTest{fixture: &Fixture{assert: assert, RunNum: 100}, tt: tt, t: t, srv: serviceMethod(tt)}
		optsV.Set(reflect.ValueOf(gt.fixture))
		gt.Do()
	}
}

type okErr struct {
	error
}

type failErr struct {
	error
}

func Ok(err error) error {
	return &okErr{err}
}

func Fail(err error) error {
	return &failErr{err}
}
