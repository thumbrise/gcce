package wire_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/composition"
	"github.com/thumbrise/gcce/ecosystem/wire"
)

func newSrv() *http.Server                          { return &http.Server{} }
func newMux() *http.ServeMux                        { return http.NewServeMux() }
func newSrvWithMux(mux *http.ServeMux) *http.Server { return &http.Server{Handler: mux} }
func newErrSrv() (*http.Server, error)              { return &http.Server{}, nil }

func TestWire_Simple(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newSrv),
	)
	require.NoError(t, err)

	steps, err := c.Compile()
	require.NoError(t, err)
	require.Len(t, steps, 1)

	var buf bytes.Buffer

	err = wire.Compile(&buf, steps, wire.Root(new(*http.Server)))
	require.NoError(t, err)

	t.Log("generated:\n", buf.String())

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err)
}

func TestWire_Chain(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newMux),
		composition.Provide(newSrvWithMux),
	)
	require.NoError(t, err)

	steps, err := c.Compile()
	require.NoError(t, err)
	require.Len(t, steps, 2)

	var buf bytes.Buffer

	err = wire.Compile(&buf, steps, wire.Root(new(*http.Server)))
	require.NoError(t, err)

	t.Log("generated:\n", buf.String())

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err)
}

func TestWire_WithError(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newErrSrv),
	)
	require.NoError(t, err)

	steps, err := c.Compile()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = wire.Compile(&buf, steps, wire.Root(new(*http.Server)))
	require.NoError(t, err)

	t.Log("generated:\n", buf.String())

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err)
}

func TestWire_CustomPackageName(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newErrSrv),
	)
	require.NoError(t, err)

	steps, err := c.Compile()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = wire.Compile(&buf, steps,
		wire.Root(new(*http.Server)),
		wire.PackageName("myapp"),
	)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "package myapp")
}

func TestWire_CustomFunctionName(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newErrSrv),
	)
	require.NoError(t, err)

	steps, err := c.Compile()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = wire.Compile(&buf, steps,
		wire.Root(new(*http.Server)),
		wire.FunctionName("InitializeKernel"),
	)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "func InitializeKernel()")
}

func TestWire_Error_MissingRoot(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newSrv),
	)
	require.NoError(t, err)

	steps, err := c.Compile()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = wire.Compile(&buf, steps, wire.Root(new(*http.Client)))
	require.Error(t, err)
	require.Contains(t, err.Error(), "no step produces")
}

func TestWire_Error_MissingRootOption(t *testing.T) {
	c, _ := composition.New()
	steps, _ := c.Compile()

	var buf bytes.Buffer

	err := wire.Compile(&buf, steps)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Root() option is required")
}
