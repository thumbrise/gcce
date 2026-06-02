package composition_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/thumbrise/gcce/composition"
)

// ─── Package-level helpers (named = usable by compiler) ───

func newMux() *http.ServeMux                        { return http.NewServeMux() }
func newSrv() *http.Server                          { return &http.Server{} }
func newSrvWithMux(mux *http.ServeMux) *http.Server { return &http.Server{Handler: mux} }
func newHandlerStr(h http.Handler) string           { return "" }

func newErrSrv() (*http.Server, error)              { return &http.Server{}, nil }
func newCleanSrv() (*http.Server, func())           { return &http.Server{}, func() {} }
func newCleanErrSrv() (*http.Server, func(), error) { return &http.Server{}, func() {}, nil }

func newSrvCount(s []*http.Server) int { return len(s) }

// deep chain
type chainA struct{}
type chainB struct{ a *chainA }
type chainC struct{ b *chainB }

func newChainA() *chainA          { return &chainA{} }
func newChainB(a *chainA) *chainB { return &chainB{a} }
func newChainC(b *chainB) *chainC { return &chainC{b} }

// cycle
func cycBool(i int64) bool   { return true }
func cycInt32(b bool) int32  { return 0 }
func cycInt64(i int32) int64 { return 0 }

// ─── Happy paths ───

func TestCompile_SimpleChain(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newMux),
		composition.Provide(newSrvWithMux),
	)
	require.NoError(t, err)

	steps, err := c.Compile(func(s *http.Server) {})
	require.NoError(t, err)
	require.Len(t, steps, 2)
	require.Equal(t, "*http.ServeMux", steps[0].TargetType)
	require.Equal(t, "*http.Server", steps[1].TargetType)
	require.Equal(t, []string{"*http.ServeMux"}, steps[1].ArgsTypes)
}

func TestCompile_InterfaceBinding(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newMux, composition.As(new(http.Handler))),
		composition.Provide(newHandlerStr),
	)
	require.NoError(t, err)

	steps, err := c.Compile(func(s string) {})
	require.NoError(t, err)
	require.Len(t, steps, 2)
	require.Equal(t, "http.Handler", steps[1].ArgsTypes[0])
}

func TestCompile_Collection(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newSrv),
		composition.Provide(newSrv),
		composition.Provide(newSrvCount),
	)
	require.NoError(t, err)

	steps, err := c.Compile(func(i int) {})
	require.NoError(t, err)
	require.Len(t, steps, 3)
	require.Equal(t, "[]*http.Server", steps[2].ArgsTypes[0])
}

func TestCompile_ConstructorWithError(t *testing.T) {
	c, err := composition.New(composition.Provide(newErrSrv))
	require.NoError(t, err)

	steps, err := c.Compile(func(s *http.Server) {})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	require.True(t, steps[0].ReturnsErr)
}

func TestCompile_ConstructorWithCleanup(t *testing.T) {
	c, err := composition.New(composition.Provide(newCleanSrv))
	require.NoError(t, err)

	steps, err := c.Compile(func(s *http.Server) {})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	require.False(t, steps[0].ReturnsErr)
}

func TestCompile_ConstructorWithCleanupAndError(t *testing.T) {
	c, err := composition.New(composition.Provide(newCleanErrSrv))
	require.NoError(t, err)

	steps, err := c.Compile(func(s *http.Server) {})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	require.True(t, steps[0].ReturnsErr)
}

func TestCompile_NoDeps(t *testing.T) {
	c, err := composition.New(composition.Provide(newSrv))
	require.NoError(t, err)

	steps, err := c.Compile(func(s *http.Server) {})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	require.Empty(t, steps[0].ArgsTypes)
}

func TestCompile_Metadata(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newSrv, composition.Meta("env", "prod")),
	)
	require.NoError(t, err)

	steps, err := c.Compile(func(s *http.Server) {})
	require.NoError(t, err)
	require.Equal(t, "prod", steps[0].Metadata["env"])
}

func TestCompile_DeepChain(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newChainA),
		composition.Provide(newChainB),
		composition.Provide(newChainC),
	)
	require.NoError(t, err)

	steps, err := c.Compile(func(c *chainC) {})
	require.NoError(t, err)
	require.Len(t, steps, 3)
	require.Equal(t, "*composition_test.chainA", steps[0].TargetType)
	require.Equal(t, "*composition_test.chainB", steps[1].TargetType)
	require.Equal(t, "*composition_test.chainC", steps[2].TargetType)
}

// ─── Error paths ───

func TestCompile_Error_DuplicateSingleton(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newSrv),
		composition.Provide(newSrv),
	)
	require.NoError(t, err)

	_, err = c.Compile(func(s *http.Server) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple definitions")
}

func TestCompile_Error_Cycle(t *testing.T) {
	c, err := composition.New(
		composition.Provide(cycBool),
		composition.Provide(cycInt32),
		composition.Provide(cycInt64),
	)
	require.NoError(t, err)

	_, err = c.Compile(func(b bool) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cycle detected")
}

func TestCompile_Error_MissingDependency(t *testing.T) {
	c, err := composition.New(
		composition.Provide(newSrvWithMux),
	)
	require.NoError(t, err)

	_, err = c.Compile(func(s *http.Server) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not exists")
}

func TestCompile_Error_NilConstructor(t *testing.T) {
	_, err := composition.New(composition.Provide(nil))
	require.Error(t, err)
}

func TestCompile_Error_NonFunctionCtor(t *testing.T) {
	_, err := composition.New(composition.Provide("string"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid constructor signature")
}

func TestCompile_Error_InvalidConstructorSignature(t *testing.T) {
	_, err := composition.New(composition.Provide(func() {}))
	require.Error(t, err)
	// rejects as anonymous before checking return count
	require.Contains(t, err.Error(), "anonymous")
}

func TestCompile_Error_NilInvocation(t *testing.T) {
	c, _ := composition.New()
	_, err := c.Compile(nil)
	require.Error(t, err)
}

func TestCompile_Error_NonFunctionInvocation(t *testing.T) {
	c, _ := composition.New()
	_, err := c.Compile(42)
	require.Error(t, err)
}

func TestCompile_Error_InvocationReturnsValue(t *testing.T) {
	c, _ := composition.New()
	_, err := c.Compile(func() *http.Server { return nil })
	require.Error(t, err)
}

func TestCompile_Error_AnonymousConstructor(t *testing.T) {
	_, err := composition.New(
		composition.Provide(func() *http.Server { return &http.Server{} }),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "anonymous")
}
