package wire

import (
	"errors"
	"fmt"
	"go/format"
	"io"
	"reflect"

	"github.com/thumbrise/gcce/composition"
)

var (
	ErrRootRequired     = errors.New("wire: Root() option is required")
	ErrRootNotFound     = errors.New("wire: no step produces root type")
	ErrMultipleRoots    = errors.New("wire: multiple steps produce root type")
	ErrStepReturnsError = errors.New("wire: step returns error but root does not")
	ErrUnresolvable     = errors.New("wire: unresolvable type")
)

type config struct {
	rootType     reflect.Type
	packageName  string
	functionName string
}

type Option interface {
	apply(cfg *config)
}

type option func(*config)

func (o option) apply(c *config) { o(c) }

// Root sets the root type for code generation.
// ptr must be a pointer to the target type, e.g. new(*Kernel).
func Root(ptr interface{}) Option {
	return option(func(c *config) {
		t := reflect.TypeOf(ptr)
		if t.Kind() != reflect.Pointer || t.Elem().Kind() == reflect.Interface {
			panic("wire.Root: must be a pointer to a type, e.g. new(*Kernel)")
		}

		c.rootType = t.Elem()
	})
}

// PackageName sets the package name of the generated file.
func PackageName(name string) Option {
	return option(func(c *config) {
		c.packageName = name
	})
}

// FunctionName sets the name of the generated function.
func FunctionName(name string) Option {
	return option(func(c *config) {
		c.functionName = name
	})
}

// Compile generates Go source code from composition steps.
func Compile(w io.Writer, steps []composition.CompileStep, opts ...Option) error {
	var cfg config
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.rootType == nil {
		return ErrRootRequired
	}

	if cfg.packageName == "" {
		cfg.packageName = "main"
	}

	if cfg.functionName == "" {
		cfg.functionName = "Provide" + typeNameForFunc(cfg.rootType.String())
	}

	src, err := generate(steps, &cfg)
	if err != nil {
		return err
	}

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("wire: go fmt failed: %w\n\n%s", err, src)
	}

	_, err = w.Write(formatted)

	return err
}
