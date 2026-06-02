package composition

import "fmt"

const (
	temporary = 1
	permanent = 2
)

func visit(s schema, node *node, marks map[*node]int) error {
	if marks[node] == permanent {
		return nil
	}

	if marks[node] == temporary {
		return ErrCycleDetected
	}

	marks[node] = temporary

	params, err := node.deps(s)
	if err != nil {
		return fmt.Errorf("%s: %w", node, err)
	}

	for _, param := range params {
		if err := visit(s, param, marks); err != nil {
			return err
		}
	}

	marks[node] = permanent

	return nil
}
