package xtest

import (
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
	wfn := Wrap(t.fn)
	for _, param := range t.params {
		xerror.Exit(wfn(param...)())
	}
	return
}

func TestFuncWith(fn interface{}) *xtest {
	if fn == nil {
		xerror.Panic(ErrParamIsNil)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		xerror.PanicF(ErrParamTypeNotFunc, "kind: %s, name: %s", reflect.TypeOf(fn).Kind(), reflect.TypeOf(fn))
	}

	return &xtest{fn: fn}
}
