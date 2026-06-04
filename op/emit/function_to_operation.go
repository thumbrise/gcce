package emit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

var (
	ErrIsNotFunction   = errors.New("is not a function")
	ErrIsAnonymous     = errors.New("is anonymous")
	ErrUnsupportedKind = errors.New("unsupported kind")
)

// FunctionToOperation
//
//nolint:funlen // Orchestration
func FunctionToOperation(fn interface{}) (schema.Operation, error) {
	operation := schema.Operation{}

	fnReflVal := reflect.ValueOf(fn)
	if fnReflVal.Kind() != reflect.Func {
		return operation, ErrIsNotFunction
	}

	fnName := runtime.FuncForPC(fnReflVal.Pointer()).Name()
	if fnName == "" || strings.Contains(fnName, ".func") {
		return operation, ErrIsAnonymous
	}

	operation.ID = fnName

	fnReflTyp := fnReflVal.Type()

	inputRail := make([]schema.Term, 0, fnReflTyp.NumIn())

	inpI := 0

	for i := range fnReflTyp.NumIn() {
		inTyp := fnReflTyp.In(i)

		term, err := reflTypToTerm(inTyp)
		if err != nil {
			if errors.Is(err, errContextOccurred) {
				continue
			}

			return operation, err
		}

		term.ID = fmt.Sprintf("input%d", inpI)
		inputRail = append(inputRail, term)
		inpI++
	}

	outputRail := make([]schema.Term, 0, fnReflTyp.NumOut())
	errorRail := make([]schema.Term, 0, fnReflTyp.NumOut())

	errorTypElem := reflect.TypeOf(new(error)).Elem()

	errI := 0
	outI := 0

	for i := range fnReflTyp.NumOut() {
		outTyp := fnReflTyp.Out(i)

		term, err := reflTypToTerm(outTyp)
		if err != nil {
			return operation, err
		}

		if outTyp.Implements(errorTypElem) {
			term.ID = fmt.Sprintf("error%d", errI)
			errorRail = append(errorRail, term)
			errI++
		} else {
			term.ID = fmt.Sprintf("output%d", outI)
			outputRail = append(outputRail, term)
			outI++
		}
	}

	operation.Input = inputRail
	operation.Output = outputRail
	operation.Error = errorRail
	operation.Trait = []schema.Term{}

	return operation, nil
}

func reflTypToTerm(typ reflect.Type) (schema.Term, error) {
	typElem := typ
	fqn := typeToFQN(typ)
	required := true

	if typ.Kind() == reflect.Pointer {
		typElem = typ.Elem()
		required = false
	}

	result := schema.Term{}

	kind, err := kindFromType(typElem)
	if err != nil {
		return result, err
	}

	result.Required = &required
	result.Kind = &kind
	result.Of = []schema.Term{}
	result.Trait = []schema.Term{
		trait.NewFQN(fqn),
	}

	return result, nil
}

var errContextOccurred = errors.New("context is occurred")

// kindFromType maps a Go reflect.Type to an OP schema.Kind.
// It handles special types like time.Time and byte slices before falling back to reflect.Kind.
//
//nolint:cyclop // No other ways
func kindFromType(t reflect.Type) (schema.Kind, error) {
	// Special case: time.Time -> datetime
	if t == reflect.TypeOf(time.Time{}) {
		return schema.KindDatetime, nil
	}

	// Special case: []byte and [N]byte -> binary
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return schema.KindBinary, nil
	}

	if t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8 {
		return schema.KindBinary, nil
	}

	switch t.Kind() {
	case reflect.String:
		return schema.KindString, nil
	case reflect.Bool:
		return schema.KindBoolean, nil
	case reflect.Float32, reflect.Float64:
		return schema.KindFloat, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return schema.KindInteger, nil
	case reflect.Array, reflect.Slice:
		// If we didn't catch []byte above, it's a generic array
		return schema.KindArray, nil
	case reflect.Struct:
		return schema.KindObject, nil
	case reflect.Map:
		// Map is represented as object in OP
		return schema.KindObject, nil
	case reflect.Interface:
		errTyp := reflect.TypeOf(new(error)).Elem()
		if t.Implements(errTyp) {
			return schema.KindObject, nil
		}

		ctxTyp := reflect.TypeOf(new(context.Context)).Elem()
		if t.Implements(ctxTyp) {
			return "", errContextOccurred
		}

		return "", fmt.Errorf("%w: only error interface supported, but given %v", ErrUnsupportedKind, t)
	case reflect.Complex64, reflect.Complex128:
		return "", fmt.Errorf("%w: complex numbers are not supported (type %v)", ErrUnsupportedKind, t)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return "", fmt.Errorf("%w: unserializable type %v", ErrUnsupportedKind, t.Kind())
	case reflect.Invalid:
		return "", fmt.Errorf("%w: literally invalid reflect kind", ErrUnsupportedKind)
	case reflect.Pointer, reflect.Uintptr:
		return "", fmt.Errorf("%w: unexpected pointer (type %v)", ErrUnsupportedKind, t)

	default:
		return "", fmt.Errorf("%w: %v", ErrUnsupportedKind, t.Kind())
	}
}

// typeToFQN returns the fully qualified name of a reflect.Type.
// It handles named types (including basic types like "int", "string"),
// pointers, slices, arrays, maps, and the special error interface.
func typeToFQN(typ reflect.Type) string {
	// Named type (including basic types with names like "int", "string")
	if name := typ.Name(); name != "" {
		if pkg := typ.PkgPath(); pkg != "" {
			return pkg + "." + name
		}

		return name
	}
	// Unnamed composite types
	//nolint:exhaustive // Defaulting
	switch typ.Kind() {
	case reflect.Pointer:
		return "*" + typeToFQN(typ.Elem())
	case reflect.Slice:
		return "[]" + typeToFQN(typ.Elem())
	case reflect.Array:
		return "[" + strconv.Itoa(typ.Len()) + "]" + typeToFQN(typ.Elem())
	case reflect.Map:
		return "map[" + typeToFQN(typ.Key()) + "]" + typeToFQN(typ.Elem())
	case reflect.Interface:
		// Special case for the built-in error interface
		if typ == reflect.TypeOf((*error)(nil)).Elem() {
			return "error"
		}

		return typ.String() // fallback, e.g., "interface {}"
	default:
		return typ.String()
	}
}
