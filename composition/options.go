package composition

import (
	"fmt"

	"github.com/thumbrise/gcce/composition/data"
)

type Option func(*Graph) error

type BindOption func(binding *data.Binding) error

func Bind(constructor interface{}, options ...BindOption) Option {
	return func(g *Graph) error {
		binding := &data.Binding{
			Constructor: constructor,
		}
		for _, option := range options {
			if err := option(binding); err != nil {
				return fmt.Errorf("failed binding option: %w", err)
			}
		}

		g.bindings = append(g.bindings, binding)

		return nil
	}
}

func Implements(interfaces ...interface{}) BindOption {
	return func(binding *data.Binding) error {
		binding.Implements = interfaces

		return nil
	}
}
