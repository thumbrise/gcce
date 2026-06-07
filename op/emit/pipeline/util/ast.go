package util

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"strings"
)

var (
	ErrNotFunction  = errors.New("not a function")
	ErrNoFuncInfo   = errors.New("cannot get function info")
	ErrFuncNotFound = errors.New("function not found in AST")
)

// FuncDecl returns the *ast.FuncDecl for the given Go function.
func FuncDecl(fn interface{}) (*ast.FuncDecl, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, ErrNotFunction
	}

	pc := v.Pointer()

	rf := runtime.FuncForPC(pc)
	if rf == nil {
		return nil, ErrNoFuncInfo
	}

	file, line := rf.FileLine(pc)

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	fullName := rf.Name()
	// Extract the short name (last part after the final dot)
	shortName := fullName[strings.LastIndexByte(fullName, '.')+1:]

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fd.Name.Name != shortName {
			continue
		}
		// Check that the line falls within the function declaration
		startPos := fset.Position(fd.Pos())

		endPos := fset.Position(fd.End())
		if line >= startPos.Line && line <= endPos.Line {
			return fd, nil
		}
	}

	return nil, ErrFuncNotFound
}
