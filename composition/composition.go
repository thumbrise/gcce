package composition

import (
	"fmt"
	"reflect"
	"strings"
)

// CompileStep is an atomic composition instruction for code generators.
type CompileStep struct {
	Index       int      `json:"index"`
	TargetType  string   `json:"target_type"`
	PackagePath string   `json:"package_path,omitempty"`
	FuncName    string   `json:"func_name,omitempty"`
	ReturnsErr  bool     `json:"returns_err"`
	ArgsTypes   []string `json:"args_types,omitempty"`
	Metadata    Metadata `json:"metadata,omitempty"`
}

// Metadata is informational baggage for the generator.
type Metadata map[string]string

// Container is a dependency graph.
type Container struct {
	schema *defaultSchema
}

// New creates a Container from options.
func New(options ...Option) (*Container, error) {
	c := &Container{schema: newDefaultSchema()}
	var cfg config
	for _, opt := range options {
		opt.apply(&cfg)
	}
	for _, p := range cfg.provides {
		if err := c.provide(p.ctor, p.options...); err != nil {
			return nil, fmt.Errorf("%s: %w", p.frame, err)
		}
	}
	return c, nil
}

// Provide registers a constructor.
func (c *Container) Provide(constructor Constructor, options ...ProvideOption) error {
	return c.provide(constructor, options...)
}

// Compile walks the graph from the invocation and returns a bottom-up plan.
func (c *Container) Compile(invocation Invocation) ([]CompileStep, error) {
	if invocation == nil {
		return nil, fmt.Errorf("%w, got nil", errInvalidInvocationSignature)
	}
	fn, valid := inspectFunction(invocation)
	if !valid || !validateInvocation(fn) {
		return nil, fmt.Errorf("%w, got %s", errInvalidInvocationSignature, reflect.TypeOf(invocation))
	}

	nodes, err := parseInvocationParameters(fn, c.schema)
	if err != nil {
		return nil, err
	}

	var steps []CompileStep
	visited := map[*node]bool{}

	for _, n := range nodes {
		if err := c.schema.prepare(n); err != nil {
			return nil, err
		}
		steps, err = c.dfsWalk(n, steps, visited)
		if err != nil {
			return nil, err
		}
	}

	// reverse to bottom-up order
	ordered := make([]CompileStep, len(steps))
	for i, s := range steps {
		s.Index = i
		ordered[len(steps)-1-i] = s
	}
	for i := range ordered {
		ordered[i].Index = i
	}

	return ordered, nil
}

func (c *Container) dfsWalk(n *node, steps []CompileStep, visited map[*node]bool) ([]CompileStep, error) {
	if n == nil || visited[n] {
		return steps, nil
	}
	visited[n] = true

	if ctor, ok := n.compiler.(*constructorCompiler); ok {
		step := CompileStep{
			TargetType: n.rt.String(),
			ReturnsErr: ctor.typ == ctorValueError || ctor.typ == ctorValueCleanupError,
			Metadata:   n.metadata,
		}
		fullName := ctor.fn.Name
		lastDot := strings.LastIndex(fullName, ".")
		if lastDot != -1 {
			step.PackagePath = fullName[:lastDot]
			step.FuncName = fullName[lastDot+1:]
		} else {
			step.FuncName = fullName
		}
		for i := 0; i < ctor.fn.NumIn(); i++ {
			step.ArgsTypes = append(step.ArgsTypes, ctor.fn.In(i).String())
		}
		steps = append(steps, step)
	}

	childNodes, err := n.deps(c.schema)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", n, err)
	}
	for _, child := range childNodes {
		steps, err = c.dfsWalk(child, steps, visited)
		if err != nil {
			return nil, err
		}
	}
	return steps, nil
}

func (c *Container) provide(constructor Constructor, options ...ProvideOption) error {
	if constructor == nil {
		return fmt.Errorf("invalid constructor, got nil")
	}
	params := ProvideParams{}
	for _, opt := range options {
		opt.applyProvide(&params)
	}
	n, err := newConstructorNode(constructor)
	if err != nil {
		return err
	}
	n.metadata = params.Metadata
	return c.provideNode(n, params)
}

func (c *Container) provideNode(n *node, params ProvideParams) error {
	c.schema.register(n)
	for _, cur := range params.Interfaces {
		i, err := inspectInterfacePointer(cur)
		if err != nil {
			return err
		}
		if !n.rt.Implements(i.Type) {
			return fmt.Errorf("%s does not implement %s", n.rt, i.Type)
		}
		c.schema.register(&node{
			rt:       i.Type,
			compiler: n.compiler,
			metadata: n.metadata,
		})
	}
	return nil
}

// --- DSL types ---

type Option interface {
	apply(*config)
}

type ProvideOption interface {
	applyProvide(*ProvideParams)
}

type Constructor interface{}
type Invocation interface{}
type Interface interface{}

// --- DSL functions ---

func Provide(ctor Constructor, options ...ProvideOption) Option {
	frame := stacktrace(0)
	return option(func(cfg *config) {
		cfg.provides = append(cfg.provides, provideOpt{
			frame:   frame,
			ctor:    ctor,
			options: options,
		})
	})
}

func As(iface Interface) ProvideOption {
	return provideOption(func(params *ProvideParams) {
		params.Interfaces = append(params.Interfaces, iface)
	})
}

func Meta(key, value string) ProvideOption {
	return provideOption(func(params *ProvideParams) {
		if params.Metadata == nil {
			params.Metadata = Metadata{}
		}
		params.Metadata[key] = value
	})
}

// --- internal types ---

type config struct {
	provides []provideOpt
}

type provideOpt struct {
	frame   callerFrame
	ctor    Constructor
	options []ProvideOption
}

type ProvideParams struct {
	Interfaces []Interface
	Metadata   Metadata
}

type option func(cfg *config)

func (o option) apply(cfg *config) { o(cfg) }

type provideOption func(params *ProvideParams)

func (o provideOption) applyProvide(params *ProvideParams) { o(params) }
