package composition

import (
	"fmt"
	"reflect"
)

type schema interface {
	find(t reflect.Type) (*node, error)
}

type defaultSchema struct {
	parents []*defaultSchema
	nodes   map[reflect.Type][]*node
}

func newDefaultSchema() *defaultSchema {
	return &defaultSchema{
		nodes: map[reflect.Type][]*node{},
	}
}

func (s *defaultSchema) register(n *node) {
	s.nodes[n.rt] = append(s.nodes[n.rt], n)
}

func (s *defaultSchema) prepare(n *node) error {
	marks := map[*node]int{}

	return visit(s, n, marks)
}

func (s *defaultSchema) find(t reflect.Type) (*node, error) {
	nodes, ok := s.list(t)
	if ok {
		if len(nodes) > 1 {
			return nil, fmt.Errorf("%s: %w, maybe you need to use group: []%s", t, ErrMultipleDefinitions, t)
		}

		return nodes[0], nil
	}

	if t.Kind() == reflect.Slice {
		return s.group(t)
	}

	return nil, fmt.Errorf("%s %w", t, ErrTypeNotExists)
}

func (s *defaultSchema) group(t reflect.Type) (*node, error) {
	elems, ok := s.list(t.Elem())
	if !ok {
		return nil, fmt.Errorf("%s %w", t, ErrTypeNotExists)
	}

	return &node{
		compiler: newGroupCompiler(t, elems),
		rt:       t,
	}, nil
}

func (s *defaultSchema) list(t reflect.Type) (nodes []*node, ok bool) {
	for _, parent := range s.parents {
		if n, o := parent.list(t); o {
			nodes = append(nodes, n...)
			ok = true
		}
	}

	if n, o := s.nodes[t]; o {
		nodes = append(nodes, n...)
		ok = true
	}

	return nodes, ok
}
