package analyze

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	ErrNotFunction    = errors.New("constructor is not function")
	ErrNoFuncForPC    = errors.New("cannot get runtime func for pc")
	ErrAnonymousFunc  = errors.New("constructor is anonymous")
	ErrNoReturnValues = errors.New("constructor must return at least one value")
)

type Constructor struct {
	rflValue reflect.Value
	Name     string
	Package  string

	TargetName    string
	TargetPackage string
}

func NewConstructor(constructor interface{}) (*Constructor, error) {
	ctor := &Constructor{}
	rflValue := reflect.ValueOf(constructor)

	if rflValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("%w: %T\n%s", ErrNotFunction, constructor, debug.Stack())
	}

	funcForPC := runtime.FuncForPC(rflValue.Pointer())
	if funcForPC == nil {
		return nil, fmt.Errorf("%w: %T", ErrNoFuncForPC, constructor)
	}

	fullFuncName := funcForPC.Name()

	if ctor.isAnonymousFunc(fullFuncName) {
		return nil, fmt.Errorf("%w: %s\n%s", ErrAnonymousFunc, fullFuncName, debug.Stack())
	}

	pkgPath, funcName := ctor.splitFuncPath(fullFuncName)

	typ := rflValue.Type()
	if typ.NumOut() == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoReturnValues, fullFuncName)
	}

	targetType := typ.Out(0)

	actualTargetType := targetType
	if targetType.Kind() == reflect.Pointer {
		actualTargetType = targetType.Elem()
	}

	ctor.rflValue = rflValue
	ctor.Package = pkgPath
	ctor.Name = funcName
	ctor.TargetPackage = actualTargetType.PkgPath()
	ctor.TargetName = actualTargetType.Name()

	return ctor, nil
}

func (ctr *Constructor) Type() reflect.Type {
	return ctr.rflValue.Type()
}

func (ctr *Constructor) TargetType() reflect.Type {
	return ctr.rflValue.Type().Out(0)
}

func (ctr *Constructor) Target() string {
	if ctr.Type().NumOut() == 0 {
		return ""
	}
	// Пропускаем через унификатор FQN
	return ctr.fqnOf(ctr.Type().Out(0))
}

func (ctr *Constructor) IsVariadic() bool {
	return ctr.Type().IsVariadic()
}

func (ctr *Constructor) Dependencies() []string {
	typ := ctr.Type()

	numIn := typ.NumIn()
	if numIn == 0 {
		return make([]string, 0)
	}

	deps := make([]string, 0, numIn)
	for i := range numIn {
		// Каждый входящий аргумент пропускаем через унификатор FQN
		deps = append(deps, ctr.fqnOf(typ.In(i)))
	}

	return deps
}

func (ctr *Constructor) splitFuncPath(fullName string) (string, string) {
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot < 0 {
		return "", fullName
	}

	return fullName[:lastDot], fullName[lastDot+1:]
}

func (ctr *Constructor) isAnonymousFunc(fullName string) bool {
	// Анонимные функции в рантайме Go всегда содержат ".func" (например, pkg.NewKernel.func1)
	if strings.Contains(fullName, ".func") {
		return true
	}

	if strings.HasSuffix(fullName, ".init") {
		return true
	}

	return false
}

func (ctr *Constructor) ReturnsError() bool {
	typ := ctr.Type()

	numOut := typ.NumOut()
	if numOut <= 1 {
		return false
	}

	lastOut := typ.Out(numOut - 1)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	return lastOut == errorInterface
}

func (ctr *Constructor) fqnOf(t reflect.Type) string {
	if t == nil {
		return ""
	}

	if t.Kind() == reflect.Pointer {
		return "*" + ctr.fqnOf(t.Elem())
	}

	if t.Kind() == reflect.Slice {
		return "[]" + ctr.fqnOf(t.Elem())
	}

	if t.PkgPath() == "" {
		return t.Name()
	}

	return t.PkgPath() + "." + t.Name()
}
