package wire

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/thumbrise/gcce/composition"
)

func resolveArg(argType composition.CompileType, typeMap map[string][]int, vars map[int]string, currentIdx int) (jen.Code, error) {
	typeStr := argType.Type

	if indices, ok := typeMap[typeStr]; ok {
		match := filterBefore(indices, currentIdx)

		matched := varsOf(match, vars)
		if len(matched) == 1 {
			return jen.Id(matched[0]), nil
		}
	}

	if isSliceType(typeStr) {
		return resolveCollection(argType, typeMap, vars, currentIdx)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnresolvable, typeStr)
}

func resolveCollection(argType composition.CompileType, typeMap map[string][]int, vars map[int]string, currentIdx int) (jen.Code, error) {
	et := elemType(argType.Type)

	indices, ok := typeMap[et]
	if !ok {
		return nil, fmt.Errorf("%w: no providers for %s", ErrUnresolvable, et)
	}

	match := filterBefore(indices, currentIdx)

	matched := varsOf(match, vars)
	if len(matched) == 0 {
		return nil, fmt.Errorf("%w: no providers for %s", ErrUnresolvable, et)
	}

	if len(matched) == 1 {
		return jen.Id(matched[0]), nil
	}

	items := make([]jen.Code, len(matched))
	for i, v := range matched {
		items[i] = jen.Id(v)
	}

	return typeExpr(argType).Values(items...), nil
}

func filterBefore(indices []int, currentIdx int) []int {
	out := make([]int, 0, len(indices))
	for _, idx := range indices {
		if idx < currentIdx {
			out = append(out, idx)
		}
	}

	return out
}

func varsOf(indices []int, vars map[int]string) []string {
	out := make([]string, 0, len(indices))
	for _, idx := range indices {
		if v, ok := vars[idx]; ok {
			out = append(out, v)
		}
	}

	return out
}

func isSliceType(t string) bool {
	return strings.HasPrefix(t, "[]")
}

func elemType(t string) string {
	return t[2:]
}
