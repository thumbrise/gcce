package analyze

import (
	"fmt"

	"github.com/thumbrise/gcce/composition/data"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

func BindingToOperation(binding *data.Binding) (op.Operation, error) {
	result := op.Operation{
		ID:      "",
		Comment: "",
		Input:   make([]op.Term, 0),
		Output:  make([]op.Term, 0),
		Error:   make([]op.Term, 0),
		Trait:   make([]op.Term, 0),
	}

	ctor, err := NewConstructor(binding.Constructor)
	if err != nil {
		return result, fmt.Errorf("failed creating constructor: %w", err)
	}

	constructorFQN := ctor.Package + "." + ctor.Name
	targetFQN := ctor.Target()

	result.ID = constructorFQN
	result.Trait = append(result.Trait, trait.NewFQN(constructorFQN))
	result.Output = append(result.Output, op.Term{
		ID:    targetFQN,
		Trait: []op.Term{trait.NewFQN(targetFQN)},
	})

	deps := ctor.Dependencies()
	isVariadic := ctor.IsVariadic()

	for i, inputFQN := range deps {
		term := op.Term{
			ID:    inputFQN,
			Trait: []op.Term{trait.NewFQN(inputFQN)},
		}

		if isVariadic && i == len(deps)-1 {
			term.Trait = append(term.Trait, trait.NewVariadic())
		}

		result.Input = append(result.Input, term)
	}

	if ctor.ReturnsError() {
		result.Error = append(result.Error, op.Term{
			ID:    "error",
			Trait: []op.Term{trait.NewFQN("error")},
		})
	}

	for _, implementationRaw := range binding.Implements {
		implementation, err := NewImplementation(implementationRaw, ctor)
		if err != nil {
			return result, fmt.Errorf("failed reflect implementation: %w", err)
		}

		implementationFQN := implementation.Package + "." + implementation.Name
		result.Output[0].Trait = append(result.Output[0].Trait, trait.NewImplements(implementationFQN))
	}

	return result, nil
}
