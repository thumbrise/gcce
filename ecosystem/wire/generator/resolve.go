package generator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

var (
	ErrRootNotFound      = errors.New("root type not found")
	ErrUnresolvable      = errors.New("unresolvable dependency")
	ErrMultipleProviders = errors.New("multiple providers for type (slice injection not yet supported)")
)

type factory struct {
	varName    string
	ctor       string
	target     string
	implements []string
	deps       []string
	ctorArgs   []string
	returnsErr bool
	root       bool
}

func resolve(operations []op.Operation, rootFQN string) ([]factory, error) {
	factories := extract(operations)
	if len(factories) == 0 {
		return nil, fmt.Errorf("%w: no operations provided", ErrRootNotFound)
	}

	index := buildIndex(factories)

	rootIdx := findRoot(rootFQN, factories)
	if rootIdx < 0 {
		return nil, fmt.Errorf("%w: %q", ErrRootNotFound, rootFQN)
	}

	visited := make(map[int]bool)
	if err := visitDeps(rootIdx, factories, index, visited); err != nil {
		return nil, err
	}

	aliases := collectAliases(factories)
	assignNames(factories, visited, aliases)

	reachable := make([]factory, 0, len(visited))
	for i, f := range factories {
		if !visited[i] {
			continue
		}

		args := make([]string, len(f.deps))
		for j, dep := range f.deps {
			providers := index[dep]
			if len(providers) == 0 {
				return nil, fmt.Errorf("%w: %q for step %q", ErrUnresolvable, dep, f.ctor)
			}

			if len(providers) > 1 {
				return nil, fmt.Errorf("%w: %q has %d providers (step %q)", ErrMultipleProviders, dep, len(providers), f.ctor)
			}

			args[j] = factories[providers[0]].varName
		}

		f.ctorArgs = args
		f.root = (i == rootIdx)
		reachable = append(reachable, f)
	}

	return reachable, nil
}

func extract(operations []op.Operation) []factory {
	factories := make([]factory, 0, len(operations))
	for _, operation := range operations {
		if len(operation.Output) == 0 {
			continue
		}

		target := operation.Output[0].ID

		deps := make([]string, 0, len(operation.Input))
		for _, in := range operation.Input {
			deps = append(deps, in.ID)
		}

		returnsErr := len(operation.Error) > 0

		implements := make([]string, 0)

		for _, t := range operation.Output[0].Trait {
			if t.ID == trait.ImplementsID {
				if val, ok := t.Value.(string); ok {
					implements = append(implements, val)
				}
			}
		}

		factories = append(factories, factory{
			ctor:       operation.ID,
			target:     target,
			implements: implements,
			deps:       deps,
			returnsErr: returnsErr,
		})
	}

	return factories
}

func buildIndex(factories []factory) map[string][]int {
	index := map[string][]int{}
	for i, f := range factories {
		index[f.target] = append(index[f.target], i)
		for _, impl := range f.implements {
			index[impl] = append(index[impl], i)
		}
	}

	return index
}

func findRoot(rootFQN string, factories []factory) int {
	for i, f := range factories {
		if f.target == rootFQN {
			return i
		}
	}

	return -1
}

func visitDeps(idx int, factories []factory, index map[string][]int, visited map[int]bool) error {
	if visited[idx] {
		return nil
	}

	visited[idx] = true

	for _, dep := range factories[idx].deps {
		providers := index[dep]
		if len(providers) == 0 {
			return fmt.Errorf("%w: %q for step %q", ErrUnresolvable, dep, factories[idx].ctor)
		}

		if len(providers) > 1 {
			return fmt.Errorf("%w: %q has %d providers (step %q)", ErrMultipleProviders, dep, len(providers), factories[idx].ctor)
		}

		if err := visitDeps(providers[0], factories, index, visited); err != nil {
			return err
		}
	}

	return nil
}

var goKeywords = map[string]bool{
	"break": true, "default": true, "func": true, "interface": true, "select": true,
	"case": true, "defer": true, "go": true, "map": true, "struct": true,
	"chan": true, "else": true, "goto": true, "package": true, "switch": true,
	"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
	"continue": true, "for": true, "import": true, "return": true, "var": true,
	"error": true, "string": true, "bool": true, "int": true, "int8": true,
	"int16": true, "int32": true, "int64": true, "uint": true, "uint8": true,
	"uint16": true, "uint32": true, "uint64": true, "uintptr": true,
	"float32": true, "float64": true, "complex64": true, "complex128": true,
	"byte": true, "rune": true, "nil": true, "true": true, "false": true,
	"iota": true, "any": true, "comparable": true,
}

func collectAliases(factories []factory) map[string]int {
	paths := make(map[string]bool)

	for _, f := range factories {
		if pkg := splitFQN(f.ctor); pkg != "" {
			paths[pkg] = true
		}

		if pkg := splitFQN(f.target); pkg != "" {
			paths[pkg] = true
		}

		for _, impl := range f.implements {
			if pkg := splitFQN(impl); pkg != "" {
				paths[pkg] = true
			}
		}

		for _, dep := range f.deps {
			if pkg := splitFQN(dep); pkg != "" {
				paths[pkg] = true
			}
		}
	}

	aliasCount := map[string]int{}

	for pkg := range paths {
		alias := packageAlias(pkg)
		aliasCount[alias]++
	}

	reserved := map[string]int{}
	for alias, count := range aliasCount {
		reserved[alias] = count
	}

	return reserved
}

func assignNames(factories []factory, visited map[int]bool, aliases map[string]int) {
	used := make(map[string]int)
	for k := range goKeywords {
		used[k] = 1
	}

	for alias, count := range aliases {
		used[alias] = count
	}

	used["err"] = 1

	for i, f := range factories {
		if !visited[i] {
			continue
		}

		name := varNameFromTarget(f.target, used)
		factories[i].varName = name
	}
}

func varNameFromTarget(target string, used map[string]int) string {
	name := strings.TrimLeft(target, "*")

	if idx := strings.LastIndex(name, "/"); idx != -1 {
		name = name[idx+1:]
	}

	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}

	if len(name) > 0 {
		runes := []rune(name)
		name = strings.ToLower(string(runes[0])) + string(runes[1:])
	}

	if n, taken := used[name]; taken {
		used[name] = n + 1

		return fmt.Sprintf("%s%d", name, n)
	}

	used[name] = 1

	return name
}

func packageAlias(pkgPath string) string {
	if idx := strings.LastIndex(pkgPath, "/"); idx != -1 {
		return pkgPath[idx+1:]
	}

	if idx := strings.LastIndex(pkgPath, "."); idx != -1 {
		return pkgPath[idx+1:]
	}

	return pkgPath
}

func splitFQN(fqn string) string {
	clean := strings.TrimLeft(fqn, "*")

	lastDot := strings.LastIndex(clean, ".")
	if lastDot < 0 {
		return ""
	}

	return clean[:lastDot]
}
