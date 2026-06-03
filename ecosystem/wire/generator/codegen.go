package generator

import (
	"fmt"
	"go/format"
	"io"
	"strings"

	"github.com/dave/jennifer/jen"
)

func generateCode(w io.Writer, factories []factory, cfg Config) error {
	f := jen.NewFile(cfg.PkgName)
	rootType := typeStmt(cfg.RootType)

	body := make([]jen.Code, 0, len(factories)*3+1)

	var rootVar string

	hasError := false

	for _, fac := range factories {
		if fac.root {
			rootVar = fac.varName
		}

		if fac.returnsErr {
			hasError = true
		}

		ctor := callStmt(fac.ctor)

		args := make([]jen.Code, len(fac.ctorArgs))
		for i, a := range fac.ctorArgs {
			args[i] = jen.Id(a)
		}

		if fac.returnsErr {
			returnNil := jen.Nil()
			if !strings.HasPrefix(cfg.RootType, "*") {
				returnNil = jen.Op("*").Add(jen.New(typeStmt(cfg.RootType)))
			}

			body = append(body,
				jen.List(jen.Id(fac.varName), jen.Err()).Op(":=").Add(ctor).Call(args...),
				jen.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.List(returnNil, jen.Err())),
				),
			)
		} else {
			body = append(body, jen.Id(fac.varName).Op(":=").Add(ctor).Call(args...))
		}
	}

	if hasError {
		body = append(body, jen.Return(jen.List(jen.Id(rootVar), jen.Nil())))
		f.Func().Id(cfg.FuncName).Params().Params(rootType, jen.Error()).Block(body...)
	} else {
		body = append(body, jen.Return(jen.Id(rootVar)))
		f.Func().Id(cfg.FuncName).Params().Params(rootType).Block(body...)
	}

	src, err := format.Source([]byte(f.GoString()))
	if err != nil {
		return fmt.Errorf("codegen: go fmt failed: %w\n\n%s", err, f.GoString())
	}

	_, err = w.Write(src)

	return err
}

func typeStmt(fqn string) *jen.Statement {
	ptr := strings.HasPrefix(fqn, "*")
	if ptr {
		fqn = fqn[1:]
	}

	lastDot := strings.LastIndex(fqn, ".")

	var stmt *jen.Statement
	if lastDot < 0 {
		stmt = jen.Id(fqn)
	} else {
		stmt = jen.Qual(fqn[:lastDot], fqn[lastDot+1:])
	}

	if ptr {
		stmt = jen.Op("*").Add(stmt)
	}

	return stmt
}

func callStmt(fqn string) *jen.Statement {
	lastDot := strings.LastIndex(fqn, ".")
	if lastDot < 0 {
		return jen.Id(fqn)
	}

	return jen.Qual(fqn[:lastDot], fqn[lastDot+1:])
}
