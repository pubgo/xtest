package xtest

import (
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"

	"reflect"
)

type xtest struct {
	fn     interface{}
	params [][]interface{}
}

func (t *xtest) In(args ...interface{}) {
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
}

func (t *xtest) Do() {
	defer xerror.RespExit()

	wfn := fx.WrapRaw(t.fn)
	for _, param := range t.params {
		_ = wfn(param...)
	}
}

func TestFunc(fn interface{}) *xtest {
	if fn == nil {
		xerror.Exit(ErrParamIsNil)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		xerror.PanicF(ErrParamTypeNotFunc, "kind: %s, name: %s", reflect.TypeOf(fn).Kind(), reflect.TypeOf(fn))
	}

	return &xtest{fn: fn}
}
