package generator

import (
	"io"

	op "github.com/thumbrise/op-universal-schema-go/schema"
)

type Config struct {
	PkgName  string
	FuncName string
	RootType string
}

func Generate(w io.Writer, operations []op.Operation, cfg Config) error {
	factories, err := resolve(operations, cfg.RootType)
	if err != nil {
		return err
	}

	return generateCode(w, factories, cfg)
}
