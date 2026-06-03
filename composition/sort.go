package composition

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

var ErrCyclicDependency = errors.New("cyclic or unresolvable dependency detected")

type CycleError struct {
	Pending []op.Operation
	Ready   map[string]bool
}

func (e *CycleError) Error() string {
	unresolvable := e.unresolvableTypes()

	var buf strings.Builder
	buf.WriteString("cyclic or unresolvable dependency detected")

	if len(unresolvable) > 0 {
		e.formatTree(&buf, unresolvable)
	} else if len(e.Pending) > 0 {
		e.formatCycle(&buf)
	}

	e.formatReady(&buf)
	e.formatCTA(&buf, len(unresolvable) > 0)

	return buf.String()
}

func (e *CycleError) producedTypes() map[string]bool {
	produced := make(map[string]bool)

	for _, op := range e.Pending {
		if len(op.Output) == 0 {
			continue
		}

		produced[op.Output[0].ID] = true

		for _, t := range op.Output[0].Trait {
			if t.ID == trait.ImplementsID {
				if val, ok := t.Value.(string); ok {
					produced[val] = true
				}
			}
		}
	}

	return produced
}

func (e *CycleError) depSet() map[string]bool {
	depSet := make(map[string]bool)

	for _, op := range e.Pending {
		for _, in := range op.Input {
			if in.ID != "" {
				depSet[in.ID] = true
			}
		}
	}

	return depSet
}

func (e *CycleError) unresolvableTypes() []string {
	produced := e.producedTypes()
	depSet := e.depSet()

	unresolvable := make([]string, 0)

	for dep := range depSet {
		if !e.Ready[dep] && !produced[dep] {
			unresolvable = append(unresolvable, dep)
		}
	}

	sort.Strings(unresolvable)

	return unresolvable
}

func (e *CycleError) formatTree(buf *strings.Builder, unresolvable []string) {
	blockedBy := make(map[string][]string)

	for _, dep := range unresolvable {
		blocked := make([]string, 0)

		for _, op := range e.Pending {
			for _, in := range op.Input {
				if in.ID == dep {
					blocked = append(blocked, op.ID)

					break
				}
			}
		}

		sort.Strings(blocked)
		blockedBy[dep] = blocked
	}

	buf.WriteString("\n\nunresolvable (types never produced):")

	for _, dep := range unresolvable {
		blocked := blockedBy[dep]

		buf.WriteString("\n  ")
		buf.WriteString(dep)
		fmt.Fprintf(buf, "\n    └── %d ops blocked:", len(blocked))

		for _, opID := range blocked {
			buf.WriteString("\n        - ")
			buf.WriteString(opID)
		}
	}
}

func (e *CycleError) formatCycle(buf *strings.Builder) {
	buf.WriteString("\n\ncycle detected:")
	buf.WriteString("\n  pending:")

	for _, op := range e.Pending {
		buf.WriteString("\n    - ")
		buf.WriteString(op.ID)

		missing := make([]string, 0, len(op.Input))

		for _, in := range op.Input {
			if !e.Ready[in.ID] {
				missing = append(missing, in.ID)
			}
		}

		if len(missing) > 0 {
			fmt.Fprintf(buf, " — waiting for %v", missing)
		}
	}
}

func (e *CycleError) formatReady(buf *strings.Builder) {
	readyList := make([]string, 0, len(e.Ready))

	for t := range e.Ready {
		readyList = append(readyList, t)
	}

	sort.Strings(readyList)

	if len(readyList) > 0 {
		buf.WriteString("\n\nready types (available):")

		for _, t := range readyList {
			buf.WriteString("\n  - ")
			buf.WriteString(t)
		}
	}
}

func (e *CycleError) formatCTA(buf *strings.Builder, hasMissing bool) {
	if hasMissing {
		buf.WriteString("\n\n💡 tip: Bind a constructor that produces each missing type above:\n")
		buf.WriteString("    composition.Bind(YourConstructor)")
	} else {
		buf.WriteString("\n\n💡 tip: Break the cycle by extracting an interface or\n")
		buf.WriteString("    removing a direct dependency between the operations above.")
	}
}

func (e *CycleError) Unwrap() error {
	return ErrCyclicDependency
}

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
			return nil, &CycleError{
				Pending: pending,
				Ready:   readyTypes,
			}
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
