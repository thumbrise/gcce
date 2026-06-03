package analyze

import (
	"fmt"
	"reflect"

	"github.com/thumbrise/gcce/composition/errs"
)

type Implementation struct {
	Name    string
	Package string
}

func NewImplementation(interfac interface{}, constructor *Constructor) (*Implementation, error) {
	if interfac == nil {
		return nil, errs.ErrNilInterface
	}

	ptrType := reflect.TypeOf(interfac)
	if ptrType.Kind() != reflect.Pointer || ptrType.Elem().Kind() != reflect.Interface {
		return nil, fmt.Errorf("%s: %w", ptrType, errs.ErrNotInterfacePtr)
	}

	interfaceType := ptrType.Elem()

	targetType := constructor.TargetType()
	if !targetType.Implements(interfaceType) {
		return nil, fmt.Errorf("%s %w %s", targetType, errs.ErrNotImplement, interfaceType)
	}

	return &Implementation{
		Name:    interfaceType.Name(),
		Package: interfaceType.PkgPath(),
	}, nil
}
