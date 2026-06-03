package composition

import (
	"fmt"

	"github.com/thumbrise/gcce/composition/data"
	"github.com/thumbrise/gcce/composition/internal/analyze"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

type Graph struct {
	bindings []*data.Binding
}

func New(options ...Option) (*Graph, error) {
	g := &Graph{}
	for _, option := range options {
		err := option(g)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return g, nil
}

func (g *Graph) Resolve() ([]op.Operation, error) {
	rawOperations := make([]op.Operation, 0, len(g.bindings))
	for _, binding := range g.bindings {
		step, err := analyze.BindingToOperation(binding)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve step: %w", err)
		}

		rawOperations = append(rawOperations, step)
	}

	return SortOperations(rawOperations)
}
