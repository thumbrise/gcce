package pass

import (
	"go/ast"
	"reflect"

	"github.com/thumbrise/op-universal-schema-go/schema"
)

type Instruction struct {
	ID         string
	Comment    string
	Version    string
	Operations []Operation
	Trait      []Term
}
type Operation struct {
	ID          string
	Comment     string
	Input       []Term
	Output      []Term
	Error       []Term
	Trait       []Term
	ReflectType reflect.Type
	ASTFuncDecl *ast.FuncDecl
}
type Term struct {
	ID          string
	Comment     string
	Required    *bool
	Kind        *schema.Kind
	Value       any
	Of          []Term
	Trait       []Term
	ReflectType reflect.Type
}
