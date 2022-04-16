package xtest

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/smartystreets/assertions"
)

type Fixture struct {
	assert *assertions.Assertion
	data   map[string]*Params
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

func (t *Fixture) InitHandlerParam(name string, pf func(p *Params), fn interface{}) {
	xerror.Assert(name == "" || pf == nil || fn == nil, "name or pf or fn is null")

	if t.data == nil {
		t.data = make(map[string]*Params)
	}

	t.data[name] = &Params{fn: fn}
	pf(t.data[name])

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

type Params struct {
	fn     interface{}
	params [][]interface{}
}

func (t *Params) In(args ...interface{}) *Params {
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
			wfn := fx.WrapReflect(h.fn)
			for _, ppp := range t.fixture.data[name].params {
				var nnn = fmt.Sprintf("%v", ppp)
				if cache[nnn] {
					continue
				}

				cache[nnn] = true

				var ret = fx.WrapRaw(t.fixture.data[name].fn)(ppp...)
				tt.Run(fmt.Sprintf("%#v", ret[0].Interface()), func(tt *testing.T) {
					resp := wfn(ret...)
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

		var gt = &grpcTest{fixture: &Fixture{assert: assert}, tt: tt, t: t, srv: serviceMethod(tt)}
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
