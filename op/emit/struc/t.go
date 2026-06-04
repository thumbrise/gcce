package struc

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/thumbrise/gcce/op/emit"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

var (
	ErrNil       = errors.New("nil pointer passed to T")
	ErrNotStruct = errors.New("not a struct")
)

// T returns operations for all exported methods of the given struct.
// Each operation is annotated with a group trait equal to the struct name.
func T(v interface{}) ([]schema.Operation, error) {
	if v == nil {
		return nil, ErrNil
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	// Ensure we always work with a pointer to the struct to see all methods.
	if typ.Kind() != reflect.Pointer {
		typ = reflect.PointerTo(typ)
	}

	elem := typ.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w, got %s", ErrNotStruct, elem.Kind())
	}

	groupName := elem.Name()

	var ops []schema.Operation

	for i := range typ.NumMethod() {
		m := typ.Method(i)
		if isMethodUnexported(m) {
			continue
		}

		fn := m.Func.Interface()

		op, err := emit.FunctionToOperation(fn)
		if err != nil {
			return nil, fmt.Errorf("struc.T: method %s: %w", m.Name, err)
		}

		op.Trait = append(op.Trait, trait.NewGroup(groupName))
		ops = append(ops, op)
	}

	return ops, nil
}

func isMethodUnexported(m reflect.Method) bool {
	// PkgPath is the package path that qualifies a lower case (unexported) method name.
	// It is empty for upper case (exported) method names.
	return m.PkgPath != ""
}
