package contract

import "github.com/thumbrise/gcce/op/emit/pipeline/pass"

type OperationPlugin interface {
	Name() string
	Visit(pass pass.OperationPass) error
}
type InstructionPlugin interface {
	Name() string
	Visit(pass pass.InstructionPass) error
}
type OperationRegistration struct {
	FN      interface{}
	Plugins []OperationPlugin
}
type InstructionRegistration struct {
	ID      string
	Version string
	Comment string
	Plugins []InstructionPlugin
}
