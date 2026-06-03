package wire //nolint:testpackage

import (
	"bytes"
	"go/parser"
	"go/token"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/composition"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

type (
	Config   struct{}
	Server   struct{}
	Handler  struct{}
	App      struct{}
	Logger   interface{ Log() }
	MyLogger struct{}
)

func (m *MyLogger) Log() {}

func NewConfig() Config           { return Config{} }
func NewServer(Config) *Server    { return &Server{} }
func NewHandler(*Server) *Handler { return &Handler{} }
func NewApp(*Handler) *App        { return &App{} }
func NewMyLogger() *MyLogger      { return &MyLogger{} }

func NewErrServer(Config) (*Server, error)     { return &Server{}, nil }
func NewErrApp(*Handler, Logger) (*App, error) { return &App{}, nil }

func NewMuxApp(mux *http.ServeMux) *App { return &App{} }

func requireValidGo(t *testing.T, src string) {
	t.Helper()

	_, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Fatalf("invalid Go source:\n%s\n\nerror: %v", src, err)
	}
}

func requireContains(t *testing.T, src, substr string) {
	t.Helper()

	if !strings.Contains(src, substr) {
		t.Fatalf("expected %q to contain %q", src, substr)
	}
}

func TestWire_SimpleChain(t *testing.T) {
	c, err := composition.New(
		composition.Bind(NewConfig),
		composition.Bind(NewServer),
		composition.Bind(NewHandler),
		composition.Bind(NewApp),
	)
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops, Root(new(*App)))
	require.NoError(t, err)

	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func ProvideApp()")
	requireContains(t, buf.String(), "return app")
}

func TestWire_WithInterface(t *testing.T) {
	c, err := composition.New(
		composition.Bind(NewMyLogger, composition.Implements((*Logger)(nil))),
		composition.Bind(NewApp),
		composition.Bind(NewHandler),
		composition.Bind(NewServer),
		composition.Bind(NewConfig),
	)
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops, Root(new(*App)))
	require.NoError(t, err)

	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "return app")
}

func TestWire_ErrorMidChain(t *testing.T) {
	c, err := composition.New(
		composition.Bind(NewErrServer),
		composition.Bind(NewHandler),
		composition.Bind(NewApp),
		composition.Bind(NewConfig),
	)
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops, Root(new(*App)))
	require.NoError(t, err)

	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "if err != nil")
	requireContains(t, buf.String(), "return nil, err")
	requireContains(t, buf.String(), "error")
}

func TestWire_CrossPackage(t *testing.T) {
	c, err := composition.New(
		composition.Bind(http.NewServeMux),
		composition.Bind(NewMuxApp),
	)
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops, Root(new(*App)))
	require.NoError(t, err)

	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "import")
	requireContains(t, buf.String(), "net/http")
}

func TestWire_CustomPackageName(t *testing.T) {
	c, err := composition.New(
		composition.Bind(NewConfig),
		composition.Bind(NewServer),
		composition.Bind(NewHandler),
		composition.Bind(NewApp),
	)
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops,
		Root(new(*App)),
		PackageName("custompkg"),
		FunctionName("InitApp"),
	)
	require.NoError(t, err)

	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "package custompkg")
	requireContains(t, buf.String(), "func InitApp")
}

func TestWire_MissingRoot(t *testing.T) {
	var buf bytes.Buffer

	err := Compile(&buf, []op.Operation{})
	require.ErrorIs(t, err, ErrRootRequired)
}

func TestWire_EmptyBindings(t *testing.T) {
	c, err := composition.New()
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops, Root(new(*App)))
	require.ErrorIs(t, err, ErrRootNotFound)
}

func TestWire_RootTypeMismatch(t *testing.T) {
	c, err := composition.New(
		composition.Bind(NewConfig),
	)
	require.NoError(t, err)

	ops, err := c.Resolve()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = Compile(&buf, ops, Root(new(*App)))
	require.ErrorIs(t, err, ErrRootNotFound)
}
