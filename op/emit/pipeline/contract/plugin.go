package contract

import "github.com/thumbrise/gcce/op/emit/pipeline/pass"

// OperationPlugin defines a plugin that visits an operation pass to enrich or
// transform the operation metadata before the instruction is assembled.
type OperationPlugin interface {
	Name() string
	Visit(pass pass.OperationPass) error
}

// InstructionPlugin defines a plugin that visits the instruction pass to enrich
// or transform instruction-level metadata after all operations are registered.
type InstructionPlugin interface {
	Name() string
	Visit(pass pass.InstructionPass) error
}

// OperationRegistration binds a Go function with its associated plugins.
// The function is introspected via emit.FunctionToOperation to derive the initial operation schema.
type OperationRegistration struct {
	FN      interface{}
	Plugins []OperationPlugin
}

// InstructionRegistration holds the static metadata for an instruction and its
// associated plugins that further enrich the instruction schema.
type InstructionRegistration struct {
	ID      string
	Version string
	Comment string
	Plugins []InstructionPlugin
}
