package composition

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type function struct {
	Name string
	reflect.Type
	reflect.Value
}

var errorInterface = reflect.TypeOf(new(error)).Elem()

func isError(typ reflect.Type) bool {
	return typ.Implements(errorInterface)
}

func isCleanup(typ reflect.Type) bool {
	return typ.Kind() == reflect.Func && typ.NumIn() == 0 && typ.NumOut() == 0
}

func inspectFunction(fn interface{}) (function, bool) {
	if reflect.ValueOf(fn).Kind() != reflect.Func {
		return function{}, false
	}
	val := reflect.ValueOf(fn)
	typ := val.Type()
	funcForPC := runtime.FuncForPC(val.Pointer())
	return function{
		Name:  funcForPC.Name(),
		Type:  typ,
		Value: val,
	}, true
}

func isAnonymous(fn function) bool {
	name := fn.Name
	if name == "" {
		return true
	}
	lastDot := strings.LastIndex(name, ".")
	if lastDot < 0 {
		return true
	}
	lastSeg := name[lastDot+1:]
	if len(lastSeg) > 4 && lastSeg[:4] == "func" && lastSeg[4] >= '0' && lastSeg[4] <= '9' {
		return true
	}
	return false
}

type link struct {
	Name string
	Type reflect.Type
}

func inspectInterfacePointer(i interface{}) (*link, error) {
	if i == nil {
		return nil, fmt.Errorf("nil: not a pointer to interface")
	}
	typ := reflect.TypeOf(i)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Interface {
		return nil, fmt.Errorf("%s: not a pointer to interface", typ)
	}
	return &link{
		Name: typ.Elem().Name(),
		Type: typ.Elem(),
	}, nil
}
