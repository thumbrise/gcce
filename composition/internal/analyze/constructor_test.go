package analyze_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/composition/internal/analyze"
)

type (
	Foo struct{}
	Bar struct{}
)

func NewFoo() Foo                { return Foo{} }
func NewBar(f Foo) Bar           { return Bar{} }
func NewBaz() (*Bar, error)      { return &Bar{}, nil }
func NewPtr() *Foo               { return &Foo{} }
func NewMulti(a Foo, b Bar) *Foo { return &Foo{} }

func TestConstructor_Valid(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewFoo)
	require.NoError(t, err)
	require.Equal(t, "NewFoo", ctor.Name)
	require.NotEmpty(t, ctor.Package)
	require.NotEmpty(t, ctor.Target())
}

func TestConstructor_NotFunc(t *testing.T) {
	_, err := analyze.NewConstructor("not a func")
	require.ErrorIs(t, err, analyze.ErrNotFunction)
}

func TestConstructor_Anonymous(t *testing.T) {
	fn := func() Foo { return Foo{} }
	_, err := analyze.NewConstructor(fn)
	require.ErrorIs(t, err, analyze.ErrAnonymousFunc)
}

func noReturn() {}

func TestConstructor_NoReturn(t *testing.T) {
	_, err := analyze.NewConstructor(noReturn)
	require.ErrorIs(t, err, analyze.ErrNoReturnValues)
}

func TestConstructor_ReturnsError(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewBaz)
	require.NoError(t, err)
	require.True(t, ctor.ReturnsError())
}

func TestConstructor_NoError(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewBar)
	require.NoError(t, err)
	require.False(t, ctor.ReturnsError())
}

func TestConstructor_PointerTarget(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewPtr)
	require.NoError(t, err)
	require.Contains(t, ctor.Target(), "*")
}

func TestConstructor_SingleDep(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewBar)
	require.NoError(t, err)
	require.Len(t, ctor.Dependencies(), 1)
	require.Contains(t, ctor.Dependencies()[0], "Foo")
}

func TestConstructor_MultiDeps(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewMulti)
	require.NoError(t, err)
	require.Len(t, ctor.Dependencies(), 2)
}

func stdlibDep(w io.Writer) error { return nil }

func NewVariadic(plugins ...*Foo) *Bar { return &Bar{} }

func TestConstructor_StdlibDep(t *testing.T) {
	ctor, err := analyze.NewConstructor(stdlibDep)
	require.NoError(t, err)
	require.Len(t, ctor.Dependencies(), 1)
	require.Equal(t, "io.Writer", ctor.Dependencies()[0])
}

func TestConstructor_IsVariadic(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewVariadic)
	require.NoError(t, err)
	require.True(t, ctor.IsVariadic())
	require.Len(t, ctor.Dependencies(), 1)
	require.Contains(t, ctor.Dependencies()[0], "[]")
	require.Contains(t, ctor.Dependencies()[0], "Foo")
}

func TestConstructor_IsNotVariadic(t *testing.T) {
	ctor, err := analyze.NewConstructor(NewMulti)
	require.NoError(t, err)
	require.False(t, ctor.IsVariadic())
}
