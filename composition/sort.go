package composition

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

var ErrCyclicDependency = errors.New("cyclic dependency detected")

type CycleError struct {
	Pending []op.Operation
	Ready   map[string]bool
}

func (e *CycleError) Error() string {
	var buf strings.Builder
	buf.WriteString("cyclic dependency detected")

	if len(e.Pending) > 0 {
		e.formatCycle(&buf)
	}

	e.formatReady(&buf)

	buf.WriteString("\n\n💡 tip: Break the cycle by extracting an interface or\n")
	buf.WriteString("    removing a direct dependency between the operations above.")

	return buf.String()
}

func (e *CycleError) formatCycle(buf *strings.Builder) {
	buf.WriteString("\n  pending:")

	for _, pendingOp := range e.Pending {
		buf.WriteString("\n    - ")
		buf.WriteString(pendingOp.ID)

		missing := make([]string, 0, len(pendingOp.Input))

		for _, in := range pendingOp.Input {
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
			unknowns := findUnknowns(pending, readyTypes)

			if len(unknowns) > 0 {
				for _, u := range unknowns {
					readyTypes[u] = true
				}

				continue
			}

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

func findUnknowns(pending []op.Operation, ready map[string]bool) []string {
	produced := pendingProduced(pending)

	unknowns := make([]string, 0)
	seen := make(map[string]bool)

	for _, op := range pending {
		for _, in := range op.Input {
			if in.ID == "" || seen[in.ID] {
				continue
			}

			seen[in.ID] = true

			if !ready[in.ID] && !produced[in.ID] {
				unknowns = append(unknowns, in.ID)
			}
		}
	}

	sort.Strings(unknowns)

	return unknowns
}

func pendingProduced(pending []op.Operation) map[string]bool {
	produced := make(map[string]bool)

	for _, op := range pending {
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
