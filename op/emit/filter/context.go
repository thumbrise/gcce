package filter

import (
	"fmt"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

// Context removes context.Context parameters from all operation inputs and renumbers the remaining inputs.
// The input slice is modified in place and returned as a convenience.
func Context(operations []schema.Operation) []schema.Operation {
	for i := range operations {
		op := &operations[i]

		filtered := make([]schema.Term, 0, len(op.Input))
		for _, term := range op.Input {
			if isContext(term) {
				continue
			}

			filtered = append(filtered, term)
		}

		op.Input = filtered
		renumberInputs(op.Input)
	}

	return operations
}

const contextTraitValue = "context.Context"

func isContext(term schema.Term) bool {
	for _, t := range term.Trait {
		if t.ID == trait.FQNID && t.Value == contextTraitValue {
			return true
		}
	}

	return false
}

func renumberInputs(inputs []schema.Term) {
	for i := range inputs {
		inputs[i].ID = fmt.Sprintf("input%d", i)
	}
}
