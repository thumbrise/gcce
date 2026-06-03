package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"strings"
)

var (
	ErrNotFunc           = errors.New("not a function")
	ErrAnonymousFunction = errors.New("anonymous function")
	ErrFunctionNotFound  = errors.New("function not found in source")
	ErrFileNotFound      = errors.New("cannot locate source file")
)

// ParamNames returns the declared parameter names of a function.
// Unnamed parameters are returned as empty strings.
// The first non-interface{} argument must be a function.
func ParamNames(fn interface{}) ([]string, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, ErrNotFunc
	}

	ptr := v.Pointer()

	fullName := runtime.FuncForPC(ptr).Name()
	if fullName == "" {
		return nil, ErrAnonymousFunction
	}
	// Compiler names anonymous functions like "pkg.Func.func1"
	if strings.Contains(fullName, ".func") {
		return nil, ErrAnonymousFunction
	}

	funcName := extractFuncName(fullName)

	file, _ := runtime.FuncForPC(ptr).FileLine(ptr)
	if file == "" {
		return nil, ErrFileNotFound
	}

	return extractParamNamesFromAST(file, funcName)
}

// extractFuncName extracts the simple function name from a fully qualified runtime name.
// e.g. "github.com/thumbrise/gcce/op/emit_test.NamedParams" -> "NamedParams"
// e.g. "github.com/thumbrise/gcce/op/emit_test.Type.Method" -> "Method"
func extractFuncName(fullName string) string {
	// Strip package path (everything up to and including the last '/')
	if idx := strings.LastIndex(fullName, "/"); idx >= 0 {
		fullName = fullName[idx+1:]
	}
	// Now we have "pkg.Func" or "pkg.Type.Method"
	// Take everything after the first dot.
	parts := strings.SplitN(fullName, ".", 2)
	if len(parts) < 2 {
		return fullName // shouldn't happen for a valid function
	}

	afterPkg := parts[1] // "Func" or "Type.Method"
	// If it's a method, the name is after the last dot.
	if idx := strings.LastIndex(afterPkg, "."); idx >= 0 {
		return afterPkg[idx+1:]
	}

	return afterPkg
}

// extractParamNamesFromAST parses the source file and returns the parameter names
// of the named function declaration.
func extractParamNamesFromAST(file string, funcName string) ([]string, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	var names []string

	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if !ok || fd.Name.Name != funcName {
			return true
		}

		found = true

		for _, field := range fd.Type.Params.List {
			for _, name := range field.Names {
				names = append(names, name.Name)
			}

			if len(field.Names) == 0 {
				names = append(names, "")
			}
		}

		return false
	})

	if !found {
		return nil, ErrFunctionNotFound
	}

	return names, nil
}
