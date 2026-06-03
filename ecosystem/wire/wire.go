package wire

import (
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/thumbrise/gcce/ecosystem/wire/generator"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

var (
	ErrRootRequired  = errors.New("wire: Root() option is required")
	ErrRootNotFound  = generator.ErrRootNotFound
	errRootMustBePtr = errors.New("wire.Root: must be a pointer to a type, e.g. new(*Kernel)")
)

type config struct {
	rootFQN      string
	packageName  string
	functionName string
}

type Option interface {
	apply(cfg *config)
}

type option func(*config)

func (o option) apply(c *config) { o(c) }

func Root(ptr interface{}) Option {
	return option(func(c *config) {
		t := reflect.TypeOf(ptr)
		if t == nil || t.Kind() != reflect.Pointer {
			panic(errRootMustBePtr)
		}

		c.rootFQN = fqnOf(t.Elem())
	})
}

func PackageName(name string) Option {
	return option(func(c *config) {
		c.packageName = name
	})
}

func FunctionName(name string) Option {
	return option(func(c *config) {
		c.functionName = name
	})
}

func Compile(w io.Writer, operations []op.Operation, opts ...Option) error {
	var cfg config
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.rootFQN == "" {
		return ErrRootRequired
	}

	if cfg.packageName == "" {
		cfg.packageName = "main"
	}

	if cfg.functionName == "" {
		short := cfg.rootFQN
		if lastDot := strings.LastIndex(cfg.rootFQN, "."); lastDot >= 0 {
			short = cfg.rootFQN[lastDot+1:]
		}

		short = strings.TrimLeft(short, "*")
		cfg.functionName = "Provide" + short
	}

	return generator.Generate(w, operations, generator.Config{
		PkgName:  cfg.packageName,
		FuncName: cfg.functionName,
		RootType: cfg.rootFQN,
	})
}

func fqnOf(t reflect.Type) string {
	if t == nil {
		return ""
	}

	if t.Kind() == reflect.Pointer {
		return "*" + fqnOf(t.Elem())
	}

	if t.PkgPath() == "" {
		return t.Name()
	}

	return t.PkgPath() + "." + t.Name()
}
