package xtest

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
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
	*testing.T
	RunNum uint
	assert *assertions.Assertion
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

type handler struct {
	fn    reflect.Value
	gen   reflect.Value
	stack string
}

func serviceMethod(val interface{}) map[string]*handler {
	var data = make(map[string]*handler)
	var t = reflect.TypeOf(val)
	var v = reflect.ValueOf(val)
	for i := t.NumMethod() - 1; i >= 0; i-- {
		if strings.HasPrefix(t.Method(i).Name, "Mock") {
			continue
		}

		var genName = fmt.Sprintf("Mock%s", t.Method(i).Name)
		var genFn = v.MethodByName(genName)
		var s = stack.Func(t.Method(i).Func.Interface())
		data[t.Method(i).Name] = &handler{
			fn:    v.Method(i),
			stack: s,
			gen:   genFn,
		}
	}
	return data
}

type Test interface {
	Setup()
	Teardown()
}

type grpcTest struct {
	name    string
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
		if !h.gen.IsValid() || h.gen.IsNil() {
			continue
		}

		fmt.Println(name+":", h.stack)
		t.t.Run(name, func(tt *testing.T) {
			wfn := fx.WrapReflect(h.fn)
			for i := uint(0); i < t.fixture.RunNum; i++ {
				ppp := h.gen.Call(nil)
				var dt, err = json.Marshal(ppp[0].Interface())
				xerror.Panic(err)
				var nnn = string(dt)
				if cache[nnn] {
					continue
				}

				cache[nnn] = true

				tt.Run(nnn, func(tt *testing.T) {
					resp := wfn(ppp...)
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

					if !resp[0].IsNil() {
						dt, err := json.Marshal(resp[0].Interface())
						xerror.Panic(err)
						tt.Logf("resp: %s", dt)
					}
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

		var fix = &Fixture{T: t, assert: assert, RunNum: 100}
		var name = reflect.TypeOf(tt).Elem().Name()

		optsV := reflect.ValueOf(tt).Elem().FieldByName("Fixture")
		if !optsV.IsValid() {
			panic("has not [Fixture] field")
		}

		var gt = &grpcTest{name: name, fixture: fix, tt: tt, t: t, srv: serviceMethod(tt)}
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
