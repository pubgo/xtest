package xtest

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xtest/model"
)

func serviceMethod(val interface{}) map[string]map[string]reflect.Value {
	var data = make(map[string]map[string]reflect.Value)
	var t = reflect.TypeOf(val)
	var name = t.String()
	data[name] = make(map[string]reflect.Value)

	var v = reflect.ValueOf(val)
	for i := t.NumMethod() - 1; i >= 0; i-- {
		data[name][t.Method(i).Name] = v.Method(i)
	}
	return data
}

type Request struct {
	Gen  interface{}
	Data [][]interface{}
}

type Test interface {
	Setup()
	Teardown()
	GenReq() map[string]Request
}

type grpcTest struct {
	tt     Test
	t      *testing.T
	name   string
	req    map[string]Request
	srv    map[string]map[string]reflect.Value
	fn     interface{}
	params [][]interface{}
}

func randVal(data ...interface{}) interface{} {
	return data[rand.Intn(len(data))]
}

func (t *grpcTest) Do() {
	defer xerror.RespExit()

	// setup
	// defer teardown
	t.t.Log("Setup", stack.Func(t.tt.Setup))
	t.tt.Setup()
	t.t.Cleanup(func() {
		t.t.Log("Teardown")
		t.tt.Teardown()
	})

	var cache = make(map[string]bool)
	//asseet := assertions.New(t.t)

	// record uuid
	for srv, methods := range t.srv {
		t.t.Run(srv, func(tt *testing.T) {
			for name, fn := range methods {

				if t.req[name].Gen == nil {
					continue
				}

				tt.Run(name, func(tt *testing.T) {
					wfn := fx.WrapReflect(fn)
					for i := 0; i < 10; i++ {
						var ppp []interface{}
						for _, ddd := range t.req[name].Data {
							ppp = append(ppp, randVal(ddd...))
						}

						var nnn = fmt.Sprintf("%v", ppp)
						if cache[nnn] {
							continue
						}

						cache[nnn] = true

						var ddddd = fx.WrapRaw(t.req[name].Gen)(ppp...)
						tt.Run(fmt.Sprintf("%#v", ddddd[0].Interface()), func(t *testing.T) {
							_ = ddddd
							_ = wfn
							//resp := wfn(ddddd...)
							//var dt = resp[0]
							//var err = resp[1]
							//_ = dt
							//// err check
							//// save result
							//convey.So(err, convey.ShouldBeNil)

							//asseet.So(nil, should.BeNil)
							fmt.Printf("%#v\n", &model.Result{Name: name, Service: srv, Request: ddddd[0].Interface()})
						})
					}
				})
			}
		})
	}
}

func Run(t *testing.T, tests ...Test) {
	for i := range tests {
		(&grpcTest{tt: tests[i], t: t, srv: serviceMethod(tests[i]), req: tests[i].GenReq()}).Do()
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
