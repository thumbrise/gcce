package composition

import (
	"errors"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

var ErrCyclicDependency = errors.New("cyclic or unresolvable dependency detected in remaining steps")

func SortOperations(raw []op.Operation) ([]op.Operation, error) {
	pending := make([]op.Operation, len(raw))
	copy(pending, raw)

	ordered := make([]op.Operation, 0, len(raw))
	readyTypes := make(map[string]bool)

	for len(pending) > 0 {
		progress := false

		for i := 0; i < len(pending); i++ {
			if !canResolve(pending[i], readyTypes) {
				continue
			}

			ordered = append(ordered, pending[i])
			markReady(pending[i], readyTypes)

			pending = append(pending[:i], pending[i+1:]...)
			i--
			progress = true
		}

		if !progress {
			return nil, ErrCyclicDependency
		}
	}

	for i := range ordered {
		ordered[i].Trait = append(ordered[i].Trait, trait.NewOrder(i))
	}

	return ordered, nil
}

func canResolve(operation op.Operation, ready map[string]bool) bool {
	for _, in := range operation.Input {
		if !ready[in.ID] {
			return false
		}
	}

	return true
}

func markReady(operation op.Operation, ready map[string]bool) {
	if len(operation.Output) == 0 {
		return
	}

	ready[operation.Output[0].ID] = true

	for _, t := range operation.Output[0].Trait {
		if t.ID == trait.ImplementsID {
			if val, ok := t.Value.(string); ok {
				ready[val] = true
			}
		}
	}
}
