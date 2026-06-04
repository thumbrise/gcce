package emit

import (
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
	ErrIsNotFunction = errors.New("is not a function")
	ErrIsAnonymous   = errors.New("is anonymous")
)

// FunctionToOperation
//

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

		term := reflTypToTerm(inTyp)

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

		term := reflTypToTerm(outTyp)

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

func reflTypToTerm(typ reflect.Type) schema.Term {
	typElem := typ
	fqn := typeToFQN(typ)
	required := true

	if typ.Kind() == reflect.Pointer {
		typElem = typ.Elem()
		required = false
	}

	result := schema.Term{}

	result.Required = &required
	result.Kind = kindFromType(typElem)
	result.Of = []schema.Term{}
	result.Trait = []schema.Term{
		trait.NewFQN(fqn),
	}

	return result
}

// kindFromType maps a Go reflect.Type to an OP schema.Kind.
// It handles special types like time.Time and byte slices before falling back to reflect.Kind.
//
//nolint:cyclop // No other ways
func kindFromType(t reflect.Type) *schema.Kind {
	// Special case: time.Time -> datetime
	if t == reflect.TypeOf(time.Time{}) {
		return new(schema.KindDatetime)
	}

	// Special case: []byte and [N]byte -> binary
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return new(schema.KindBinary)
	}

	if t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8 {
		return new(schema.KindBinary)
	}

	switch t.Kind() {
	case reflect.String:
		return new(schema.KindString)
	case reflect.Bool:
		return new(schema.KindBoolean)
	case reflect.Float32, reflect.Float64:
		return new(schema.KindFloat)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(schema.KindInteger)
	case reflect.Array, reflect.Slice:
		// If we didn't catch []byte above, it's a generic array
		return new(schema.KindArray)
	case reflect.Struct:
		return new(schema.KindObject)
	case reflect.Map:
		// Map is represented as object in OP
		return new(schema.KindObject)
	case reflect.Invalid,
		reflect.Chan, reflect.Func, reflect.UnsafePointer,
		reflect.Complex64, reflect.Complex128,
		reflect.Interface,
		reflect.Pointer, reflect.Uintptr:
		return nil
	default:
		return nil
	}
}

// typeToFQN returns the fully qualified name of a reflect.Type.
// It handles named types (including basic types like "int", "string"),
// pointers, slices, arrays, maps.
func typeToFQN(typ reflect.Type) string {
	if name := typ.Name(); name != "" {
		if pkg := typ.PkgPath(); pkg != "" {
			return pkg + "." + name
		}

		return name
	}

	switch typ.Kind() {
	case reflect.Pointer:
		return "*" + typeToFQN(typ.Elem())
	case reflect.Slice:
		return "[]" + typeToFQN(typ.Elem())
	case reflect.Array:
		return "[" + strconv.Itoa(typ.Len()) + "]" + typeToFQN(typ.Elem())
	case reflect.Map:
		return "map[" + typeToFQN(typ.Key()) + "]" + typeToFQN(typ.Elem())
	default:
		return typ.String()
	}
}
