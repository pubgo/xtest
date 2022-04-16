package xtest

import (
	"fmt"
	"github.com/pubgo/x/stack"
	"reflect"
	"testing"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type Param [][]interface{}

type Fixture struct {
	data map[string]*Params
}

func (t *Fixture) AddParamHandleFunc(name string, params Param, fn interface{}) {
	if t.data == nil {
		t.data = make(map[string]*Params)
	}
	var p = &Params{fn: fn}
	for i := range params {
		p.In(params[i]...)
	}
	t.data[name] = p
}

func serviceMethod(val interface{}) map[string]map[string]reflect.Value {
	var data = make(map[string]map[string]reflect.Value)
	var t = reflect.TypeOf(val)
	var name = t.String()
	data[name] = make(map[string]reflect.Value)

	var v = reflect.ValueOf(val)
	for i := t.NumMethod() - 1; i >= 0; i-- {
		fmt.Println(stack.Func(t.Method(i).Func.Interface()))
		data[name][t.Method(i).Name] = v.Method(i)
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
	srv     map[string]map[string]reflect.Value
	fn      interface{}
}

func (t *grpcTest) Do() {
	defer xerror.RespExit()

	t.t.Log("Setup")
	t.tt.Setup()
	t.t.Cleanup(func() {
		t.t.Log("Teardown")
		t.tt.Teardown()
	})

	var cache = make(map[string]bool)
	asseet := assertions.New(t.t)
	//var id = time.Now().Unix()

	// record uuid
	for srv, methods := range t.srv {
		t.t.Run(srv, func(tt *testing.T) {
			for name, fn := range methods {

				if t.fixture.data[name] == nil {
					continue
				}

				tt.Run(name, func(tt *testing.T) {
					wfn := fx.WrapReflect(fn)
					for _, ppp := range t.fixture.data[name].params {
						var nnn = fmt.Sprintf("%v", ppp)
						if cache[nnn] {
							continue
						}

						cache[nnn] = true

						var ret = fx.WrapRaw(t.fixture.data[name].fn)(ppp...)
						tt.Run(fmt.Sprintf("%#v", ret[0].Interface()), func(tt *testing.T) {
							resp := wfn(ret...)
							var dt = resp[0].Interface()
							var err = resp[1].Interface()
							t.t.Log(err, dt)
							asseet.So(err, should.BeNil)
							asseet.So(dt, should.NotBeNil)
							fmt.Println(stack.Func(fn.Interface()))
							//fmt.Printf("%#v\n", &model.Result{
							//	Name: name, Service: srv, Request: ret[0].Interface(), Id: int(id)})
						})
					}
				})
			}
		})
	}
}

func Run(t *testing.T, tests ...Test) {
	for i := range tests {
		var fix = &Fixture{}
		optsV := reflect.ValueOf(tests[i]).Elem().FieldByName("Fixture")
		optsV.Set(reflect.ValueOf(fix))
		(&grpcTest{
			fixture: fix,
			tt:      tests[i],
			t:       t,
			srv:     serviceMethod(tests[i]),
		}).Do()
	}
}

func TestFunc(name string, fn interface{}) *grpcTest {
	if fn == nil {
		xerror.Exit(ErrParamIsNil)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		xerror.PanicF(ErrParamTypeNotFunc, "kind: %s, name: %s", reflect.TypeOf(fn).Kind(), reflect.TypeOf(fn))
	}

	return &grpcTest{fn: fn}
}
