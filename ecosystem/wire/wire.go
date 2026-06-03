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
	ErrRootRequired = errors.New("wire: Root() option is required")
	ErrRootNotFound = generator.ErrRootNotFound
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
		if t.Kind() != reflect.Pointer {
			panic("wire.Root: must be a pointer to a type, e.g. new(*Kernel)")
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
		lastDot := stringsLastDot(cfg.rootFQN)

		short := cfg.rootFQN
		if lastDot >= 0 {
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

func stringsLastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}

	return -1
}
